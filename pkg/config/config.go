package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Database DatabaseConfig
	Redis    RedisConfig
	Server   ServerConfig
	Actor    ActorConfig
	Observability ObservabilityConfig
	LoadTest LoadTestConfig
	System   SystemConfig
	Logging  LoggingConfig
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port int
	Host string
	Mode string
}

// ActorConfig holds actor system configuration
type ActorConfig struct {
	PoolSize           int
	MessageBufferSize  int
	HeartbeatInterval  time.Duration
	CleanupInterval    time.Duration
}

// ObservabilityConfig holds observability configuration
type ObservabilityConfig struct {
	TracingEnabled    bool
	TracingEndpoint   string
	TracingServiceName string
	MetricsEnabled    bool
	MetricsInterval   time.Duration
}

// LoadTestConfig holds load testing configuration
type LoadTestConfig struct {
	ConcurrentUsers int
	Duration        time.Duration
	RampUp          time.Duration
}

// SystemConfig holds system-wide configuration
type SystemConfig struct {
	Mode string // "actor" or "traditional"
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string
	Format string
	Output string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			Name:     getEnv("DB_NAME", "actor_observability"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Server: ServerConfig{
			Port: getEnvAsInt("SERVER_PORT", 8080),
			Host: getEnv("SERVER_HOST", "localhost"),
			Mode: getEnv("SERVER_MODE", "development"),
		},
		Actor: ActorConfig{
			PoolSize:          getEnvAsInt("ACTOR_POOL_SIZE", 100),
			MessageBufferSize: getEnvAsInt("ACTOR_MESSAGE_BUFFER_SIZE", 1000),
			HeartbeatInterval: getEnvAsDuration("ACTOR_HEARTBEAT_INTERVAL", 30*time.Second),
			CleanupInterval:   getEnvAsDuration("ACTOR_CLEANUP_INTERVAL", 5*time.Minute),
		},
		Observability: ObservabilityConfig{
			TracingEnabled:     getEnvAsBool("TRACING_ENABLED", true),
			TracingEndpoint:    getEnv("TRACING_ENDPOINT", "http://localhost:14268/api/traces"),
			TracingServiceName: getEnv("TRACING_SERVICE_NAME", "actor-observability"),
			MetricsEnabled:     getEnvAsBool("METRICS_ENABLED", true),
			MetricsInterval:    getEnvAsDuration("METRICS_INTERVAL", 10*time.Second),
		},
		LoadTest: LoadTestConfig{
			ConcurrentUsers: getEnvAsInt("LOAD_TEST_CONCURRENT_USERS", 100),
			Duration:        getEnvAsDuration("LOAD_TEST_DURATION", 5*time.Minute),
			RampUp:          getEnvAsDuration("LOAD_TEST_RAMP_UP", 30*time.Second),
		},
		System: SystemConfig{
			Mode: getEnv("SYSTEM_MODE", "actor"),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
			Output: getEnv("LOG_OUTPUT", "stdout"),
		},
	}, nil
}

// Helper functions for environment variable parsing

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// GetDatabaseDSN returns the database connection string
func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

// GetRedisAddr returns the Redis connection address
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}

// GetServerAddr returns the server address
func (c *Config) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// IsActorMode returns true if the system is running in actor mode
func (c *Config) IsActorMode() bool {
	return c.System.Mode == "actor"
}

// IsTraditionalMode returns true if the system is running in traditional mode
func (c *Config) IsTraditionalMode() bool {
	return c.System.Mode == "traditional"
}