package observability

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"actor-model-observability/internal/config"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// OTelMonitor implements OpenTelemetry-based monitoring
type OTelMonitor struct {
	config         *config.OpenTelemetryConfig
	logger         *logrus.Entry
	meterProvider  *sdkmetric.MeterProvider
	tracerProvider *sdktrace.TracerProvider
	meter          metric.Meter
	tracer         trace.Tracer

	// Metrics instruments
	httpRequestsTotal     metric.Int64Counter
	httpRequestDuration   metric.Float64Histogram
	databaseQueriesTotal  metric.Int64Counter
	databaseQueryDuration metric.Float64Histogram
	cacheOperationsTotal  metric.Int64Counter
	cacheOperationDuration metric.Float64Histogram
	systemCPUUsage        metric.Float64Histogram
	systemMemoryUsage     metric.Float64Histogram
	systemDiskUsage       metric.Float64Histogram
	businessMetrics       map[string]metric.Float64Histogram
}

// NewOTelMonitor creates a new OpenTelemetry monitor
func NewOTelMonitor(cfg *config.OpenTelemetryConfig) (*OTelMonitor, error) {
	monitor := &OTelMonitor{
		config:          cfg,
		logger:          logrus.WithField("component", "otel_monitor"),
		businessMetrics: make(map[string]metric.Float64Histogram),
	}

	if err := monitor.initializeResource(); err != nil {
		return nil, fmt.Errorf("failed to initialize resource: %w", err)
	}

	if cfg.MetricsEnabled {
		if err := monitor.initializeMetrics(); err != nil {
			return nil, fmt.Errorf("failed to initialize metrics: %w", err)
		}
	}

	if cfg.TracingEnabled {
		if err := monitor.initializeTracing(); err != nil {
			return nil, fmt.Errorf("failed to initialize tracing: %w", err)
		}
	}

	return monitor, nil
}

// initializeResource creates the OpenTelemetry resource
func (om *OTelMonitor) initializeResource() error {
	attributes := []attribute.KeyValue{
		semconv.ServiceName(om.config.ServiceName),
		semconv.ServiceVersion(om.config.ServiceVersion),
		semconv.DeploymentEnvironment(om.config.Environment),
	}

	// Add custom resource attributes
	for key, value := range om.config.ResourceAttributes {
		attributes = append(attributes, attribute.String(key, value))
	}

	_, err := resource.New(context.Background(),
		resource.WithAttributes(attributes...),
	)
	return err
}

// initializeMetrics sets up OpenTelemetry metrics
func (om *OTelMonitor) initializeMetrics() error {
	var reader sdkmetric.Reader
	var err error

	switch om.config.MetricsExporter {
	case "prometheus":
		reader, err = prometheus.New()
		if err != nil {
			return fmt.Errorf("failed to create Prometheus reader: %w", err)
		}
	default:
		return fmt.Errorf("unsupported metrics exporter: %s", om.config.MetricsExporter)
	}

	om.meterProvider = sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
	)

	otel.SetMeterProvider(om.meterProvider)
	om.meter = om.meterProvider.Meter(om.config.ServiceName)

	// Initialize metric instruments
	if err := om.createMetricInstruments(); err != nil {
		return fmt.Errorf("failed to create metric instruments: %w", err)
	}

	om.logger.Info("OpenTelemetry metrics initialized")
	return nil
}

// initializeTracing sets up OpenTelemetry tracing
func (om *OTelMonitor) initializeTracing() error {
	var exporter sdktrace.SpanExporter
	var err error

	switch om.config.TracingExporter {
	case "jaeger":
		exporter, err = jaeger.New(jaeger.WithCollectorEndpoint(
			jaeger.WithEndpoint(om.config.JaegerEndpoint),
		))
		if err != nil {
			return fmt.Errorf("failed to create Jaeger exporter: %w", err)
		}
	default:
		return fmt.Errorf("unsupported tracing exporter: %s", om.config.TracingExporter)
	}

	om.tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(om.config.SampleRate)),
	)

	otel.SetTracerProvider(om.tracerProvider)
	om.tracer = om.tracerProvider.Tracer(om.config.ServiceName)

	om.logger.Info("OpenTelemetry tracing initialized")
	return nil
}

// createMetricInstruments creates all the metric instruments
func (om *OTelMonitor) createMetricInstruments() error {
	var err error

	// HTTP metrics
	om.httpRequestsTotal, err = om.meter.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
	)
	if err != nil {
		return err
	}

	om.httpRequestDuration, err = om.meter.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}

	// Database metrics
	om.databaseQueriesTotal, err = om.meter.Int64Counter(
		"database_queries_total",
		metric.WithDescription("Total number of database queries"),
	)
	if err != nil {
		return err
	}

	om.databaseQueryDuration, err = om.meter.Float64Histogram(
		"database_query_duration_seconds",
		metric.WithDescription("Database query duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}

	// Cache metrics
	om.cacheOperationsTotal, err = om.meter.Int64Counter(
		"cache_operations_total",
		metric.WithDescription("Total number of cache operations"),
	)
	if err != nil {
		return err
	}

	om.cacheOperationDuration, err = om.meter.Float64Histogram(
		"cache_operation_duration_seconds",
		metric.WithDescription("Cache operation duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}

	// System metrics
	om.systemCPUUsage, err = om.meter.Float64Histogram(
		"system_cpu_usage_percent",
		metric.WithDescription("System CPU usage percentage"),
		metric.WithUnit("%"),
	)
	if err != nil {
		return err
	}

	om.systemMemoryUsage, err = om.meter.Float64Histogram(
		"system_memory_usage_percent",
		metric.WithDescription("System memory usage percentage"),
		metric.WithUnit("%"),
	)
	if err != nil {
		return err
	}

	om.systemDiskUsage, err = om.meter.Float64Histogram(
		"system_disk_usage_percent",
		metric.WithDescription("System disk usage percentage"),
		metric.WithUnit("%"),
	)
	if err != nil {
		return err
	}

	return nil
}

// RecordRequest records HTTP request metrics and creates spans
func (om *OTelMonitor) RecordRequest(ctx context.Context, endpoint, method string, duration time.Duration, statusCode int) {
	// Record metrics
	if om.config.MetricsEnabled {
		attrs := []attribute.KeyValue{
			attribute.String("endpoint", endpoint),
			attribute.String("method", method),
			attribute.Int("status_code", statusCode),
		}

		om.httpRequestsTotal.Add(ctx, 1, metric.WithAttributes(attrs...))
		om.httpRequestDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
	}

	// Create span if tracing is enabled
	if om.config.TracingEnabled {
		_, span := om.tracer.Start(ctx, fmt.Sprintf("%s %s", method, endpoint))
		span.SetAttributes(
			attribute.String("http.method", method),
			attribute.String("http.route", endpoint),
			attribute.Int("http.status_code", statusCode),
			attribute.Float64("http.duration", duration.Seconds()),
		)
		span.End()
	}
}

// RecordDatabaseOperation records database operation metrics and spans
func (om *OTelMonitor) RecordDatabaseOperation(ctx context.Context, operation, table string, duration time.Duration, success bool) {
	status := "success"
	if !success {
		status = "error"
	}

	// Record metrics
	if om.config.MetricsEnabled {
		attrs := []attribute.KeyValue{
			attribute.String("operation", operation),
			attribute.String("table", table),
			attribute.String("status", status),
		}

		om.databaseQueriesTotal.Add(ctx, 1, metric.WithAttributes(attrs...))
		om.databaseQueryDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
	}

	// Create span if tracing is enabled
	if om.config.TracingEnabled {
		_, span := om.tracer.Start(ctx, fmt.Sprintf("db.%s", operation))
		span.SetAttributes(
			attribute.String("db.operation", operation),
			attribute.String("db.table", table),
			attribute.Bool("db.success", success),
			attribute.Float64("db.duration", duration.Seconds()),
		)
		span.End()
	}
}

// RecordCacheOperation records cache operation metrics and spans
func (om *OTelMonitor) RecordCacheOperation(ctx context.Context, operation string, hit bool, duration time.Duration) {
	hitStatus := "miss"
	if hit {
		hitStatus = "hit"
	}

	// Record metrics
	if om.config.MetricsEnabled {
		attrs := []attribute.KeyValue{
			attribute.String("operation", operation),
			attribute.String("result", hitStatus),
		}

		om.cacheOperationsTotal.Add(ctx, 1, metric.WithAttributes(attrs...))
		om.cacheOperationDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
	}

	// Create span if tracing is enabled
	if om.config.TracingEnabled {
		_, span := om.tracer.Start(ctx, fmt.Sprintf("cache.%s", operation))
		span.SetAttributes(
			attribute.String("cache.operation", operation),
			attribute.Bool("cache.hit", hit),
			attribute.Float64("cache.duration", duration.Seconds()),
		)
		span.End()
	}
}

// RecordSystemMetrics records system-level metrics
func (om *OTelMonitor) RecordSystemMetrics(ctx context.Context, cpuUsage, memoryUsage, diskUsage float64) {
	if !om.config.MetricsEnabled {
		return
	}

	om.systemCPUUsage.Record(ctx, cpuUsage)
	om.systemMemoryUsage.Record(ctx, memoryUsage)
	om.systemDiskUsage.Record(ctx, diskUsage)
}

// RecordBusinessMetrics records business-specific metrics
func (om *OTelMonitor) RecordBusinessMetrics(ctx context.Context, metricName string, value float64, tags map[string]string) {
	if !om.config.MetricsEnabled {
		return
	}

	// Create histogram if it doesn't exist
	if _, exists := om.businessMetrics[metricName]; !exists {
		histogram, err := om.meter.Float64Histogram(
			metricName,
			metric.WithDescription(fmt.Sprintf("Business metric: %s", metricName)),
		)
		if err != nil {
			om.logger.WithError(err).Errorf("Failed to create business metric: %s", metricName)
			return
		}
		om.businessMetrics[metricName] = histogram
	}

	// Convert tags to attributes
	var attrs []attribute.KeyValue
	for key, val := range tags {
		attrs = append(attrs, attribute.String(key, val))
	}

	om.businessMetrics[metricName].Record(ctx, value, metric.WithAttributes(attrs...))
}

// GetPrometheusHandler returns the Prometheus metrics handler
func (om *OTelMonitor) GetPrometheusHandler() http.Handler {
	return promhttp.Handler()
}

// StartSpan starts a new tracing span
func (om *OTelMonitor) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if !om.config.TracingEnabled {
		return ctx, trace.SpanFromContext(ctx)
	}
	return om.tracer.Start(ctx, name, opts...)
}

// Shutdown gracefully shuts down the monitor
func (om *OTelMonitor) Shutdown(ctx context.Context) error {
	var errs []error

	if om.meterProvider != nil {
		if err := om.meterProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to shutdown meter provider: %w", err))
		}
	}

	if om.tracerProvider != nil {
		if err := om.tracerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to shutdown tracer provider: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}

	om.logger.Info("OpenTelemetry monitor shutdown complete")
	return nil
}