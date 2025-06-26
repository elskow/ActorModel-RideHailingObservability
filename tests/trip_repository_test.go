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
	"github.com/stretchr/testify/assert"
)

func TestTripRepository_Create_Success(t *testing.T) {
	db, mock := SetupMockDB(t)
	defer db.Close()

	repo := postgres.NewTripRepository(db)

	tripID := uuid.New()
	passengerID := uuid.New()
	driverID := uuid.New()
	now := time.Now()

	trip := &models.Trip{
		ID:                   tripID,
		PassengerID:          passengerID,
		DriverID:             &driverID,
		PickupLatitude:       40.7128,
		PickupLongitude:      -74.0060,
		DestinationLatitude:  40.7589,
		DestinationLongitude: -73.9851,
		PickupAddress:        stringPtr("123 Main St"),
		DestinationAddress:   stringPtr("456 Broadway"),
		Status:               models.TripStatusRequested,
		RequestedAt:          now,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	mock.ExpectExec(`INSERT INTO trips`).
		WithArgs(
			tripID, passengerID, driverID, models.TripStatusRequested,
			40.7128, -74.0060, 40.7589, -73.9851,
			"123 Main St", "456 Broadway",
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Create(context.Background(), trip)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTripRepository_GetByID_Success(t *testing.T) {
	db, mock := SetupMockDB(t)
	defer db.Close()

	repo := postgres.NewTripRepository(db)

	tripID := uuid.New()
	passengerID := uuid.New()
	driverID := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "passenger_id", "driver_id", "status",
		"pickup_latitude", "pickup_longitude", "destination_latitude", "destination_longitude",
		"pickup_address", "destination_address", "fare_amount", "distance_km",
		"duration_minutes", "requested_at", "matched_at", "accepted_at", "pickup_at", "completed_at", "cancelled_at",
		"created_at", "updated_at",
	}).AddRow(
		tripID, passengerID, driverID, models.TripStatusRequested,
		40.7128, -74.0060, 40.7589, -73.9851,
		"123 Main St", "456 Broadway", nil, nil,
		nil, now, nil, nil, nil, nil, nil,
		now, now,
	)

	mock.ExpectQuery(`SELECT (.+) FROM trips WHERE id = \$1`).
		WithArgs(tripID.String()).
		WillReturnRows(rows)

	trip, err := repo.GetByID(context.Background(), tripID.String())

	assert.NoError(t, err)
	assert.NotNil(t, trip)
	assert.Equal(t, tripID, trip.ID)
	assert.Equal(t, passengerID, trip.PassengerID)
	assert.Equal(t, driverID, *trip.DriverID)
	assert.Equal(t, models.TripStatusRequested, trip.Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTripRepository_GetByID_NotFound(t *testing.T) {
	db, mock := SetupMockDB(t)
	defer db.Close()

	repo := postgres.NewTripRepository(db)
	tripID := uuid.New()

	mock.ExpectQuery(`SELECT (.+) FROM trips WHERE id = \$1`).
		WithArgs(tripID.String()).
		WillReturnError(sql.ErrNoRows)

	trip, err := repo.GetByID(context.Background(), tripID.String())

	assert.Error(t, err)
	assert.Nil(t, trip)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTripRepository_Update_Success(t *testing.T) {
	db, mock := SetupMockDB(t)
	defer db.Close()

	repo := postgres.NewTripRepository(db)

	tripID := uuid.New()
	passengerID := uuid.New()
	driverID := uuid.New()
	now := time.Now()

	trip := &models.Trip{
		ID:                   tripID,
		PassengerID:          passengerID,
		DriverID:             &driverID,
		PickupLatitude:       40.7128,
		PickupLongitude:      -74.0060,
		DestinationLatitude:  40.7589,
		DestinationLongitude: -73.9851,
		Status:               models.TripStatusMatched,
		RequestedAt:          now,
		MatchedAt:            &now,
		UpdatedAt:            now,
	}

	mock.ExpectExec(`UPDATE trips SET`).
		WithArgs(
			tripID, passengerID, driverID, models.TripStatusMatched,
			40.7128, -74.0060, 40.7589, -73.9851,
			sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Update(context.Background(), trip)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTripRepository_GetByPassengerID_Success(t *testing.T) {
	db, mock := SetupMockDB(t)
	defer db.Close()

	repo := postgres.NewTripRepository(db)

	passengerID := uuid.New()
	tripID1 := uuid.New()
	tripID2 := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "passenger_id", "driver_id", "status",
		"pickup_latitude", "pickup_longitude", "destination_latitude", "destination_longitude",
		"pickup_address", "destination_address", "fare_amount", "distance_km",
		"duration_minutes", "requested_at", "matched_at", "accepted_at", "pickup_at", "completed_at", "cancelled_at",
		"created_at", "updated_at",
	}).AddRow(
		tripID1, passengerID, nil, models.TripStatusRequested,
		40.7128, -74.0060, 40.7589, -73.9851,
		"123 Main St", "456 Broadway", nil, nil,
		nil, now, nil, nil, nil, nil, nil,
		now, now,
	).AddRow(
		tripID2, passengerID, nil, models.TripStatusCompleted,
		40.7500, -74.0000, 40.7600, -73.9800,
		"789 Oak St", "321 Pine St", nil, nil,
		nil, now, nil, nil, nil, &now, nil,
		now, now,
	)

	mock.ExpectQuery(`SELECT (.+) FROM trips WHERE passenger_id = \$1`).
		WithArgs(passengerID.String(), 10, 0).
		WillReturnRows(rows)

	trips, err := repo.GetByPassengerID(context.Background(), passengerID.String(), 10, 0)

	assert.NoError(t, err)
	assert.Len(t, trips, 2)
	assert.Equal(t, tripID1, trips[0].ID)
	assert.Equal(t, tripID2, trips[1].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTripRepository_GetActiveTrips_Success(t *testing.T) {
	db, mock := SetupMockDB(t)
	defer db.Close()

	repo := postgres.NewTripRepository(db)

	tripID := uuid.New()
	passengerID := uuid.New()
	driverID := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "passenger_id", "driver_id", "status",
		"pickup_latitude", "pickup_longitude", "destination_latitude", "destination_longitude",
		"pickup_address", "destination_address", "fare_amount", "distance_km",
		"duration_minutes", "requested_at", "matched_at", "accepted_at", "pickup_at", "completed_at", "cancelled_at",
		"created_at", "updated_at",
	}).AddRow(
		tripID, passengerID, driverID, models.TripStatusInProgress,
		40.7128, -74.0060, 40.7589, -73.9851,
		"123 Main St", "456 Broadway", nil, nil,
		nil, now, &now, &now, &now, nil, nil,
		now, now,
	)

	mock.ExpectQuery(`SELECT (.+) FROM trips WHERE status IN`).
		WillReturnRows(rows)

	trips, err := repo.GetActiveTrips(context.Background())

	assert.NoError(t, err)
	assert.Len(t, trips, 1)
	assert.Equal(t, tripID, trips[0].ID)
	assert.Equal(t, models.TripStatusInProgress, trips[0].Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTripRepository_GetTripsByStatus_Success(t *testing.T) {
	db, mock := SetupMockDB(t)
	defer db.Close()

	repo := postgres.NewTripRepository(db)

	tripID := uuid.New()
	passengerID := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "passenger_id", "driver_id", "status",
		"pickup_latitude", "pickup_longitude", "destination_latitude", "destination_longitude",
		"pickup_address", "destination_address", "fare_amount", "distance_km",
		"duration_minutes", "requested_at", "matched_at", "accepted_at", "pickup_at", "completed_at", "cancelled_at",
		"created_at", "updated_at",
	}).AddRow(
		tripID, passengerID, nil, models.TripStatusRequested,
		40.7128, -74.0060, 40.7589, -73.9851,
		"123 Main St", "456 Broadway", nil, nil,
		nil, now, nil, nil, nil, nil, nil,
		now, now,
	)

	mock.ExpectQuery(`SELECT (.+) FROM trips WHERE status = \$1`).
		WithArgs(models.TripStatusRequested, 10, 0).
		WillReturnRows(rows)

	trips, err := repo.GetTripsByStatus(context.Background(), models.TripStatusRequested, 10, 0)

	assert.NoError(t, err)
	assert.Len(t, trips, 1)
	assert.Equal(t, tripID, trips[0].ID)
	assert.Equal(t, models.TripStatusRequested, trips[0].Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTripRepository_Delete_Success(t *testing.T) {
	db, mock := SetupMockDB(t)
	defer db.Close()

	repo := postgres.NewTripRepository(db)
	tripID := uuid.New()

	mock.ExpectExec(`DELETE FROM trips WHERE id = \$1`).
		WithArgs(tripID.String()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Delete(context.Background(), tripID.String())

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
