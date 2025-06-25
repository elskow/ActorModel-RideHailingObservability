package traditional

import (
	"context"
	"net/http"
	"time"

	"actor-model-observability/internal/logging"
	"actor-model-observability/internal/observability"
)

// TraditionalMonitor implements OpenTelemetry-based monitoring
// This replaces the previous custom monitoring implementation
type TraditionalMonitor struct {
	otelMonitor *observability.OTelMonitor
	logger      *logging.Logger
	ctx         context.Context
	cancel      context.CancelFunc
}

// ServiceHealth represents the health status of a service
type ServiceHealth struct {
	ServiceName   string            `json:"service_name"`
	Status        string            `json:"status"`
	Uptime        time.Duration     `json:"uptime" swaggertype:"integer"`
	LastCheck     time.Time         `json:"last_check"`
	HealthChecks  map[string]string `json:"health_checks"`
	Dependencies  []string          `json:"dependencies"`
	Metrics       map[string]float64 `json:"metrics"`
}

// NewTraditionalMonitor creates a new OpenTelemetry-based monitoring system
func NewTraditionalMonitor(logger *logging.Logger, otelMonitor *observability.OTelMonitor) *TraditionalMonitor {
	return &TraditionalMonitor{
		otelMonitor: otelMonitor,
		logger:      logger.WithComponent("traditional_monitor"),
	}
}

// Start begins the OpenTelemetry monitoring process
func (tm *TraditionalMonitor) Start(ctx context.Context) error {
	tm.ctx, tm.cancel = context.WithCancel(ctx)
	tm.logger.Info("OpenTelemetry monitor started")
	return nil
}

// Stop gracefully shuts down the OpenTelemetry monitor
func (tm *TraditionalMonitor) Stop() error {
	tm.logger.Info("Stopping OpenTelemetry monitor")
	if tm.cancel != nil {
		tm.cancel()
	}

	if tm.otelMonitor != nil {
		// Create a fresh context with timeout for shutdown
		// Don't use tm.ctx as it might already be canceled
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		
		if err := tm.otelMonitor.Shutdown(shutdownCtx); err != nil {
			tm.logger.WithError(err).Error("Failed to shutdown OpenTelemetry monitor")
			return err
		}
	}

	tm.logger.Info("OpenTelemetry monitor stopped")
	return nil
}

// RecordRequest records a request for monitoring using OpenTelemetry
func (tm *TraditionalMonitor) RecordRequest(endpoint, method string, duration time.Duration, statusCode int) {
	if tm.otelMonitor != nil {
		tm.otelMonitor.RecordRequest(tm.ctx, endpoint, method, duration, statusCode)
	}
}

// RecordDatabaseOperation records database operation metrics using OpenTelemetry
func (tm *TraditionalMonitor) RecordDatabaseOperation(operation, table string, duration time.Duration, success bool) {
	if tm.otelMonitor != nil {
		tm.otelMonitor.RecordDatabaseOperation(tm.ctx, operation, table, duration, success)
	}
}

// RecordCacheOperation records cache operation metrics using OpenTelemetry
func (tm *TraditionalMonitor) RecordCacheOperation(operation string, hit bool, duration time.Duration) {
	if tm.otelMonitor != nil {
		tm.otelMonitor.RecordCacheOperation(tm.ctx, operation, hit, duration)
	}
}

// RecordSystemMetrics records system-level metrics using OpenTelemetry
func (tm *TraditionalMonitor) RecordSystemMetrics(cpuUsage, memoryUsage, diskUsage float64) {
	if tm.otelMonitor != nil {
		tm.otelMonitor.RecordSystemMetrics(tm.ctx, cpuUsage, memoryUsage, diskUsage)
	}
}

// RecordBusinessMetrics records business-specific metrics using OpenTelemetry
func (tm *TraditionalMonitor) RecordBusinessMetrics(metricName string, value float64, tags map[string]string) {
	if tm.otelMonitor != nil {
		tm.otelMonitor.RecordBusinessMetrics(tm.ctx, metricName, value, tags)
	}
}

// GetPrometheusHandler returns the Prometheus metrics handler
func (tm *TraditionalMonitor) GetPrometheusHandler() http.Handler {
	if tm.otelMonitor != nil {
		return tm.otelMonitor.GetPrometheusHandler()
	}
	return nil
}