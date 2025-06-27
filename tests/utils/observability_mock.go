package utils

import (
	"actor-model-observability/internal/handlers"
	"actor-model-observability/internal/models"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
)

// MockObservabilityRepository is a mock implementation of ObservabilityRepository
type MockObservabilityRepository struct {
	mock.Mock
}

func (m *MockObservabilityRepository) CreateActorInstance(ctx context.Context, instance *models.ActorInstance) error {
	args := m.Called(ctx, instance)
	return args.Error(0)
}

func (m *MockObservabilityRepository) GetActorInstance(ctx context.Context, id string) (*models.ActorInstance, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.ActorInstance), args.Error(1)
}

func (m *MockObservabilityRepository) UpdateActorInstance(ctx context.Context, instance *models.ActorInstance) error {
	args := m.Called(ctx, instance)
	return args.Error(0)
}

func (m *MockObservabilityRepository) ListActorInstances(ctx context.Context, actorType string, limit, offset int) ([]*models.ActorInstance, error) {
	args := m.Called(ctx, actorType, limit, offset)
	return args.Get(0).([]*models.ActorInstance), args.Error(1)
}

func (m *MockObservabilityRepository) CreateActorMessage(ctx context.Context, message *models.ActorMessage) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockObservabilityRepository) GetActorMessage(ctx context.Context, id string) (*models.ActorMessage, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.ActorMessage), args.Error(1)
}

func (m *MockObservabilityRepository) ListActorMessages(ctx context.Context, fromActor, toActor string, limit, offset int) ([]*models.ActorMessage, error) {
	args := m.Called(ctx, fromActor, toActor, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ActorMessage), args.Error(1)
}

func (m *MockObservabilityRepository) GetMessagesByTimeRange(ctx context.Context, startTime, endTime string, limit, offset int) ([]*models.ActorMessage, error) {
	args := m.Called(ctx, startTime, endTime, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ActorMessage), args.Error(1)
}

func (m *MockObservabilityRepository) CreateSystemMetric(ctx context.Context, metric *models.SystemMetric) error {
	args := m.Called(ctx, metric)
	return args.Error(0)
}

func (m *MockObservabilityRepository) GetSystemMetric(ctx context.Context, id string) (*models.SystemMetric, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.SystemMetric), args.Error(1)
}

func (m *MockObservabilityRepository) ListSystemMetrics(ctx context.Context, metricType string, limit, offset int) ([]*models.SystemMetric, error) {
	args := m.Called(ctx, metricType, limit, offset)
	return args.Get(0).([]*models.SystemMetric), args.Error(1)
}

func (m *MockObservabilityRepository) GetMetricsByTimeRange(ctx context.Context, startTime, endTime string, limit, offset int) ([]*models.SystemMetric, error) {
	args := m.Called(ctx, startTime, endTime, limit, offset)
	return args.Get(0).([]*models.SystemMetric), args.Error(1)
}

func (m *MockObservabilityRepository) CreateDistributedTrace(ctx context.Context, trace *models.DistributedTrace) error {
	args := m.Called(ctx, trace)
	return args.Error(0)
}

func (m *MockObservabilityRepository) GetDistributedTrace(ctx context.Context, id string) (*models.DistributedTrace, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.DistributedTrace), args.Error(1)
}

func (m *MockObservabilityRepository) GetTracesByTraceID(ctx context.Context, traceID string) ([]*models.DistributedTrace, error) {
	args := m.Called(ctx, traceID)
	return args.Get(0).([]*models.DistributedTrace), args.Error(1)
}

func (m *MockObservabilityRepository) ListDistributedTraces(ctx context.Context, operation string, limit, offset int) ([]*models.DistributedTrace, error) {
	args := m.Called(ctx, operation, limit, offset)
	return args.Get(0).([]*models.DistributedTrace), args.Error(1)
}

func (m *MockObservabilityRepository) CreateEventLog(ctx context.Context, log *models.EventLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockObservabilityRepository) GetEventLog(ctx context.Context, id string) (*models.EventLog, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.EventLog), args.Error(1)
}

func (m *MockObservabilityRepository) ListEventLogs(ctx context.Context, eventType, source string, limit, offset int) ([]*models.EventLog, error) {
	args := m.Called(ctx, eventType, source, limit, offset)
	return args.Get(0).([]*models.EventLog), args.Error(1)
}

func (m *MockObservabilityRepository) GetEventLogsByTimeRange(ctx context.Context, startTime, endTime string, limit, offset int) ([]*models.EventLog, error) {
	args := m.Called(ctx, startTime, endTime, limit, offset)
	return args.Get(0).([]*models.EventLog), args.Error(1)
}

// SetupObservabilityHandler creates a test setup for observability handler
func SetupObservabilityHandler() (*gin.Engine, *MockObservabilityRepository, *MockTraditionalRepository, *handlers.ObservabilityHandler) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockObsRepo := &MockObservabilityRepository{}
	mockTradRepo := &MockTraditionalRepository{}

	obsHandler := handlers.NewObservabilityHandler(mockObsRepo, mockTradRepo)

	return router, mockObsRepo, mockTradRepo, obsHandler
}
