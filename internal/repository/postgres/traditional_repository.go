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

// TraditionalRepositoryImpl implements the TraditionalRepository interface using PostgreSQL
type TraditionalRepositoryImpl struct {
	db *sqlx.DB
}

// NewTraditionalRepository creates a new instance of TraditionalRepositoryImpl
func NewTraditionalRepository(db *sqlx.DB) repository.TraditionalRepository {
	return &TraditionalRepositoryImpl{db: db}
}

// Traditional Metrics methods

// CreateTraditionalMetric creates a new traditional metric record
func (r *TraditionalRepositoryImpl) CreateTraditionalMetric(ctx context.Context, metric *models.TraditionalMetric) error {
	query := `
		INSERT INTO traditional_metrics (id, metric_name, metric_type, metric_value, labels, 
			service_name, instance_id, timestamp, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(ctx, query,
		metric.ID,
		metric.MetricName,
		metric.MetricType,
		metric.MetricValue,
		metric.Labels,
		metric.ServiceName,
		metric.InstanceID,
		metric.Timestamp,
		metric.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create traditional metric: %w", err)
	}

	return nil
}

// GetTraditionalMetric retrieves a traditional metric by ID
func (r *TraditionalRepositoryImpl) GetTraditionalMetric(ctx context.Context, id string) (*models.TraditionalMetric, error) {
	query := `
		SELECT id, metric_name, metric_type, metric_value, labels, service_name, 
			instance_id, timestamp, created_at
		FROM traditional_metrics
		WHERE id = $1
	`

	metric := &models.TraditionalMetric{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&metric.ID,
		&metric.MetricName,
		&metric.MetricType,
		&metric.MetricValue,
		&metric.Labels,
		&metric.ServiceName,
		&metric.InstanceID,
		&metric.Timestamp,
		&metric.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &models.NotFoundError{
				Resource: "traditional_metric",
				ID:       id,
			}
		}
		return nil, fmt.Errorf("failed to get traditional metric: %w", err)
	}

	return metric, nil
}

// ListTraditionalMetrics retrieves traditional metrics with optional filtering
func (r *TraditionalRepositoryImpl) ListTraditionalMetrics(ctx context.Context, metricType, serviceName string, limit, offset int) ([]*models.TraditionalMetric, error) {
	var query string
	var args []interface{}

	if metricType != "" && serviceName != "" {
		query = `
			SELECT id, metric_name, metric_type, metric_value, labels, service_name, 
				instance_id, timestamp, created_at
			FROM traditional_metrics
			WHERE metric_type = $1 AND service_name = $2
			ORDER BY timestamp DESC
			LIMIT $3 OFFSET $4
		`
		args = []interface{}{metricType, serviceName, limit, offset}
	} else if metricType != "" {
		query = `
			SELECT id, metric_name, metric_type, metric_value, labels, service_name, 
				instance_id, timestamp, created_at
			FROM traditional_metrics
			WHERE metric_type = $1
			ORDER BY timestamp DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{metricType, limit, offset}
	} else if serviceName != "" {
		query = `
			SELECT id, metric_name, metric_type, metric_value, labels, service_name, 
				instance_id, timestamp, created_at
			FROM traditional_metrics
			WHERE service_name = $1
			ORDER BY timestamp DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{serviceName, limit, offset}
	} else {
		query = `
			SELECT id, metric_name, metric_type, metric_value, labels, service_name, 
				instance_id, timestamp, created_at
			FROM traditional_metrics
			ORDER BY timestamp DESC
			LIMIT $1 OFFSET $2
		`
		args = []interface{}{limit, offset}
	}

	return r.scanTraditionalMetrics(ctx, query, args...)
}

// GetTraditionalMetricsByTimeRange retrieves traditional metrics within a time range
func (r *TraditionalRepositoryImpl) GetTraditionalMetricsByTimeRange(ctx context.Context, startTime, endTime string, limit, offset int) ([]*models.TraditionalMetric, error) {
	startTimeParsed, err := time.Parse(time.RFC3339, startTime)
	if err != nil {
		return nil, fmt.Errorf("invalid start time format: %w", err)
	}

	endTimeParsed, err := time.Parse(time.RFC3339, endTime)
	if err != nil {
		return nil, fmt.Errorf("invalid end time format: %w", err)
	}

	query := `
		SELECT id, metric_name, metric_type, metric_value, labels, service_name, 
			instance_id, timestamp, created_at
		FROM traditional_metrics
		WHERE timestamp >= $1 AND timestamp <= $2
		ORDER BY timestamp DESC
		LIMIT $3 OFFSET $4
	`

	return r.scanTraditionalMetrics(ctx, query, startTimeParsed, endTimeParsed, limit, offset)
}

// Traditional Logs methods

// CreateTraditionalLog creates a new traditional log record
func (r *TraditionalRepositoryImpl) CreateTraditionalLog(ctx context.Context, log *models.TraditionalLog) error {
	query := `
		INSERT INTO traditional_logs (id, level, message, service_name, instance_id, 
			fields, timestamp, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		log.ID,
		log.Level,
		log.Message,
		log.ServiceName,
		log.InstanceID,
		log.Fields,
		log.Timestamp,
		log.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create traditional log: %w", err)
	}

	return nil
}

// GetTraditionalLog retrieves a traditional log by ID
func (r *TraditionalRepositoryImpl) GetTraditionalLog(ctx context.Context, id string) (*models.TraditionalLog, error) {
	query := `
		SELECT id, level, message, service_name, instance_id, fields, 
			timestamp, created_at
		FROM traditional_logs
		WHERE id = $1
	`

	log := &models.TraditionalLog{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&log.ID,
		&log.Level,
		&log.Message,
		&log.ServiceName,
		&log.InstanceID,
		&log.Fields,
		&log.Timestamp,
		&log.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &models.NotFoundError{
				Resource: "traditional_log",
				ID:       id,
			}
		}
		return nil, fmt.Errorf("failed to get traditional log: %w", err)
	}

	return log, nil
}

// ListTraditionalLogs retrieves traditional logs with optional filtering
func (r *TraditionalRepositoryImpl) ListTraditionalLogs(ctx context.Context, level, serviceName string, limit, offset int) ([]*models.TraditionalLog, error) {
	var query string
	var args []interface{}

	if level != "" && serviceName != "" {
		query = `
			SELECT id, level, message, service_name, instance_id, fields, 
				timestamp, created_at
			FROM traditional_logs
			WHERE level = $1 AND service_name = $2
			ORDER BY timestamp DESC
			LIMIT $3 OFFSET $4
		`
		args = []interface{}{level, serviceName, limit, offset}
	} else if level != "" {
		query = `
			SELECT id, level, message, service_name, instance_id, fields, 
				timestamp, created_at
			FROM traditional_logs
			WHERE level = $1
			ORDER BY timestamp DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{level, limit, offset}
	} else if serviceName != "" {
		query = `
			SELECT id, level, message, service_name, instance_id, fields, 
				timestamp, created_at
			FROM traditional_logs
			WHERE service_name = $1
			ORDER BY timestamp DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{serviceName, limit, offset}
	} else {
		query = `
			SELECT id, level, message, service_name, instance_id, fields, 
				timestamp, created_at
			FROM traditional_logs
			ORDER BY timestamp DESC
			LIMIT $1 OFFSET $2
		`
		args = []interface{}{limit, offset}
	}

	return r.scanTraditionalLogs(ctx, query, args...)
}

// GetTraditionalLogsByTimeRange retrieves traditional logs within a time range
func (r *TraditionalRepositoryImpl) GetTraditionalLogsByTimeRange(ctx context.Context, startTime, endTime string, limit, offset int) ([]*models.TraditionalLog, error) {
	startTimeParsed, err := time.Parse(time.RFC3339, startTime)
	if err != nil {
		return nil, fmt.Errorf("invalid start time format: %w", err)
	}

	endTimeParsed, err := time.Parse(time.RFC3339, endTime)
	if err != nil {
		return nil, fmt.Errorf("invalid end time format: %w", err)
	}

	query := `
		SELECT id, level, message, service_name, instance_id, fields, 
			timestamp, created_at
		FROM traditional_logs
		WHERE timestamp >= $1 AND timestamp <= $2
		ORDER BY timestamp DESC
		LIMIT $3 OFFSET $4
	`

	return r.scanTraditionalLogs(ctx, query, startTimeParsed, endTimeParsed, limit, offset)
}

// Service Health methods

// CreateServiceHealth creates a new service health record
func (r *TraditionalRepositoryImpl) CreateServiceHealth(ctx context.Context, health *models.ServiceHealth) error {
	query := `
		INSERT INTO service_health (id, service_name, status, response_time_ms, 
			cpu_usage_percent, memory_usage_percent, disk_usage_percent, 
			active_connections, error_rate_percent, last_error, timestamp, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.db.ExecContext(ctx, query,
		health.ID,
		health.ServiceName,
		health.Status,
		health.ResponseTimeMs,
		health.CPUUsagePercent,
		health.MemoryUsagePercent,
		health.DiskUsagePercent,
		health.ActiveConnections,
		health.ErrorRatePercent,
		health.LastError,
		health.Timestamp,
		health.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create service health: %w", err)
	}

	return nil
}

// GetServiceHealth retrieves a service health record by ID
func (r *TraditionalRepositoryImpl) GetServiceHealth(ctx context.Context, id string) (*models.ServiceHealth, error) {
	query := `
		SELECT id, service_name, status, response_time_ms, cpu_usage_percent, 
			memory_usage_percent, disk_usage_percent, active_connections, 
			error_rate_percent, last_error, timestamp, created_at
		FROM service_health
		WHERE id = $1
	`

	health := &models.ServiceHealth{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&health.ID,
		&health.ServiceName,
		&health.Status,
		&health.ResponseTimeMs,
		&health.CPUUsagePercent,
		&health.MemoryUsagePercent,
		&health.DiskUsagePercent,
		&health.ActiveConnections,
		&health.ErrorRatePercent,
		&health.LastError,
		&health.Timestamp,
		&health.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &models.NotFoundError{
				Resource: "service_health",
				ID:       id,
			}
		}
		return nil, fmt.Errorf("failed to get service health: %w", err)
	}

	return health, nil
}

// GetLatestServiceHealth retrieves the latest health status for a service
func (r *TraditionalRepositoryImpl) GetLatestServiceHealth(ctx context.Context, serviceName string) (*models.ServiceHealth, error) {
	query := `
		SELECT id, service_name, status, response_time_ms, cpu_usage_percent, 
			memory_usage_percent, disk_usage_percent, active_connections, 
			error_rate_percent, last_error, timestamp, created_at
		FROM service_health
		WHERE service_name = $1
		ORDER BY timestamp DESC
		LIMIT 1
	`

	health := &models.ServiceHealth{}
	err := r.db.QueryRowContext(ctx, query, serviceName).Scan(
		&health.ID,
		&health.ServiceName,
		&health.Status,
		&health.ResponseTimeMs,
		&health.CPUUsagePercent,
		&health.MemoryUsagePercent,
		&health.DiskUsagePercent,
		&health.ActiveConnections,
		&health.ErrorRatePercent,
		&health.LastError,
		&health.Timestamp,
		&health.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &models.NotFoundError{
				Resource: "service_health",
				ID:       serviceName,
			}
		}
		return nil, fmt.Errorf("failed to get latest service health: %w", err)
	}

	return health, nil
}

// ListServiceHealth retrieves service health records with optional filtering
func (r *TraditionalRepositoryImpl) ListServiceHealth(ctx context.Context, serviceName string, limit, offset int) ([]*models.ServiceHealth, error) {
	var query string
	var args []interface{}

	if serviceName != "" {
		query = `
			SELECT id, service_name, status, response_time_ms, cpu_usage_percent, 
				memory_usage_percent, disk_usage_percent, active_connections, 
				error_rate_percent, last_error, timestamp, created_at
			FROM service_health
			WHERE service_name = $1
			ORDER BY timestamp DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{serviceName, limit, offset}
	} else {
		query = `
			SELECT id, service_name, status, response_time_ms, cpu_usage_percent, 
				memory_usage_percent, disk_usage_percent, active_connections, 
				error_rate_percent, last_error, timestamp, created_at
			FROM service_health
			ORDER BY timestamp DESC
			LIMIT $1 OFFSET $2
		`
		args = []interface{}{limit, offset}
	}

	return r.scanServiceHealth(ctx, query, args...)
}

// GetServiceHealthByTimeRange retrieves service health records within a time range
func (r *TraditionalRepositoryImpl) GetServiceHealthByTimeRange(ctx context.Context, serviceName, startTime, endTime string, limit, offset int) ([]*models.ServiceHealth, error) {
	startTimeParsed, err := time.Parse(time.RFC3339, startTime)
	if err != nil {
		return nil, fmt.Errorf("invalid start time format: %w", err)
	}

	endTimeParsed, err := time.Parse(time.RFC3339, endTime)
	if err != nil {
		return nil, fmt.Errorf("invalid end time format: %w", err)
	}

	var query string
	var args []interface{}

	if serviceName != "" {
		query = `
			SELECT id, service_name, status, response_time_ms, cpu_usage_percent, 
				memory_usage_percent, disk_usage_percent, active_connections, 
				error_rate_percent, last_error, timestamp, created_at
			FROM service_health
			WHERE service_name = $1 AND timestamp >= $2 AND timestamp <= $3
			ORDER BY timestamp DESC
			LIMIT $4 OFFSET $5
		`
		args = []interface{}{serviceName, startTimeParsed, endTimeParsed, limit, offset}
	} else {
		query = `
			SELECT id, service_name, status, response_time_ms, cpu_usage_percent, 
				memory_usage_percent, disk_usage_percent, active_connections, 
				error_rate_percent, last_error, timestamp, created_at
			FROM service_health
			WHERE timestamp >= $1 AND timestamp <= $2
			ORDER BY timestamp DESC
			LIMIT $3 OFFSET $4
		`
		args = []interface{}{startTimeParsed, endTimeParsed, limit, offset}
	}

	return r.scanServiceHealth(ctx, query, args...)
}

// Helper methods for scanning results

func (r *TraditionalRepositoryImpl) scanTraditionalMetrics(ctx context.Context, query string, args ...interface{}) ([]*models.TraditionalMetric, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var metrics []*models.TraditionalMetric
	for rows.Next() {
		metric := &models.TraditionalMetric{}
		err := rows.Scan(
			&metric.ID,
			&metric.MetricName,
			&metric.MetricType,
			&metric.MetricValue,
			&metric.Labels,
			&metric.ServiceName,
			&metric.InstanceID,
			&metric.Timestamp,
			&metric.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan traditional metric: %w", err)
		}
		metrics = append(metrics, metric)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating traditional metrics: %w", err)
	}

	return metrics, nil
}

func (r *TraditionalRepositoryImpl) scanTraditionalLogs(ctx context.Context, query string, args ...interface{}) ([]*models.TraditionalLog, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var logs []*models.TraditionalLog
	for rows.Next() {
		log := &models.TraditionalLog{}
		err := rows.Scan(
			&log.ID,
			&log.Level,
			&log.Message,
			&log.ServiceName,
			&log.InstanceID,
			&log.Fields,
			&log.Timestamp,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan traditional log: %w", err)
		}
		logs = append(logs, log)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating traditional logs: %w", err)
	}

	return logs, nil
}

func (r *TraditionalRepositoryImpl) scanServiceHealth(ctx context.Context, query string, args ...interface{}) ([]*models.ServiceHealth, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var healthRecords []*models.ServiceHealth
	for rows.Next() {
		health := &models.ServiceHealth{}
		err := rows.Scan(
			&health.ID,
			&health.ServiceName,
			&health.Status,
			&health.ResponseTimeMs,
			&health.CPUUsagePercent,
			&health.MemoryUsagePercent,
			&health.DiskUsagePercent,
			&health.ActiveConnections,
			&health.ErrorRatePercent,
			&health.LastError,
			&health.Timestamp,
			&health.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan service health: %w", err)
		}
		healthRecords = append(healthRecords, health)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating service health records: %w", err)
	}

	return healthRecords, nil
}
