package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"actor-model-observability/internal/config"
	"actor-model-observability/internal/logging"

	"github.com/lib/pq"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// PostgresDB wraps sql.DB with additional functionality
type PostgresDB struct {
	*sql.DB
	config *config.DatabaseConfig
	logger *logging.Logger
}

// NewPostgresConnection creates a new PostgreSQL database connection
func NewPostgresConnection(cfg *config.DatabaseConfig, logger *logging.Logger) (*PostgresDB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	pgDB := &PostgresDB{
		DB:     db,
		config: cfg,
		logger: logger,
	}

	logger.WithComponent("database").Info("PostgreSQL connection established")

	return pgDB, nil
}

// Close closes the database connection
func (db *PostgresDB) Close() error {
	db.logger.WithComponent("database").Info("Closing PostgreSQL connection")
	return db.DB.Close()
}

// Ping checks if the database connection is alive
func (db *PostgresDB) Ping(ctx context.Context) error {
	start := time.Now()
	err := db.DB.PingContext(ctx)
	duration := time.Since(start).Milliseconds()

	db.logger.LogDatabaseOperation("ping", "", duration, err, nil)
	return err
}

// ExecContext executes a query with context and logging
func (db *PostgresDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := db.DB.ExecContext(ctx, query, args...)
	duration := time.Since(start).Milliseconds()

	db.logger.LogDatabaseOperation("exec", extractTableName(query), duration, err, logging.Fields{
		"query": query,
		"args":  args,
	})

	return result, err
}

// QueryContext executes a query with context and logging
func (db *PostgresDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := db.DB.QueryContext(ctx, query, args...)
	duration := time.Since(start).Milliseconds()

	db.logger.LogDatabaseOperation("query", extractTableName(query), duration, err, logging.Fields{
		"query": query,
		"args":  args,
	})

	return rows, err
}

// QueryRowContext executes a query that returns a single row with context and logging
func (db *PostgresDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := db.DB.QueryRowContext(ctx, query, args...)
	duration := time.Since(start).Milliseconds()

	db.logger.LogDatabaseOperation("query_row", extractTableName(query), duration, nil, logging.Fields{
		"query": query,
		"args":  args,
	})

	return row
}

// BeginTx starts a transaction with context and logging
func (db *PostgresDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	start := time.Now()
	tx, err := db.DB.BeginTx(ctx, opts)
	duration := time.Since(start).Milliseconds()

	db.logger.LogDatabaseOperation("begin_tx", "", duration, err, nil)

	return tx, err
}

// GetStats returns database connection statistics
func (db *PostgresDB) GetStats() sql.DBStats {
	return db.DB.Stats()
}

// HealthCheck performs a comprehensive health check
func (db *PostgresDB) HealthCheck(ctx context.Context) error {
	// Check basic connectivity
	if err := db.Ping(ctx); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	// Check if we can execute a simple query
	var result int
	err := db.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("simple query failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("unexpected query result: %d", result)
	}

	// Check connection pool stats
	stats := db.GetStats()
	if stats.OpenConnections == 0 {
		return fmt.Errorf("no open connections")
	}

	return nil
}

// IsConnectionError checks if an error is a connection-related error
func IsConnectionError(err error) bool {
	if err == nil {
		return false
	}

	// Check for PostgreSQL-specific connection errors
	if pqErr, ok := err.(*pq.Error); ok {
		// Connection errors typically have these codes
		switch pqErr.Code {
		case "08000", "08003", "08006", "08001", "08004":
			return true
		}
	}

	// Check for common connection error messages
	errorMsg := err.Error()
	connectionErrors := []string{
		"connection refused",
		"connection reset",
		"connection timeout",
		"no such host",
		"network is unreachable",
		"connection lost",
	}

	for _, connErr := range connectionErrors {
		if contains(errorMsg, connErr) {
			return true
		}
	}

	return false
}

// IsDuplicateKeyError checks if an error is a duplicate key constraint violation
func IsDuplicateKeyError(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505" // unique_violation
	}
	return false
}

// IsForeignKeyError checks if an error is a foreign key constraint violation
func IsForeignKeyError(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23503" // foreign_key_violation
	}
	return false
}

// IsNotNullError checks if an error is a not null constraint violation
func IsNotNullError(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23502" // not_null_violation
	}
	return false
}

// Helper functions

func extractTableName(query string) string {
	// Simple table name extraction - could be improved with proper SQL parsing
	// This is a basic implementation for logging purposes
	if len(query) < 10 {
		return "unknown"
	}

	// Convert to lowercase for easier matching
	q := query
	if len(q) > 100 {
		q = q[:100] // Limit length for performance
	}

	// Look for common SQL patterns
	patterns := map[string][]string{
		"users":                {"FROM users", "INTO users", "UPDATE users", "DELETE FROM users"},
		"drivers":              {"FROM drivers", "INTO drivers", "UPDATE drivers", "DELETE FROM drivers"},
		"passengers":           {"FROM passengers", "INTO passengers", "UPDATE passengers", "DELETE FROM passengers"},
		"trips":                {"FROM trips", "INTO trips", "UPDATE trips", "DELETE FROM trips"},
		"actor_instances":      {"FROM actor_instances", "INTO actor_instances", "UPDATE actor_instances", "DELETE FROM actor_instances"},
		"actor_messages":       {"FROM actor_messages", "INTO actor_messages", "UPDATE actor_messages", "DELETE FROM actor_messages"},
		"system_metrics":       {"FROM system_metrics", "INTO system_metrics", "UPDATE system_metrics", "DELETE FROM system_metrics"},
		"distributed_traces":   {"FROM distributed_traces", "INTO distributed_traces", "UPDATE distributed_traces", "DELETE FROM distributed_traces"},
		"event_logs":           {"FROM event_logs", "INTO event_logs", "UPDATE event_logs", "DELETE FROM event_logs"},
		"traditional_metrics": {"FROM traditional_metrics", "INTO traditional_metrics", "UPDATE traditional_metrics", "DELETE FROM traditional_metrics"},
		"traditional_logs":     {"FROM traditional_logs", "INTO traditional_logs", "UPDATE traditional_logs", "DELETE FROM traditional_logs"},
		"service_health":       {"FROM service_health", "INTO service_health", "UPDATE service_health", "DELETE FROM service_health"},
	}

	for table, tablePatterns := range patterns {
		for _, pattern := range tablePatterns {
			if contains(q, pattern) {
				return table
			}
		}
	}

	return "unknown"
}

func contains(s, substr string) bool {
	// Simple case-insensitive contains check
	s = toLower(s)
	substr = toLower(substr)
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i, b := range []byte(s) {
		if b >= 'A' && b <= 'Z' {
			result[i] = b + 32
		} else {
			result[i] = b
		}
	}
	return string(result)
}

func findSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}

	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}