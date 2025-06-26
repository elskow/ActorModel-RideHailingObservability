package tests

import (
	"context"
	"testing"

	"actor-model-observability/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// SetupMockDB creates a mock database connection for testing
func SetupMockDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(mockDB, "postgres")
	return sqlxDB, mock
}

// stringPtr returns a pointer to the given string
func stringPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to the given int
func IntPtr(i int) *int {
	return &i
}

// Mock repositories
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByPhone(ctx context.Context, phone string) (*models.User, error) {
	args := m.Called(ctx, phone)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*models.User), args.Error(1)
}

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
