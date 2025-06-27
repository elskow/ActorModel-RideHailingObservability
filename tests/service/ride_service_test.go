package service

import (
	"context"
	"errors"
	"testing"

	"actor-model-observability/internal/actor"
	"actor-model-observability/internal/config"
	"actor-model-observability/internal/logging"
	"actor-model-observability/internal/models"
	"actor-model-observability/internal/observability"
	"actor-model-observability/internal/service"
	"actor-model-observability/internal/traditional"
	"actor-model-observability/tests/utils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock repositories are now defined in common.go

// Mock components are no longer needed as we use real implementations

func TestRideService_RequestRide_ActorModel_Success(t *testing.T) {
	// Setup mocks
	userRepo := &utils.MockUserRepository{}
	driverRepo := &utils.MockDriverRepository{}
	passengerRepo := &utils.MockPassengerRepository{}
	tripRepo := &utils.MockTripRepository{}

	// Create logger with proper config
	loggerCfg := &config.LoggingConfig{
		Level:  "debug",
		Format: "text",
		Output: "stdout",
	}
	logger, err := logging.NewLogger(loggerCfg)
	require.NoError(t, err)

	// Create real dependencies for service
	actorSystemReal := actor.NewActorSystem("test-system")
	err = actorSystemReal.Start(context.Background())
	require.NoError(t, err)
	defer actorSystemReal.Stop()

	metricsCollectorReal := observability.NewMetricsCollector(nil, nil, &config.Config{}, logger)
	traditionalMonitorReal := traditional.NewTraditionalMonitor(logger, nil)

	// Create service with actor model enabled
	rideService := service.NewRideService(
		userRepo, driverRepo, passengerRepo, tripRepo,
		actorSystemReal, metricsCollectorReal, traditionalMonitorReal,
		logger, true, // useActorModel = true
	)

	// Test data
	passengerID := uuid.New()
	passenger := &models.Passenger{
		ID:     passengerID,
		UserID: uuid.New(),
	}
	pickup := models.Location{Latitude: 40.7128, Longitude: -74.0060}
	dropoff := models.Location{Latitude: 40.7589, Longitude: -73.9851}
	pickupAddr := "123 Main St"
	dropoffAddr := "456 Broadway"

	// Setup expectations
	passengerRepo.On("GetByID", mock.Anything, passengerID.String()).Return(passenger, nil)
	tripRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Trip")).Return(nil)

	// Execute
	trip, err := rideService.RequestRide(context.Background(), passengerID.String(), pickup, dropoff, pickupAddr, dropoffAddr)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, trip)
	assert.Equal(t, passengerID, trip.PassengerID)
	assert.Equal(t, pickup.Latitude, trip.PickupLatitude)
	assert.Equal(t, pickup.Longitude, trip.PickupLongitude)
	assert.Equal(t, dropoff.Latitude, trip.DestinationLatitude)
	assert.Equal(t, dropoff.Longitude, trip.DestinationLongitude)
	assert.Equal(t, models.TripStatusRequested, trip.Status)

	// Verify mocks
	passengerRepo.AssertExpectations(t)
	tripRepo.AssertExpectations(t)
}

func TestRideService_RequestRide_Traditional_Success(t *testing.T) {
	// Setup mocks
	userRepo := &utils.MockUserRepository{}
	driverRepo := &utils.MockDriverRepository{}
	passengerRepo := &utils.MockPassengerRepository{}
	tripRepo := &utils.MockTripRepository{}

	// Create logger with proper config
	loggerCfg := &config.LoggingConfig{
		Level:  "debug",
		Format: "text",
		Output: "stdout",
	}
	logger, err := logging.NewLogger(loggerCfg)
	require.NoError(t, err)

	// Create real dependencies for service
	actorSystemReal := actor.NewActorSystem("test-system")
	err = actorSystemReal.Start(context.Background())
	require.NoError(t, err)
	defer actorSystemReal.Stop()

	metricsCollectorReal := observability.NewMetricsCollector(nil, nil, &config.Config{}, logger)
	traditionalMonitorReal := traditional.NewTraditionalMonitor(logger, nil)

	// Create service with traditional approach
	rideService := service.NewRideService(
		userRepo, driverRepo, passengerRepo, tripRepo,
		actorSystemReal, metricsCollectorReal, traditionalMonitorReal,
		logger, false, // useActorModel = false
	)

	// Test data
	passengerID := uuid.New()
	driverID := uuid.New()
	passenger := &models.Passenger{
		ID:     passengerID,
		UserID: uuid.New(),
	}
	lat := 40.7100
	lng := -74.0050
	driver := &models.Driver{
		ID:               driverID,
		UserID:           uuid.New(),
		CurrentLatitude:  &lat,
		CurrentLongitude: &lng,
		Status:           models.DriverStatusOnline,
	}
	pickup := models.Location{Latitude: 40.7128, Longitude: -74.0060}
	dropoff := models.Location{Latitude: 40.7589, Longitude: -73.9851}
	pickupAddr := "123 Main St"
	dropoffAddr := "456 Broadway"

	// Setup expectations
	passengerRepo.On("GetByID", mock.Anything, passengerID.String()).Return(passenger, nil)
	tripRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Trip")).Return(nil)
	driverRepo.On("GetOnlineDrivers", mock.Anything).Return([]*models.Driver{driver}, nil)
	tripRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Trip")).Return(nil)
	driverRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Driver")).Return(nil)

	// Execute
	trip, err := rideService.RequestRide(context.Background(), passengerID.String(), pickup, dropoff, pickupAddr, dropoffAddr)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, trip)
	assert.Equal(t, passengerID, trip.PassengerID)
	assert.Equal(t, driverID, *trip.DriverID)
	assert.Equal(t, models.TripStatusMatched, trip.Status)
	assert.NotNil(t, trip.MatchedAt)

	// Verify mocks
	passengerRepo.AssertExpectations(t)
	tripRepo.AssertExpectations(t)
	driverRepo.AssertExpectations(t)
}

func TestRideService_RequestRide_PassengerNotFound(t *testing.T) {
	// Setup mocks
	userRepo := &utils.MockUserRepository{}
	driverRepo := &utils.MockDriverRepository{}
	passengerRepo := &utils.MockPassengerRepository{}
	tripRepo := &utils.MockTripRepository{}

	// Create logger with proper config
	loggerCfg := &config.LoggingConfig{
		Level:  "debug",
		Format: "text",
		Output: "stdout",
	}
	logger, err := logging.NewLogger(loggerCfg)
	require.NoError(t, err)

	// Create real dependencies for service
	actorSystemReal := actor.NewActorSystem("test-system")
	err = actorSystemReal.Start(context.Background())
	require.NoError(t, err)
	defer actorSystemReal.Stop()

	metricsCollectorReal := observability.NewMetricsCollector(nil, nil, &config.Config{}, logger)
	traditionalMonitorReal := traditional.NewTraditionalMonitor(logger, nil)

	rideService := service.NewRideService(
		userRepo, driverRepo, passengerRepo, tripRepo,
		actorSystemReal, metricsCollectorReal, traditionalMonitorReal,
		logger, true,
	)

	passengerID := uuid.New()
	pickup := models.Location{Latitude: 40.7128, Longitude: -74.0060}
	dropoff := models.Location{Latitude: 40.7589, Longitude: -73.9851}

	// Setup expectations
	passengerRepo.On("GetByID", mock.Anything, passengerID.String()).Return((*models.Passenger)(nil), errors.New("passenger not found"))

	// Execute
	trip, err := rideService.RequestRide(context.Background(), passengerID.String(), pickup, dropoff, "pickup", "dropoff")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, trip)
	assert.Contains(t, err.Error(), "passenger not found")

	// Verify mocks
	passengerRepo.AssertExpectations(t)
}

func TestRideService_RequestRide_InvalidPassengerID(t *testing.T) {
	// Setup mocks
	userRepo := &utils.MockUserRepository{}
	driverRepo := &utils.MockDriverRepository{}
	passengerRepo := &utils.MockPassengerRepository{}
	tripRepo := &utils.MockTripRepository{}

	// Create logger with proper config
	loggerCfg := &config.LoggingConfig{
		Level:  "debug",
		Format: "text",
		Output: "stdout",
	}
	logger, err := logging.NewLogger(loggerCfg)
	require.NoError(t, err)

	// Create real dependencies for service
	actorSystemReal := actor.NewActorSystem("test-system")
	err = actorSystemReal.Start(context.Background())
	require.NoError(t, err)
	defer actorSystemReal.Stop()

	metricsCollectorReal := observability.NewMetricsCollector(nil, nil, &config.Config{}, logger)
	traditionalMonitorReal := traditional.NewTraditionalMonitor(logger, nil)

	rideService := service.NewRideService(
		userRepo, driverRepo, passengerRepo, tripRepo,
		actorSystemReal, metricsCollectorReal, traditionalMonitorReal,
		logger, true,
	)

	passenger := &models.Passenger{
		ID:     uuid.New(),
		UserID: uuid.New(),
	}
	pickup := models.Location{Latitude: 40.7128, Longitude: -74.0060}
	dropoff := models.Location{Latitude: 40.7589, Longitude: -73.9851}

	// Setup expectations
	passengerRepo.On("GetByID", mock.Anything, "invalid-uuid").Return(passenger, nil)

	// Execute
	trip, err := rideService.RequestRide(context.Background(), "invalid-uuid", pickup, dropoff, "pickup", "dropoff")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, trip)
	assert.Contains(t, err.Error(), "invalid passenger ID")

	// Verify mocks
	passengerRepo.AssertExpectations(t)
}

func TestRideService_RequestRide_TripCreateFailed(t *testing.T) {
	// Setup mocks
	userRepo := &utils.MockUserRepository{}
	driverRepo := &utils.MockDriverRepository{}
	passengerRepo := &utils.MockPassengerRepository{}
	tripRepo := &utils.MockTripRepository{}

	// Create logger with proper config
	loggerCfg := &config.LoggingConfig{
		Level:  "debug",
		Format: "text",
		Output: "stdout",
	}
	logger, err := logging.NewLogger(loggerCfg)
	require.NoError(t, err)

	// Create real dependencies for service
	actorSystemReal := actor.NewActorSystem("test-system")
	err = actorSystemReal.Start(context.Background())
	require.NoError(t, err)
	defer actorSystemReal.Stop()

	metricsCollectorReal := observability.NewMetricsCollector(nil, nil, &config.Config{}, logger)
	traditionalMonitorReal := traditional.NewTraditionalMonitor(logger, nil)

	rideService := service.NewRideService(
		userRepo, driverRepo, passengerRepo, tripRepo,
		actorSystemReal, metricsCollectorReal, traditionalMonitorReal,
		logger, true,
	)

	passengerID := uuid.New()
	passenger := &models.Passenger{
		ID:     passengerID,
		UserID: uuid.New(),
	}
	pickup := models.Location{Latitude: 40.7128, Longitude: -74.0060}
	dropoff := models.Location{Latitude: 40.7589, Longitude: -73.9851}

	// Setup expectations
	passengerRepo.On("GetByID", mock.Anything, passengerID.String()).Return(passenger, nil)
	tripRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Trip")).Return(errors.New("database error"))

	// Execute
	trip, err := rideService.RequestRide(context.Background(), passengerID.String(), pickup, dropoff, "pickup", "dropoff")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, trip)
	assert.Contains(t, err.Error(), "failed to create trip")

	// Verify mocks
	passengerRepo.AssertExpectations(t)
	tripRepo.AssertExpectations(t)
}

func TestRideService_RequestRide_Traditional_NoDriversAvailable(t *testing.T) {
	// Setup mocks
	userRepo := &utils.MockUserRepository{}
	driverRepo := &utils.MockDriverRepository{}
	passengerRepo := &utils.MockPassengerRepository{}
	tripRepo := &utils.MockTripRepository{}

	// Create logger with proper config
	loggerCfg := &config.LoggingConfig{
		Level:  "debug",
		Format: "text",
		Output: "stdout",
	}
	logger, err := logging.NewLogger(loggerCfg)
	require.NoError(t, err)

	// Create real dependencies for service
	actorSystemReal := actor.NewActorSystem("test-system")
	err = actorSystemReal.Start(context.Background())
	require.NoError(t, err)
	defer actorSystemReal.Stop()

	metricsCollectorReal := observability.NewMetricsCollector(nil, nil, &config.Config{}, logger)
	traditionalMonitorReal := traditional.NewTraditionalMonitor(logger, nil)

	rideService := service.NewRideService(
		userRepo, driverRepo, passengerRepo, tripRepo,
		actorSystemReal, metricsCollectorReal, traditionalMonitorReal,
		logger, false, // traditional approach
	)

	passengerID := uuid.New()
	passenger := &models.Passenger{
		ID:     passengerID,
		UserID: uuid.New(),
	}
	pickup := models.Location{Latitude: 40.7128, Longitude: -74.0060}
	dropoff := models.Location{Latitude: 40.7589, Longitude: -73.9851}

	// Setup expectations
	passengerRepo.On("GetByID", mock.Anything, passengerID.String()).Return(passenger, nil)
	tripRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Trip")).Return(nil)
	driverRepo.On("GetOnlineDrivers", mock.Anything).Return([]*models.Driver{}, nil)

	// Execute
	trip, err := rideService.RequestRide(context.Background(), passengerID.String(), pickup, dropoff, "pickup", "dropoff")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, trip)
	assert.Contains(t, err.Error(), "no available drivers found")

	// Verify mocks
	passengerRepo.AssertExpectations(t)
	tripRepo.AssertExpectations(t)
	driverRepo.AssertExpectations(t)
}

func TestRideService_CancelRide_Success(t *testing.T) {
	// Setup mocks
	userRepo := &utils.MockUserRepository{}
	driverRepo := &utils.MockDriverRepository{}
	passengerRepo := &utils.MockPassengerRepository{}
	tripRepo := &utils.MockTripRepository{}

	// Create logger with proper config
	loggerCfg := &config.LoggingConfig{
		Level:  "debug",
		Format: "text",
		Output: "stdout",
	}
	logger, err := logging.NewLogger(loggerCfg)
	require.NoError(t, err)

	// Create real dependencies for service
	actorSystemReal := actor.NewActorSystem("test-system")
	err = actorSystemReal.Start(context.Background())
	require.NoError(t, err)
	defer actorSystemReal.Stop()

	metricsCollectorReal := observability.NewMetricsCollector(nil, nil, &config.Config{}, logger)
	traditionalMonitorReal := traditional.NewTraditionalMonitor(logger, nil)

	rideService := service.NewRideService(
		userRepo, driverRepo, passengerRepo, tripRepo,
		actorSystemReal, metricsCollectorReal, traditionalMonitorReal,
		logger, true,
	)

	tripID := uuid.New()
	trip := &models.Trip{
		ID:     tripID,
		Status: models.TripStatusRequested,
	}

	// Setup expectations
	tripRepo.On("GetByID", mock.Anything, tripID.String()).Return(trip, nil)
	tripRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Trip")).Return(nil)

	// Execute
	err = rideService.CancelRide(context.Background(), tripID.String(), "passenger cancelled")

	// Assert
	assert.NoError(t, err)

	// Verify mocks
	tripRepo.AssertExpectations(t)
}

func TestRideService_GetTripStatus_Success(t *testing.T) {
	// Setup mocks
	userRepo := &utils.MockUserRepository{}
	driverRepo := &utils.MockDriverRepository{}
	passengerRepo := &utils.MockPassengerRepository{}
	tripRepo := &utils.MockTripRepository{}

	// Create logger with proper config
	loggerCfg := &config.LoggingConfig{
		Level:  "debug",
		Format: "text",
		Output: "stdout",
	}
	logger, err := logging.NewLogger(loggerCfg)
	require.NoError(t, err)

	// Create real dependencies for service
	actorSystemReal := actor.NewActorSystem("test-system")
	err = actorSystemReal.Start(context.Background())
	require.NoError(t, err)
	defer actorSystemReal.Stop()

	metricsCollectorReal := observability.NewMetricsCollector(nil, nil, &config.Config{}, logger)
	traditionalMonitorReal := traditional.NewTraditionalMonitor(logger, nil)

	rideService := service.NewRideService(
		userRepo, driverRepo, passengerRepo, tripRepo,
		actorSystemReal, metricsCollectorReal, traditionalMonitorReal,
		logger, true,
	)

	tripID := uuid.New()
	trip := &models.Trip{
		ID:     tripID,
		Status: models.TripStatusInProgress,
	}

	// Setup expectations
	tripRepo.On("GetByID", mock.Anything, tripID.String()).Return(trip, nil)

	// Execute
	result, err := rideService.GetTripStatus(context.Background(), tripID.String())

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, tripID, result.ID)
	assert.Equal(t, models.TripStatusInProgress, result.Status)

	// Verify mocks
	tripRepo.AssertExpectations(t)
}

func TestRideService_ListRides_Success(t *testing.T) {
	// Setup mocks
	userRepo := &utils.MockUserRepository{}
	driverRepo := &utils.MockDriverRepository{}
	passengerRepo := &utils.MockPassengerRepository{}
	tripRepo := &utils.MockTripRepository{}

	// Create logger with proper config
	loggerCfg := &config.LoggingConfig{
		Level:  "debug",
		Format: "text",
		Output: "stdout",
	}
	logger, err := logging.NewLogger(loggerCfg)
	require.NoError(t, err)

	// Create real dependencies for service
	actorSystemReal := actor.NewActorSystem("test-system")
	err = actorSystemReal.Start(context.Background())
	require.NoError(t, err)
	defer actorSystemReal.Stop()

	metricsCollectorReal := observability.NewMetricsCollector(nil, nil, &config.Config{}, logger)
	traditionalMonitorReal := traditional.NewTraditionalMonitor(logger, nil)

	rideService := service.NewRideService(
		userRepo, driverRepo, passengerRepo, tripRepo,
		actorSystemReal, metricsCollectorReal, traditionalMonitorReal,
		logger, true,
	)

	passengerID := uuid.New()
	trips := []*models.Trip{
		{ID: uuid.New(), PassengerID: passengerID, Status: models.TripStatusCompleted},
		{ID: uuid.New(), PassengerID: passengerID, Status: models.TripStatusRequested},
	}

	// Setup expectations
	tripRepo.On("List", mock.Anything, 10, 0).Return(trips, nil)

	// Execute
	passengerIDStr := passengerID.String()
	result, count, err := rideService.ListRides(context.Background(), &passengerIDStr, nil, nil, 10, 0)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, int64(2), count)
	assert.Equal(t, trips[0].ID, result[0].ID)
	assert.Equal(t, trips[1].ID, result[1].ID)

	// Verify mocks
	tripRepo.AssertExpectations(t)
}
