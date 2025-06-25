package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ActorType represents the type of actor
type ActorType string

const (
	ActorTypePassenger     ActorType = "passenger"
	ActorTypeDriver        ActorType = "driver"
	ActorTypeTrip          ActorType = "trip"
	ActorTypeMatching      ActorType = "matching"
	ActorTypeObservability ActorType = "observability"
)

// ActorStatus represents the status of an actor
type ActorStatus string

const (
	ActorStatusActive   ActorStatus = "active"
	ActorStatusInactive ActorStatus = "inactive"
	ActorStatusError    ActorStatus = "error"
)

// ActorInstance represents an actor instance in the system
type ActorInstance struct {
	ID            uuid.UUID    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ActorType     ActorType    `json:"actor_type" gorm:"not null;check:actor_type IN ('passenger', 'driver', 'trip', 'matching', 'observability')"`
	ActorID       string       `json:"actor_id" gorm:"not null"`
	EntityID      *uuid.UUID   `json:"entity_id" gorm:"type:uuid"`
	Status        ActorStatus  `json:"status" gorm:"default:'active';check:status IN ('active', 'inactive', 'error')"`
	LastHeartbeat time.Time    `json:"last_heartbeat" gorm:"default:CURRENT_TIMESTAMP"`
	CreatedAt     time.Time    `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time    `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
}

// TableName returns the table name for ActorInstance
func (ActorInstance) TableName() string {
	return "actor_instances"
}

// MessageStatus represents the status of a message
type MessageStatus string

const (
	MessageStatusSent      MessageStatus = "sent"
	MessageStatusReceived  MessageStatus = "received"
	MessageStatusProcessed MessageStatus = "processed"
	MessageStatusFailed    MessageStatus = "failed"
)

// ActorMessage represents a message between actors
type ActorMessage struct {
	ID                   uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TraceID              uuid.UUID      `json:"trace_id" gorm:"type:uuid;not null;index"`
	SpanID               uuid.UUID      `json:"span_id" gorm:"type:uuid;not null;index"`
	ParentSpanID         *uuid.UUID     `json:"parent_span_id" gorm:"type:uuid;index"`
	SenderActorType      ActorType      `json:"sender_actor_type" gorm:"not null"`
	SenderActorID        string         `json:"sender_actor_id" gorm:"not null"`
	ReceiverActorType    ActorType      `json:"receiver_actor_type" gorm:"not null"`
	ReceiverActorID      string         `json:"receiver_actor_id" gorm:"not null"`
	MessageType          string         `json:"message_type" gorm:"not null"`
	MessagePayload       json.RawMessage `json:"message_payload" gorm:"type:jsonb" swaggertype:"object"`
	Status               MessageStatus  `json:"status" gorm:"default:'sent';check:status IN ('sent', 'received', 'processed', 'failed')"`
	SentAt               time.Time      `json:"sent_at" gorm:"default:CURRENT_TIMESTAMP"`
	ReceivedAt           *time.Time     `json:"received_at"`
	ProcessedAt          *time.Time     `json:"processed_at"`
	ProcessingDurationMs *int           `json:"processing_duration_ms"`
	ErrorMessage         *string        `json:"error_message"`
	CreatedAt            time.Time      `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
}

// TableName returns the table name for ActorMessage
func (ActorMessage) TableName() string {
	return "actor_messages"
}

// MetricType represents the type of metric
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
)

// SystemMetric represents a system metric
type SystemMetric struct {
	ID          uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	MetricName  string          `json:"metric_name" gorm:"not null;index"`
	MetricType  MetricType      `json:"metric_type" gorm:"not null;check:metric_type IN ('counter', 'gauge', 'histogram')"`
	MetricValue float64         `json:"metric_value" gorm:"type:decimal(15,6);not null"`
	Labels      json.RawMessage `json:"labels" gorm:"type:jsonb" swaggertype:"object"`
	ActorType   *ActorType      `json:"actor_type"`
	ActorID     *string         `json:"actor_id"`
	Timestamp   time.Time       `json:"timestamp" gorm:"default:CURRENT_TIMESTAMP;index"`
	CreatedAt   time.Time       `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
}

// TableName returns the table name for SystemMetric
func (SystemMetric) TableName() string {
	return "system_metrics"
}

// TraceStatus represents the status of a trace
type TraceStatus string

const (
	TraceStatusOK      TraceStatus = "ok"
	TraceStatusError   TraceStatus = "error"
	TraceStatusTimeout TraceStatus = "timeout"
)

// DistributedTrace represents a distributed trace span
type DistributedTrace struct {
	ID            uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TraceID       uuid.UUID       `json:"trace_id" gorm:"type:uuid;not null;index"`
	SpanID        uuid.UUID       `json:"span_id" gorm:"type:uuid;not null;index"`
	ParentSpanID  *uuid.UUID      `json:"parent_span_id" gorm:"type:uuid;index"`
	OperationName string          `json:"operation_name" gorm:"not null"`
	ActorType     *ActorType      `json:"actor_type"`
	ActorID       *string         `json:"actor_id"`
	StartTime     time.Time       `json:"start_time" gorm:"not null;index"`
	EndTime       *time.Time      `json:"end_time"`
	DurationMs    *int            `json:"duration_ms"`
	Status        TraceStatus     `json:"status" gorm:"default:'ok';check:status IN ('ok', 'error', 'timeout')"`
	Tags          json.RawMessage `json:"tags" gorm:"type:jsonb" swaggertype:"object"`
	Logs          json.RawMessage `json:"logs" gorm:"type:jsonb" swaggertype:"object"`
	CreatedAt     time.Time       `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
}

// TableName returns the table name for DistributedTrace
func (DistributedTrace) TableName() string {
	return "distributed_traces"
}

// EventCategory represents the category of an event
type EventCategory string

const (
	EventCategoryBusiness    EventCategory = "business"
	EventCategorySystem      EventCategory = "system"
	EventCategoryError       EventCategory = "error"
	EventCategoryPerformance EventCategory = "performance"
	EventCategorySecurity    EventCategory = "security"
)

// EventSeverity represents the severity of an event
type EventSeverity string

const (
	EventSeverityDebug EventSeverity = "debug"
	EventSeverityInfo  EventSeverity = "info"
	EventSeverityWarn  EventSeverity = "warn"
	EventSeverityError EventSeverity = "error"
	EventSeverityFatal EventSeverity = "fatal"
)

// EventLog represents an event log entry
type EventLog struct {
	ID            uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TraceID       *uuid.UUID      `json:"trace_id" gorm:"type:uuid;index"`
	EventType     string          `json:"event_type" gorm:"not null;index"`
	EventCategory EventCategory   `json:"event_category" gorm:"not null;check:event_category IN ('business', 'system', 'error', 'performance', 'security')"`
	ActorType     *ActorType      `json:"actor_type"`
	ActorID       *string         `json:"actor_id"`
	EntityType    *string         `json:"entity_type"`
	EntityID      *uuid.UUID      `json:"entity_id" gorm:"type:uuid"`
	EventData     json.RawMessage `json:"event_data" gorm:"type:jsonb" swaggertype:"object"`
	Severity      EventSeverity   `json:"severity" gorm:"default:'info';check:severity IN ('debug', 'info', 'warn', 'error', 'fatal')"`
	Message       string          `json:"message"`
	Timestamp     time.Time       `json:"timestamp" gorm:"default:CURRENT_TIMESTAMP;index"`
	CreatedAt     time.Time       `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
}

// TableName returns the table name for EventLog
func (EventLog) TableName() string {
	return "event_logs"
}

// Helper methods for ActorMessage

// MarkReceived marks the message as received
func (am *ActorMessage) MarkReceived() {
	am.Status = MessageStatusReceived
	now := time.Now()
	am.ReceivedAt = &now
}

// MarkProcessed marks the message as processed
func (am *ActorMessage) MarkProcessed(processingDuration time.Duration) {
	am.Status = MessageStatusProcessed
	now := time.Now()
	am.ProcessedAt = &now
	durationMs := int(processingDuration.Milliseconds())
	am.ProcessingDurationMs = &durationMs
}

// MarkFailed marks the message as failed
func (am *ActorMessage) MarkFailed(errorMsg string) {
	am.Status = MessageStatusFailed
	am.ErrorMessage = &errorMsg
}

// Helper methods for DistributedTrace

// Finish finishes the trace span
func (dt *DistributedTrace) Finish() {
	now := time.Now()
	dt.EndTime = &now
	duration := int(now.Sub(dt.StartTime).Milliseconds())
	dt.DurationMs = &duration
}

// FinishWithError finishes the trace span with an error
func (dt *DistributedTrace) FinishWithError(err error) {
	dt.Finish()
	dt.Status = TraceStatusError
	// Add error to logs
	errorLog := map[string]interface{}{
		"error": err.Error(),
		"timestamp": time.Now(),
	}
	if dt.Logs == nil {
		dt.Logs = json.RawMessage("[]")
	}
	// Append error to existing logs
	var logs []interface{}
	json.Unmarshal(dt.Logs, &logs)
	logs = append(logs, errorLog)
	dt.Logs, _ = json.Marshal(logs)
}

// AddTag adds a tag to the trace
func (dt *DistributedTrace) AddTag(key string, value interface{}) {
	if dt.Tags == nil {
		dt.Tags = json.RawMessage("{}")
	}
	var tags map[string]interface{}
	json.Unmarshal(dt.Tags, &tags)
	if tags == nil {
		tags = make(map[string]interface{})
	}
	tags[key] = value
	dt.Tags, _ = json.Marshal(tags)
}

// AddLog adds a log entry to the trace
func (dt *DistributedTrace) AddLog(message string, fields map[string]interface{}) {
	logEntry := map[string]interface{}{
		"message": message,
		"timestamp": time.Now(),
	}
	for k, v := range fields {
		logEntry[k] = v
	}
	
	if dt.Logs == nil {
		dt.Logs = json.RawMessage("[]")
	}
	var logs []interface{}
	json.Unmarshal(dt.Logs, &logs)
	logs = append(logs, logEntry)
	dt.Logs, _ = json.Marshal(logs)
}

// TraditionalMetric represents a traditional monitoring metric
type TraditionalMetric struct {
	ID          uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	MetricName  string          `json:"metric_name" gorm:"not null;index"`
	MetricType  MetricType      `json:"metric_type" gorm:"not null;check:metric_type IN ('counter', 'gauge', 'histogram')"`
	MetricValue float64         `json:"metric_value" gorm:"type:decimal(15,6);not null"`
	Labels      json.RawMessage `json:"labels" gorm:"type:jsonb" swaggertype:"object"`
	ServiceName string          `json:"service_name" gorm:"not null"`
	InstanceID  string          `json:"instance_id"`
	Timestamp   time.Time       `json:"timestamp" gorm:"default:CURRENT_TIMESTAMP;index"`
	CreatedAt   time.Time       `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
}

// TableName returns the table name for TraditionalMetric
func (TraditionalMetric) TableName() string {
	return "traditional_metrics"
}

// LogLevel represents the level of a log entry
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
	LogLevelFatal LogLevel = "fatal"
)

// TraditionalLog represents a traditional log entry
type TraditionalLog struct {
	ID          uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Level       LogLevel        `json:"level" gorm:"not null;check:level IN ('debug', 'info', 'warn', 'error', 'fatal')"`
	Message     string          `json:"message" gorm:"not null"`
	ServiceName string          `json:"service_name" gorm:"not null"`
	InstanceID  string          `json:"instance_id"`
	Fields      json.RawMessage `json:"fields" gorm:"type:jsonb" swaggertype:"object"`
	Timestamp   time.Time       `json:"timestamp" gorm:"default:CURRENT_TIMESTAMP;index"`
	CreatedAt   time.Time       `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
}

// TableName returns the table name for TraditionalLog
func (TraditionalLog) TableName() string {
	return "traditional_logs"
}

// ServiceHealth represents a service health record
type ServiceHealth struct {
	ID                  uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ServiceName         string    `json:"service_name" gorm:"not null"`
	Status              string    `json:"status" gorm:"not null"`
	ResponseTimeMs      float64   `json:"response_time_ms"`
	CPUUsagePercent     float64   `json:"cpu_usage_percent"`
	MemoryUsagePercent  float64   `json:"memory_usage_percent"`
	DiskUsagePercent    float64   `json:"disk_usage_percent"`
	ActiveConnections   int       `json:"active_connections"`
	ErrorRatePercent    float64   `json:"error_rate_percent"`
	LastError           string    `json:"last_error"`
	Timestamp           time.Time `json:"timestamp" gorm:"default:CURRENT_TIMESTAMP;index"`
	CreatedAt           time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
}

// TableName returns the table name for ServiceHealth
func (ServiceHealth) TableName() string {
	return "service_health"
}