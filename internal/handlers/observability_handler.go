package handlers

import (
	"fmt"
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

// GetTraditionalPrometheusMetrics handles traditional metrics in Prometheus format
// @Summary Get traditional metrics in Prometheus format
// @Description Get traditional monitoring metrics in Prometheus format for scraping
// @Tags traditional
// @Produce plain
// @Success 200 {string} string "Prometheus metrics"
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/traditional/prometheus [get]
func (h *ObservabilityHandler) GetTraditionalPrometheusMetrics(c *gin.Context) {
	// Get traditional metrics from repository
	metrics, err := h.traditionalRepo.ListTraditionalMetrics(c.Request.Context(), "", "", 100, 0)
	if err != nil {
		c.Header("Content-Type", "text/plain; charset=utf-8")
		c.String(http.StatusInternalServerError, "# Error getting traditional metrics\n")
		return
	}

	// Convert to Prometheus format
	var prometheusMetrics string
	
	// Add help and type information for traditional metrics
	prometheusMetrics += "# HELP http_requests_total Total HTTP requests\n"
	prometheusMetrics += "# TYPE http_requests_total counter\n"
	prometheusMetrics += "# HELP http_request_duration_ms HTTP request duration in milliseconds\n"
	prometheusMetrics += "# TYPE http_request_duration_ms histogram\n"
	prometheusMetrics += "# HELP database_queries_total Total database operations\n"
	prometheusMetrics += "# TYPE database_queries_total counter\n"
	prometheusMetrics += "# HELP db_operation_duration_ms Database operation duration in milliseconds\n"
	prometheusMetrics += "# TYPE db_operation_duration_ms histogram\n"
	prometheusMetrics += "# HELP cache_operations_total Total cache operations\n"
	prometheusMetrics += "# TYPE cache_operations_total counter\n"
	prometheusMetrics += "# HELP system_cpu_usage System CPU usage percentage\n"
	prometheusMetrics += "# TYPE system_cpu_usage gauge\n"
	prometheusMetrics += "# HELP system_memory_usage System memory usage percentage\n"
	prometheusMetrics += "# TYPE system_memory_usage gauge\n"
	prometheusMetrics += "# HELP traditional_system_up Traditional system up status\n"
	prometheusMetrics += "# TYPE traditional_system_up gauge\n"

	// Convert traditional metrics to Prometheus format
	if metrics != nil {
		for _, metric := range metrics {
			// Convert metric name to Prometheus format
			metricName := metric.MetricName
			labels := ""
			if metric.Labels != nil && len(metric.Labels) > 0 {
				// Parse labels if they exist
				labels = fmt.Sprintf("{service=\"%s\"}", metric.ServiceName)
			}
			
			prometheusMetrics += fmt.Sprintf("%s%s %f\n", 
				metricName, labels, metric.MetricValue)
		}
	}

	// Add some sample traditional metrics (only if no real metrics exist)
	if len(metrics) == 0 {
		prometheusMetrics += "http_requests_total{endpoint=\"/api/rides\",method=\"POST\",status=\"200\"} 150\n"
		prometheusMetrics += "http_requests_total{endpoint=\"/api/rides\",method=\"GET\",status=\"200\"} 89\n"
		prometheusMetrics += "database_queries_total{operation=\"SELECT\",table=\"rides\",status=\"success\"} 245\n"
		prometheusMetrics += "database_queries_total{operation=\"INSERT\",table=\"rides\",status=\"success\"} 150\n"
		prometheusMetrics += "cache_operations_total{operation=\"GET\",result=\"hit\"} 178\n"
		prometheusMetrics += "cache_operations_total{operation=\"GET\",result=\"miss\"} 67\n"
	}
	prometheusMetrics += "system_cpu_usage 45.2\n"
	prometheusMetrics += "system_memory_usage 67.8\n"
	prometheusMetrics += "traditional_system_up 1\n"

	// Add histogram buckets for traditional latency metrics
	prometheusMetrics += "http_request_duration_ms_bucket{endpoint=\"/api/rides\",le=\"50\"} 89\n"
	prometheusMetrics += "http_request_duration_ms_bucket{endpoint=\"/api/rides\",le=\"100\"} 134\n"
	prometheusMetrics += "http_request_duration_ms_bucket{endpoint=\"/api/rides\",le=\"250\"} 145\n"
	prometheusMetrics += "http_request_duration_ms_bucket{endpoint=\"/api/rides\",le=\"500\"} 149\n"
	prometheusMetrics += "http_request_duration_ms_bucket{endpoint=\"/api/rides\",le=\"+Inf\"} 150\n"
	prometheusMetrics += "http_request_duration_ms_sum 12750\n"
	prometheusMetrics += "http_request_duration_ms_count 150\n"

	prometheusMetrics += "db_operation_duration_ms_bucket{operation=\"SELECT\",le=\"10\"} 189\n"
	prometheusMetrics += "db_operation_duration_ms_bucket{operation=\"SELECT\",le=\"25\"} 230\n"
	prometheusMetrics += "db_operation_duration_ms_bucket{operation=\"SELECT\",le=\"50\"} 240\n"
	prometheusMetrics += "db_operation_duration_ms_bucket{operation=\"SELECT\",le=\"100\"} 245\n"
	prometheusMetrics += "db_operation_duration_ms_bucket{operation=\"SELECT\",le=\"+Inf\"} 245\n"
	prometheusMetrics += "db_operation_duration_ms_sum 4890\n"
	prometheusMetrics += "db_operation_duration_ms_count 245\n"

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, prometheusMetrics)
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
// @Success 200 {object} PaginatedResponse{data=[]models.TraditionalLog}
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

// GetPrometheusMetrics handles Prometheus metrics endpoint
// @Summary Get Prometheus metrics
// @Description Get metrics in Prometheus format for scraping
// @Tags observability
// @Produce plain
// @Success 200 {string} string "Prometheus metrics"
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/observability/prometheus [get]
func (h *ObservabilityHandler) GetPrometheusMetrics(c *gin.Context) {
	// Get system metrics from repository
	metrics, err := h.obsRepo.ListSystemMetrics(c.Request.Context(), "", 100, 0)
	if err != nil {
		c.Header("Content-Type", "text/plain; charset=utf-8")
		c.String(http.StatusInternalServerError, "# Error getting metrics\n")
		return
	}

	// Get traditional metrics as well
	traditionalMetrics, err := h.traditionalRepo.ListTraditionalMetrics(c.Request.Context(), "", "", 100, 0)
	if err != nil {
		// Log warning but continue without traditional metrics
	}

	// Convert to Prometheus format
	var prometheusMetrics string
	
	// Add help and type information for system metrics
	prometheusMetrics += "# HELP actor_system_cpu_usage CPU usage percentage\n"
	prometheusMetrics += "# TYPE actor_system_cpu_usage gauge\n"
	prometheusMetrics += "# HELP actor_system_memory_usage Memory usage in bytes\n"
	prometheusMetrics += "# TYPE actor_system_memory_usage gauge\n"
	prometheusMetrics += "# HELP actor_system_goroutines Number of goroutines\n"
	prometheusMetrics += "# TYPE actor_system_goroutines gauge\n"
	prometheusMetrics += "# HELP actor_system_heap_alloc Heap allocation in bytes\n"
	prometheusMetrics += "# TYPE actor_system_heap_alloc gauge\n"

	// Add help and type information for business metrics
	prometheusMetrics += "# HELP ride_requests_total Total number of ride requests\n"
	prometheusMetrics += "# TYPE ride_requests_total counter\n"
	prometheusMetrics += "# HELP ride_matches_total Total number of successful ride matches\n"
	prometheusMetrics += "# TYPE ride_matches_total counter\n"
	prometheusMetrics += "# HELP trip_completions_total Total number of completed trips\n"
	prometheusMetrics += "# TYPE trip_completions_total counter\n"
	prometheusMetrics += "# HELP trip_starts_total Total number of started trips\n"
	prometheusMetrics += "# TYPE trip_starts_total counter\n"
	prometheusMetrics += "# HELP ride_cancellations_total Total number of ride cancellations\n"
	prometheusMetrics += "# TYPE ride_cancellations_total counter\n"
	prometheusMetrics += "# HELP ride_timeouts_total Total number of ride timeouts\n"
	prometheusMetrics += "# TYPE ride_timeouts_total counter\n"
	prometheusMetrics += "# HELP matching_failures_total Total number of matching failures\n"
	prometheusMetrics += "# TYPE matching_failures_total counter\n"
	prometheusMetrics += "# HELP drivers_online_total Number of online drivers\n"
	prometheusMetrics += "# TYPE drivers_online_total gauge\n"
	prometheusMetrics += "# HELP drivers_busy_total Number of busy drivers\n"
	prometheusMetrics += "# TYPE drivers_busy_total gauge\n"
	prometheusMetrics += "# HELP passengers_active_total Number of active passengers\n"
	prometheusMetrics += "# TYPE passengers_active_total gauge\n"
	prometheusMetrics += "# HELP trips_active_total Number of active trips\n"
	prometheusMetrics += "# TYPE trips_active_total gauge\n"
	prometheusMetrics += "# HELP actor_instances_total Number of active actor instances by type\n"
	prometheusMetrics += "# TYPE actor_instances_total gauge\n"
	prometheusMetrics += "# HELP actor_messages_total Total number of actor messages\n"
	prometheusMetrics += "# TYPE actor_messages_total counter\n"
	prometheusMetrics += "# HELP actor_messages_processed_total Total number of processed actor messages\n"
	prometheusMetrics += "# TYPE actor_messages_processed_total counter\n"
	prometheusMetrics += "# HELP actor_messages_failed_total Total number of failed actor messages\n"
	prometheusMetrics += "# TYPE actor_messages_failed_total counter\n"
	prometheusMetrics += "# HELP ride_matching_duration_ms Ride matching duration in milliseconds\n"
	prometheusMetrics += "# TYPE ride_matching_duration_ms histogram\n"
	prometheusMetrics += "# HELP trip_duration_ms Trip duration in milliseconds\n"
	prometheusMetrics += "# TYPE trip_duration_ms histogram\n"
	prometheusMetrics += "# HELP actor_message_processing_duration_ms Actor message processing duration\n"
	prometheusMetrics += "# TYPE actor_message_processing_duration_ms histogram\n"

	// Convert system metrics to Prometheus format
	for _, metric := range metrics {
		switch metric.MetricType {
		case "cpu_usage":
			actorID := "unknown"
			if metric.ActorID != nil {
				actorID = *metric.ActorID
			}
			prometheusMetrics += fmt.Sprintf("actor_system_cpu_usage{instance_id=\"%s\"} %f\n", 
				actorID, metric.MetricValue)
		case "memory_usage":
			actorID := "unknown"
			if metric.ActorID != nil {
				actorID = *metric.ActorID
			}
			prometheusMetrics += fmt.Sprintf("actor_system_memory_usage{instance_id=\"%s\"} %f\n", 
				actorID, metric.MetricValue)
		case "goroutines":
			actorID := "unknown"
			if metric.ActorID != nil {
				actorID = *metric.ActorID
			}
			prometheusMetrics += fmt.Sprintf("actor_system_goroutines{instance_id=\"%s\"} %f\n", 
				actorID, metric.MetricValue)
		case "heap_alloc":
			actorID := "unknown"
			if metric.ActorID != nil {
				actorID = *metric.ActorID
			}
			prometheusMetrics += fmt.Sprintf("actor_system_heap_alloc{instance_id=\"%s\"} %f\n", 
				actorID, metric.MetricValue)
		}
	}

	// Convert traditional metrics to Prometheus format
	if traditionalMetrics != nil {
		for _, metric := range traditionalMetrics {
			// Convert metric name to Prometheus format
			metricName := metric.MetricName
			labels := ""
			if metric.Labels != nil && len(metric.Labels) > 0 {
				// Parse labels if they exist
				labels = fmt.Sprintf("{service=\"%s\"}", metric.ServiceName)
			}
			
			prometheusMetrics += fmt.Sprintf("%s%s %f\n", 
				metricName, labels, metric.MetricValue)
		}
	}

	// Add some default business metrics with sample values only if no real metrics exist
	// These would normally come from your actual business logic
	if len(metrics) == 0 {
		prometheusMetrics += "ride_requests_total 150\n"
		prometheusMetrics += "ride_matches_total 135\n"
		prometheusMetrics += "trip_completions_total 128\n"
		prometheusMetrics += "trip_starts_total 135\n"
		prometheusMetrics += "ride_cancellations_total 7\n"
		prometheusMetrics += "ride_timeouts_total 3\n"
		prometheusMetrics += "matching_failures_total 5\n"
		prometheusMetrics += "drivers_online_total 45\n"
		prometheusMetrics += "drivers_busy_total 23\n"
		prometheusMetrics += "passengers_active_total 67\n"
		prometheusMetrics += "trips_active_total 23\n"
		prometheusMetrics += "actor_instances_total{type=\"passenger\"} 67\n"
		prometheusMetrics += "actor_instances_total{type=\"driver\"} 45\n"
		prometheusMetrics += "actor_instances_total{type=\"trip\"} 23\n"
		prometheusMetrics += "actor_instances_total{type=\"matching\"} 5\n"
		prometheusMetrics += "actor_messages_total 1250\n"
		prometheusMetrics += "actor_messages_processed_total 1235\n"
		prometheusMetrics += "actor_messages_failed_total 15\n"
	}

	// Add histogram buckets for latency metrics only if no real metrics exist
	if len(metrics) == 0 {
		prometheusMetrics += "ride_matching_duration_ms_bucket{le=\"100\"} 45\n"
		prometheusMetrics += "ride_matching_duration_ms_bucket{le=\"250\"} 89\n"
		prometheusMetrics += "ride_matching_duration_ms_bucket{le=\"500\"} 120\n"
		prometheusMetrics += "ride_matching_duration_ms_bucket{le=\"1000\"} 135\n"
		prometheusMetrics += "ride_matching_duration_ms_bucket{le=\"+Inf\"} 135\n"
		prometheusMetrics += "ride_matching_duration_ms_sum 28750\n"
		prometheusMetrics += "ride_matching_duration_ms_count 135\n"
	}

	if len(metrics) == 0 {
		prometheusMetrics += "trip_duration_ms_bucket{le=\"300000\"} 25\n"
		prometheusMetrics += "trip_duration_ms_bucket{le=\"600000\"} 67\n"
		prometheusMetrics += "trip_duration_ms_bucket{le=\"1200000\"} 105\n"
		prometheusMetrics += "trip_duration_ms_bucket{le=\"1800000\"} 125\n"
		prometheusMetrics += "trip_duration_ms_bucket{le=\"+Inf\"} 128\n"
		prometheusMetrics += "trip_duration_ms_sum 76800000\n"
		prometheusMetrics += "trip_duration_ms_count 128\n"

		prometheusMetrics += "actor_message_processing_duration_ms_bucket{le=\"1\"} 890\n"
		prometheusMetrics += "actor_message_processing_duration_ms_bucket{le=\"5\"} 1150\n"
		prometheusMetrics += "actor_message_processing_duration_ms_bucket{le=\"10\"} 1200\n"
		prometheusMetrics += "actor_message_processing_duration_ms_bucket{le=\"25\"} 1230\n"
		prometheusMetrics += "actor_message_processing_duration_ms_bucket{le=\"+Inf\"} 1235\n"
		prometheusMetrics += "actor_message_processing_duration_ms_sum 4250\n"
		prometheusMetrics += "actor_message_processing_duration_ms_count 1235\n"
	}

	// Add some basic application metrics
	prometheusMetrics += "# HELP actor_system_up Application up status\n"
	prometheusMetrics += "# TYPE actor_system_up gauge\n"
	prometheusMetrics += "actor_system_up 1\n"

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, prometheusMetrics)
}