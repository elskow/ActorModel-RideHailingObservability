package tests

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"actor-model-observability/internal/models"
	"actor-model-observability/internal/repository/postgres"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestObservabilityRepository_CreateActorInstance_Success(t *testing.T) {
	db, mock := SetupMockDB(t)
	defer db.Close()

	repo := postgres.NewObservabilityRepository(db)

	actorID := uuid.New()
	now := time.Now()

	entityID := uuid.New()
	instance := &models.ActorInstance{
		ID:            actorID,
		ActorType:     models.ActorTypePassenger,
		ActorID:       "passenger-123",
		EntityID:      &entityID,
		Status:        models.ActorStatusActive,
		CreatedAt:     now,
		UpdatedAt:     now,
		LastHeartbeat: now,
	}

	mock.ExpectExec(`INSERT INTO actor_instances`).
		WithArgs(
			actorID, models.ActorTypePassenger, "passenger-123", &entityID, models.ActorStatusActive,
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.CreateActorInstance(context.Background(), instance)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestObservabilityRepository_GetActorInstance_Success(t *testing.T) {
	db, mock := SetupMockDB(t)
	defer db.Close()

	repo := postgres.NewObservabilityRepository(db)

	actorID := uuid.New()
	entityID := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "actor_type", "actor_id", "entity_id", "status", "last_heartbeat", "created_at", "updated_at",
	}).AddRow(
		actorID, models.ActorTypePassenger, "passenger-123", &entityID, models.ActorStatusActive, now, now, now,
	)

	mock.ExpectQuery(`SELECT (.+) FROM actor_instances WHERE id = \$1`).
		WithArgs(actorID.String()).
		WillReturnRows(rows)

	instance, err := repo.GetActorInstance(context.Background(), actorID.String())

	assert.NoError(t, err)
	assert.NotNil(t, instance)
	assert.Equal(t, actorID, instance.ID)
	assert.Equal(t, models.ActorType("passenger"), instance.ActorType)
	assert.Equal(t, "passenger-123", instance.ActorID)
	assert.Equal(t, models.ActorStatus("active"), instance.Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestObservabilityRepository_CreateActorMessage_Success(t *testing.T) {
	db, mock := SetupMockDB(t)
	defer db.Close()

	repo := postgres.NewObservabilityRepository(db)

	messageID := uuid.New()
	now := time.Now()

	traceID := uuid.New()
	spanID := uuid.New()

	message := &models.ActorMessage{
		ID:                messageID,
		TraceID:           traceID,
		SpanID:            spanID,
		MessageType:       "ride_request",
		SenderActorType:   "passenger",
		SenderActorID:     "passenger-123",
		ReceiverActorType: "driver",
		ReceiverActorID:   "driver-456",
		SentAt:            now,
		MessagePayload:    json.RawMessage(`{"pickup_lat": 40.7128, "pickup_lng": -74.0060}`),
		CreatedAt:         now,
	}

	mock.ExpectExec(`INSERT INTO actor_messages`).
		WithArgs(
			messageID, traceID, spanID, sqlmock.AnyArg(), "passenger", "passenger-123",
			"driver", "driver-456", "ride_request", sqlmock.AnyArg(), sqlmock.AnyArg(),
			now, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), now,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.CreateActorMessage(context.Background(), message)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestObservabilityRepository_ListActorMessages_Success(t *testing.T) {
	db, mock := SetupMockDB(t)
	defer db.Close()

	repo := postgres.NewObservabilityRepository(db)

	messageID1 := uuid.New()
	messageID2 := uuid.New()
	now := time.Now()

	traceID := uuid.New()
	spanID1 := uuid.New()
	spanID2 := uuid.New()

	rows := sqlmock.NewRows([]string{
		"id", "trace_id", "span_id", "parent_span_id", "sender_actor_type", "sender_actor_id", "receiver_actor_type", "receiver_actor_id", "message_type", "message_payload", "status", "sent_at", "received_at", "processed_at", "processing_duration_ms", "error_message", "created_at",
	}).AddRow(
		messageID1, traceID, spanID1, nil, models.ActorTypePassenger, "passenger-123", models.ActorTypeDriver, "driver-456", "ride_request", json.RawMessage(`{"pickup_lat": 40.7128}`), models.MessageStatusSent, now, nil, nil, nil, nil, now,
	).AddRow(
		messageID2, traceID, spanID2, nil, models.ActorTypeDriver, "driver-456", models.ActorTypePassenger, "passenger-123", "ride_accepted", json.RawMessage(`{"eta": 5}`), models.MessageStatusSent, now, nil, nil, nil, nil, now,
	)

	mock.ExpectQuery(`SELECT (.+) FROM actor_messages WHERE sender_actor_id = \$1 AND receiver_actor_id = \$2`).
		WithArgs("passenger-123", "driver-456", 10, 0).
		WillReturnRows(rows)

	messages, err := repo.ListActorMessages(context.Background(), "passenger-123", "driver-456", 10, 0)

	assert.NoError(t, err)
	assert.Len(t, messages, 2)
	assert.Equal(t, messageID1, messages[0].ID)
	assert.Equal(t, "ride_request", messages[0].MessageType)
	assert.Equal(t, messageID2, messages[1].ID)
	assert.Equal(t, "ride_accepted", messages[1].MessageType)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestObservabilityRepository_CreateSystemMetric_Success(t *testing.T) {
	db, mock := SetupMockDB(t)
	defer db.Close()

	repo := postgres.NewObservabilityRepository(db)

	metricID := uuid.New()
	now := time.Now()

	metric := &models.SystemMetric{
		ID:          metricID,
		MetricName:  "cpu_usage",
		MetricType:  models.MetricTypeCounter,
		MetricValue: 75.5,
		Labels:      json.RawMessage(`{"host": "server-1", "core": "0"}`),
		Timestamp:   now,
		CreatedAt:   now,
	}

	mock.ExpectExec(`INSERT INTO system_metrics`).
		WithArgs(
			metricID, "cpu_usage", models.MetricTypeCounter, 75.5,
			json.RawMessage(`{"host": "server-1", "core": "0"}`),
			nil, nil, sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.CreateSystemMetric(context.Background(), metric)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestObservabilityRepository_ListSystemMetrics_Success(t *testing.T) {
	db, mock := SetupMockDB(t)
	defer db.Close()

	repo := postgres.NewObservabilityRepository(db)

	metricID1 := uuid.New()
	metricID2 := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "metric_name", "metric_type", "metric_value", "labels", "actor_type", "actor_id", "timestamp", "created_at",
	}).AddRow(
		metricID1, "system.cpu.percent", models.MetricTypeCounter, 75.5, json.RawMessage(`{"host": "server-1"}`), nil, nil, now, now,
	).AddRow(
		metricID2, "system.cpu.percent", models.MetricTypeCounter, 80.2, json.RawMessage(`{"host": "server-2"}`), nil, nil, now, now,
	)

	mock.ExpectQuery(`SELECT (.+) FROM system_metrics WHERE metric_type = \$1`).
		WithArgs("counter", 10, 0).
		WillReturnRows(rows)

	metrics, err := repo.ListSystemMetrics(context.Background(), "counter", 10, 0)

	assert.NoError(t, err)
	assert.Len(t, metrics, 2)
	assert.Equal(t, metricID1, metrics[0].ID)
	assert.Equal(t, models.MetricTypeCounter, metrics[0].MetricType)
	assert.Equal(t, 75.5, metrics[0].MetricValue)
	assert.Equal(t, metricID2, metrics[1].ID)
	assert.Equal(t, 80.2, metrics[1].MetricValue)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestObservabilityRepository_CreateDistributedTrace_Success(t *testing.T) {
	db, mock := SetupMockDB(t)
	defer db.Close()

	repo := postgres.NewObservabilityRepository(db)

	traceID := uuid.New()
	now := time.Now()

	spanID := uuid.New()
	parentSpanID := uuid.New()

	trace := &models.DistributedTrace{
		ID:            traceID,
		TraceID:       traceID,
		SpanID:        spanID,
		ParentSpanID:  &parentSpanID,
		OperationName: "ride_request",
		StartTime:     now,
		EndTime:       &now,
		DurationMs:    IntPtr(100),
		Status:        models.TraceStatusOK,
		Tags:          json.RawMessage(`{"user_id": "123", "trip_id": "456"}`),
		Logs:          json.RawMessage(`{"events": []}`),
		CreatedAt:     now,
	}

	mock.ExpectExec(`INSERT INTO distributed_traces`).
		WithArgs(
			traceID, traceID, spanID, &parentSpanID, "ride_request",
			nil, nil, now, &now, IntPtr(100), models.TraceStatusOK,
			json.RawMessage(`{"user_id": "123", "trip_id": "456"}`),
			json.RawMessage(`{"events": []}`), now,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.CreateDistributedTrace(context.Background(), trace)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestObservabilityRepository_GetTracesByTraceID_Success(t *testing.T) {
	db, mock := SetupMockDB(t)
	defer db.Close()

	repo := postgres.NewObservabilityRepository(db)

	traceID1 := uuid.New()
	traceID2 := uuid.New()
	now := time.Now()

	traceUUID := uuid.New()
	spanID1 := uuid.New()
	spanID2 := uuid.New()
	parentSpanID := uuid.New()

	rows := sqlmock.NewRows([]string{
		"id", "trace_id", "span_id", "parent_span_id", "operation_name", "actor_type", "actor_id",
		"start_time", "end_time", "duration_ms", "status", "tags", "logs", "created_at",
	}).AddRow(
		traceID1, traceUUID, spanID1, &parentSpanID, "ride_request", nil, nil,
		now, &now, 100, models.TraceStatusOK, json.RawMessage(`{"user_id": "123"}`), json.RawMessage(`{}`), now,
	).AddRow(
		traceID2, traceUUID, spanID2, &spanID1, "find_driver", nil, nil,
		now, &now, 50, models.TraceStatusOK, json.RawMessage(`{"driver_id": "456"}`), json.RawMessage(`{}`), now,
	)

	mock.ExpectQuery(`SELECT (.+) FROM distributed_traces WHERE trace_id = \$1`).
		WithArgs(traceUUID).
		WillReturnRows(rows)

	traces, err := repo.GetTracesByTraceID(context.Background(), traceUUID.String())

	assert.NoError(t, err)
	assert.Len(t, traces, 2)
	assert.Equal(t, traceID1, traces[0].ID)
	assert.Equal(t, "ride_request", traces[0].OperationName)
	assert.Equal(t, traceID2, traces[1].ID)
	assert.Equal(t, "find_driver", traces[1].OperationName)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestObservabilityRepository_CreateEventLog_Success(t *testing.T) {
	db, mock := SetupMockDB(t)
	defer db.Close()

	repo := postgres.NewObservabilityRepository(db)

	eventID := uuid.New()
	now := time.Now()

	eventLog := &models.EventLog{
		ID:            eventID,
		EventType:     "ride_requested",
		EventCategory: models.EventCategoryBusiness,
		Message:       "Passenger requested a ride",
		Severity:      models.EventSeverityInfo,
		EventData:     json.RawMessage(`{"passenger_id": "123", "pickup_lat": 40.7128}`),
		Timestamp:     now,
		CreatedAt:     now,
	}

	mock.ExpectExec(`INSERT INTO event_logs`).
		WithArgs(
			eventID, nil, "ride_requested", models.EventCategoryBusiness,
			nil, nil, nil, nil,
			json.RawMessage(`{"passenger_id": "123", "pickup_lat": 40.7128}`),
			models.EventSeverityInfo, "Passenger requested a ride", now, now,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.CreateEventLog(context.Background(), eventLog)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestObservabilityRepository_ListEventLogs_Success(t *testing.T) {
	db, mock := SetupMockDB(t)
	defer db.Close()

	repo := postgres.NewObservabilityRepository(db)

	eventID1 := uuid.New()
	eventID2 := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "trace_id", "event_type", "event_category", "actor_type", "actor_id", "entity_type", "entity_id", "event_data", "severity", "message", "timestamp", "created_at",
	}).AddRow(
		eventID1, nil, "ride_requested", models.EventCategoryBusiness, nil, nil, nil, nil, json.RawMessage(`{"passenger_id": "123"}`), models.EventSeverityInfo, "Passenger requested a ride", now, now,
	).AddRow(
		eventID2, nil, "ride_requested", models.EventCategoryBusiness, nil, nil, nil, nil, json.RawMessage(`{"passenger_id": "456"}`), models.EventSeverityInfo, "Another ride requested", now, now,
	)

	mock.ExpectQuery(`SELECT (.+) FROM event_logs WHERE event_type = \$1`).
		WithArgs("ride_requested", "business", 10, 0).
		WillReturnRows(rows)

	eventLogs, err := repo.ListEventLogs(context.Background(), "ride_requested", "business", 10, 0)

	assert.NoError(t, err)
	assert.Len(t, eventLogs, 2)
	assert.Equal(t, eventID1, eventLogs[0].ID)
	assert.Equal(t, "ride_requested", eventLogs[0].EventType)
	assert.Equal(t, models.EventCategoryBusiness, eventLogs[0].EventCategory)
	assert.Equal(t, eventID2, eventLogs[1].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestObservabilityRepository_GetEventLogsByTimeRange_Success(t *testing.T) {
	db, mock := SetupMockDB(t)
	defer db.Close()

	repo := postgres.NewObservabilityRepository(db)

	eventID := uuid.New()
	now := time.Now()
	startTime := now.Add(-1 * time.Hour)
	endTime := now

	rows := sqlmock.NewRows([]string{
		"id", "trace_id", "event_type", "event_category", "actor_type", "actor_id", "entity_type", "entity_id", "event_data", "severity", "message", "timestamp", "created_at",
	}).AddRow(
		eventID, nil, "ride_requested", models.EventCategoryBusiness, nil, nil, nil, nil, json.RawMessage(`{"passenger_id": "123"}`), models.EventSeverityInfo, "Passenger requested a ride", now, now,
	)

	mock.ExpectQuery(`SELECT (.+) FROM event_logs WHERE (.+)`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	eventLogs, err := repo.GetEventLogsByTimeRange(context.Background(), startTime.Format(time.RFC3339), endTime.Format(time.RFC3339), 10, 0)

	assert.NoError(t, err)
	assert.Len(t, eventLogs, 1)
	assert.Equal(t, eventID, eventLogs[0].ID)
	assert.Equal(t, "ride_requested", eventLogs[0].EventType)
	assert.NoError(t, mock.ExpectationsWereMet())
}
