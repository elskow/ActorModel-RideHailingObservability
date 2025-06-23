package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"actor-model-observability/internal/models"
	"actor-model-observability/internal/repository"

	"github.com/lib/pq"
)

// UserRepositoryImpl implements the UserRepository interface using PostgreSQL
type UserRepositoryImpl struct {
	db *sql.DB
}

// NewUserRepository creates a new instance of UserRepositoryImpl
func NewUserRepository(db *sql.DB) repository.UserRepository {
	return &UserRepositoryImpl{db: db}
}

// Create creates a new user in the database
func (r *UserRepositoryImpl) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, email, phone, name, user_type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	
	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Phone,
		user.Name,
		user.UserType,
		user.CreatedAt,
		user.UpdatedAt,
	)
	
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				if pqErr.Constraint == "users_email_key" {
					return &models.ValidationError{
						Field:   "email",
						Message: "email already exists",
					}
				}
				if pqErr.Constraint == "users_phone_key" {
					return &models.ValidationError{
						Field:   "phone",
						Message: "phone number already exists",
					}
				}
			}
		}
		return fmt.Errorf("failed to create user: %w", err)
	}
	
	return nil
}

// GetByID retrieves a user by ID
func (r *UserRepositoryImpl) GetByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT id, email, phone, name, user_type, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Phone,
		&user.Name,
		&user.UserType,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &models.NotFoundError{
				Resource: "user",
				ID:       id,
			}
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	
	return user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepositoryImpl) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, phone, name, user_type, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Phone,
		&user.Name,
		&user.UserType,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &models.NotFoundError{
				Resource: "user",
				ID:       email,
			}
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	
	return user, nil
}

// GetByPhone retrieves a user by phone number
func (r *UserRepositoryImpl) GetByPhone(ctx context.Context, phone string) (*models.User, error) {
	query := `
		SELECT id, email, phone, name, user_type, created_at, updated_at
		FROM users
		WHERE phone = $1
	`
	
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, phone).Scan(
		&user.ID,
		&user.Email,
		&user.Phone,
		&user.Name,
		&user.UserType,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &models.NotFoundError{
				Resource: "user",
				ID:       phone,
			}
		}
		return nil, fmt.Errorf("failed to get user by phone: %w", err)
	}
	
	return user, nil
}

// Update updates an existing user
func (r *UserRepositoryImpl) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET email = $2, phone = $3, name = $4, user_type = $5, updated_at = $6
		WHERE id = $1
	`
	
	result, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Phone,
		user.Name,
		user.UserType,
		user.UpdatedAt,
	)
	
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				if pqErr.Constraint == "users_email_key" {
					return &models.ValidationError{
						Field:   "email",
						Message: "email already exists",
					}
				}
				if pqErr.Constraint == "users_phone_key" {
					return &models.ValidationError{
						Field:   "phone",
						Message: "phone number already exists",
					}
				}
			}
		}
		return fmt.Errorf("failed to update user: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return &models.NotFoundError{
			Resource: "user",
			ID:       user.ID.String(),
		}
	}
	
	return nil
}

// Delete deletes a user by ID
func (r *UserRepositoryImpl) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return &models.NotFoundError{
			Resource: "user",
			ID:       id,
		}
	}
	
	return nil
}

// List retrieves a list of users with pagination
func (r *UserRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	query := `
		SELECT id, email, phone, name, user_type, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()
	
	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Phone,
			&user.Name,
			&user.UserType,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}
	
	return users, nil
}