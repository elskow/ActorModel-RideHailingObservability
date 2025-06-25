package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Server        ServerConfig
	Database      DatabaseConfig
	Redis         RedisConfig
	Actor         ActorConfig
	Logging       LoggingConfig
	Observability ObservabilityConfig
	Metrics       MetricsConfig
	OpenTelemetry OpenTelemetryConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port         string
	Host         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	Mode         string // gin mode: debug, release, test
}

// DatabaseConfig holds PostgreSQL database configuration
type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host         string
	Port         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// ActorConfig holds actor system configuration
type ActorConfig struct {
	MaxActors           int
	SupervisionStrategy string // restart, stop, ignore
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level          string // debug, info, warn, error
	Format         string // json, text
	Output         string // stdout, file
	FilePath       string
	MaxSize        int // megabytes
	MaxBackups     int
	MaxAge         int // days
	Compress       bool
	SkipPaths      []string // HTTP paths to skip logging
	SkipUserAgents []string // User agents to skip logging
}

// ObservabilityConfig holds observability configuration
type ObservabilityConfig struct {
	MetricsInterval time.Duration
}

// MetricsConfig holds metrics collection configuration
type MetricsConfig struct {
	CollectInterval time.Duration
	FlushInterval   time.Duration
	RetentionPeriod time.Duration
	BatchSize       int
}

// OpenTelemetryConfig holds OpenTelemetry configuration
type OpenTelemetryConfig struct {
	ServiceName        string
	ServiceVersion     string
	Environment        string
	MetricsEnabled     bool
	TracingEnabled     bool
	MetricsExporter    string // prometheus, otlp
	TracingExporter    string // jaeger, otlp
	JaegerEndpoint     string
	OTLPEndpoint       string
	SampleRate         float64
	MetricsInterval    time.Duration
	ResourceAttributes map[string]string
}

// Load loads configuration from environment variables with defaults
func Load() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getDurationEnv("SERVER_IDLE_TIMEOUT", 60*time.Second),
			Mode:         getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "postgres"),
			DBName:          getEnv("DB_NAME", "actor_observability"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    getIntEnv("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getIntEnv("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getDurationEnv("DB_CONN_MAX_LIFETIME", 5*time.Minute),
			ConnMaxIdleTime: getDurationEnv("DB_CONN_MAX_IDLE_TIME", 5*time.Minute),
		},
		Redis: RedisConfig{
			Host:         getEnv("REDIS_HOST", "localhost"),
			Port:         getEnv("REDIS_PORT", "6379"),
			Password:     getEnv("REDIS_PASSWORD", ""),
			DB:           getIntEnv("REDIS_DB", 0),
			PoolSize:     getIntEnv("REDIS_POOL_SIZE", 10),
			MinIdleConns: getIntEnv("REDIS_MIN_IDLE_CONNS", 2),
			DialTimeout:  getDurationEnv("REDIS_DIAL_TIMEOUT", 5*time.Second),
			ReadTimeout:  getDurationEnv("REDIS_READ_TIMEOUT", 3*time.Second),
			WriteTimeout: getDurationEnv("REDIS_WRITE_TIMEOUT", 3*time.Second),
		},
		Actor: ActorConfig{
			MaxActors:           getIntEnv("ACTOR_MAX_ACTORS", 10000),
			SupervisionStrategy: getEnv("ACTOR_SUPERVISION_STRATEGY", "restart"),
		},
		Logging: LoggingConfig{
			Level:          getEnv("LOG_LEVEL", "info"),
			Format:         getEnv("LOG_FORMAT", "json"),
			Output:         getEnv("LOG_OUTPUT", "stdout"),
			FilePath:       getEnv("LOG_FILE_PATH", "./logs/app.log"),
			MaxSize:        getIntEnv("LOG_MAX_SIZE", 100),
			MaxBackups:     getIntEnv("LOG_MAX_BACKUPS", 3),
			MaxAge:         getIntEnv("LOG_MAX_AGE", 28),
			Compress:       getBoolEnv("LOG_COMPRESS", true),
			SkipPaths:      getStringSliceEnv("LOG_SKIP_PATHS", []string{"/metrics", "/health", "/prometheus"}),
			SkipUserAgents: getStringSliceEnv("LOG_SKIP_USER_AGENTS", []string{"Prometheus", "kube-probe"}),
		},
		Observability: ObservabilityConfig{
			MetricsInterval: getDurationEnv("METRICS_INTERVAL", 10*time.Second),
		},
		Metrics: MetricsConfig{
			CollectInterval: getDurationEnv("METRICS_COLLECT_INTERVAL", 30*time.Second),
			FlushInterval:   getDurationEnv("METRICS_FLUSH_INTERVAL", 5*time.Minute),
			RetentionPeriod: getDurationEnv("METRICS_RETENTION_PERIOD", 7*24*time.Hour),
			BatchSize:       getIntEnv("METRICS_BATCH_SIZE", 100),
		},
		OpenTelemetry: OpenTelemetryConfig{
			ServiceName:        getEnv("OTEL_SERVICE_NAME", "actor-model-observability"),
			ServiceVersion:     getEnv("OTEL_SERVICE_VERSION", "1.0.0"),
			Environment:        getEnv("OTEL_ENVIRONMENT", "development"),
			MetricsEnabled:     getBoolEnv("OTEL_METRICS_ENABLED", true),
			TracingEnabled:     getBoolEnv("OTEL_TRACING_ENABLED", true),
			MetricsExporter:    getEnv("OTEL_METRICS_EXPORTER", "prometheus"),
			TracingExporter:    getEnv("OTEL_TRACING_EXPORTER", "otlp"),
			JaegerEndpoint:     getEnv("OTEL_JAEGER_ENDPOINT", "http://localhost:14268/api/traces"),
			OTLPEndpoint:       getEnv("OTEL_OTLP_ENDPOINT", "localhost:4318"),
			SampleRate:         getFloatEnv("OTEL_SAMPLE_RATE", 1.0),
			MetricsInterval:    getDurationEnv("OTEL_METRICS_INTERVAL", 10*time.Second),
			ResourceAttributes: getMapEnv("OTEL_RESOURCE_ATTRIBUTES"),
		},
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate server config
	if c.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}
	if c.Server.Host == "" {
		return fmt.Errorf("server host is required")
	}
	if c.Server.Mode != "debug" && c.Server.Mode != "release" && c.Server.Mode != "test" {
		return fmt.Errorf("invalid server mode: %s", c.Server.Mode)
	}

	// Validate database config
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.Port == "" {
		return fmt.Errorf("database port is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("database user is required")
	}
	if c.Database.DBName == "" {
		return fmt.Errorf("database name is required")
	}
	if c.Database.MaxOpenConns <= 0 {
		return fmt.Errorf("database max open connections must be positive")
	}
	if c.Database.MaxIdleConns <= 0 {
		return fmt.Errorf("database max idle connections must be positive")
	}

	// Validate Redis config
	if c.Redis.Host == "" {
		return fmt.Errorf("redis host is required")
	}
	if c.Redis.Port == "" {
		return fmt.Errorf("redis port is required")
	}
	if c.Redis.DB < 0 || c.Redis.DB > 15 {
		return fmt.Errorf("redis database must be between 0 and 15")
	}
	if c.Redis.PoolSize <= 0 {
		return fmt.Errorf("redis pool size must be positive")
	}

	// Validate actor config
	if c.Actor.MaxActors <= 0 {
		return fmt.Errorf("actor max actors must be positive")
	}
	if c.Actor.SupervisionStrategy != "restart" && c.Actor.SupervisionStrategy != "stop" && c.Actor.SupervisionStrategy != "ignore" {
		return fmt.Errorf("invalid actor supervision strategy: %s", c.Actor.SupervisionStrategy)
	}

	// Validate logging config
	if c.Logging.Level != "debug" && c.Logging.Level != "info" && c.Logging.Level != "warn" && c.Logging.Level != "error" {
		return fmt.Errorf("invalid log level: %s", c.Logging.Level)
	}
	if c.Logging.Format != "json" && c.Logging.Format != "text" {
		return fmt.Errorf("invalid log format: %s", c.Logging.Format)
	}
	if c.Logging.Output != "stdout" && c.Logging.Output != "file" {
		return fmt.Errorf("invalid log output: %s", c.Logging.Output)
	}
	if c.Logging.Output == "file" && c.Logging.FilePath == "" {
		return fmt.Errorf("log file path is required when output is file")
	}

	// Validate metrics config
	if c.Metrics.BatchSize <= 0 {
		return fmt.Errorf("metrics batch size must be positive")
	}

	return nil
}

// GetDSN returns the PostgreSQL data source name
func (c *Config) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}

// GetRedisAddr returns the Redis address
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%s", c.Redis.Host, c.Redis.Port)
}

// GetServerAddr returns the server address
func (c *Config) GetServerAddr() string {
	return fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port)
}

// Helper functions for environment variable parsing

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getFloatEnv(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return defaultValue
}

func getMapEnv(key string) map[string]string {
	result := make(map[string]string)
	if value := os.Getenv(key); value != "" {
		// Parse comma-separated key=value pairs
		pairs := strings.Split(value, ",")
		for _, pair := range pairs {
			if kv := strings.SplitN(strings.TrimSpace(pair), "=", 2); len(kv) == 2 {
				result[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
			}
		}
	}
	return result
}

// getStringSliceEnv gets a string slice from environment variable (comma-separated)
func getStringSliceEnv(key string, defaultValue []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	var result []string
	parts := strings.Split(value, ",")
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// Development returns a configuration suitable for development
func Development() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         "8080",
			Host:         "localhost",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
			Mode:         "debug",
		},
		Database: DatabaseConfig{
			Host:            "localhost",
			Port:            "5432",
			User:            "postgres",
			Password:        "postgres",
			DBName:          "actor_observability",
			SSLMode:         "disable",
			MaxOpenConns:    10,
			MaxIdleConns:    2,
			ConnMaxLifetime: 5 * time.Minute,
			ConnMaxIdleTime: 5 * time.Minute,
		},
		Redis: RedisConfig{
			Host:         "localhost",
			Port:         "6379",
			Password:     "",
			DB:           0,
			PoolSize:     5,
			MinIdleConns: 1,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
		},
		Actor: ActorConfig{
			MaxActors:           1000,
			SupervisionStrategy: "restart",
		},
		Logging: LoggingConfig{
			Level:          "debug",
			Format:         "text",
			Output:         "stdout",
			FilePath:       "",
			MaxSize:        50,
			MaxBackups:     2,
			MaxAge:         7,
			Compress:       false,
			SkipPaths:      []string{"/metrics", "/health", "/prometheus"},
			SkipUserAgents: []string{"Prometheus", "kube-probe"},
		},
		Metrics: MetricsConfig{
			CollectInterval: 30 * time.Second,
			FlushInterval:   5 * time.Minute,
			RetentionPeriod: 24 * time.Hour,
			BatchSize:       50,
		},
	}
}

// Production returns a configuration suitable for production
func Production() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         "8080",
			Host:         "0.0.0.0",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
			Mode:         "release",
		},
		Database: DatabaseConfig{
			Host:            "localhost",
			Port:            "5432",
			User:            "postgres",
			Password:        "postgres",
			DBName:          "actor_observability",
			SSLMode:         "require",
			MaxOpenConns:    50,
			MaxIdleConns:    10,
			ConnMaxLifetime: 10 * time.Minute,
			ConnMaxIdleTime: 10 * time.Minute,
		},
		Redis: RedisConfig{
			Host:         "localhost",
			Port:         "6379",
			Password:     "",
			DB:           0,
			PoolSize:     20,
			MinIdleConns: 5,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
		},
		Actor: ActorConfig{
			MaxActors:           50000,
			SupervisionStrategy: "restart",
		},
		Logging: LoggingConfig{
			Level:          "info",
			Format:         "json",
			Output:         "file",
			FilePath:       "/var/log/actor-observability/app.log",
			MaxSize:        200,
			MaxBackups:     10,
			MaxAge:         30,
			Compress:       true,
			SkipPaths:      []string{"/metrics", "/health", "/prometheus"},
			SkipUserAgents: []string{"Prometheus", "kube-probe"},
		},
		Metrics: MetricsConfig{
			CollectInterval: 60 * time.Second,
			FlushInterval:   10 * time.Minute,
			RetentionPeriod: 7 * 24 * time.Hour,
			BatchSize:       100,
		},
	}
}
