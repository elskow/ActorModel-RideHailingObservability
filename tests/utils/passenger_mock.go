package utils

import (
	"actor-model-observability/internal/models"
	"context"
	"github.com/stretchr/testify/mock"
)

// MockPassengerRepository Mock repositories for Passenger
type MockPassengerRepository struct {
	mock.Mock
}

func (m *MockPassengerRepository) Create(ctx context.Context, passenger *models.Passenger) error {
	args := m.Called(ctx, passenger)
	return args.Error(0)
}

func (m *MockPassengerRepository) GetByID(ctx context.Context, id string) (*models.Passenger, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Passenger), args.Error(1)
}

func (m *MockPassengerRepository) GetByUserID(ctx context.Context, userID string) (*models.Passenger, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*models.Passenger), args.Error(1)
}

func (m *MockPassengerRepository) Update(ctx context.Context, passenger *models.Passenger) error {
	args := m.Called(ctx, passenger)
	return args.Error(0)
}

func (m *MockPassengerRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPassengerRepository) List(ctx context.Context, limit, offset int) ([]*models.Passenger, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*models.Passenger), args.Error(1)
}
