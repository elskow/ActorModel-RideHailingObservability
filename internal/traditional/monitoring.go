package traditional

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"actor-model-observability/internal/config"
	"actor-model-observability/internal/models"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	gorm "gorm.io/gorm"
)

// TraditionalMonitor implements traditional centralized monitoring
type TraditionalMonitor struct {
	db          *gorm.DB
	redis       *redis.Client
	logger      *logrus.Entry
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	config      *config.Config

	// Centralized metrics storage
	metrics     map[string]*TraditionalMetric
	logs        []*models.TraditionalLog
	metricsLock sync.RWMutex

	// Collection intervals
	collectionInterval time.Duration
	flushInterval      time.Duration

	// Performance counters
	requestCount    int64
	errorCount      int64
	responseTime    time.Duration
	throughput      float64
	lastCollection  time.Time
	countersLock    sync.RWMutex
}

// TraditionalMetric represents a traditional monitoring metric
type TraditionalMetric struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Value       float64                `json:"value"`
	Unit        string                 `json:"unit"`
	Tags        map[string]string      `json:"tags"`
	Timestamp   time.Time              `json:"timestamp"`
	Source      string                 `json:"source"`
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

// NewTraditionalMonitor creates a new traditional monitoring system
func NewTraditionalMonitor(db *gorm.DB, redis *redis.Client, cfg *config.Config) *TraditionalMonitor {
	return &TraditionalMonitor{
		db:                 db,
		redis:              redis,
		logger:             logrus.WithField("component", "traditional_monitor"),
		config:             cfg,
		metrics:            make(map[string]*TraditionalMetric),
		logs:               make([]*models.TraditionalLog, 0),
		collectionInterval: cfg.Metrics.CollectInterval,
		flushInterval:      cfg.Metrics.FlushInterval,
		lastCollection:     time.Now(),
	}
}

// Start begins the traditional monitoring process
func (tm *TraditionalMonitor) Start(ctx context.Context) error {
	tm.ctx, tm.cancel = context.WithCancel(ctx)

	// Start monitoring goroutines
	tm.wg.Add(3)
	go tm.metricsCollectionLoop()
	go tm.flushLoop()
	go tm.healthCheckLoop()

	tm.logger.Info("Traditional monitor started")
	return nil
}

// Stop gracefully shuts down the traditional monitor
func (tm *TraditionalMonitor) Stop() error {
	tm.logger.Info("Stopping traditional monitor")

	if tm.cancel != nil {
		tm.cancel()
	}

	// Flush remaining data
	tm.flushMetrics()

	// Wait for goroutines to finish
	tm.wg.Wait()

	tm.logger.Info("Traditional monitor stopped")
	return nil
}

// RecordRequest records a request for monitoring
func (tm *TraditionalMonitor) RecordRequest(endpoint, method string, duration time.Duration, statusCode int) {
	tm.countersLock.Lock()
	defer tm.countersLock.Unlock()

	tm.requestCount++
	if statusCode >= 400 {
		tm.errorCount++
	}

	// Update average response time
	tm.responseTime = (tm.responseTime + duration) / 2

	// Record detailed metrics
	tm.recordMetric("http_requests_total", "counter", 1, "requests", map[string]string{
		"endpoint": endpoint,
		"method":   method,
		"status":   fmt.Sprintf("%d", statusCode),
	}, "http_server")

	tm.recordMetric("http_request_duration", "histogram", float64(duration.Milliseconds()), "ms", map[string]string{
		"endpoint": endpoint,
		"method":   method,
	}, "http_server")

	// Log the request
	tm.logRequest(endpoint, method, duration, statusCode)
}

// RecordDatabaseOperation records database operation metrics
func (tm *TraditionalMonitor) RecordDatabaseOperation(operation, table string, duration time.Duration, success bool) {
	status := "success"
	if !success {
		status = "error"
	}

	tm.recordMetric("db_operations_total", "counter", 1, "operations", map[string]string{
		"operation": operation,
		"table":     table,
		"status":    status,
	}, "database")

	tm.recordMetric("db_operation_duration", "histogram", float64(duration.Milliseconds()), "ms", map[string]string{
		"operation": operation,
		"table":     table,
	}, "database")

	tm.logDatabaseOperation(operation, table, duration, success)
}

// RecordCacheOperation records cache operation metrics
func (tm *TraditionalMonitor) RecordCacheOperation(operation string, hit bool, duration time.Duration) {
	hitStatus := "miss"
	if hit {
		hitStatus = "hit"
	}

	tm.recordMetric("cache_operations_total", "counter", 1, "operations", map[string]string{
		"operation": operation,
		"result":    hitStatus,
	}, "cache")

	tm.recordMetric("cache_operation_duration", "histogram", float64(duration.Microseconds()), "μs", map[string]string{
		"operation": operation,
	}, "cache")

	tm.logCacheOperation(operation, hit, duration)
}

// RecordSystemMetrics records system-level metrics
func (tm *TraditionalMonitor) RecordSystemMetrics(cpuUsage, memoryUsage, diskUsage float64) {
	tm.recordMetric("system_cpu_usage", "gauge", cpuUsage, "percent", nil, "system")
	tm.recordMetric("system_memory_usage", "gauge", memoryUsage, "percent", nil, "system")
	tm.recordMetric("system_disk_usage", "gauge", diskUsage, "percent", nil, "system")
}

// RecordBusinessMetrics records business-specific metrics
func (tm *TraditionalMonitor) RecordBusinessMetrics(metricName string, value float64, tags map[string]string) {
	tm.recordMetric(metricName, "gauge", value, "count", tags, "business")
}

// recordMetric is a helper to record metrics
func (tm *TraditionalMonitor) recordMetric(name, metricType string, value float64, unit string, tags map[string]string, source string) {
	tm.metricsLock.Lock()
	defer tm.metricsLock.Unlock()

	metricID := fmt.Sprintf("%s_%d", name, time.Now().UnixNano())
	metric := &TraditionalMetric{
		ID:        metricID,
		Name:      name,
		Type:      metricType,
		Value:     value,
		Unit:      unit,
		Tags:      tags,
		Timestamp: time.Now(),
		Source:    source,
	}

	tm.metrics[metricID] = metric

	// Store in Redis for real-time access
	tm.storeMetricInRedis(metric)
}

// logRequest logs HTTP request details
func (tm *TraditionalMonitor) logRequest(endpoint, method string, duration time.Duration, statusCode int) {
	fields, _ := json.Marshal(map[string]interface{}{
		"endpoint":     endpoint,
		"method":       method,
		"status_code":  statusCode,
		"duration_ms":  duration.Milliseconds(),
	})

	logEntry := &models.TraditionalLog{
		Level:       models.LogLevelInfo,
		Message:     fmt.Sprintf("%s %s - %d", method, endpoint, statusCode),
		ServiceName: "http_server",
		InstanceID:  "main",
		Fields:      fields,
		Timestamp:   time.Now(),
	}

	tm.metricsLock.Lock()
	tm.logs = append(tm.logs, logEntry)
	tm.metricsLock.Unlock()
}

// logDatabaseOperation logs database operation details
func (tm *TraditionalMonitor) logDatabaseOperation(operation, table string, duration time.Duration, success bool) {
	level := models.LogLevelInfo
	if !success {
		level = models.LogLevelError
	}

	fields, _ := json.Marshal(map[string]interface{}{
		"operation":   operation,
		"table":       table,
		"success":     success,
		"duration_ms": duration.Milliseconds(),
	})

	logEntry := &models.TraditionalLog{
		Level:       level,
		Message:     fmt.Sprintf("DB %s on %s - %v", operation, table, success),
		ServiceName: "database",
		InstanceID:  "main",
		Fields:      fields,
		Timestamp:   time.Now(),
	}

	tm.metricsLock.Lock()
	tm.logs = append(tm.logs, logEntry)
	tm.metricsLock.Unlock()
}

// logCacheOperation logs cache operation details
func (tm *TraditionalMonitor) logCacheOperation(operation string, hit bool, duration time.Duration) {
	fields, _ := json.Marshal(map[string]interface{}{
		"operation":   operation,
		"hit":         hit,
		"duration_us": duration.Microseconds(),
	})

	logEntry := &models.TraditionalLog{
		Level:       models.LogLevelDebug,
		Message:     fmt.Sprintf("Cache %s - %v", operation, hit),
		ServiceName: "cache",
		InstanceID:  "main",
		Fields:      fields,
		Timestamp:   time.Now(),
	}

	tm.metricsLock.Lock()
	tm.logs = append(tm.logs, logEntry)
	tm.metricsLock.Unlock()
}

// metricsCollectionLoop runs the periodic metrics collection
func (tm *TraditionalMonitor) metricsCollectionLoop() {
	defer tm.wg.Done()

	ticker := time.NewTicker(tm.collectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			tm.collectSystemMetrics()
			tm.calculateThroughput()
		case <-tm.ctx.Done():
			return
		}
	}
}

// flushLoop periodically flushes metrics to the database
func (tm *TraditionalMonitor) flushLoop() {
	defer tm.wg.Done()

	ticker := time.NewTicker(tm.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			tm.flushMetrics()
		case <-tm.ctx.Done():
			return
		}
	}
}

// healthCheckLoop performs periodic health checks
func (tm *TraditionalMonitor) healthCheckLoop() {
	defer tm.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			tm.performHealthChecks()
		case <-tm.ctx.Done():
			return
		}
	}
}

// collectSystemMetrics collects system-level metrics
func (tm *TraditionalMonitor) collectSystemMetrics() {
	// Simulate system metrics collection
	// In a real implementation, you would use system APIs or libraries
	cpuUsage := 45.0 + (float64(time.Now().Unix()%10) * 2.0) // Simulated CPU usage
	memoryUsage := 60.0 + (float64(time.Now().Unix()%5) * 3.0) // Simulated memory usage
	diskUsage := 30.0 + (float64(time.Now().Unix()%3) * 1.0)   // Simulated disk usage

	tm.RecordSystemMetrics(cpuUsage, memoryUsage, diskUsage)
}

// calculateThroughput calculates current throughput
func (tm *TraditionalMonitor) calculateThroughput() {
	tm.countersLock.Lock()
	defer tm.countersLock.Unlock()

	now := time.Now()
	duration := now.Sub(tm.lastCollection)
	if duration.Seconds() > 0 {
		tm.throughput = float64(tm.requestCount) / duration.Seconds()
	}
	tm.lastCollection = now

	// Record throughput metric
	tm.recordMetric("system_throughput", "gauge", tm.throughput, "requests/sec", nil, "system")
}

// performHealthChecks performs health checks on system components
func (tm *TraditionalMonitor) performHealthChecks() {
	// Check database health
	dbHealth := tm.checkDatabaseHealth()
	// Check Redis health
	redisHealth := tm.checkRedisHealth()
	// Check overall system health
	systemHealth := tm.checkSystemHealth()

	// Record health metrics
	tm.recordMetric("health_database", "gauge", boolToFloat(dbHealth), "status", nil, "health")
	tm.recordMetric("health_redis", "gauge", boolToFloat(redisHealth), "status", nil, "health")
	tm.recordMetric("health_system", "gauge", boolToFloat(systemHealth), "status", nil, "health")
}

// checkDatabaseHealth checks database connectivity
func (tm *TraditionalMonitor) checkDatabaseHealth() bool {
	ctx, cancel := context.WithTimeout(tm.ctx, 5*time.Second)
	defer cancel()

	var result int
	err := tm.db.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error
	return err == nil
}

// checkRedisHealth checks Redis connectivity
func (tm *TraditionalMonitor) checkRedisHealth() bool {
	ctx, cancel := context.WithTimeout(tm.ctx, 5*time.Second)
	defer cancel()

	err := tm.redis.Ping(ctx).Err()
	return err == nil
}

// checkSystemHealth checks overall system health
func (tm *TraditionalMonitor) checkSystemHealth() bool {
	// Simple health check based on error rate
	tm.countersLock.RLock()
	errorRate := float64(tm.errorCount) / float64(tm.requestCount+1)
	tm.countersLock.RUnlock()

	return errorRate < 0.05 // Less than 5% error rate
}

// flushMetrics flushes collected metrics to the database
func (tm *TraditionalMonitor) flushMetrics() {
	tm.metricsLock.Lock()
	defer tm.metricsLock.Unlock()

	start := time.Now()

	// Convert metrics to database format and flush
	if len(tm.metrics) > 0 {
		tm.flushTraditionalMetrics()
	}

	// Flush logs
	if len(tm.logs) > 0 {
		tm.flushLogs()
	}

	flushDuration := time.Since(start)
	tm.logger.WithField("flush_duration", flushDuration).Debug("Traditional metrics flushed")
}

// flushTraditionalMetrics flushes metrics to database
func (tm *TraditionalMonitor) flushTraditionalMetrics() {
	// Skip database operations if db is nil (e.g., during benchmarks)
	if tm.db == nil {
		// Clear metrics
		tm.metrics = make(map[string]*TraditionalMetric)
		return
	}

	var dbMetrics []*models.TraditionalMetric
	for _, metric := range tm.metrics {
		tagsJSON, _ := json.Marshal(metric.Tags)
		
		// Convert metric type to models.MetricType
		var metricType models.MetricType
		switch metric.Type {
		case "counter":
			metricType = models.MetricTypeCounter
		case "gauge":
			metricType = models.MetricTypeGauge
		case "histogram":
			metricType = models.MetricTypeHistogram
		default:
			metricType = models.MetricTypeGauge
		}

		dbMetric := &models.TraditionalMetric{
			MetricName:  metric.Name,
			MetricType:  metricType,
			MetricValue: metric.Value,
			Labels:      tagsJSON,
			ServiceName: metric.Source,
			InstanceID:  "main",
			Timestamp:   metric.Timestamp,
		}
		dbMetrics = append(dbMetrics, dbMetric)
	}

	if err := tm.db.CreateInBatches(dbMetrics, 100).Error; err != nil {
		tm.logger.WithError(err).Error("Failed to flush traditional metrics")
	} else {
		tm.logger.WithField("count", len(dbMetrics)).Debug("Traditional metrics flushed")
	}

	// Clear metrics
	tm.metrics = make(map[string]*TraditionalMetric)
}

// flushLogs flushes logs to database
func (tm *TraditionalMonitor) flushLogs() {
	// Skip database operations if db is nil (e.g., during benchmarks)
	if tm.db == nil {
		// Clear logs
		tm.logs = make([]*models.TraditionalLog, 0)
		return
	}

	if err := tm.db.CreateInBatches(tm.logs, 100).Error; err != nil {
		tm.logger.WithError(err).Error("Failed to flush traditional logs")
	} else {
		tm.logger.WithField("count", len(tm.logs)).Debug("Traditional logs flushed")
	}

	// Clear logs
	tm.logs = tm.logs[:0]
}

// storeMetricInRedis stores metric in Redis for real-time access
func (tm *TraditionalMonitor) storeMetricInRedis(metric *TraditionalMetric) {
	// Skip Redis operations if client is nil (e.g., during benchmarks)
	if tm.redis == nil {
		return
	}

	key := fmt.Sprintf("traditional:metric:%s", metric.Name)
	data, err := json.Marshal(metric)
	if err != nil {
		tm.logger.WithError(err).Error("Failed to marshal metric for Redis")
		return
	}

	if err := tm.redis.Set(tm.ctx, key, data, time.Hour).Err(); err != nil {
		tm.logger.WithError(err).Error("Failed to store metric in Redis")
	}
}

// GetMetrics returns current metrics
func (tm *TraditionalMonitor) GetMetrics() map[string]interface{} {
	tm.countersLock.RLock()
	tm.metricsLock.RLock()
	defer tm.countersLock.RUnlock()
	defer tm.metricsLock.RUnlock()

	return map[string]interface{}{
		"request_count":   tm.requestCount,
		"error_count":     tm.errorCount,
		"response_time":   tm.responseTime.Milliseconds(),
		"throughput":      tm.throughput,
		"metrics_count":   len(tm.metrics),
		"logs_count":      len(tm.logs),
		"last_collection": tm.lastCollection,
	}
}

// boolToFloat converts boolean to float for metrics
func boolToFloat(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}