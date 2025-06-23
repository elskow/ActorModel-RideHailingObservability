package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"actor-model-observability/internal/config"
	"actor-model-observability/internal/logging"

	"github.com/go-redis/redis/v8"
)

// RedisClient wraps redis.Client with additional functionality
type RedisClient struct {
	*redis.Client
	config *config.RedisConfig
	logger *logging.Logger
}

// NewRedisConnection creates a new Redis client connection
func NewRedisConnection(cfg *config.RedisConfig, logger *logging.Logger) (*RedisClient, error) {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		rdb.Close()
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	redisClient := &RedisClient{
		Client: rdb,
		config: cfg,
		logger: logger,
	}

	logger.WithComponent("redis").Info("Redis connection established")

	return redisClient, nil
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	r.logger.WithComponent("redis").Info("Closing Redis connection")
	return r.Client.Close()
}

// Ping checks if the Redis connection is alive
func (r *RedisClient) Ping(ctx context.Context) error {
	start := time.Now()
	err := r.Client.Ping(ctx).Err()
	duration := time.Since(start).Milliseconds()

	r.logger.LogDatabaseOperation("ping", "redis", duration, err, nil)
	return err
}

// Set sets a key-value pair with optional expiration
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	start := time.Now()
	err := r.Client.Set(ctx, key, value, expiration).Err()
	duration := time.Since(start).Milliseconds()

	r.logger.LogDatabaseOperation("set", "redis", duration, err, logging.Fields{
		"key":        key,
		"expiration": expiration,
	})

	return err
}

// Get retrieves a value by key
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	start := time.Now()
	result, err := r.Client.Get(ctx, key).Result()
	duration := time.Since(start).Milliseconds()

	r.logger.LogDatabaseOperation("get", "redis", duration, err, logging.Fields{
		"key": key,
	})

	return result, err
}

// Del deletes one or more keys
func (r *RedisClient) Del(ctx context.Context, keys ...string) error {
	start := time.Now()
	err := r.Client.Del(ctx, keys...).Err()
	duration := time.Since(start).Milliseconds()

	r.logger.LogDatabaseOperation("del", "redis", duration, err, logging.Fields{
		"keys": keys,
	})

	return err
}

// Exists checks if keys exist
func (r *RedisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	start := time.Now()
	result, err := r.Client.Exists(ctx, keys...).Result()
	duration := time.Since(start).Milliseconds()

	r.logger.LogDatabaseOperation("exists", "redis", duration, err, logging.Fields{
		"keys": keys,
	})

	return result, err
}

// Expire sets expiration for a key
func (r *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	start := time.Now()
	err := r.Client.Expire(ctx, key, expiration).Err()
	duration := time.Since(start).Milliseconds()

	r.logger.LogDatabaseOperation("expire", "redis", duration, err, logging.Fields{
		"key":        key,
		"expiration": expiration,
	})

	return err
}

// HSet sets field in hash
func (r *RedisClient) HSet(ctx context.Context, key string, values ...interface{}) error {
	start := time.Now()
	err := r.Client.HSet(ctx, key, values...).Err()
	duration := time.Since(start).Milliseconds()

	r.logger.LogDatabaseOperation("hset", "redis", duration, err, logging.Fields{
		"key":    key,
		"fields": len(values) / 2,
	})

	return err
}

// HGet gets field from hash
func (r *RedisClient) HGet(ctx context.Context, key, field string) (string, error) {
	start := time.Now()
	result, err := r.Client.HGet(ctx, key, field).Result()
	duration := time.Since(start).Milliseconds()

	r.logger.LogDatabaseOperation("hget", "redis", duration, err, logging.Fields{
		"key":   key,
		"field": field,
	})

	return result, err
}

// HGetAll gets all fields from hash
func (r *RedisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	start := time.Now()
	result, err := r.Client.HGetAll(ctx, key).Result()
	duration := time.Since(start).Milliseconds()

	r.logger.LogDatabaseOperation("hgetall", "redis", duration, err, logging.Fields{
		"key": key,
	})

	return result, err
}

// HDel deletes fields from hash
func (r *RedisClient) HDel(ctx context.Context, key string, fields ...string) error {
	start := time.Now()
	err := r.Client.HDel(ctx, key, fields...).Err()
	duration := time.Since(start).Milliseconds()

	r.logger.LogDatabaseOperation("hdel", "redis", duration, err, logging.Fields{
		"key":    key,
		"fields": fields,
	})

	return err
}

// LPush pushes elements to the head of list
func (r *RedisClient) LPush(ctx context.Context, key string, values ...interface{}) error {
	start := time.Now()
	err := r.Client.LPush(ctx, key, values...).Err()
	duration := time.Since(start).Milliseconds()

	r.logger.LogDatabaseOperation("lpush", "redis", duration, err, logging.Fields{
		"key":   key,
		"count": len(values),
	})

	return err
}

// RPop removes and returns the last element of list
func (r *RedisClient) RPop(ctx context.Context, key string) (string, error) {
	start := time.Now()
	result, err := r.Client.RPop(ctx, key).Result()
	duration := time.Since(start).Milliseconds()

	r.logger.LogDatabaseOperation("rpop", "redis", duration, err, logging.Fields{
		"key": key,
	})

	return result, err
}

// LLen returns the length of list
func (r *RedisClient) LLen(ctx context.Context, key string) (int64, error) {
	start := time.Now()
	result, err := r.Client.LLen(ctx, key).Result()
	duration := time.Since(start).Milliseconds()

	r.logger.LogDatabaseOperation("llen", "redis", duration, err, logging.Fields{
		"key": key,
	})

	return result, err
}

// Incr increments the integer value of key
func (r *RedisClient) Incr(ctx context.Context, key string) (int64, error) {
	start := time.Now()
	result, err := r.Client.Incr(ctx, key).Result()
	duration := time.Since(start).Milliseconds()

	r.logger.LogDatabaseOperation("incr", "redis", duration, err, logging.Fields{
		"key": key,
	})

	return result, err
}

// IncrBy increments the integer value of key by increment
func (r *RedisClient) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	start := time.Now()
	result, err := r.Client.IncrBy(ctx, key, value).Result()
	duration := time.Since(start).Milliseconds()

	r.logger.LogDatabaseOperation("incrby", "redis", duration, err, logging.Fields{
		"key":   key,
		"value": value,
	})

	return result, err
}

// SetJSON sets a JSON-encoded value
func (r *RedisClient) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return r.Set(ctx, key, data, expiration)
}

// GetJSON gets and JSON-decodes a value
func (r *RedisClient) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := r.Get(ctx, key)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}

// SetMetric stores a metric value with timestamp
func (r *RedisClient) SetMetric(ctx context.Context, metricName string, value interface{}, timestamp time.Time) error {
	key := fmt.Sprintf("metric:%s:%d", metricName, timestamp.Unix())
	return r.SetJSON(ctx, key, value, 24*time.Hour) // Keep metrics for 24 hours
}

// GetMetrics retrieves metrics within a time range
func (r *RedisClient) GetMetrics(ctx context.Context, metricName string, startTime, endTime time.Time) ([]interface{}, error) {
	pattern := fmt.Sprintf("metric:%s:*", metricName)
	keys, err := r.Client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	var metrics []interface{}
	for _, key := range keys {
		var metric interface{}
		if err := r.GetJSON(ctx, key, &metric); err != nil {
			continue // Skip invalid metrics
		}
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

// SetActorState stores actor state
func (r *RedisClient) SetActorState(ctx context.Context, actorID string, state interface{}) error {
	key := fmt.Sprintf("actor:state:%s", actorID)
	return r.SetJSON(ctx, key, state, time.Hour) // Keep actor state for 1 hour
}

// GetActorState retrieves actor state
func (r *RedisClient) GetActorState(ctx context.Context, actorID string, dest interface{}) error {
	key := fmt.Sprintf("actor:state:%s", actorID)
	return r.GetJSON(ctx, key, dest)
}

// DeleteActorState removes actor state
func (r *RedisClient) DeleteActorState(ctx context.Context, actorID string) error {
	key := fmt.Sprintf("actor:state:%s", actorID)
	return r.Del(ctx, key)
}

// IncrementCounter increments a counter metric
func (r *RedisClient) IncrementCounter(ctx context.Context, counterName string) error {
	key := fmt.Sprintf("counter:%s", counterName)
	_, err := r.Incr(ctx, key)
	return err
}

// GetCounter gets counter value
func (r *RedisClient) GetCounter(ctx context.Context, counterName string) (int64, error) {
	key := fmt.Sprintf("counter:%s", counterName)
	result, err := r.Get(ctx, key)
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	var count int64
	if err := json.Unmarshal([]byte(result), &count); err != nil {
		return 0, err
	}

	return count, nil
}

// HealthCheck performs a comprehensive health check
func (r *RedisClient) HealthCheck(ctx context.Context) error {
	// Check basic connectivity
	if err := r.Ping(ctx); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	// Test set and get operations
	testKey := "health_check_test"
	testValue := "test_value"

	if err := r.Set(ctx, testKey, testValue, time.Minute); err != nil {
		return fmt.Errorf("set operation failed: %w", err)
	}

	result, err := r.Get(ctx, testKey)
	if err != nil {
		return fmt.Errorf("get operation failed: %w", err)
	}

	if result != testValue {
		return fmt.Errorf("unexpected value: got %s, want %s", result, testValue)
	}

	// Clean up test key
	if err := r.Del(ctx, testKey); err != nil {
		r.logger.WithComponent("redis").WithError(err).Warn("Failed to clean up test key")
	}

	return nil
}

// GetPoolStats returns Redis connection pool statistics
func (r *RedisClient) GetPoolStats() *redis.PoolStats {
	return r.Client.PoolStats()
}

// IsConnectionError checks if an error is a connection-related error
func IsRedisConnectionError(err error) bool {
	if err == nil {
		return false
	}

	// Check for common Redis connection error messages
	errorMsg := err.Error()
	connectionErrors := []string{
		"connection refused",
		"connection reset",
		"connection timeout",
		"no such host",
		"network is unreachable",
		"connection lost",
		"EOF",
		"broken pipe",
	}

	for _, connErr := range connectionErrors {
		if contains(errorMsg, connErr) {
			return true
		}
	}

	return false
}