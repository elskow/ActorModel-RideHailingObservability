package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"actor-model-observability/internal/actor"
	"actor-model-observability/internal/config"
	"actor-model-observability/internal/database"
	"actor-model-observability/internal/logging"
	"actor-model-observability/internal/models"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// MetricsCollector collects and stores observability data
type MetricsCollector struct {
	db     *database.PostgresDB
	redis  *redis.Client
	logger *logging.Logger
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	config *config.Config

	// Metrics storage
	actorMetrics   map[string]*models.ActorInstance
	messageMetrics []*models.ActorMessage
	systemMetrics  []*models.SystemMetric
	traces         []*models.DistributedTrace
	eventLogs      []*models.EventLog
	metricsLock    sync.RWMutex

	// Collection intervals
	collectionInterval time.Duration
	flushInterval      time.Duration

	// Batch processing
	batchSize int
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(db *database.PostgresDB, redis *redis.Client, cfg *config.Config, logger *logging.Logger) *MetricsCollector {
	return &MetricsCollector{
		db:                 db,
		redis:              redis,
		logger:             logger.WithComponent("metrics_collector"),
		config:             cfg,
		actorMetrics:       make(map[string]*models.ActorInstance),
		messageMetrics:     make([]*models.ActorMessage, 0),
		systemMetrics:      make([]*models.SystemMetric, 0),
		traces:             make([]*models.DistributedTrace, 0),
		eventLogs:          make([]*models.EventLog, 0),
		collectionInterval: cfg.Observability.MetricsInterval,
		flushInterval:      cfg.Observability.MetricsInterval, // use same interval for flushing
		batchSize:          100,                               // default batch size
	}
}

// Start begins the metrics collection process
func (mc *MetricsCollector) Start(ctx context.Context) error {
	mc.ctx, mc.cancel = context.WithCancel(ctx)

	// Start collection goroutines
	mc.wg.Add(2)
	go mc.metricsCollectionLoop()
	go mc.flushLoop()

	mc.logger.Info("Metrics collector started")
	return nil
}

// Stop gracefully shuts down the metrics collector
func (mc *MetricsCollector) Stop() error {
	mc.logger.Info("Stopping metrics collector")

	if mc.cancel != nil {
		mc.cancel()
	}

	// Flush remaining data
	mc.flushMetrics()

	// Wait for goroutines to finish
	mc.wg.Wait()

	mc.logger.Info("Metrics collector stopped")
	return nil
}

// CollectActorMetrics collects metrics from an actor system
func (mc *MetricsCollector) CollectActorMetrics(system *actor.ActorSystem) {
	mc.metricsLock.Lock()
	defer mc.metricsLock.Unlock()

	// Collect system-level metrics
	systemMetrics := system.GetMetrics()
	mc.recordSystemMetrics(systemMetrics)

	// Collect individual actor metrics
	actors := system.ListActors()
	for _, actorRef := range actors {
		mc.recordActorMetrics(actorRef)
	}
}

// RecordMessage records a message exchange between actors
func (mc *MetricsCollector) RecordMessage(from, to, messageType string, payload interface{}, timestamp time.Time) {
	mc.metricsLock.Lock()
	defer mc.metricsLock.Unlock()

	payloadJSON, _ := json.Marshal(payload)

	message := &models.ActorMessage{
		ID:                uuid.New(),
		TraceID:           uuid.New(),
		SpanID:            uuid.New(),
		SenderActorType:   models.ActorTypeObservability,
		SenderActorID:     from,
		ReceiverActorType: models.ActorTypeObservability,
		ReceiverActorID:   to,
		MessageType:       messageType,
		MessagePayload:    payloadJSON,
		Status:            models.MessageStatusSent,
		SentAt:            timestamp,
		CreatedAt:         time.Now(),
	}

	mc.messageMetrics = append(mc.messageMetrics, message)

	// Also store in Redis for real-time access
	mc.storeMessageInRedis(message)
}

// RecordTrace records a distributed trace
func (mc *MetricsCollector) RecordTrace(traceID, spanID, parentSpanID, operation string, startTime, endTime time.Time, tags map[string]string) {
	mc.metricsLock.Lock()
	defer mc.metricsLock.Unlock()

	tagsJSON, _ := json.Marshal(tags)

	// Parse string UUIDs to uuid.UUID
	traceUUID, _ := uuid.Parse(traceID)
	spanUUID, _ := uuid.Parse(spanID)
	var parentSpanUUID *uuid.UUID
	if parentSpanID != "" {
		parsed, _ := uuid.Parse(parentSpanID)
		parentSpanUUID = &parsed
	}

	trace := &models.DistributedTrace{
		ID:            uuid.New(),
		TraceID:       traceUUID,
		SpanID:        spanUUID,
		ParentSpanID:  parentSpanUUID,
		OperationName: operation,
		StartTime:     startTime,
		EndTime:       &endTime,
		DurationMs:    func() *int { d := int(endTime.Sub(startTime).Milliseconds()); return &d }(),
		Tags:          tagsJSON,
		CreatedAt:     time.Now(),
	}

	mc.traces = append(mc.traces, trace)
}

// RecordEvent records a system event
func (mc *MetricsCollector) RecordEvent(eventType, source, description string, metadata map[string]interface{}) {
	mc.metricsLock.Lock()
	defer mc.metricsLock.Unlock()

	metadataJSON, _ := json.Marshal(metadata)

	event := &models.EventLog{
		ID:            uuid.New(),
		EventType:     eventType,
		EventCategory: models.EventCategorySystem,
		Message:       description,
		EventData:     metadataJSON,
		Severity:      models.EventSeverityInfo,
		Timestamp:     time.Now(),
		CreatedAt:     time.Now(),
	}

	mc.eventLogs = append(mc.eventLogs, event)

	// Log critical events
	if eventType == "error" || eventType == "critical" {
		mc.logger.WithFields(logging.Fields{
			"event_type":  eventType,
			"source":      source,
			"description": description,
		}).Error("Critical event recorded")
	}
}

// recordActorMetrics records metrics for a specific actor
func (mc *MetricsCollector) recordActorMetrics(actorRef *actor.ActorRef) {
	// Get actor metrics and state (currently not used but available for future implementation)
	// metrics := actorRef.Actor.GetMetrics()
	// state := actorRef.Actor.GetState()

	// Parse actorRef.ID to UUID
	actorUUID, _ := uuid.Parse(actorRef.ID)

	// Convert actorRef.Type to ActorType
	var actorType models.ActorType
	switch actorRef.Type {
	case "passenger":
		actorType = models.ActorTypePassenger
	case "driver":
		actorType = models.ActorTypeDriver
	case "trip":
		actorType = models.ActorTypeTrip
	case "matching":
		actorType = models.ActorTypeMatching
	case "observability":
		actorType = models.ActorTypeObservability
	default:
		actorType = models.ActorTypePassenger // default fallback
	}

	actorInstance := &models.ActorInstance{
		ID:            actorUUID,
		ActorType:     actorType,
		ActorID:       actorRef.ID,
		Status:        models.ActorStatusActive,
		LastHeartbeat: time.Now(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	mc.actorMetrics[actorRef.ID] = actorInstance

	// Store in Redis for real-time monitoring
	mc.storeActorMetricsInRedis(actorInstance)
}

// recordSystemMetrics records system-level metrics
func (mc *MetricsCollector) recordSystemMetrics(metrics actor.SystemMetrics) {
	systemMetric := &models.SystemMetric{
		ID:          uuid.New(),
		MetricName:  "system_performance",
		MetricType:  models.MetricTypeGauge,
		MetricValue: float64(metrics.TotalActors),
		Timestamp:   time.Now(),
		CreatedAt:   time.Now(),
	}

	mc.systemMetrics = append(mc.systemMetrics, systemMetric)

	// Store in Redis for real-time dashboards
	mc.storeSystemMetricsInRedis(systemMetric)
}

// metricsCollectionLoop runs the periodic metrics collection
func (mc *MetricsCollector) metricsCollectionLoop() {
	defer mc.wg.Done()

	ticker := time.NewTicker(mc.collectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Collection is triggered externally via CollectActorMetrics
			// This loop just maintains the ticker for consistency
		case <-mc.ctx.Done():
			return
		}
	}
}

// flushLoop periodically flushes metrics to the database
func (mc *MetricsCollector) flushLoop() {
	defer mc.wg.Done()

	ticker := time.NewTicker(mc.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mc.flushMetrics()
		case <-mc.ctx.Done():
			return
		}
	}
}

// flushMetrics flushes collected metrics to the database
func (mc *MetricsCollector) flushMetrics() {
	mc.metricsLock.Lock()
	defer mc.metricsLock.Unlock()

	start := time.Now()

	// Flush actor instances
	if len(mc.actorMetrics) > 0 {
		mc.flushActorMetrics()
	}

	// Flush messages in batches
	if len(mc.messageMetrics) > 0 {
		mc.flushMessageMetrics()
	}

	// Flush system metrics
	if len(mc.systemMetrics) > 0 {
		mc.flushSystemMetrics()
	}

	// Flush traces
	if len(mc.traces) > 0 {
		mc.flushTraces()
	}

	// Flush event logs
	if len(mc.eventLogs) > 0 {
		mc.flushEventLogs()
	}

	flushDuration := time.Since(start)
	mc.logger.WithField("flush_duration", flushDuration).Debug("Metrics flushed to database")
}

// flushActorMetrics flushes actor metrics to database
func (mc *MetricsCollector) flushActorMetrics() {
	var instances []*models.ActorInstance
	for _, instance := range mc.actorMetrics {
		instances = append(instances, instance)
	}

	if len(instances) > 0 {
		if err := mc.insertActorInstancesBatch(instances); err != nil {
			mc.logger.WithError(err).Error("Failed to flush actor metrics")
		} else {
			mc.logger.WithField("count", len(instances)).Debug("Actor metrics flushed")
		}
	}

	// Clear the map
	mc.actorMetrics = make(map[string]*models.ActorInstance)
}

// flushMessageMetrics flushes message metrics to database
func (mc *MetricsCollector) flushMessageMetrics() {
	if len(mc.messageMetrics) > 0 {
		if err := mc.insertMessagesBatch(mc.messageMetrics); err != nil {
			mc.logger.WithError(err).Error("Failed to flush message metrics")
		} else {
			mc.logger.WithField("count", len(mc.messageMetrics)).Debug("Message metrics flushed")
		}
	}

	// Clear the slice
	mc.messageMetrics = mc.messageMetrics[:0]
}

// flushSystemMetrics flushes system metrics to database
func (mc *MetricsCollector) flushSystemMetrics() {
	if len(mc.systemMetrics) > 0 {
		if err := mc.insertSystemMetricsBatch(mc.systemMetrics); err != nil {
			mc.logger.WithError(err).Error("Failed to flush system metrics")
		} else {
			mc.logger.WithField("count", len(mc.systemMetrics)).Debug("System metrics flushed")
		}
	}

	// Clear the slice
	mc.systemMetrics = mc.systemMetrics[:0]
}

// flushTraces flushes traces to database
func (mc *MetricsCollector) flushTraces() {
	if len(mc.traces) > 0 {
		if err := mc.insertTracesBatch(mc.traces); err != nil {
			mc.logger.WithError(err).Error("Failed to flush traces")
		} else {
			mc.logger.WithField("count", len(mc.traces)).Debug("Traces flushed")
		}
	}

	// Clear the slice
	mc.traces = mc.traces[:0]
}

// flushEventLogs flushes event logs to database
func (mc *MetricsCollector) flushEventLogs() {
	if len(mc.eventLogs) > 0 {
		if err := mc.insertEventLogsBatch(mc.eventLogs); err != nil {
			mc.logger.WithError(err).Error("Failed to flush event logs")
		} else {
			mc.logger.WithField("count", len(mc.eventLogs)).Debug("Event logs flushed")
		}
	}

	// Clear the slice
	mc.eventLogs = mc.eventLogs[:0]
}

// storeActorMetricsInRedis stores actor metrics in Redis for real-time access
func (mc *MetricsCollector) storeActorMetricsInRedis(instance *models.ActorInstance) {
	// Skip Redis operations if Redis client is not available (e.g., in tests)
	if mc.redis == nil {
		return
	}

	key := fmt.Sprintf("actor:metrics:%s", instance.ID)
	data, err := json.Marshal(instance)
	if err != nil {
		mc.logger.WithError(err).Error("Failed to marshal actor metrics for Redis")
		return
	}

	if err := mc.redis.Set(mc.ctx, key, data, time.Hour).Err(); err != nil {
		mc.logger.WithError(err).Error("Failed to store actor metrics in Redis")
	}
}

// storeMessageInRedis stores message in Redis for real-time access
func (mc *MetricsCollector) storeMessageInRedis(message *models.ActorMessage) {
	// Skip Redis operations if Redis client is not available (e.g., in tests)
	if mc.redis == nil {
		return
	}

	key := fmt.Sprintf("message:%s", message.ID)
	data, err := json.Marshal(message)
	if err != nil {
		mc.logger.WithError(err).Error("Failed to marshal message for Redis")
		return
	}

	if err := mc.redis.Set(mc.ctx, key, data, 30*time.Minute).Err(); err != nil {
		mc.logger.WithError(err).Error("Failed to store message in Redis")
	}

	// Also add to recent messages list
	listKey := "messages:recent"
	mc.redis.LPush(mc.ctx, listKey, message.ID)
	mc.redis.LTrim(mc.ctx, listKey, 0, 1000) // Keep only last 1000 messages
	mc.redis.Expire(mc.ctx, listKey, time.Hour)
}

// storeSystemMetricsInRedis stores system metrics in Redis
func (mc *MetricsCollector) storeSystemMetricsInRedis(metric *models.SystemMetric) {
	// Skip Redis operations if Redis client is not available (e.g., in tests)
	if mc.redis == nil {
		return
	}

	key := "system:metrics:latest"
	data, err := json.Marshal(metric)
	if err != nil {
		mc.logger.WithError(err).Error("Failed to marshal system metrics for Redis")
		return
	}

	if err := mc.redis.Set(mc.ctx, key, data, time.Hour).Err(); err != nil {
		mc.logger.WithError(err).Error("Failed to store system metrics in Redis")
	}
}

// GetRealtimeMetrics returns real-time metrics from Redis
func (mc *MetricsCollector) GetRealtimeMetrics() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Return empty result if Redis client is not available (e.g., in tests)
	if mc.redis == nil {
		return result, nil
	}

	// Get latest system metrics
	systemData, err := mc.redis.Get(mc.ctx, "system:metrics:latest").Result()
	if err == nil {
		var systemMetrics models.SystemMetric
		if err := json.Unmarshal([]byte(systemData), &systemMetrics); err == nil {
			result["system"] = systemMetrics
		}
	}

	// Get recent messages count
	messageCount, err := mc.redis.LLen(mc.ctx, "messages:recent").Result()
	if err == nil {
		result["recent_messages_count"] = messageCount
	}

	// Get active actors count
	actorKeys, err := mc.redis.Keys(mc.ctx, "actor:metrics:*").Result()
	if err == nil {
		result["active_actors_count"] = len(actorKeys)
	}

	return result, nil
}

// insertActorInstancesBatch inserts actor instances in batch using sqlx
func (mc *MetricsCollector) insertActorInstancesBatch(instances []*models.ActorInstance) error {
	if len(instances) == 0 {
		return nil
	}

	query := `INSERT INTO actor_instances (id, actor_type, actor_id, entity_id, status, last_heartbeat, created_at, updated_at) 
			  VALUES (:id, :actor_type, :actor_id, :entity_id, :status, :last_heartbeat, :created_at, :updated_at)`

	_, err := mc.db.NamedExec(query, instances)
	return err
}

// insertEventLogsBatch inserts event logs in batch using sqlx
func (mc *MetricsCollector) insertEventLogsBatch(logs []*models.EventLog) error {
	if len(logs) == 0 {
		return nil
	}

	query := `INSERT INTO event_logs (id, trace_id, event_type, event_category, actor_type, actor_id, entity_type, entity_id, event_data, severity, message, timestamp, created_at) 
			  VALUES (:id, :trace_id, :event_type, :event_category, :actor_type, :actor_id, :entity_type, :entity_id, :event_data, :severity, :message, :timestamp, :created_at)`

	_, err := mc.db.NamedExec(query, logs)
	return err
}

// insertTracesBatch inserts distributed traces in batch using sqlx
func (mc *MetricsCollector) insertTracesBatch(traces []*models.DistributedTrace) error {
	if len(traces) == 0 {
		return nil
	}

	query := `INSERT INTO distributed_traces (id, trace_id, span_id, parent_span_id, operation_name, actor_type, actor_id, start_time, end_time, duration_ms, status, tags, logs, created_at) 
			  VALUES (:id, :trace_id, :span_id, :parent_span_id, :operation_name, :actor_type, :actor_id, :start_time, :end_time, :duration_ms, :status, :tags, :logs, :created_at)`

	_, err := mc.db.NamedExec(query, traces)
	return err
}

// insertSystemMetricsBatch inserts system metrics in batch using sqlx
func (mc *MetricsCollector) insertSystemMetricsBatch(metrics []*models.SystemMetric) error {
	if len(metrics) == 0 {
		return nil
	}

	query := `INSERT INTO system_metrics (id, metric_name, metric_type, metric_value, labels, actor_type, actor_id, timestamp, created_at) 
			  VALUES (:id, :metric_name, :metric_type, :metric_value, :labels, :actor_type, :actor_id, :timestamp, :created_at)`

	_, err := mc.db.NamedExec(query, metrics)
	return err
}

// insertMessagesBatch inserts actor messages in batch using sqlx
func (mc *MetricsCollector) insertMessagesBatch(messages []*models.ActorMessage) error {
	if len(messages) == 0 {
		return nil
	}

	query := `INSERT INTO actor_messages (id, trace_id, span_id, parent_span_id, sender_actor_type, sender_actor_id, receiver_actor_type, receiver_actor_id, message_type, message_payload, status, sent_at, received_at, processed_at, processing_duration_ms, error_message, created_at) 
			  VALUES (:id, :trace_id, :span_id, :parent_span_id, :sender_actor_type, :sender_actor_id, :receiver_actor_type, :receiver_actor_id, :message_type, :message_payload, :status, :sent_at, :received_at, :processed_at, :processing_duration_ms, :error_message, :created_at)`

	_, err := mc.db.NamedExec(query, messages)
	return err
}
