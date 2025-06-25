package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"actor-model-observability/internal/models"
	"actor-model-observability/internal/repository"

	"github.com/jmoiron/sqlx"
)

// ObservabilityRepositoryImpl implements the ObservabilityRepository interface using PostgreSQL
type ObservabilityRepositoryImpl struct {
	db *sqlx.DB
}

// NewObservabilityRepository creates a new instance of ObservabilityRepositoryImpl
func NewObservabilityRepository(db *sqlx.DB) repository.ObservabilityRepository {
	return &ObservabilityRepositoryImpl{db: db}
}

// Actor Instances methods

// CreateActorInstance creates a new actor instance record
func (r *ObservabilityRepositoryImpl) CreateActorInstance(ctx context.Context, instance *models.ActorInstance) error {
	query := `
		INSERT INTO actor_instances (id, actor_type, actor_id, entity_id, status, last_heartbeat, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		instance.ID,
		instance.ActorType,
		instance.ActorID,
		instance.EntityID,
		instance.Status,
		instance.LastHeartbeat,
		instance.CreatedAt,
		instance.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create actor instance: %w", err)
	}

	return nil
}

// GetActorInstance retrieves an actor instance by ID
func (r *ObservabilityRepositoryImpl) GetActorInstance(ctx context.Context, id string) (*models.ActorInstance, error) {
	query := `
		SELECT id, actor_type, actor_id, entity_id, status, last_heartbeat, created_at, updated_at
		FROM actor_instances
		WHERE id = $1
	`

	instance := &models.ActorInstance{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&instance.ID,
		&instance.ActorType,
		&instance.ActorID,
		&instance.EntityID,
		&instance.Status,
		&instance.LastHeartbeat,
		&instance.CreatedAt,
		&instance.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &models.NotFoundError{
				Resource: "actor_instance",
				ID:       id,
			}
		}
		return nil, fmt.Errorf("failed to get actor instance: %w", err)
	}

	return instance, nil
}

// UpdateActorInstance updates an existing actor instance
func (r *ObservabilityRepositoryImpl) UpdateActorInstance(ctx context.Context, instance *models.ActorInstance) error {
	query := `
		UPDATE actor_instances
		SET actor_type = $2, actor_id = $3, entity_id = $4, status = $5, 
			last_heartbeat = $6, updated_at = $7
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		instance.ID,
		instance.ActorType,
		instance.ActorID,
		instance.EntityID,
		instance.Status,
		instance.LastHeartbeat,
		instance.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update actor instance: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &models.NotFoundError{
			Resource: "actor_instance",
			ID:       instance.ID.String(),
		}
	}

	return nil
}

// ListActorInstances retrieves actor instances by type with pagination
func (r *ObservabilityRepositoryImpl) ListActorInstances(ctx context.Context, actorType string, limit, offset int) ([]*models.ActorInstance, error) {
	var query string
	var args []interface{}

	if actorType != "" {
		query = `
			SELECT id, actor_type, actor_id, entity_id, status, last_heartbeat, created_at, updated_at
			FROM actor_instances
			WHERE actor_type = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{actorType, limit, offset}
	} else {
		query = `
			SELECT id, actor_type, actor_id, entity_id, status, last_heartbeat, created_at, updated_at
			FROM actor_instances
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2
		`
		args = []interface{}{limit, offset}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list actor instances: %w", err)
	}
	defer rows.Close()

	var instances []*models.ActorInstance
	for rows.Next() {
		instance := &models.ActorInstance{}
		err := rows.Scan(
			&instance.ID,
			&instance.ActorType,
			&instance.ActorID,
			&instance.EntityID,
			&instance.Status,
			&instance.LastHeartbeat,
			&instance.CreatedAt,
			&instance.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan actor instance: %w", err)
		}
		instances = append(instances, instance)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating actor instances: %w", err)
	}

	return instances, nil
}

// Actor Messages methods

// CreateActorMessage creates a new actor message record
func (r *ObservabilityRepositoryImpl) CreateActorMessage(ctx context.Context, message *models.ActorMessage) error {
	query := `
		INSERT INTO actor_messages (id, trace_id, span_id, parent_span_id, sender_actor_type, 
			sender_actor_id, receiver_actor_type, receiver_actor_id, message_type, message_payload, 
			status, sent_at, received_at, processed_at, processing_duration_ms, error_message, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`

	_, err := r.db.ExecContext(ctx, query,
		message.ID,
		message.TraceID,
		message.SpanID,
		message.ParentSpanID,
		message.SenderActorType,
		message.SenderActorID,
		message.ReceiverActorType,
		message.ReceiverActorID,
		message.MessageType,
		message.MessagePayload,
		message.Status,
		message.SentAt,
		message.ReceivedAt,
		message.ProcessedAt,
		message.ProcessingDurationMs,
		message.ErrorMessage,
		message.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create actor message: %w", err)
	}

	return nil
}

// GetActorMessage retrieves an actor message by ID
func (r *ObservabilityRepositoryImpl) GetActorMessage(ctx context.Context, id string) (*models.ActorMessage, error) {
	query := `
		SELECT id, trace_id, span_id, parent_span_id, sender_actor_type, sender_actor_id, 
			receiver_actor_type, receiver_actor_id, message_type, message_payload, status, 
			sent_at, received_at, processed_at, processing_duration_ms, error_message, created_at
		FROM actor_messages
		WHERE id = $1
	`

	message := &models.ActorMessage{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&message.ID,
		&message.TraceID,
		&message.SpanID,
		&message.ParentSpanID,
		&message.SenderActorType,
		&message.SenderActorID,
		&message.ReceiverActorType,
		&message.ReceiverActorID,
		&message.MessageType,
		&message.MessagePayload,
		&message.Status,
		&message.SentAt,
		&message.ReceivedAt,
		&message.ProcessedAt,
		&message.ProcessingDurationMs,
		&message.ErrorMessage,
		&message.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &models.NotFoundError{
				Resource: "actor_message",
				ID:       id,
			}
		}
		return nil, fmt.Errorf("failed to get actor message: %w", err)
	}

	return message, nil
}

// ListActorMessages retrieves actor messages with optional filtering
func (r *ObservabilityRepositoryImpl) ListActorMessages(ctx context.Context, fromActor, toActor string, limit, offset int) ([]*models.ActorMessage, error) {
	var query string
	var args []interface{}

	if fromActor != "" && toActor != "" {
		query = `
			SELECT id, trace_id, span_id, parent_span_id, sender_actor_type, sender_actor_id, 
				receiver_actor_type, receiver_actor_id, message_type, message_payload, status, 
				sent_at, received_at, processed_at, processing_duration_ms, error_message, created_at
			FROM actor_messages
			WHERE sender_actor_id = $1 AND receiver_actor_id = $2
			ORDER BY created_at DESC
			LIMIT $3 OFFSET $4
		`
		args = []interface{}{fromActor, toActor, limit, offset}
	} else if fromActor != "" {
		query = `
			SELECT id, trace_id, span_id, parent_span_id, sender_actor_type, sender_actor_id, 
				receiver_actor_type, receiver_actor_id, message_type, message_payload, status, 
				sent_at, received_at, processed_at, processing_duration_ms, error_message, created_at
			FROM actor_messages
			WHERE sender_actor_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{fromActor, limit, offset}
	} else if toActor != "" {
		query = `
			SELECT id, trace_id, span_id, parent_span_id, sender_actor_type, sender_actor_id, 
				receiver_actor_type, receiver_actor_id, message_type, message_payload, status, 
				sent_at, received_at, processed_at, processing_duration_ms, error_message, created_at
			FROM actor_messages
			WHERE receiver_actor_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{toActor, limit, offset}
	} else {
		query = `
			SELECT id, trace_id, span_id, parent_span_id, sender_actor_type, sender_actor_id, 
				receiver_actor_type, receiver_actor_id, message_type, message_payload, status, 
				sent_at, received_at, processed_at, processing_duration_ms, error_message, created_at
			FROM actor_messages
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2
		`
		args = []interface{}{limit, offset}
	}

	return r.scanActorMessages(ctx, query, args...)
}

// GetMessagesByTimeRange retrieves messages within a time range
func (r *ObservabilityRepositoryImpl) GetMessagesByTimeRange(ctx context.Context, startTime, endTime string, limit, offset int) ([]*models.ActorMessage, error) {
	startTimeParsed, err := time.Parse(time.RFC3339, startTime)
	if err != nil {
		return nil, fmt.Errorf("invalid start time format: %w", err)
	}

	endTimeParsed, err := time.Parse(time.RFC3339, endTime)
	if err != nil {
		return nil, fmt.Errorf("invalid end time format: %w", err)
	}

	query := `
		SELECT id, trace_id, span_id, parent_span_id, sender_actor_type, sender_actor_id, 
			receiver_actor_type, receiver_actor_id, message_type, message_payload, status, 
			sent_at, received_at, processed_at, processing_duration_ms, error_message, created_at
		FROM actor_messages
		WHERE created_at >= $1 AND created_at <= $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	return r.scanActorMessages(ctx, query, startTimeParsed, endTimeParsed, limit, offset)
}

// System Metrics methods

// CreateSystemMetric creates a new system metric record
func (r *ObservabilityRepositoryImpl) CreateSystemMetric(ctx context.Context, metric *models.SystemMetric) error {
	query := `
		INSERT INTO system_metrics (id, metric_name, metric_type, metric_value, labels, 
			actor_type, actor_id, timestamp, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(ctx, query,
		metric.ID,
		metric.MetricName,
		metric.MetricType,
		metric.MetricValue,
		metric.Labels,
		metric.ActorType,
		metric.ActorID,
		metric.Timestamp,
		metric.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create system metric: %w", err)
	}

	return nil
}

// GetSystemMetric retrieves a system metric by ID
func (r *ObservabilityRepositoryImpl) GetSystemMetric(ctx context.Context, id string) (*models.SystemMetric, error) {
	query := `
		SELECT id, metric_name, metric_type, metric_value, labels, actor_type, actor_id, timestamp, created_at
		FROM system_metrics
		WHERE id = $1
	`

	metric := &models.SystemMetric{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&metric.ID,
		&metric.MetricName,
		&metric.MetricType,
		&metric.MetricValue,
		&metric.Labels,
		&metric.ActorType,
		&metric.ActorID,
		&metric.Timestamp,
		&metric.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &models.NotFoundError{
				Resource: "system_metric",
				ID:       id,
			}
		}
		return nil, fmt.Errorf("failed to get system metric: %w", err)
	}

	return metric, nil
}

// ListSystemMetrics retrieves system metrics with optional filtering
func (r *ObservabilityRepositoryImpl) ListSystemMetrics(ctx context.Context, metricType string, limit, offset int) ([]*models.SystemMetric, error) {
	var query string
	var args []interface{}

	if metricType != "" {
		query = `
			SELECT id, metric_name, metric_type, metric_value, labels, actor_type, actor_id, timestamp, created_at
			FROM system_metrics
			WHERE metric_type = $1
			ORDER BY timestamp DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{metricType, limit, offset}
	} else {
		query = `
			SELECT id, metric_name, metric_type, metric_value, labels, actor_type, actor_id, timestamp, created_at
			FROM system_metrics
			ORDER BY timestamp DESC
			LIMIT $1 OFFSET $2
		`
		args = []interface{}{limit, offset}
	}

	return r.scanSystemMetrics(ctx, query, args...)
}

// GetMetricsByTimeRange retrieves metrics within a time range
func (r *ObservabilityRepositoryImpl) GetMetricsByTimeRange(ctx context.Context, startTime, endTime string, limit, offset int) ([]*models.SystemMetric, error) {
	startTimeParsed, err := time.Parse(time.RFC3339, startTime)
	if err != nil {
		return nil, fmt.Errorf("invalid start time format: %w", err)
	}

	endTimeParsed, err := time.Parse(time.RFC3339, endTime)
	if err != nil {
		return nil, fmt.Errorf("invalid end time format: %w", err)
	}

	query := `
		SELECT id, metric_name, metric_type, metric_value, labels, actor_type, actor_id, timestamp, created_at
		FROM system_metrics
		WHERE timestamp >= $1 AND timestamp <= $2
		ORDER BY timestamp DESC
		LIMIT $3 OFFSET $4
	`

	return r.scanSystemMetrics(ctx, query, startTimeParsed, endTimeParsed, limit, offset)
}

// Distributed Traces methods

// CreateDistributedTrace creates a new distributed trace record
func (r *ObservabilityRepositoryImpl) CreateDistributedTrace(ctx context.Context, trace *models.DistributedTrace) error {
	query := `
		INSERT INTO distributed_traces (id, trace_id, span_id, parent_span_id, operation_name, 
			actor_type, actor_id, start_time, end_time, duration_ms, status, tags, logs, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	_, err := r.db.ExecContext(ctx, query,
		trace.ID,
		trace.TraceID,
		trace.SpanID,
		trace.ParentSpanID,
		trace.OperationName,
		trace.ActorType,
		trace.ActorID,
		trace.StartTime,
		trace.EndTime,
		trace.DurationMs,
		trace.Status,
		trace.Tags,
		trace.Logs,
		trace.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create distributed trace: %w", err)
	}

	return nil
}

// GetDistributedTrace retrieves a distributed trace by ID
func (r *ObservabilityRepositoryImpl) GetDistributedTrace(ctx context.Context, id string) (*models.DistributedTrace, error) {
	query := `
		SELECT id, trace_id, span_id, parent_span_id, operation_name, actor_type, actor_id, 
			start_time, end_time, duration_ms, status, tags, logs, created_at
		FROM distributed_traces
		WHERE id = $1
	`

	trace := &models.DistributedTrace{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&trace.ID,
		&trace.TraceID,
		&trace.SpanID,
		&trace.ParentSpanID,
		&trace.OperationName,
		&trace.ActorType,
		&trace.ActorID,
		&trace.StartTime,
		&trace.EndTime,
		&trace.DurationMs,
		&trace.Status,
		&trace.Tags,
		&trace.Logs,
		&trace.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &models.NotFoundError{
				Resource: "distributed_trace",
				ID:       id,
			}
		}
		return nil, fmt.Errorf("failed to get distributed trace: %w", err)
	}

	return trace, nil
}

// GetTracesByTraceID retrieves all spans for a specific trace ID
func (r *ObservabilityRepositoryImpl) GetTracesByTraceID(ctx context.Context, traceID string) ([]*models.DistributedTrace, error) {
	query := `
		SELECT id, trace_id, span_id, parent_span_id, operation_name, actor_type, actor_id, 
			start_time, end_time, duration_ms, status, tags, logs, created_at
		FROM distributed_traces
		WHERE trace_id = $1
		ORDER BY start_time ASC
	`

	return r.scanDistributedTraces(ctx, query, traceID)
}

// ListDistributedTraces retrieves distributed traces with optional filtering
func (r *ObservabilityRepositoryImpl) ListDistributedTraces(ctx context.Context, operation string, limit, offset int) ([]*models.DistributedTrace, error) {
	var query string
	var args []interface{}

	if operation != "" {
		query = `
			SELECT id, trace_id, span_id, parent_span_id, operation_name, actor_type, actor_id, 
				start_time, end_time, duration_ms, status, tags, logs, created_at
			FROM distributed_traces
			WHERE operation_name = $1
			ORDER BY start_time DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{operation, limit, offset}
	} else {
		query = `
			SELECT id, trace_id, span_id, parent_span_id, operation_name, actor_type, actor_id, 
				start_time, end_time, duration_ms, status, tags, logs, created_at
			FROM distributed_traces
			ORDER BY start_time DESC
			LIMIT $1 OFFSET $2
		`
		args = []interface{}{limit, offset}
	}

	return r.scanDistributedTraces(ctx, query, args...)
}

// Event Logs methods

// CreateEventLog creates a new event log record
func (r *ObservabilityRepositoryImpl) CreateEventLog(ctx context.Context, log *models.EventLog) error {
	query := `
		INSERT INTO event_logs (id, trace_id, event_type, event_category, actor_type, actor_id, 
			entity_type, entity_id, event_data, severity, message, timestamp, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.db.ExecContext(ctx, query,
		log.ID,
		log.TraceID,
		log.EventType,
		log.EventCategory,
		log.ActorType,
		log.ActorID,
		log.EntityType,
		log.EntityID,
		log.EventData,
		log.Severity,
		log.Message,
		log.Timestamp,
		log.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create event log: %w", err)
	}

	return nil
}

// GetEventLog retrieves an event log by ID
func (r *ObservabilityRepositoryImpl) GetEventLog(ctx context.Context, id string) (*models.EventLog, error) {
	query := `
		SELECT id, trace_id, event_type, event_category, actor_type, actor_id, entity_type, 
			entity_id, event_data, severity, message, timestamp, created_at
		FROM event_logs
		WHERE id = $1
	`

	log := &models.EventLog{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&log.ID,
		&log.TraceID,
		&log.EventType,
		&log.EventCategory,
		&log.ActorType,
		&log.ActorID,
		&log.EntityType,
		&log.EntityID,
		&log.EventData,
		&log.Severity,
		&log.Message,
		&log.Timestamp,
		&log.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &models.NotFoundError{
				Resource: "event_log",
				ID:       id,
			}
		}
		return nil, fmt.Errorf("failed to get event log: %w", err)
	}

	return log, nil
}

// ListEventLogs retrieves event logs with optional filtering
func (r *ObservabilityRepositoryImpl) ListEventLogs(ctx context.Context, eventType, source string, limit, offset int) ([]*models.EventLog, error) {
	var query string
	var args []interface{}

	if eventType != "" && source != "" {
		query = `
			SELECT id, trace_id, event_type, event_category, actor_type, actor_id, entity_type, 
				entity_id, event_data, severity, message, timestamp, created_at
			FROM event_logs
			WHERE event_type = $1 AND actor_id = $2
			ORDER BY timestamp DESC
			LIMIT $3 OFFSET $4
		`
		args = []interface{}{eventType, source, limit, offset}
	} else if eventType != "" {
		query = `
			SELECT id, trace_id, event_type, event_category, actor_type, actor_id, entity_type, 
				entity_id, event_data, severity, message, timestamp, created_at
			FROM event_logs
			WHERE event_type = $1
			ORDER BY timestamp DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{eventType, limit, offset}
	} else if source != "" {
		query = `
			SELECT id, trace_id, event_type, event_category, actor_type, actor_id, entity_type, 
				entity_id, event_data, severity, message, timestamp, created_at
			FROM event_logs
			WHERE actor_id = $1
			ORDER BY timestamp DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{source, limit, offset}
	} else {
		query = `
			SELECT id, trace_id, event_type, event_category, actor_type, actor_id, entity_type, 
				entity_id, event_data, severity, message, timestamp, created_at
			FROM event_logs
			ORDER BY timestamp DESC
			LIMIT $1 OFFSET $2
		`
		args = []interface{}{limit, offset}
	}

	return r.scanEventLogs(ctx, query, args...)
}

// GetEventLogsByTimeRange retrieves event logs within a time range
func (r *ObservabilityRepositoryImpl) GetEventLogsByTimeRange(ctx context.Context, startTime, endTime string, limit, offset int) ([]*models.EventLog, error) {
	startTimeParsed, err := time.Parse(time.RFC3339, startTime)
	if err != nil {
		return nil, fmt.Errorf("invalid start time format: %w", err)
	}

	endTimeParsed, err := time.Parse(time.RFC3339, endTime)
	if err != nil {
		return nil, fmt.Errorf("invalid end time format: %w", err)
	}

	query := `
		SELECT id, trace_id, event_type, event_category, actor_type, actor_id, entity_type, 
			entity_id, event_data, severity, message, timestamp, created_at
		FROM event_logs
		WHERE timestamp >= $1 AND timestamp <= $2
		ORDER BY timestamp DESC
		LIMIT $3 OFFSET $4
	`

	return r.scanEventLogs(ctx, query, startTimeParsed, endTimeParsed, limit, offset)
}

// Helper methods for scanning results

func (r *ObservabilityRepositoryImpl) scanActorMessages(ctx context.Context, query string, args ...interface{}) ([]*models.ActorMessage, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var messages []*models.ActorMessage
	for rows.Next() {
		message := &models.ActorMessage{}
		err := rows.Scan(
			&message.ID,
			&message.TraceID,
			&message.SpanID,
			&message.ParentSpanID,
			&message.SenderActorType,
			&message.SenderActorID,
			&message.ReceiverActorType,
			&message.ReceiverActorID,
			&message.MessageType,
			&message.MessagePayload,
			&message.Status,
			&message.SentAt,
			&message.ReceivedAt,
			&message.ProcessedAt,
			&message.ProcessingDurationMs,
			&message.ErrorMessage,
			&message.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan actor message: %w", err)
		}
		messages = append(messages, message)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating actor messages: %w", err)
	}

	return messages, nil
}

func (r *ObservabilityRepositoryImpl) scanSystemMetrics(ctx context.Context, query string, args ...interface{}) ([]*models.SystemMetric, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var metrics []*models.SystemMetric
	for rows.Next() {
		metric := &models.SystemMetric{}
		err := rows.Scan(
			&metric.ID,
			&metric.MetricName,
			&metric.MetricType,
			&metric.MetricValue,
			&metric.Labels,
			&metric.ActorType,
			&metric.ActorID,
			&metric.Timestamp,
			&metric.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan system metric: %w", err)
		}
		metrics = append(metrics, metric)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating system metrics: %w", err)
	}

	return metrics, nil
}

func (r *ObservabilityRepositoryImpl) scanDistributedTraces(ctx context.Context, query string, args ...interface{}) ([]*models.DistributedTrace, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var traces []*models.DistributedTrace
	for rows.Next() {
		trace := &models.DistributedTrace{}
		err := rows.Scan(
			&trace.ID,
			&trace.TraceID,
			&trace.SpanID,
			&trace.ParentSpanID,
			&trace.OperationName,
			&trace.ActorType,
			&trace.ActorID,
			&trace.StartTime,
			&trace.EndTime,
			&trace.DurationMs,
			&trace.Status,
			&trace.Tags,
			&trace.Logs,
			&trace.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan distributed trace: %w", err)
		}
		traces = append(traces, trace)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating distributed traces: %w", err)
	}

	return traces, nil
}

func (r *ObservabilityRepositoryImpl) scanEventLogs(ctx context.Context, query string, args ...interface{}) ([]*models.EventLog, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var logs []*models.EventLog
	for rows.Next() {
		log := &models.EventLog{}
		err := rows.Scan(
			&log.ID,
			&log.TraceID,
			&log.EventType,
			&log.EventCategory,
			&log.ActorType,
			&log.ActorID,
			&log.EntityType,
			&log.EntityID,
			&log.EventData,
			&log.Severity,
			&log.Message,
			&log.Timestamp,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event log: %w", err)
		}
		logs = append(logs, log)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating event logs: %w", err)
	}

	return logs, nil
}
