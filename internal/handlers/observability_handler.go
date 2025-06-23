package handlers

import (
	"net/http"
	"strconv"

	"actor-model-observability/internal/repository"
	"github.com/gin-gonic/gin"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Limit      int         `json:"limit"`
	Offset     int         `json:"offset"`
	HasMore    bool        `json:"has_more"`
}

// ObservabilityHandler handles observability-related HTTP requests
type ObservabilityHandler struct {
	obsRepo         repository.ObservabilityRepository
	traditionalRepo repository.TraditionalRepository
}

// NewObservabilityHandler creates a new ObservabilityHandler instance
func NewObservabilityHandler(
	obsRepo repository.ObservabilityRepository,
	traditionalRepo repository.TraditionalRepository,
) *ObservabilityHandler {
	return &ObservabilityHandler{
		obsRepo:         obsRepo,
		traditionalRepo: traditionalRepo,
	}
}

// GetActorInstances handles actor instances listing
// @Summary List actor instances
// @Description Get a paginated list of actor instances
// @Tags observability
// @Produce json
// @Param actor_type query string false "Filter by actor type"
// @Param limit query int false "Number of items per page" default(20)
// @Param offset query int false "Number of items to skip" default(0)
// @Success 200 {object} PaginatedResponse{data=[]models.ActorInstance}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/observability/actors [get]
func (h *ObservabilityHandler) GetActorInstances(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	actorType := c.Query("actor_type")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid limit",
			Message: "Limit must be a positive integer between 1 and 100",
		})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid offset",
			Message: "Offset must be a non-negative integer",
		})
		return
	}

	// Get actor instances from repository
	actors, err := h.obsRepo.ListActorInstances(c.Request.Context(), actorType, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to list actor instances",
		})
		return
	}

	// Return paginated response
	c.JSON(http.StatusOK, PaginatedResponse{
		Data:   actors,
		Limit:  limit,
		Offset: offset,
		Total:      int64(len(actors)),
	})
}

// GetActorMessages handles actor messages listing
// @Summary List actor messages
// @Description Get a paginated list of actor messages with optional filtering
// @Tags observability
// @Produce json
// @Param from_actor query string false "Filter by sender actor ID"
// @Param to_actor query string false "Filter by receiver actor ID"
// @Param start_time query string false "Start time (RFC3339 format)"
// @Param end_time query string false "End time (RFC3339 format)"
// @Param limit query int false "Number of items per page" default(20)
// @Param offset query int false "Number of items to skip" default(0)
// @Success 200 {object} PaginatedResponse{data=[]models.ActorMessage}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/observability/messages [get]
func (h *ObservabilityHandler) GetActorMessages(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	fromActor := c.Query("from_actor")
	toActor := c.Query("to_actor")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid limit",
			Message: "Limit must be a positive integer between 1 and 100",
		})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid offset",
			Message: "Offset must be a non-negative integer",
		})
		return
	}

	// Get messages from repository
	var messages interface{}
	if startTime != "" && endTime != "" {
		messages, err = h.obsRepo.GetMessagesByTimeRange(c.Request.Context(), startTime, endTime, limit, offset)
	} else {
		messages, err = h.obsRepo.ListActorMessages(c.Request.Context(), fromActor, toActor, limit, offset)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to list actor messages",
		})
		return
	}

	// Return paginated response
	c.JSON(http.StatusOK, PaginatedResponse{
		Data:   messages,
		Limit:  limit,
		Offset: offset,
		Total:  0, // Would need separate count query in real implementation
	})
}

// GetSystemMetrics handles system metrics listing
// @Summary List system metrics
// @Description Get a paginated list of system metrics with optional filtering
// @Tags observability
// @Produce json
// @Param metric_type query string false "Filter by metric type"
// @Param start_time query string false "Start time (RFC3339 format)"
// @Param end_time query string false "End time (RFC3339 format)"
// @Param limit query int false "Number of items per page" default(20)
// @Param offset query int false "Number of items to skip" default(0)
// @Success 200 {object} PaginatedResponse{data=[]models.SystemMetric}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/observability/metrics [get]
func (h *ObservabilityHandler) GetSystemMetrics(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	metricType := c.Query("metric_type")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid limit",
			Message: "Limit must be a positive integer between 1 and 100",
		})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid offset",
			Message: "Offset must be a non-negative integer",
		})
		return
	}

	// Get metrics from repository
	var metrics interface{}
	if startTime != "" && endTime != "" {
		metrics, err = h.obsRepo.GetMetricsByTimeRange(c.Request.Context(), startTime, endTime, limit, offset)
	} else {
		metrics, err = h.obsRepo.ListSystemMetrics(c.Request.Context(), metricType, limit, offset)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to list system metrics",
		})
		return
	}

	// Return paginated response
	c.JSON(http.StatusOK, PaginatedResponse{
		Data:   metrics,
		Limit:  limit,
		Offset: offset,
		Total:  0, // Would need separate count query in real implementation
	})
}

// GetDistributedTraces handles distributed traces listing
// @Summary List distributed traces
// @Description Get a paginated list of distributed traces with optional filtering
// @Tags observability
// @Produce json
// @Param operation query string false "Filter by operation name"
// @Param trace_id query string false "Get all spans for a specific trace ID"
// @Param limit query int false "Number of items per page" default(20)
// @Param offset query int false "Number of items to skip" default(0)
// @Success 200 {object} PaginatedResponse{data=[]models.DistributedTrace}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/observability/traces [get]
func (h *ObservabilityHandler) GetDistributedTraces(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	operation := c.Query("operation")
	traceID := c.Query("trace_id")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid limit",
			Message: "Limit must be a positive integer between 1 and 100",
		})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid offset",
			Message: "Offset must be a non-negative integer",
		})
		return
	}

	// Get traces from repository
	var traces interface{}
	if traceID != "" {
		traces, err = h.obsRepo.GetTracesByTraceID(c.Request.Context(), traceID)
	} else {
		traces, err = h.obsRepo.ListDistributedTraces(c.Request.Context(), operation, limit, offset)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to list distributed traces",
		})
		return
	}

	// Return paginated response
	c.JSON(http.StatusOK, PaginatedResponse{
		Data:   traces,
		Limit:  limit,
		Offset: offset,
		Total:  0, // Would need separate count query in real implementation
	})
}

// GetEventLogs handles event logs listing
// @Summary List event logs
// @Description Get a paginated list of event logs with optional filtering
// @Tags observability
// @Produce json
// @Param event_type query string false "Filter by event type"
// @Param source query string false "Filter by source (actor ID)"
// @Param start_time query string false "Start time (RFC3339 format)"
// @Param end_time query string false "End time (RFC3339 format)"
// @Param limit query int false "Number of items per page" default(20)
// @Param offset query int false "Number of items to skip" default(0)
// @Success 200 {object} PaginatedResponse{data=[]models.EventLog}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/observability/events [get]
func (h *ObservabilityHandler) GetEventLogs(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	eventType := c.Query("event_type")
	source := c.Query("source")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid limit",
			Message: "Limit must be a positive integer between 1 and 100",
		})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid offset",
			Message: "Offset must be a non-negative integer",
		})
		return
	}

	// Get event logs from repository
	var logs interface{}
	if startTime != "" && endTime != "" {
		logs, err = h.obsRepo.GetEventLogsByTimeRange(c.Request.Context(), startTime, endTime, limit, offset)
	} else {
		logs, err = h.obsRepo.ListEventLogs(c.Request.Context(), eventType, source, limit, offset)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to list event logs",
		})
		return
	}

	// Return paginated response
	c.JSON(http.StatusOK, PaginatedResponse{
		Data:   logs,
		Limit:  limit,
		Offset: offset,
		Total:  0, // Would need separate count query in real implementation
	})
}

// GetTraditionalMetrics handles traditional metrics listing
// @Summary List traditional metrics
// @Description Get a paginated list of traditional monitoring metrics
// @Tags traditional
// @Produce json
// @Param metric_type query string false "Filter by metric type"
// @Param service_name query string false "Filter by service name"
// @Param start_time query string false "Start time (RFC3339 format)"
// @Param end_time query string false "End time (RFC3339 format)"
// @Param limit query int false "Number of items per page" default(20)
// @Param offset query int false "Number of items to skip" default(0)
// @Success 200 {object} PaginatedResponse{data=[]models.TraditionalMetric}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/traditional/metrics [get]
func (h *ObservabilityHandler) GetTraditionalMetrics(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	metricType := c.Query("metric_type")
	serviceName := c.Query("service_name")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid limit",
			Message: "Limit must be a positive integer between 1 and 100",
		})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid offset",
			Message: "Offset must be a non-negative integer",
		})
		return
	}

	// Get traditional metrics from repository
	var metrics interface{}
	if startTime != "" && endTime != "" {
		metrics, err = h.traditionalRepo.GetTraditionalMetricsByTimeRange(c.Request.Context(), startTime, endTime, limit, offset)
	} else {
		metrics, err = h.traditionalRepo.ListTraditionalMetrics(c.Request.Context(), metricType, serviceName, limit, offset)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to list traditional metrics",
		})
		return
	}

	// Return paginated response
	c.JSON(http.StatusOK, PaginatedResponse{
		Data:   metrics,
		Limit:  limit,
		Offset: offset,
		Total:  0, // Would need separate count query in real implementation
	})
}

// GetTraditionalLogs handles traditional logs listing
// @Summary List traditional logs
// @Description Get a paginated list of traditional application logs
// @Tags traditional
// @Produce json
// @Param level query string false "Filter by log level"
// @Param service_name query string false "Filter by service name"
// @Param start_time query string false "Start time (RFC3339 format)"
// @Param end_time query string false "End time (RFC3339 format)"
// @Param limit query int false "Number of items per page" default(20)
// @Param offset query int false "Number of items to skip" default(0)
// @Success 200 {object} PaginatedResponse{data=[]traditional.TraditionalLog}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/traditional/logs [get]
func (h *ObservabilityHandler) GetTraditionalLogs(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	level := c.Query("level")
	serviceName := c.Query("service_name")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid limit",
			Message: "Limit must be a positive integer between 1 and 100",
		})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid offset",
			Message: "Offset must be a non-negative integer",
		})
		return
	}

	// Get traditional logs from repository
	var logs interface{}
	if startTime != "" && endTime != "" {
		logs, err = h.traditionalRepo.GetTraditionalLogsByTimeRange(c.Request.Context(), startTime, endTime, limit, offset)
	} else {
		logs, err = h.traditionalRepo.ListTraditionalLogs(c.Request.Context(), level, serviceName, limit, offset)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to list traditional logs",
		})
		return
	}

	// Return paginated response
	c.JSON(http.StatusOK, PaginatedResponse{
		Data:   logs,
		Limit:  limit,
		Offset: offset,
		Total:  0, // Would need separate count query in real implementation
	})
}

// GetServiceHealth handles service health status retrieval
// @Summary Get service health
// @Description Get the current health status of services
// @Tags traditional
// @Produce json
// @Param service_name query string false "Filter by service name"
// @Param start_time query string false "Start time (RFC3339 format)"
// @Param end_time query string false "End time (RFC3339 format)"
// @Param limit query int false "Number of items per page" default(20)
// @Param offset query int false "Number of items to skip" default(0)
// @Success 200 {object} PaginatedResponse{data=[]traditional.ServiceHealth}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/traditional/health [get]
func (h *ObservabilityHandler) GetServiceHealth(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	_ = c.Query("service_name")     // serviceName - unused for now
	_ = c.Query("start_time")       // startTime - unused for now
	_ = c.Query("end_time")         // endTime - unused for now

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid limit",
			Message: "Limit must be a positive integer between 1 and 100",
		})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid offset",
			Message: "Offset must be a non-negative integer",
		})
		return
	}

	// Get service health from repository
	var health interface{}
	// TODO: Implement service health methods in TraditionalRepository
	// if startTime != "" && endTime != "" {
	//	health, err = h.traditionalRepo.GetServiceHealthByTimeRange(c.Request.Context(), serviceName, startTime, endTime, limit, offset)
	// } else {
	//	health, err = h.traditionalRepo.ListServiceHealth(c.Request.Context(), serviceName, limit, offset)
	// }
	health = []interface{}{} // Placeholder empty slice
	err = nil

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to get service health",
		})
		return
	}

	// Return paginated response
	c.JSON(http.StatusOK, PaginatedResponse{
		Data:   health,
		Limit:  limit,
		Offset: offset,
		Total:  0, // Would need separate count query in real implementation
	})
}

// GetLatestServiceHealth handles latest service health status retrieval
// @Summary Get latest service health
// @Description Get the latest health status for a specific service
// @Tags traditional
// @Produce json
// @Param service_name path string true "Service name"
// @Success 200 {object} traditional.ServiceHealth
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/traditional/health/{service_name}/latest [get]
func (h *ObservabilityHandler) GetLatestServiceHealth(c *gin.Context) {
	serviceName := c.Param("service_name")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid service name",
			Message: "Service name is required",
		})
		return
	}

	// TODO: Implement GetLatestServiceHealth in TraditionalRepository
	// health, err := h.traditionalRepo.GetLatestServiceHealth(c.Request.Context(), serviceName)
	// if err != nil {
	//	switch err.(type) {
	//	case *models.NotFoundError:
	//		c.JSON(http.StatusNotFound, ErrorResponse{
	//			Error:   "Service health not found",
	//			Message: err.Error(),
	//		})
	//	default:
	//		c.JSON(http.StatusInternalServerError, ErrorResponse{
	//			Error:   "Internal server error",
	//			Message: "Failed to get service health",
	//		})
	//	}
	//	return
	// }

	c.JSON(http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"timestamp": "2024-01-01T00:00:00Z", // Placeholder
	})
}