package utils

import (
	"actor-model-observability/internal/models"
	"context"
	"github.com/stretchr/testify/mock"
)

// MockDriverRepository MockUserRepository Mock repositories
type MockDriverRepository struct {
	mock.Mock
}

func (m *MockDriverRepository) Create(ctx context.Context, driver *models.Driver) error {
	args := m.Called(ctx, driver)
	return args.Error(0)
}

func (m *MockDriverRepository) GetByID(ctx context.Context, id string) (*models.Driver, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Driver), args.Error(1)
}

func (m *MockDriverRepository) GetByUserID(ctx context.Context, userID string) (*models.Driver, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*models.Driver), args.Error(1)
}

func (m *MockDriverRepository) Update(ctx context.Context, driver *models.Driver) error {
	args := m.Called(ctx, driver)
	return args.Error(0)
}

func (m *MockDriverRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDriverRepository) GetOnlineDrivers(ctx context.Context) ([]*models.Driver, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.Driver), args.Error(1)
}

func (m *MockDriverRepository) GetDriversInRadius(ctx context.Context, lat, lng, radiusKm float64) ([]*models.Driver, error) {
	args := m.Called(ctx, lat, lng, radiusKm)
	return args.Get(0).([]*models.Driver), args.Error(1)
}

func (m *MockDriverRepository) UpdateLocation(ctx context.Context, driverID string, lat, lng float64) error {
	args := m.Called(ctx, driverID, lat, lng)
	return args.Error(0)
}

func (m *MockDriverRepository) UpdateStatus(ctx context.Context, driverID string, status models.DriverStatus) error {
	args := m.Called(ctx, driverID, status)
	return args.Error(0)
}

func (m *MockDriverRepository) List(ctx context.Context, limit, offset int) ([]*models.Driver, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*models.Driver), args.Error(1)
}
