package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"actor-model-observability/internal/models"
	"actor-model-observability/internal/repository"

	"github.com/lib/pq"
)

// TripRepositoryImpl implements the TripRepository interface using PostgreSQL
type TripRepositoryImpl struct {
	db *sql.DB
}

// NewTripRepository creates a new instance of TripRepositoryImpl
func NewTripRepository(db *sql.DB) repository.TripRepository {
	return &TripRepositoryImpl{db: db}
}

// Create creates a new trip in the database
func (r *TripRepositoryImpl) Create(ctx context.Context, trip *models.Trip) error {
	query := `
		INSERT INTO trips (id, passenger_id, driver_id, status, pickup_latitude, pickup_longitude, 
			destination_latitude, destination_longitude, pickup_address, destination_address, fare_amount, distance_km, 
			duration_minutes, requested_at, matched_at, pickup_at, completed_at, cancelled_at, 
			created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
	`
	
	_, err := r.db.ExecContext(ctx, query,
		trip.ID,
		trip.PassengerID,
		trip.DriverID,
		trip.Status,
		trip.PickupLatitude,
		trip.PickupLongitude,
		trip.DestinationLatitude,
		trip.DestinationLongitude,
		trip.PickupAddress,
		trip.DestinationAddress,
		trip.FareAmount,
		trip.DistanceKm,
		trip.DurationMinutes,
		trip.RequestedAt,
		trip.MatchedAt,
		trip.PickupAt,
		trip.CompletedAt,
		trip.CancelledAt,
		trip.CreatedAt,
		trip.UpdatedAt,
	)
	
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23503": // foreign_key_violation
				if pqErr.Constraint == "trips_passenger_id_fkey" {
					return &models.ValidationError{
						Field:   "passenger_id",
						Message: "passenger does not exist",
					}
				}
				if pqErr.Constraint == "trips_driver_id_fkey" {
					return &models.ValidationError{
						Field:   "driver_id",
						Message: "driver does not exist",
					}
				}
			}
		}
		return fmt.Errorf("failed to create trip: %w", err)
	}
	
	return nil
}

// GetByID retrieves a trip by ID
func (r *TripRepositoryImpl) GetByID(ctx context.Context, id string) (*models.Trip, error) {
	query := `
		SELECT id, passenger_id, driver_id, status, pickup_latitude, pickup_longitude, 
			destination_latitude, destination_longitude, pickup_address, destination_address, fare_amount, distance_km, 
			duration_minutes, requested_at, matched_at, accepted_at, pickup_at, completed_at, cancelled_at, 
			created_at, updated_at
		FROM trips
		WHERE id = $1
	`
	
	trip := &models.Trip{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&trip.ID,
		&trip.PassengerID,
		&trip.DriverID,
		&trip.Status,
		&trip.PickupLatitude,
		&trip.PickupLongitude,
		&trip.DestinationLatitude,
		&trip.DestinationLongitude,
		&trip.PickupAddress,
		&trip.DestinationAddress,
		&trip.FareAmount,
		&trip.DistanceKm,
		&trip.DurationMinutes,
		&trip.RequestedAt,
		&trip.MatchedAt,
		&trip.AcceptedAt,
		&trip.PickupAt,
		&trip.CompletedAt,
		&trip.CancelledAt,
		&trip.CreatedAt,
		&trip.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &models.NotFoundError{
				Resource: "trip",
				ID:       id,
			}
		}
		return nil, fmt.Errorf("failed to get trip by ID: %w", err)
	}
	
	return trip, nil
}

// Update updates an existing trip
func (r *TripRepositoryImpl) Update(ctx context.Context, trip *models.Trip) error {
	query := `
		UPDATE trips
		SET passenger_id = $2, driver_id = $3, status = $4, pickup_latitude = $5, pickup_longitude = $6,
			destination_latitude = $7, destination_longitude = $8, pickup_address = $9, destination_address = $10,
			fare_amount = $11, distance_km = $12, duration_minutes = $13, requested_at = $14, matched_at = $15,
			accepted_at = $16, pickup_at = $17, completed_at = $18, cancelled_at = $19,
			updated_at = $20
		WHERE id = $1
	`
	
	result, err := r.db.ExecContext(ctx, query,
		trip.ID,
		trip.PassengerID,
		trip.DriverID,
		trip.Status,
		trip.PickupLatitude,
		trip.PickupLongitude,
		trip.DestinationLatitude,
		trip.DestinationLongitude,
		trip.PickupAddress,
		trip.DestinationAddress,
		trip.FareAmount,
		trip.DistanceKm,
		trip.DurationMinutes,
		trip.RequestedAt,
		trip.MatchedAt,
		trip.AcceptedAt,
		trip.PickupAt,
		trip.CompletedAt,
		trip.CancelledAt,
		trip.UpdatedAt,
	)
	
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23503": // foreign_key_violation
				if pqErr.Constraint == "trips_passenger_id_fkey" {
					return &models.ValidationError{
						Field:   "passenger_id",
						Message: "passenger does not exist",
					}
				}
				if pqErr.Constraint == "trips_driver_id_fkey" {
					return &models.ValidationError{
						Field:   "driver_id",
						Message: "driver does not exist",
					}
				}
			}
		}
		return fmt.Errorf("failed to update trip: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return &models.NotFoundError{
			Resource: "trip",
			ID:       trip.ID.String(),
		}
	}
	
	return nil
}

// Delete deletes a trip by ID
func (r *TripRepositoryImpl) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM trips WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete trip: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return &models.NotFoundError{
			Resource: "trip",
			ID:       id,
		}
	}
	
	return nil
}

// GetByPassengerID retrieves trips by passenger ID with pagination
func (r *TripRepositoryImpl) GetByPassengerID(ctx context.Context, passengerID string, limit, offset int) ([]*models.Trip, error) {
	query := `
		SELECT id, passenger_id, driver_id, status, pickup_latitude, pickup_longitude, 
			destination_latitude, destination_longitude, pickup_address, destination_address, fare_amount, distance_km, 
			duration_minutes, requested_at, matched_at, accepted_at, pickup_at, completed_at, cancelled_at, 
			created_at, updated_at
		FROM trips
		WHERE passenger_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	
	return r.scanTrips(ctx, query, passengerID, limit, offset)
}

// GetByDriverID retrieves trips by driver ID with pagination
func (r *TripRepositoryImpl) GetByDriverID(ctx context.Context, driverID string, limit, offset int) ([]*models.Trip, error) {
	query := `
		SELECT id, passenger_id, driver_id, status, pickup_latitude, pickup_longitude, 
			destination_latitude, destination_longitude, pickup_address, destination_address, fare_amount, distance_km, 
			duration_minutes, requested_at, matched_at, accepted_at, pickup_at, completed_at, cancelled_at, 
			created_at, updated_at
		FROM trips
		WHERE driver_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	
	return r.scanTrips(ctx, query, driverID, limit, offset)
}

// GetActiveTrips retrieves all active trips (requested, matched, started)
func (r *TripRepositoryImpl) GetActiveTrips(ctx context.Context) ([]*models.Trip, error) {
	query := `
		SELECT id, passenger_id, driver_id, status, pickup_latitude, pickup_longitude, 
			destination_latitude, destination_longitude, pickup_address, destination_address, fare_amount, distance_km, 
			duration_minutes, requested_at, matched_at, accepted_at, pickup_at, completed_at, cancelled_at, 
			created_at, updated_at
		FROM trips
		WHERE status IN ('requested', 'matched', 'started')
		ORDER BY created_at DESC
	`
	
	return r.scanTripsNoParams(ctx, query)
}

// GetTripsByStatus retrieves trips by status with pagination
func (r *TripRepositoryImpl) GetTripsByStatus(ctx context.Context, status models.TripStatus, limit, offset int) ([]*models.Trip, error) {
	query := `
		SELECT id, passenger_id, driver_id, status, pickup_latitude, pickup_longitude, 
			destination_latitude, destination_longitude, pickup_address, destination_address, fare_amount, distance_km, 
			duration_minutes, requested_at, matched_at, accepted_at, pickup_at, completed_at, cancelled_at, 
			created_at, updated_at
		FROM trips
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	
	return r.scanTrips(ctx, query, status, limit, offset)
}

// GetTripsByDateRange retrieves trips within a date range with pagination
func (r *TripRepositoryImpl) GetTripsByDateRange(ctx context.Context, startDate, endDate string, limit, offset int) ([]*models.Trip, error) {
	// Parse dates
	startTime, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date format: %w", err)
	}
	
	endTime, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date format: %w", err)
	}
	
	// Add 24 hours to end date to include the entire day
	endTime = endTime.Add(24 * time.Hour)
	
	query := `
		SELECT id, passenger_id, driver_id, status, pickup_latitude, pickup_longitude, 
			destination_latitude, destination_longitude, pickup_address, destination_address, fare_amount, distance_km, 
			duration_minutes, requested_at, matched_at, accepted_at, pickup_at, completed_at, cancelled_at, 
			created_at, updated_at
		FROM trips
		WHERE created_at >= $1 AND created_at < $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`
	
	return r.scanTrips(ctx, query, startTime, endTime, limit, offset)
}

// List retrieves a list of trips with pagination
func (r *TripRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*models.Trip, error) {
	query := `
		SELECT id, passenger_id, driver_id, status, pickup_latitude, pickup_longitude, 
			destination_latitude, destination_longitude, pickup_address, destination_address, fare_amount, distance_km, 
			duration_minutes, requested_at, matched_at, accepted_at, pickup_at, completed_at, cancelled_at, 
			created_at, updated_at
		FROM trips
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	
	return r.scanTrips(ctx, query, limit, offset)
}

// scanTrips is a helper method to scan trip results with parameters
func (r *TripRepositoryImpl) scanTrips(ctx context.Context, query string, args ...interface{}) ([]*models.Trip, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()
	
	return r.scanTripRows(rows)
}

// scanTripsNoParams is a helper method to scan trip results without parameters
func (r *TripRepositoryImpl) scanTripsNoParams(ctx context.Context, query string) ([]*models.Trip, error) {
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()
	
	return r.scanTripRows(rows)
}

// scanTripRows scans trip rows from a result set
func (r *TripRepositoryImpl) scanTripRows(rows *sql.Rows) ([]*models.Trip, error) {
	var trips []*models.Trip
	for rows.Next() {
		trip := &models.Trip{}
		err := rows.Scan(
			&trip.ID,
			&trip.PassengerID,
			&trip.DriverID,
			&trip.Status,
			&trip.PickupLatitude,
			&trip.PickupLongitude,
			&trip.DestinationLatitude,
			&trip.DestinationLongitude,
			&trip.PickupAddress,
			&trip.DestinationAddress,
			&trip.FareAmount,
			&trip.DistanceKm,
			&trip.DurationMinutes,
			&trip.RequestedAt,
			&trip.MatchedAt,
			&trip.AcceptedAt,
			&trip.PickupAt,
			&trip.CompletedAt,
			&trip.CancelledAt,
			&trip.CreatedAt,
			&trip.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trip: %w", err)
		}
		trips = append(trips, trip)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating trips: %w", err)
	}
	
	return trips, nil
}