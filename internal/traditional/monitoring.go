package traditional

import (
	"context"
	"net/http"
	"time"

	"actor-model-observability/internal/config"
	"actor-model-observability/internal/observability"

	"github.com/sirupsen/logrus"
)

// TraditionalMonitor implements OpenTelemetry-based monitoring
// This replaces the previous custom monitoring implementation
type TraditionalMonitor struct {
	otelMonitor *observability.OTelMonitor
	logger      *logrus.Entry
	ctx         context.Context
	cancel      context.CancelFunc
}

// ServiceHealth represents the health status of a service
type ServiceHealth struct {
	ServiceName   string            `json:"service_name"`
	Status        string            `json:"status"`
	Uptime        time.Duration     `json:"uptime"`
	LastCheck     time.Time         `json:"last_check"`
	HealthChecks  map[string]string `json:"health_checks"`
	Dependencies  []string          `json:"dependencies"`
	Metrics       map[string]float64 `json:"metrics"`
}

// NewTraditionalMonitor creates a new OpenTelemetry-based monitoring system
func NewTraditionalMonitor(cfg *config.Config) (*TraditionalMonitor, error) {
	otelMonitor, err := observability.NewOTelMonitor(&cfg.OpenTelemetry)
	if err != nil {
		return nil, err
	}

	return &TraditionalMonitor{
		otelMonitor: otelMonitor,
		logger:      logrus.WithField("component", "traditional_monitor"),
	}, nil
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
		if err := tm.otelMonitor.Shutdown(tm.ctx); err != nil {
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