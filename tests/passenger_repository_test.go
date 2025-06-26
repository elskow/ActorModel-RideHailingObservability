package tests

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"actor-model-observability/internal/models"
	"actor-model-observability/internal/repository/postgres"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestPassengerRepository_GetByID_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := postgres.NewPassengerRepository(db)

	// Setup test data
	passengerID := uuid.New()
	userID := uuid.New()
	expectedPassenger := &models.Passenger{
		ID:         passengerID,
		UserID:     userID,
		Rating:     4.5,
		TotalTrips: 25,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Setup mock expectations
	rows := sqlmock.NewRows([]string{
		"id", "user_id", "rating", "total_trips", "created_at", "updated_at",
	}).AddRow(
		expectedPassenger.ID, expectedPassenger.UserID, expectedPassenger.Rating, expectedPassenger.TotalTrips,
		expectedPassenger.CreatedAt, expectedPassenger.UpdatedAt,
	)

	mock.ExpectQuery(`SELECT (.+) FROM passengers WHERE id = \$1`).
		WithArgs(passengerID.String()).
		WillReturnRows(rows)

	// Execute
	result, err := repo.GetByID(context.Background(), passengerID.String())

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedPassenger.ID, result.ID)
	assert.Equal(t, expectedPassenger.UserID, result.UserID)
	assert.Equal(t, expectedPassenger.Rating, result.Rating)
	assert.Equal(t, expectedPassenger.TotalTrips, result.TotalTrips)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPassengerRepository_GetByID_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := postgres.NewPassengerRepository(db)

	passengerID := uuid.New()

	// Setup mock expectations - no rows returned
	mock.ExpectQuery(`SELECT (.+) FROM passengers WHERE id = \$1`).
		WithArgs(passengerID.String()).
		WillReturnError(sql.ErrNoRows)

	// Execute
	result, err := repo.GetByID(context.Background(), passengerID.String())

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.IsType(t, &models.NotFoundError{}, err)

	notFoundErr := err.(*models.NotFoundError)
	assert.Equal(t, "passenger", notFoundErr.Resource)
	assert.Equal(t, passengerID.String(), notFoundErr.ID)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPassengerRepository_GetByUserID_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := postgres.NewPassengerRepository(db)

	// Setup test data
	passengerID := uuid.New()
	userID := uuid.New()
	expectedPassenger := &models.Passenger{
		ID:         passengerID,
		UserID:     userID,
		Rating:     4.7,
		TotalTrips: 25,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Setup mock expectations
	rows := sqlmock.NewRows([]string{
		"id", "user_id", "rating", "total_trips", "created_at", "updated_at",
	}).AddRow(
		expectedPassenger.ID, expectedPassenger.UserID, expectedPassenger.Rating, expectedPassenger.TotalTrips,
		expectedPassenger.CreatedAt, expectedPassenger.UpdatedAt,
	)

	mock.ExpectQuery(`SELECT (.+) FROM passengers WHERE user_id = \$1`).
		WithArgs(userID.String()).
		WillReturnRows(rows)

	// Execute
	result, err := repo.GetByUserID(context.Background(), userID.String())

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedPassenger.ID, result.ID)
	assert.Equal(t, expectedPassenger.UserID, result.UserID)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPassengerRepository_Create_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := postgres.NewPassengerRepository(db)

	// Setup test data
	passenger := &models.Passenger{
		ID:         uuid.New(),
		UserID:     uuid.New(),
		Rating:     0.0,
		TotalTrips: 0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Setup mock expectations
	mock.ExpectExec(`INSERT INTO passengers`).
		WithArgs(
			passenger.ID, passenger.UserID, passenger.Rating, passenger.TotalTrips,
			passenger.CreatedAt, passenger.UpdatedAt,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Execute
	err := repo.Create(context.Background(), passenger)

	// Assert
	assert.NoError(t, err)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPassengerRepository_Create_UserIDExists(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := postgres.NewPassengerRepository(db)

	// Setup test data
	passenger := &models.Passenger{
		ID:         uuid.New(),
		UserID:     uuid.New(),
		Rating:     0.0,
		TotalTrips: 0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Setup mock expectations - unique violation on user_id
	mock.ExpectExec(`INSERT INTO passengers`).
		WithArgs(
			passenger.ID, passenger.UserID, passenger.Rating, passenger.TotalTrips,
			passenger.CreatedAt, passenger.UpdatedAt,
		).
		WillReturnError(&pq.Error{
			Code:       "23505", // unique_violation
			Constraint: "passengers_user_id_key",
		})

	// Execute
	err := repo.Create(context.Background(), passenger)

	// Assert
	assert.Error(t, err)
	assert.IsType(t, &models.ValidationError{}, err)

	validationErr := err.(*models.ValidationError)
	assert.Equal(t, "user_id", validationErr.Field)
	assert.Equal(t, "passenger already exists for this user", validationErr.Message)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPassengerRepository_Update_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := postgres.NewPassengerRepository(db)

	// Setup test data
	passenger := &models.Passenger{
		ID:         uuid.New(),
		UserID:     uuid.New(),
		Rating:     4.8,
		TotalTrips: 30,
		CreatedAt:  time.Now().Add(-time.Hour),
		UpdatedAt:  time.Now(),
	}

	// Setup mock expectations
	mock.ExpectExec(`UPDATE passengers SET`).
		WithArgs(
			passenger.ID,
			passenger.Rating, passenger.TotalTrips, passenger.UpdatedAt,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Execute
	err := repo.Update(context.Background(), passenger)

	// Assert
	assert.NoError(t, err)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPassengerRepository_Update_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := postgres.NewPassengerRepository(db)

	// Setup test data
	passenger := &models.Passenger{
		ID:         uuid.New(),
		UserID:     uuid.New(),
		Rating:     4.8,
		TotalTrips: 30,
		CreatedAt:  time.Now().Add(-time.Hour),
		UpdatedAt:  time.Now(),
	}

	// Setup mock expectations - no rows affected
	mock.ExpectExec(`UPDATE passengers SET`).
		WithArgs(
			passenger.ID,
			passenger.Rating, passenger.TotalTrips, passenger.UpdatedAt,
		).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Execute
	err := repo.Update(context.Background(), passenger)

	// Assert
	assert.Error(t, err)

	var notFoundErr *models.NotFoundError
	if assert.True(t, errors.As(err, &notFoundErr)) {
		assert.Equal(t, "passenger", notFoundErr.Resource)
		assert.Equal(t, passenger.ID.String(), notFoundErr.ID)
	}

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPassengerRepository_List_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := postgres.NewPassengerRepository(db)

	// Setup test data
	passenger1 := &models.Passenger{
		ID:         uuid.New(),
		UserID:     uuid.New(),
		Rating:     4.7,
		TotalTrips: 25,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	passenger2 := &models.Passenger{
		ID:         uuid.New(),
		UserID:     uuid.New(),
		Rating:     4.9,
		TotalTrips: 40,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Setup mock expectations
	rows := sqlmock.NewRows([]string{
		"id", "user_id", "rating", "total_trips", "created_at", "updated_at",
	}).AddRow(
		passenger1.ID, passenger1.UserID, passenger1.Rating, passenger1.TotalTrips,
		passenger1.CreatedAt, passenger1.UpdatedAt,
	).AddRow(
		passenger2.ID, passenger2.UserID, passenger2.Rating, passenger2.TotalTrips,
		passenger2.CreatedAt, passenger2.UpdatedAt,
	)

	mock.ExpectQuery(`SELECT (.+) FROM passengers ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(10, 0).
		WillReturnRows(rows)

	// Execute
	result, err := repo.List(context.Background(), 10, 0)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, passenger1.ID, result[0].ID)
	assert.Equal(t, passenger2.ID, result[1].ID)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPassengerRepository_Delete_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := postgres.NewPassengerRepository(db)

	passengerID := uuid.New()

	// Setup mock expectations
	mock.ExpectExec(`DELETE FROM passengers WHERE id = \$1`).
		WithArgs(passengerID.String()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Execute
	err := repo.Delete(context.Background(), passengerID.String())

	// Assert
	assert.NoError(t, err)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPassengerRepository_Delete_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := postgres.NewPassengerRepository(db)

	passengerID := uuid.New()

	// Setup mock expectations - no rows affected
	mock.ExpectExec(`DELETE FROM passengers WHERE id = \$1`).
		WithArgs(passengerID.String()).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Execute
	err := repo.Delete(context.Background(), passengerID.String())

	// Assert
	assert.Error(t, err)
	assert.IsType(t, &models.NotFoundError{}, err)

	notFoundErr := err.(*models.NotFoundError)
	assert.Equal(t, "passenger", notFoundErr.Resource)
	assert.Equal(t, passengerID.String(), notFoundErr.ID)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}
