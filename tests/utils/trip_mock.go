package utils

import (
	"actor-model-observability/internal/models"
	"context"
	"github.com/stretchr/testify/mock"
)

// MockTripRepository Mock repositories for Trip
type MockTripRepository struct {
	mock.Mock
}

func (m *MockTripRepository) Create(ctx context.Context, trip *models.Trip) error {
	args := m.Called(ctx, trip)
	return args.Error(0)
}

func (m *MockTripRepository) GetByID(ctx context.Context, id string) (*models.Trip, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Trip), args.Error(1)
}

func (m *MockTripRepository) Update(ctx context.Context, trip *models.Trip) error {
	args := m.Called(ctx, trip)
	return args.Error(0)
}

func (m *MockTripRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTripRepository) GetByPassengerID(ctx context.Context, passengerID string, limit, offset int) ([]*models.Trip, error) {
	args := m.Called(ctx, passengerID, limit, offset)
	return args.Get(0).([]*models.Trip), args.Error(1)
}

func (m *MockTripRepository) GetByDriverID(ctx context.Context, driverID string, limit, offset int) ([]*models.Trip, error) {
	args := m.Called(ctx, driverID, limit, offset)
	return args.Get(0).([]*models.Trip), args.Error(1)
}

func (m *MockTripRepository) GetActiveTrips(ctx context.Context) ([]*models.Trip, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.Trip), args.Error(1)
}

func (m *MockTripRepository) GetTripsByStatus(ctx context.Context, status models.TripStatus, limit, offset int) ([]*models.Trip, error) {
	args := m.Called(ctx, status, limit, offset)
	return args.Get(0).([]*models.Trip), args.Error(1)
}

func (m *MockTripRepository) GetTripsByDateRange(ctx context.Context, startDate, endDate string, limit, offset int) ([]*models.Trip, error) {
	args := m.Called(ctx, startDate, endDate, limit, offset)
	return args.Get(0).([]*models.Trip), args.Error(1)
}

func (m *MockTripRepository) List(ctx context.Context, limit, offset int) ([]*models.Trip, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*models.Trip), args.Error(1)
}
