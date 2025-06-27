package utils

import (
	"actor-model-observability/internal/models"
	"context"
	"github.com/stretchr/testify/mock"
)

// MockTraditionalRepository is a mock implementation of TraditionalRepository
type MockTraditionalRepository struct {
	mock.Mock
}

func (m *MockTraditionalRepository) CreateTraditionalMetric(ctx context.Context, metric *models.TraditionalMetric) error {
	args := m.Called(ctx, metric)
	return args.Error(0)
}

func (m *MockTraditionalRepository) GetTraditionalMetric(ctx context.Context, id string) (*models.TraditionalMetric, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.TraditionalMetric), args.Error(1)
}

func (m *MockTraditionalRepository) ListTraditionalMetrics(ctx context.Context, name, metricType string, limit, offset int) ([]*models.TraditionalMetric, error) {
	args := m.Called(ctx, name, metricType, limit, offset)
	return args.Get(0).([]*models.TraditionalMetric), args.Error(1)
}

func (m *MockTraditionalRepository) GetTraditionalMetricsByTimeRange(ctx context.Context, startTime, endTime string, limit, offset int) ([]*models.TraditionalMetric, error) {
	args := m.Called(ctx, startTime, endTime, limit, offset)
	return args.Get(0).([]*models.TraditionalMetric), args.Error(1)
}

func (m *MockTraditionalRepository) CreateTraditionalLog(ctx context.Context, log *models.TraditionalLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockTraditionalRepository) GetTraditionalLog(ctx context.Context, id string) (*models.TraditionalLog, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.TraditionalLog), args.Error(1)
}

func (m *MockTraditionalRepository) ListTraditionalLogs(ctx context.Context, level, source string, limit, offset int) ([]*models.TraditionalLog, error) {
	args := m.Called(ctx, level, source, limit, offset)
	return args.Get(0).([]*models.TraditionalLog), args.Error(1)
}

func (m *MockTraditionalRepository) GetTraditionalLogsByTimeRange(ctx context.Context, startTime, endTime string, limit, offset int) ([]*models.TraditionalLog, error) {
	args := m.Called(ctx, startTime, endTime, limit, offset)
	return args.Get(0).([]*models.TraditionalLog), args.Error(1)
}
