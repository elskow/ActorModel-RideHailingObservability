package handler

import (
	"actor-model-observability/tests/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"actor-model-observability/internal/handlers"
	"actor-model-observability/internal/models"
)

// Test GetActorInstances endpoint
func TestObservabilityHandler_GetActorInstances_Success(t *testing.T) {
	router, mockObsRepo, _, obsHandler := utils.SetupObservabilityHandler()

	// Setup route
	router.GET("/api/v1/observability/actors", obsHandler.GetActorInstances)

	// Mock data
	actorID := "passenger-123"
	instances := []*models.ActorInstance{
		{
			ID:            uuid.New(),
			ActorType:     models.ActorTypePassenger,
			ActorID:       actorID,
			Status:        models.ActorStatusActive,
			LastHeartbeat: time.Now(),
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}

	// Setup mock expectations
	mockObsRepo.On("ListActorInstances", mock.Anything, "passenger", 20, 0).Return(instances, nil)

	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/observability/actors?actor_type=passenger", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response handlers.PaginatedResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 20, response.Limit)
	assert.Equal(t, 0, response.Offset)
	assert.NotNil(t, response.Data)

	mockObsRepo.AssertExpectations(t)
}

func TestObservabilityHandler_GetActorInstances_InvalidLimit(t *testing.T) {
	router, _, _, obsHandler := utils.SetupObservabilityHandler()

	// Setup route
	router.GET("/api/v1/observability/actors", obsHandler.GetActorInstances)

	// Create request with invalid limit
	req, _ := http.NewRequest("GET", "/api/v1/observability/actors?limit=0", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response handlers.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid limit", response.Error)
}

// Test GetActorMessages endpoint
func TestObservabilityHandler_GetActorMessages_Success(t *testing.T) {
	router, mockObsRepo, _, obsHandler := utils.SetupObservabilityHandler()

	// Setup route
	router.GET("/api/v1/observability/messages", obsHandler.GetActorMessages)

	// Mock data
	messages := []*models.ActorMessage{
		{
			ID:                uuid.New(),
			TraceID:           uuid.New(),
			SpanID:            uuid.New(),
			SenderActorType:   models.ActorTypePassenger,
			SenderActorID:     "passenger-123",
			ReceiverActorType: models.ActorTypeMatching,
			ReceiverActorID:   "matching-service",
			MessageType:       "ride_request",
			Status:            models.MessageStatusProcessed,
			SentAt:            time.Now(),
			CreatedAt:         time.Now(),
		},
	}

	// Setup mock expectations
	mockObsRepo.On("ListActorMessages", mock.Anything, "passenger-123", "", 20, 0).Return(messages, nil)

	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/observability/messages?from_actor=passenger-123", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response handlers.PaginatedResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 20, response.Limit)
	assert.Equal(t, 0, response.Offset)

	mockObsRepo.AssertExpectations(t)
}

func TestObservabilityHandler_GetActorMessages_WithTimeRange(t *testing.T) {
	router, mockObsRepo, _, obsHandler := utils.SetupObservabilityHandler()

	// Setup route
	router.GET("/api/v1/observability/messages", obsHandler.GetActorMessages)

	// Mock data
	var messages []*models.ActorMessage

	// Setup mock expectations
	mockObsRepo.On("GetMessagesByTimeRange", mock.Anything, "2024-01-01T00:00:00Z", "2024-01-02T00:00:00Z", 20, 0).Return(messages, nil)

	// Create request with time range
	req, _ := http.NewRequest("GET", "/api/v1/observability/messages?start_time=2024-01-01T00:00:00Z&end_time=2024-01-02T00:00:00Z", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	mockObsRepo.AssertExpectations(t)
}

// Test GetSystemMetrics endpoint
func TestObservabilityHandler_GetSystemMetrics_Success(t *testing.T) {
	router, mockObsRepo, _, obsHandler := utils.SetupObservabilityHandler()

	// Setup route
	router.GET("/api/v1/observability/metrics", obsHandler.GetSystemMetrics)

	// Mock data
	actorID := "passenger-123"
	actorType := models.ActorTypePassenger
	metrics := []*models.SystemMetric{
		{
			ID:          uuid.New(),
			MetricName:  "cpu_usage",
			MetricType:  models.MetricTypeGauge,
			MetricValue: 45.2,
			ActorType:   &actorType,
			ActorID:     &actorID,
			Timestamp:   time.Now(),
			CreatedAt:   time.Now(),
		},
	}

	// Setup mock expectations
	mockObsRepo.On("ListSystemMetrics", mock.Anything, "gauge", 20, 0).Return(metrics, nil)

	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/observability/metrics?metric_type=gauge", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response handlers.PaginatedResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 20, response.Limit)
	assert.Equal(t, 0, response.Offset)

	mockObsRepo.AssertExpectations(t)
}

// Test GetDistributedTraces endpoint
func TestObservabilityHandler_GetDistributedTraces_Success(t *testing.T) {
	router, mockObsRepo, _, obsHandler := utils.SetupObservabilityHandler()

	// Setup route
	router.GET("/api/v1/observability/traces", obsHandler.GetDistributedTraces)

	// Mock data
	actorID := "passenger-123"
	actorType := models.ActorTypePassenger
	traces := []*models.DistributedTrace{
		{
			ID:            uuid.New(),
			TraceID:       uuid.New(),
			SpanID:        uuid.New(),
			OperationName: "ride_request",
			ActorType:     &actorType,
			ActorID:       &actorID,
			StartTime:     time.Now(),
			Status:        models.TraceStatusOK,
			CreatedAt:     time.Now(),
		},
	}

	// Setup mock expectations
	mockObsRepo.On("ListDistributedTraces", mock.Anything, "ride_request", 20, 0).Return(traces, nil)

	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/observability/traces?operation=ride_request", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response handlers.PaginatedResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 20, response.Limit)
	assert.Equal(t, 0, response.Offset)

	mockObsRepo.AssertExpectations(t)
}

func TestObservabilityHandler_GetDistributedTraces_ByTraceID(t *testing.T) {
	router, mockObsRepo, _, obsHandler := utils.SetupObservabilityHandler()

	// Setup route
	router.GET("/api/v1/observability/traces", obsHandler.GetDistributedTraces)

	// Mock data
	traceID := uuid.New().String()
	var traces []*models.DistributedTrace

	// Setup mock expectations
	mockObsRepo.On("GetTracesByTraceID", mock.Anything, traceID).Return(traces, nil)

	// Create request with trace_id
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/observability/traces?trace_id=%s", traceID), nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	mockObsRepo.AssertExpectations(t)
}

// Test GetEventLogs endpoint
func TestObservabilityHandler_GetEventLogs_Success(t *testing.T) {
	router, mockObsRepo, _, obsHandler := utils.SetupObservabilityHandler()

	// Setup route
	router.GET("/api/v1/observability/events", obsHandler.GetEventLogs)

	// Mock data
	actorID := "passenger-123"
	actorType := models.ActorTypePassenger
	logs := []*models.EventLog{
		{
			ID:            uuid.New(),
			EventType:     "ride_requested",
			EventCategory: models.EventCategoryBusiness,
			ActorType:     &actorType,
			ActorID:       &actorID,
			Severity:      models.EventSeverityInfo,
			Message:       "Ride request submitted",
			Timestamp:     time.Now(),
			CreatedAt:     time.Now(),
		},
	}

	// Setup mock expectations
	mockObsRepo.On("ListEventLogs", mock.Anything, "ride_requested", "passenger-123", 20, 0).Return(logs, nil)

	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/observability/events?event_type=ride_requested&source=passenger-123", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response handlers.PaginatedResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 20, response.Limit)
	assert.Equal(t, 0, response.Offset)

	mockObsRepo.AssertExpectations(t)
}

// Test GetTraditionalMetrics endpoint
func TestObservabilityHandler_GetTraditionalMetrics_Success(t *testing.T) {
	router, _, mockTradRepo, obsHandler := utils.SetupObservabilityHandler()

	// Setup route
	router.GET("/api/v1/traditional/metrics", obsHandler.GetTraditionalMetrics)

	// Mock data
	metrics := []*models.TraditionalMetric{
		{
			ID:          uuid.New(),
			MetricName:  "http_requests_total",
			MetricType:  models.MetricTypeCounter,
			MetricValue: 150.0,
			ServiceName: "ride-service",
			InstanceID:  "instance-1",
			Timestamp:   time.Now(),
			CreatedAt:   time.Now(),
		},
	}

	// Setup mock expectations
	mockTradRepo.On("ListTraditionalMetrics", mock.Anything, "counter", "ride-service", 20, 0).Return(metrics, nil)

	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/traditional/metrics?metric_type=counter&service_name=ride-service", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response handlers.PaginatedResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 20, response.Limit)
	assert.Equal(t, 0, response.Offset)

	mockTradRepo.AssertExpectations(t)
}

// Test GetTraditionalLogs endpoint
func TestObservabilityHandler_GetTraditionalLogs_Success(t *testing.T) {
	router, _, mockTradRepo, obsHandler := utils.SetupObservabilityHandler()

	// Setup route
	router.GET("/api/v1/traditional/logs", obsHandler.GetTraditionalLogs)

	// Mock data
	logs := []*models.TraditionalLog{
		{
			ID:          uuid.New(),
			Level:       models.LogLevelInfo,
			Message:     "Request processed successfully",
			ServiceName: "ride-service",
			InstanceID:  "instance-1",
			Timestamp:   time.Now(),
			CreatedAt:   time.Now(),
		},
	}

	// Setup mock expectations
	mockTradRepo.On("ListTraditionalLogs", mock.Anything, "info", "ride-service", 20, 0).Return(logs, nil)

	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/traditional/logs?level=info&service_name=ride-service", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response handlers.PaginatedResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 20, response.Limit)
	assert.Equal(t, 0, response.Offset)

	mockTradRepo.AssertExpectations(t)
}

// Test GetServiceHealth endpoint
func TestObservabilityHandler_GetServiceHealth_Success(t *testing.T) {
	router, _, _, obsHandler := utils.SetupObservabilityHandler()

	// Setup route
	router.GET("/api/v1/traditional/health", obsHandler.GetServiceHealth)

	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/traditional/health", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response handlers.PaginatedResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 20, response.Limit)
	assert.Equal(t, 0, response.Offset)
	assert.NotNil(t, response.Data)
}

// Test GetLatestServiceHealth endpoint
func TestObservabilityHandler_GetLatestServiceHealth_Success(t *testing.T) {
	router, _, _, obsHandler := utils.SetupObservabilityHandler()

	// Setup route
	router.GET("/api/v1/traditional/health/:service_name/latest", obsHandler.GetLatestServiceHealth)

	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/traditional/health/ride-service/latest", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
	assert.NotNil(t, response["timestamp"])
}

func TestObservabilityHandler_GetLatestServiceHealth_EmptyServiceName(t *testing.T) {
	router, _, _, obsHandler := utils.SetupObservabilityHandler()

	// Setup route
	router.GET("/api/v1/traditional/health/:service_name/latest", obsHandler.GetLatestServiceHealth)

	// Create request with empty service name
	req, _ := http.NewRequest("GET", "/api/v1/traditional/health//latest", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response handlers.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid service name", response.Error)
}

// Test GetPrometheusMetrics endpoint
func TestObservabilityHandler_GetPrometheusMetrics_Success(t *testing.T) {
	router, mockObsRepo, mockTradRepo, obsHandler := utils.SetupObservabilityHandler()

	// Setup route
	router.GET("/api/v1/observability/prometheus", obsHandler.GetPrometheusMetrics)

	// Mock data
	actorID := "passenger-123"
	actorType := models.ActorTypePassenger
	systemMetrics := []*models.SystemMetric{
		{
			ID:          uuid.New(),
			MetricName:  "cpu_usage",
			MetricType:  models.MetricTypeGauge,
			MetricValue: 45.2,
			ActorType:   &actorType,
			ActorID:     &actorID,
			Timestamp:   time.Now(),
			CreatedAt:   time.Now(),
		},
	}

	tradMetrics := []*models.TraditionalMetric{
		{
			ID:          uuid.New(),
			MetricName:  "http_requests_total",
			MetricType:  models.MetricTypeCounter,
			MetricValue: 150.0,
			ServiceName: "ride-service",
			Timestamp:   time.Now(),
			CreatedAt:   time.Now(),
		},
	}

	// Setup mock expectations
	mockObsRepo.On("ListSystemMetrics", mock.Anything, "", 100, 0).Return(systemMetrics, nil)
	mockTradRepo.On("ListTraditionalMetrics", mock.Anything, "", "", 100, 0).Return(tradMetrics, nil)

	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/observability/prometheus", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/plain; charset=utf-8", w.Header().Get("Content-Type"))

	// Check that response contains Prometheus format
	body := w.Body.String()
	assert.Contains(t, body, "# HELP")
	assert.Contains(t, body, "# TYPE")
	assert.Contains(t, body, "actor_system_cpu_usage")
	assert.Contains(t, body, "http_requests_total")

	mockObsRepo.AssertExpectations(t)
	mockTradRepo.AssertExpectations(t)
}

// Test GetTraditionalPrometheusMetrics endpoint
func TestObservabilityHandler_GetTraditionalPrometheusMetrics_Success(t *testing.T) {
	router, _, mockTradRepo, obsHandler := utils.SetupObservabilityHandler()

	// Setup route
	router.GET("/api/v1/traditional/prometheus", obsHandler.GetTraditionalPrometheusMetrics)

	// Mock data
	metrics := []*models.TraditionalMetric{
		{
			ID:          uuid.New(),
			MetricName:  "http_requests_total",
			MetricType:  models.MetricTypeCounter,
			MetricValue: 150.0,
			ServiceName: "ride-service",
			Timestamp:   time.Now(),
			CreatedAt:   time.Now(),
		},
	}

	// Setup mock expectations
	mockTradRepo.On("ListTraditionalMetrics", mock.Anything, "", "", 100, 0).Return(metrics, nil)

	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/traditional/prometheus", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/plain; charset=utf-8", w.Header().Get("Content-Type"))

	// Check that response contains Prometheus format
	body := w.Body.String()
	assert.Contains(t, body, "# HELP")
	assert.Contains(t, body, "# TYPE")
	assert.Contains(t, body, "http_requests_total")
	assert.Contains(t, body, "traditional_system_up 1")

	mockTradRepo.AssertExpectations(t)
}

// Test error scenarios
func TestObservabilityHandler_GetActorInstances_RepositoryError(t *testing.T) {
	router, mockObsRepo, _, obsHandler := utils.SetupObservabilityHandler()

	// Setup route
	router.GET("/api/v1/observability/actors", obsHandler.GetActorInstances)

	// Setup mock expectations with error
	mockObsRepo.On("ListActorInstances", mock.Anything, "", 20, 0).Return([]*models.ActorInstance{}, fmt.Errorf("database error"))

	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/observability/actors", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response handlers.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Internal server error", response.Error)
	assert.Equal(t, "Failed to list actor instances", response.Message)

	mockObsRepo.AssertExpectations(t)
}

func TestObservabilityHandler_GetSystemMetrics_InvalidOffset(t *testing.T) {
	router, _, _, obsHandler := utils.SetupObservabilityHandler()

	// Setup route
	router.GET("/api/v1/observability/metrics", obsHandler.GetSystemMetrics)

	// Create request with invalid offset
	req, _ := http.NewRequest("GET", "/api/v1/observability/metrics?offset=-1", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response handlers.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid offset", response.Error)
}

func TestObservabilityHandler_GetPrometheusMetrics_RepositoryError(t *testing.T) {
	router, mockObsRepo, _, obsHandler := utils.SetupObservabilityHandler()

	// Setup route
	router.GET("/api/v1/observability/prometheus", obsHandler.GetPrometheusMetrics)

	// Setup mock expectations with error
	mockObsRepo.On("ListSystemMetrics", mock.Anything, "", 100, 0).Return([]*models.SystemMetric{}, fmt.Errorf("database error"))

	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/observability/prometheus", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "text/plain; charset=utf-8", w.Header().Get("Content-Type"))

	body := w.Body.String()
	assert.Contains(t, body, "# Error getting metrics")

	mockObsRepo.AssertExpectations(t)
}
