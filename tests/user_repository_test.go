package tests

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"actor-model-observability/internal/models"
	"actor-model-observability/internal/repository/postgres"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupMockDB creates a mock database connection for testing
func setupMockDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(mockDB, "postgres")
	return sqlxDB, mock
}

func TestUserRepository_GetByID_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := postgres.NewUserRepository(db)

	// Setup test data
	userID := uuid.New()
	expectedUser := &models.User{
		ID:        userID,
		Email:     "test@example.com",
		Phone:     "+1234567890",
		Name:      "Test User",
		UserType:  models.UserTypePassenger,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Setup mock expectations
	rows := sqlmock.NewRows([]string{"id", "email", "phone", "name", "user_type", "created_at", "updated_at"}).
		AddRow(expectedUser.ID, expectedUser.Email, expectedUser.Phone, expectedUser.Name, expectedUser.UserType, expectedUser.CreatedAt, expectedUser.UpdatedAt)

	mock.ExpectQuery(`SELECT id, email, phone, name, user_type, created_at, updated_at\s+FROM users\s+WHERE id = \$1`).
		WithArgs(userID.String()).
		WillReturnRows(rows)

	// Execute
	result, err := repo.GetByID(context.Background(), userID.String())

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedUser.ID, result.ID)
	assert.Equal(t, expectedUser.Email, result.Email)
	assert.Equal(t, expectedUser.Name, result.Name)
	assert.Equal(t, expectedUser.UserType, result.UserType)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := postgres.NewUserRepository(db)

	userID := uuid.New()

	// Setup mock expectations - no rows returned
	mock.ExpectQuery(`SELECT id, email, phone, name, user_type, created_at, updated_at FROM users WHERE id = \$1`).
		WithArgs(userID.String()).
		WillReturnError(sql.ErrNoRows)

	// Execute
	result, err := repo.GetByID(context.Background(), userID.String())

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.IsType(t, &models.NotFoundError{}, err)

	notFoundErr := err.(*models.NotFoundError)
	assert.Equal(t, "user", notFoundErr.Resource)
	assert.Equal(t, userID.String(), notFoundErr.ID)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByID_DatabaseError(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := postgres.NewUserRepository(db)

	userID := uuid.New()

	// Setup mock expectations - database error
	mock.ExpectQuery(`SELECT id, email, phone, name, user_type, created_at, updated_at FROM users WHERE id = \$1`).
		WithArgs(userID.String()).
		WillReturnError(sql.ErrConnDone)

	// Execute
	result, err := repo.GetByID(context.Background(), userID.String())

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get user by ID")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Create_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := postgres.NewUserRepository(db)

	// Setup test data
	user := &models.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		Phone:     "+1234567890",
		Name:      "Test User",
		UserType:  models.UserTypePassenger,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Setup mock expectations
	mock.ExpectExec(`INSERT INTO users`).
		WithArgs(user.ID, user.Email, user.Phone, user.Name, user.UserType, user.CreatedAt, user.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Execute
	err := repo.Create(context.Background(), user)

	// Assert
	assert.NoError(t, err)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Create_EmailExists(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := postgres.NewUserRepository(db)

	// Setup test data
	user := &models.User{
		ID:        uuid.New(),
		Email:     "existing@example.com",
		Phone:     "+1234567890",
		Name:      "Test User",
		UserType:  models.UserTypePassenger,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Setup mock expectations - unique violation on email
	mock.ExpectExec(`INSERT INTO users`).
		WithArgs(user.ID, user.Email, user.Phone, user.Name, user.UserType, user.CreatedAt, user.UpdatedAt).
		WillReturnError(&pq.Error{
			Code:       "23505", // unique_violation
			Constraint: "users_email_key",
		})

	// Execute
	err := repo.Create(context.Background(), user)

	// Assert
	assert.Error(t, err)
	assert.IsType(t, &models.ValidationError{}, err)

	validationErr := err.(*models.ValidationError)
	assert.Equal(t, "email", validationErr.Field)
	assert.Equal(t, "email already exists", validationErr.Message)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Update_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := postgres.NewUserRepository(db)

	// Setup test data
	user := &models.User{
		ID:        uuid.New(),
		Email:     "updated@example.com",
		Phone:     "+1234567890",
		Name:      "Updated User",
		UserType:  models.UserTypePassenger,
		CreatedAt: time.Now().Add(-time.Hour),
		UpdatedAt: time.Now(),
	}

	// Setup mock expectations
	mock.ExpectExec(`UPDATE users SET`).
		WithArgs(user.Email, user.Phone, user.Name, user.UserType, user.UpdatedAt, user.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Execute
	err := repo.Update(context.Background(), user)

	// Assert
	assert.NoError(t, err)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Update_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := postgres.NewUserRepository(db)

	// Setup test data
	user := &models.User{
		ID:        uuid.New(),
		Email:     "updated@example.com",
		Phone:     "+1234567890",
		Name:      "Updated User",
		UserType:  models.UserTypePassenger,
		CreatedAt: time.Now().Add(-time.Hour),
		UpdatedAt: time.Now(),
	}

	// Setup mock expectations - no rows affected
	mock.ExpectExec(`UPDATE users SET`).
		WithArgs(user.Email, user.Phone, user.Name, user.UserType, user.UpdatedAt, user.ID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Execute
	err := repo.Update(context.Background(), user)

	// Assert
	assert.Error(t, err)
	assert.IsType(t, &models.NotFoundError{}, err)

	notFoundErr := err.(*models.NotFoundError)
	assert.Equal(t, "user", notFoundErr.Resource)
	assert.Equal(t, user.ID.String(), notFoundErr.ID)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_List_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := postgres.NewUserRepository(db)

	// Setup test data
	user1 := &models.User{
		ID:        uuid.New(),
		Email:     "user1@example.com",
		Phone:     "+1234567890",
		Name:      "User 1",
		UserType:  models.UserTypePassenger,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	user2 := &models.User{
		ID:        uuid.New(),
		Email:     "user2@example.com",
		Phone:     "+1234567891",
		Name:      "User 2",
		UserType:  models.UserTypeDriver,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Setup mock expectations
	rows := sqlmock.NewRows([]string{"id", "email", "phone", "name", "user_type", "created_at", "updated_at"}).
		AddRow(user1.ID, user1.Email, user1.Phone, user1.Name, user1.UserType, user1.CreatedAt, user1.UpdatedAt).
		AddRow(user2.ID, user2.Email, user2.Phone, user2.Name, user2.UserType, user2.CreatedAt, user2.UpdatedAt)

	mock.ExpectQuery(`SELECT id, email, phone, name, user_type, created_at, updated_at FROM users ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(10, 0).
		WillReturnRows(rows)

	// Execute
	result, err := repo.List(context.Background(), 10, 0)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, user1.ID, result[0].ID)
	assert.Equal(t, user2.ID, result[1].ID)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}
