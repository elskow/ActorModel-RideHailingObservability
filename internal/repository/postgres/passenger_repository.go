package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"actor-model-observability/internal/models"
	"actor-model-observability/internal/repository"

	"github.com/lib/pq"
)

// PassengerRepositoryImpl implements the PassengerRepository interface using PostgreSQL
type PassengerRepositoryImpl struct {
	db *sql.DB
}

// NewPassengerRepository creates a new instance of PassengerRepositoryImpl
func NewPassengerRepository(db *sql.DB) repository.PassengerRepository {
	return &PassengerRepositoryImpl{db: db}
}

// Create creates a new passenger in the database
func (r *PassengerRepositoryImpl) Create(ctx context.Context, passenger *models.Passenger) error {
	query := `
		INSERT INTO passengers (id, user_id, rating, total_trips, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	
	_, err := r.db.ExecContext(ctx, query,
		passenger.ID,
		passenger.UserID,
		passenger.Rating,
		passenger.TotalTrips,
		passenger.CreatedAt,
		passenger.UpdatedAt,
	)
	
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				if pqErr.Constraint == "passengers_user_id_key" {
					return &models.ValidationError{
						Field:   "user_id",
						Message: "passenger already exists for this user",
					}
				}
			case "23503": // foreign_key_violation
				if pqErr.Constraint == "passengers_user_id_fkey" {
					return &models.ValidationError{
						Field:   "user_id",
						Message: "user does not exist",
					}
				}
			}
		}
		return fmt.Errorf("failed to create passenger: %w", err)
	}
	
	return nil
}

// GetByID retrieves a passenger by ID
func (r *PassengerRepositoryImpl) GetByID(ctx context.Context, id string) (*models.Passenger, error) {
	query := `
		SELECT id, user_id, rating, total_trips, created_at, updated_at
		FROM passengers
		WHERE id = $1
	`
	
	passenger := &models.Passenger{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&passenger.ID,
		&passenger.UserID,
		&passenger.Rating,
		&passenger.TotalTrips,
		&passenger.CreatedAt,
		&passenger.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &models.NotFoundError{
				Resource: "passenger",
				ID:       id,
			}
		}
		return nil, fmt.Errorf("failed to get passenger by ID: %w", err)
	}
	
	return passenger, nil
}

// GetByUserID retrieves a passenger by user ID
func (r *PassengerRepositoryImpl) GetByUserID(ctx context.Context, userID string) (*models.Passenger, error) {
	query := `
		SELECT id, user_id, rating, total_trips, created_at, updated_at
		FROM passengers
		WHERE user_id = $1
	`
	
	passenger := &models.Passenger{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&passenger.ID,
		&passenger.UserID,
		&passenger.Rating,
		&passenger.TotalTrips,
		&passenger.CreatedAt,
		&passenger.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &models.NotFoundError{
				Resource: "passenger",
				ID:       userID,
			}
		}
		return nil, fmt.Errorf("failed to get passenger by user ID: %w", err)
	}
	
	return passenger, nil
}

// Update updates an existing passenger
func (r *PassengerRepositoryImpl) Update(ctx context.Context, passenger *models.Passenger) error {
	query := `
		UPDATE passengers
		SET rating = $2, total_trips = $3, updated_at = $4
		WHERE id = $1
	`
	
	result, err := r.db.ExecContext(ctx, query,
		passenger.ID,
		passenger.Rating,
		passenger.TotalTrips,
		passenger.UpdatedAt,
	)
	
	if err != nil {
		return fmt.Errorf("failed to update passenger: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return &models.NotFoundError{
			Resource: "passenger",
			ID:       passenger.ID.String(),
		}
	}
	
	return nil
}

// Delete deletes a passenger by ID
func (r *PassengerRepositoryImpl) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM passengers WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete passenger: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return &models.NotFoundError{
			Resource: "passenger",
			ID:       id,
		}
	}
	
	return nil
}

// List retrieves a list of passengers with pagination
func (r *PassengerRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*models.Passenger, error) {
	query := `
		SELECT id, user_id, rating, total_trips, created_at, updated_at
		FROM passengers
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list passengers: %w", err)
	}
	defer rows.Close()
	
	var passengers []*models.Passenger
	for rows.Next() {
		passenger := &models.Passenger{}
		err := rows.Scan(
			&passenger.ID,
			&passenger.UserID,
			&passenger.Rating,
			&passenger.TotalTrips,
			&passenger.CreatedAt,
			&passenger.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan passenger: %w", err)
		}
		passengers = append(passengers, passenger)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating passengers: %w", err)
	}
	
	return passengers, nil
}