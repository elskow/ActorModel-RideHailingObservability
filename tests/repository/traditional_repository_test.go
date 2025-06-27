package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"actor-model-observability/internal/models"
	"actor-model-observability/internal/repository/postgres"
	"actor-model-observability/tests/utils"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestTraditionalRepository_CreateTraditionalMetric_Success(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	repo := postgres.NewTraditionalRepository(db)

	metricID := uuid.New()
	now := time.Now()

	labels, _ := json.Marshal(map[string]string{"method": "GET", "endpoint": "/api/rides", "status": "200"})
	metric := &models.TraditionalMetric{
		ID:          metricID,
		MetricName:  "http_requests_total",
		MetricType:  "counter",
		MetricValue: 1250.0,
		Labels:      labels,
		ServiceName: "ride-service",
		InstanceID:  "instance-1",
		Timestamp:   now,
		CreatedAt:   now,
	}

	mock.ExpectExec(`INSERT INTO traditional_metrics`).
		WithArgs(
			metricID, "http_requests_total", "counter", 1250.0,
			sqlmock.AnyArg(), "ride-service", "instance-1",
			sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.CreateTraditionalMetric(context.Background(), metric)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTraditionalRepository_GetTraditionalMetric_Success(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	repo := postgres.NewTraditionalRepository(db)

	metricID := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "metric_name", "metric_type", "metric_value", "labels", "service_name", "instance_id", "timestamp", "created_at",
	}).AddRow(
		metricID, "http_requests_total", "counter", 1250.0,
		[]byte(`{"method": "GET", "endpoint": "/api/rides"}`), "ride-service", "instance-1", now, now,
	)

	mock.ExpectQuery(`SELECT (.+) FROM traditional_metrics WHERE id = \$1`).
		WithArgs(metricID.String()).
		WillReturnRows(rows)

	metric, err := repo.GetTraditionalMetric(context.Background(), metricID.String())

	if err != nil {
		t.Fatalf("Error getting traditional metric: %v", err)
	}
	if metric == nil {
		t.Fatal("Metric is nil")
	}
	assert.NoError(t, err)
	assert.NotNil(t, metric)
	assert.Equal(t, metricID, metric.ID)
	assert.Equal(t, "http_requests_total", metric.MetricName)
	assert.Equal(t, models.MetricType("counter"), metric.MetricType)
	assert.Equal(t, 1250.0, metric.MetricValue)
	assert.Equal(t, "ride-service", metric.ServiceName)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTraditionalRepository_GetTraditionalMetric_NotFound(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	repo := postgres.NewTraditionalRepository(db)
	metricID := uuid.New()

	mock.ExpectQuery(`SELECT (.+) FROM traditional_metrics WHERE id = \$1`).
		WithArgs(metricID.String()).
		WillReturnError(sql.ErrNoRows)

	metric, err := repo.GetTraditionalMetric(context.Background(), metricID.String())

	assert.Error(t, err)
	assert.Nil(t, metric)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTraditionalRepository_ListTraditionalMetrics_Success(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	repo := postgres.NewTraditionalRepository(db)

	metricID1 := uuid.New()
	metricID2 := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "metric_name", "metric_type", "metric_value", "labels", "service_name", "instance_id", "timestamp", "created_at",
	}).AddRow(
		metricID1, "http_requests_total", "counter", 1250.0,
		[]byte(`{"method": "GET"}`), "ride-service", "instance-1", now, now,
	).AddRow(
		metricID2, "http_requests_total", "counter", 850.0,
		[]byte(`{"method": "POST"}`), "ride-service", "instance-1", now, now,
	)

	mock.ExpectQuery(`SELECT (.+) FROM traditional_metrics WHERE metric_type = \$1 AND service_name = \$2`).
		WithArgs("counter", "ride-service", 10, 0).
		WillReturnRows(rows)

	metrics, err := repo.ListTraditionalMetrics(context.Background(), "counter", "ride-service", 10, 0)

	assert.NoError(t, err)
	assert.Len(t, metrics, 2)
	assert.Equal(t, metricID1, metrics[0].ID)
	assert.Equal(t, "http_requests_total", metrics[0].MetricName)
	assert.Equal(t, 1250.0, metrics[0].MetricValue)
	assert.Equal(t, metricID2, metrics[1].ID)
	assert.Equal(t, 850.0, metrics[1].MetricValue)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTraditionalRepository_GetTraditionalMetricsByTimeRange_Success(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	repo := postgres.NewTraditionalRepository(db)

	metricID := uuid.New()
	now := time.Now()
	startTime := now.Add(-1 * time.Hour)
	endTime := now

	rows := sqlmock.NewRows([]string{
		"id", "metric_name", "metric_type", "metric_value", "labels", "service_name", "instance_id", "timestamp", "created_at",
	}).AddRow(
		metricID, "response_time_ms", "histogram", 125.5,
		[]byte(`{"endpoint": "/api/rides"}`), "ride-service", "instance-1", now, now,
	)

	mock.ExpectQuery(`SELECT (.+) FROM traditional_metrics WHERE timestamp >= \$1 AND timestamp <= \$2`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), 10, 0).
		WillReturnRows(rows)

	metrics, err := repo.GetTraditionalMetricsByTimeRange(context.Background(), startTime.Format(time.RFC3339), endTime.Format(time.RFC3339), 10, 0)

	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, metricID, metrics[0].ID)
	assert.Equal(t, "response_time_ms", metrics[0].MetricName)
	assert.Equal(t, models.MetricType("histogram"), metrics[0].MetricType)
	assert.Equal(t, 125.5, metrics[0].MetricValue)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTraditionalRepository_CreateTraditionalLog_Success(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	repo := postgres.NewTraditionalRepository(db)

	logID := uuid.New()
	now := time.Now()

	fieldsData, _ := json.Marshal(map[string]interface{}{"trip_id": "123", "passenger_id": "456", "duration_ms": 150})
	log := &models.TraditionalLog{
		ID:          logID,
		Level:       "info",
		Message:     "Ride request processed successfully",
		ServiceName: "ride-service",
		InstanceID:  "instance-1",
		Fields:      fieldsData,
		Timestamp:   now,
		CreatedAt:   now,
	}

	mock.ExpectExec(`INSERT INTO traditional_logs`).
		WithArgs(
			logID, "info", "Ride request processed successfully", "ride-service", "instance-1",
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.CreateTraditionalLog(context.Background(), log)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTraditionalRepository_GetTraditionalLog_Success(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	repo := postgres.NewTraditionalRepository(db)

	logID := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "level", "message", "service_name", "instance_id", "fields", "timestamp", "created_at",
	}).AddRow(
		logID, "info", "Ride request processed successfully", "ride-service", "instance-1",
		[]byte(`{"trip_id": "123", "passenger_id": "456"}`), now, now,
	)

	mock.ExpectQuery(`SELECT (.+) FROM traditional_logs WHERE id = \$1`).
		WithArgs(logID.String()).
		WillReturnRows(rows)

	log, err := repo.GetTraditionalLog(context.Background(), logID.String())

	assert.NoError(t, err)
	assert.NotNil(t, log)
	assert.Equal(t, logID, log.ID)
	assert.Equal(t, models.LogLevel("info"), log.Level)
	assert.Equal(t, "Ride request processed successfully", log.Message)
	assert.Equal(t, "ride-service", log.ServiceName)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTraditionalRepository_GetTraditionalLog_NotFound(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	repo := postgres.NewTraditionalRepository(db)
	logID := uuid.New()

	mock.ExpectQuery(`SELECT (.+) FROM traditional_logs WHERE id = \$1`).
		WithArgs(logID.String()).
		WillReturnError(sql.ErrNoRows)

	log, err := repo.GetTraditionalLog(context.Background(), logID.String())

	assert.Error(t, err)
	assert.Nil(t, log)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTraditionalRepository_ListTraditionalLogs_Success(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	repo := postgres.NewTraditionalRepository(db)

	logID1 := uuid.New()
	logID2 := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "level", "message", "service_name", "instance_id", "fields", "timestamp", "created_at",
	}).AddRow(
		logID1, "info", "Ride request processed", "ride-service", "instance-1",
		[]byte(`{"trip_id": "123"}`), now, now,
	).AddRow(
		logID2, "info", "Driver matched successfully", "ride-service", "instance-1",
		[]byte(`{"driver_id": "456"}`), now, now,
	)

	mock.ExpectQuery(`SELECT (.+) FROM traditional_logs WHERE level = \$1 AND service_name = \$2`).
		WithArgs("info", "ride-service", 10, 0).
		WillReturnRows(rows)

	logs, err := repo.ListTraditionalLogs(context.Background(), "info", "ride-service", 10, 0)

	assert.NoError(t, err)
	assert.Len(t, logs, 2)
	assert.Equal(t, logID1, logs[0].ID)
	assert.Equal(t, models.LogLevel("info"), logs[0].Level)
	assert.Equal(t, "Ride request processed", logs[0].Message)
	assert.Equal(t, logID2, logs[1].ID)
	assert.Equal(t, "Driver matched successfully", logs[1].Message)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTraditionalRepository_GetTraditionalLogsByTimeRange_Success(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	repo := postgres.NewTraditionalRepository(db)

	logID := uuid.New()
	now := time.Now()
	startTime := now.Add(-1 * time.Hour)
	endTime := now

	rows := sqlmock.NewRows([]string{
		"id", "level", "message", "service_name", "instance_id", "fields", "timestamp", "created_at",
	}).AddRow(
		logID, "error", "Failed to process ride request", "ride-service", "instance-1",
		[]byte(`{"error": "driver not found", "trip_id": "123"}`), now, now,
	)

	mock.ExpectQuery(`SELECT (.+) FROM traditional_logs WHERE timestamp >= \$1 AND timestamp <= \$2`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), 10, 0).
		WillReturnRows(rows)

	logs, err := repo.GetTraditionalLogsByTimeRange(context.Background(), startTime.Format(time.RFC3339), endTime.Format(time.RFC3339), 10, 0)

	assert.NoError(t, err)
	assert.Len(t, logs, 1)
	assert.Equal(t, logID, logs[0].ID)
	assert.Equal(t, models.LogLevel("error"), logs[0].Level)
	assert.Equal(t, "Failed to process ride request", logs[0].Message)
	assert.Equal(t, "ride-service", logs[0].ServiceName)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTraditionalRepository_CreateTraditionalLog_DatabaseError(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	repo := postgres.NewTraditionalRepository(db)

	logID := uuid.New()
	now := time.Now()

	log := &models.TraditionalLog{
		ID:          logID,
		Level:       "info",
		Message:     "Test log message",
		ServiceName: "test-service",
		InstanceID:  "instance-1",
		Fields:      []byte(`{"test": "data"}`),
		Timestamp:   now,
		CreatedAt:   now,
	}

	mock.ExpectExec(`INSERT INTO traditional_logs`).
		WithArgs(
			logID, "info", "Test log message", "test-service", "instance-1",
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnError(sql.ErrConnDone)

	err := repo.CreateTraditionalLog(context.Background(), log)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), sql.ErrConnDone.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTraditionalRepository_ListTraditionalLogs_EmptyResult(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	repo := postgres.NewTraditionalRepository(db)

	rows := sqlmock.NewRows([]string{
		"id", "level", "message", "service_name", "instance_id", "fields", "timestamp", "created_at",
	})

	mock.ExpectQuery(`SELECT (.+) FROM traditional_logs WHERE level = \$1 AND service_name = \$2`).
		WithArgs("debug", "non-existent-service", 10, 0).
		WillReturnRows(rows)

	logs, err := repo.ListTraditionalLogs(context.Background(), "debug", "non-existent-service", 10, 0)

	assert.NoError(t, err)
	assert.Len(t, logs, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}
