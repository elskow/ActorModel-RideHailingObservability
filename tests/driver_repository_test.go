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

func TestDriverRepository_GetByID_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := postgres.NewDriverRepository(db)

	// Setup test data
	driverID := uuid.New()
	userID := uuid.New()
	lat := 37.7749
	lng := -122.4194
	expectedDriver := &models.Driver{
		ID:               driverID,
		UserID:           userID,
		LicenseNumber:    "DL123456789",
		VehicleType:      "sedan",
		VehiclePlate:     "ABC123",
		Status:           models.DriverStatusOnline,
		CurrentLatitude:  &lat,
		CurrentLongitude: &lng,
		Rating:           4.8,
		TotalTrips:       150,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Setup mock expectations
	rows := sqlmock.NewRows([]string{
		"id", "user_id", "license_number", "vehicle_type", "vehicle_plate",
		"status", "current_latitude", "current_longitude", "rating", "total_trips", "created_at", "updated_at",
	}).AddRow(
		expectedDriver.ID, expectedDriver.UserID, expectedDriver.LicenseNumber,
		expectedDriver.VehicleType, expectedDriver.VehiclePlate, expectedDriver.Status,
		expectedDriver.CurrentLatitude, expectedDriver.CurrentLongitude, expectedDriver.Rating, expectedDriver.TotalTrips,
		expectedDriver.CreatedAt, expectedDriver.UpdatedAt,
	)

	mock.ExpectQuery(`SELECT (.+) FROM drivers WHERE id = \$1`).
		WithArgs(driverID.String()).
		WillReturnRows(rows)

	// Execute
	result, err := repo.GetByID(context.Background(), driverID.String())

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedDriver.ID, result.ID)
	assert.Equal(t, expectedDriver.UserID, result.UserID)
	assert.Equal(t, expectedDriver.LicenseNumber, result.LicenseNumber)
	assert.Equal(t, expectedDriver.VehicleType, result.VehicleType)
	assert.Equal(t, expectedDriver.Status, result.Status)
	assert.Equal(t, expectedDriver.Rating, result.Rating)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDriverRepository_GetByID_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := postgres.NewDriverRepository(db)

	driverID := uuid.New()

	// Setup mock expectations - no rows returned
	mock.ExpectQuery(`SELECT (.+) FROM drivers WHERE id = \$1`).
		WithArgs(driverID.String()).
		WillReturnError(sql.ErrNoRows)

	// Execute
	result, err := repo.GetByID(context.Background(), driverID.String())

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.IsType(t, &models.NotFoundError{}, err)

	notFoundErr := err.(*models.NotFoundError)
	assert.Equal(t, "driver", notFoundErr.Resource)
	assert.Equal(t, driverID.String(), notFoundErr.ID)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDriverRepository_GetOnlineDrivers_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := postgres.NewDriverRepository(db)

	// Setup test data
	lat := 37.7749
	lng := -122.4194
	expectedDriver := &models.Driver{
		ID:               uuid.New(),
		UserID:           uuid.New(),
		LicenseNumber:    "DL123456789",
		VehicleType:      "sedan",
		VehiclePlate:     "ABC123",
		Status:           models.DriverStatusOnline,
		CurrentLatitude:  &lat,
		CurrentLongitude: &lng,
		Rating:           4.8,
		TotalTrips:       150,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Setup mock expectations
	rows := sqlmock.NewRows([]string{
		"id", "user_id", "license_number", "vehicle_type", "vehicle_plate",
		"status", "current_latitude", "current_longitude", "rating", "total_trips", "created_at", "updated_at",
	}).AddRow(
		expectedDriver.ID, expectedDriver.UserID, expectedDriver.LicenseNumber,
		expectedDriver.VehicleType, expectedDriver.VehiclePlate, expectedDriver.Status,
		expectedDriver.CurrentLatitude, expectedDriver.CurrentLongitude, expectedDriver.Rating, expectedDriver.TotalTrips,
		expectedDriver.CreatedAt, expectedDriver.UpdatedAt,
	)

	mock.ExpectQuery(`SELECT (.+) FROM drivers WHERE status = 'online'`).
		WillReturnRows(rows)

	// Execute
	result, err := repo.GetOnlineDrivers(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, expectedDriver.ID, result[0].ID)
	assert.Equal(t, models.DriverStatusOnline, result[0].Status)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDriverRepository_GetOnlineDrivers_Empty(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := postgres.NewDriverRepository(db)

	// Setup mock expectations - empty result
	rows := sqlmock.NewRows([]string{
		"id", "user_id", "license_number", "vehicle_type", "vehicle_plate",
		"status", "current_latitude", "current_longitude", "rating", "total_trips", "created_at", "updated_at",
	})

	mock.ExpectQuery(`SELECT (.+) FROM drivers WHERE status = 'online'`).
		WillReturnRows(rows)

	// Execute
	result, err := repo.GetOnlineDrivers(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 0)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDriverRepository_Create_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := postgres.NewDriverRepository(db)

	// Setup test data
	lat := 40.7128
	lng := -74.0060
	driver := &models.Driver{
		ID:               uuid.New(),
		UserID:           uuid.New(),
		LicenseNumber:    "DL987654321",
		VehicleType:      "sedan",
		VehiclePlate:     "XYZ789",
		Status:           models.DriverStatusOffline,
		CurrentLatitude:  &lat,
		CurrentLongitude: &lng,
		Rating:           5.0,
		TotalTrips:       0,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Setup mock expectations
	mock.ExpectExec(`INSERT INTO drivers`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Execute
	err := repo.Create(context.Background(), driver)

	// Assert
	assert.NoError(t, err)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDriverRepository_Create_LicenseExists(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := postgres.NewDriverRepository(db)

	// Setup test data
	lat := 0.0
	lng := 0.0
	driver := &models.Driver{
		ID:               uuid.New(),
		UserID:           uuid.New(),
		LicenseNumber:    "DL123456789",
		VehicleType:      "sedan",
		VehiclePlate:     "ABC123",
		Status:           models.DriverStatusOffline,
		CurrentLatitude:  &lat,
		CurrentLongitude: &lng,
		Rating:           0.0,
		TotalTrips:       0,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Setup mock expectations - unique violation on license
	mock.ExpectExec(`INSERT INTO drivers`).
		WillReturnError(&pq.Error{
			Code:       "23505", // unique_violation
			Constraint: "drivers_license_number_key",
		})

	// Execute
	err := repo.Create(context.Background(), driver)

	// Assert
	assert.Error(t, err)

	var validationErr *models.ValidationError
	if assert.True(t, errors.As(err, &validationErr)) {
		assert.Equal(t, "license_number", validationErr.Field)
		assert.Equal(t, "license number already exists", validationErr.Message)
	}

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDriverRepository_UpdateLocation_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := postgres.NewDriverRepository(db)

	// Setup test data
	driverID := uuid.New()
	newLat := 40.7589
	newLng := -73.9851

	// Setup mock expectations
	mock.ExpectExec("UPDATE drivers SET current_latitude = \\$2, current_longitude = \\$3, updated_at = CURRENT_TIMESTAMP WHERE id = \\$1").
		WithArgs(driverID, newLat, newLng).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Execute
	err := repo.UpdateLocation(context.Background(), driverID.String(), newLat, newLng)

	// Assert
	assert.NoError(t, err)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDriverRepository_UpdateOnlineStatus_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := postgres.NewDriverRepository(db)

	// Setup test data
	driverID := uuid.New()
	status := models.DriverStatusOnline

	// Setup mock expectations
	mock.ExpectExec("UPDATE drivers SET status = \\$2, updated_at = CURRENT_TIMESTAMP WHERE id = \\$1").
		WithArgs(driverID, status).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Execute
	err := repo.UpdateStatus(context.Background(), driverID.String(), status)

	// Assert
	assert.NoError(t, err)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}
