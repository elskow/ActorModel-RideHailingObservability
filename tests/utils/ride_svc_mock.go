package utils

import (
	"actor-model-observability/internal/handlers"
	"actor-model-observability/internal/models"
	"actor-model-observability/internal/service"
	"context"
	"github.com/stretchr/testify/mock"
)

// MockRideService is a mock implementation of RideServiceInterface
type MockRideService struct {
	mock.Mock
}

var _ service.RideServiceInterface = (*MockRideService)(nil)

func (m *MockRideService) RequestRide(ctx context.Context, passengerID string, pickup, dropoff models.Location, pickupAddr, dropoffAddr string) (*models.Trip, error) {
	args := m.Called(ctx, passengerID, pickup, dropoff, pickupAddr, dropoffAddr)
	return args.Get(0).(*models.Trip), args.Error(1)
}

func (m *MockRideService) CancelRide(ctx context.Context, tripID string, reason string) error {
	args := m.Called(ctx, tripID, reason)
	return args.Error(0)
}

func (m *MockRideService) GetTripStatus(ctx context.Context, tripID string) (*models.Trip, error) {
	args := m.Called(ctx, tripID)
	return args.Get(0).(*models.Trip), args.Error(1)
}

func (m *MockRideService) ListRides(ctx context.Context, passengerID, driverID *string, status *string, limit, offset int) ([]*models.Trip, int64, error) {
	args := m.Called(ctx, passengerID, driverID, status, limit, offset)
	return args.Get(0).([]*models.Trip), args.Get(1).(int64), args.Error(2)
}

// SetupRideHandler creates a test handler with mocked dependencies
func SetupRideHandler() (*handlers.RideHandler, *MockRideService) {
	mockService := new(MockRideService)
	handler := handlers.NewRideHandler(mockService)
	return handler, mockService
}
