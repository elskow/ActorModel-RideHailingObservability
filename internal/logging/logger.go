package logging

import (
	"io"
	"os"
	"path/filepath"

	"actor-model-observability/internal/config"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger wraps logrus.Logger with additional functionality
type Logger struct {
	*logrus.Logger
	config *config.LoggingConfig
}

// Fields type for structured logging
type Fields map[string]interface{}

// NewLogger creates a new logger instance based on configuration
func NewLogger(cfg *config.LoggingConfig) (*Logger, error) {
	logger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}
	logger.SetLevel(level)

	// Set formatter
	switch cfg.Format {
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
				logrus.FieldKeyFunc:  "function",
				logrus.FieldKeyFile:  "file",
			},
		})
	case "text":
		logger.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			FullTimestamp:   true,
			ForceColors:     true,
		})
	default:
		logger.SetFormatter(&logrus.JSONFormatter{})
	}

	// Set output
	var output io.Writer
	switch cfg.Output {
	case "stdout":
		output = os.Stdout
	case "file":
		// Create directory if it doesn't exist
		dir := filepath.Dir(cfg.FilePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}

		// Use lumberjack for log rotation
		output = &lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}
	default:
		output = os.Stdout
	}

	logger.SetOutput(output)

	// Enable caller reporting for better debugging
	logger.SetReportCaller(true)

	return &Logger{
		Logger: logger,
		config: cfg,
	}, nil
}

// WithFields creates a new logger entry with the specified fields
func (l *Logger) WithFields(fields Fields) *logrus.Entry {
	return l.Logger.WithFields(logrus.Fields(fields))
}

// WithField creates a new logger entry with a single field
func (l *Logger) WithField(key string, value interface{}) *logrus.Entry {
	return l.Logger.WithField(key, value)
}

// WithError creates a new logger entry with an error field
func (l *Logger) WithError(err error) *logrus.Entry {
	return l.Logger.WithError(err)
}

// WithComponent creates a new logger entry with a component field
func (l *Logger) WithComponent(component string) *logrus.Entry {
	return l.Logger.WithField("component", component)
}

// WithActor creates a new logger entry with actor-related fields
func (l *Logger) WithActor(actorID, actorType string) *logrus.Entry {
	return l.Logger.WithFields(logrus.Fields{
		"actor_id":   actorID,
		"actor_type": actorType,
		"component":  "actor",
	})
}

// WithMessage creates a new logger entry with message-related fields
func (l *Logger) WithMessage(messageID, messageType, fromActor, toActor string) *logrus.Entry {
	return l.Logger.WithFields(logrus.Fields{
		"message_id":   messageID,
		"message_type": messageType,
		"from_actor":   fromActor,
		"to_actor":     toActor,
		"component":    "message",
	})
}

// WithRequest creates a new logger entry with HTTP request fields
func (l *Logger) WithRequest(method, path, userAgent, requestID string) *logrus.Entry {
	return l.Logger.WithFields(logrus.Fields{
		"method":     method,
		"path":       path,
		"user_agent": userAgent,
		"request_id": requestID,
		"component":  "http",
	})
}

// WithDatabase creates a new logger entry with database operation fields
func (l *Logger) WithDatabase(operation, table string, duration int64) *logrus.Entry {
	return l.Logger.WithFields(logrus.Fields{
		"operation": operation,
		"table":     table,
		"duration":  duration,
		"component": "database",
	})
}

// WithMetrics creates a new logger entry with metrics fields
func (l *Logger) WithMetrics(metricType, metricName string, value interface{}) *logrus.Entry {
	return l.Logger.WithFields(logrus.Fields{
		"metric_type": metricType,
		"metric_name": metricName,
		"value":       value,
		"component":   "metrics",
	})
}

// WithTrace creates a new logger entry with distributed tracing fields
func (l *Logger) WithTrace(traceID, spanID, operation string) *logrus.Entry {
	return l.Logger.WithFields(logrus.Fields{
		"trace_id":  traceID,
		"span_id":   spanID,
		"operation": operation,
		"component": "trace",
	})
}

// LogActorEvent logs an actor lifecycle event
func (l *Logger) LogActorEvent(actorID, actorType, event string, fields Fields) {
	entry := l.WithActor(actorID, actorType).WithField("event", event)
	if fields != nil {
		entry = entry.WithFields(logrus.Fields(fields))
	}
	entry.Info("Actor event")
}

// LogMessageEvent logs a message processing event
func (l *Logger) LogMessageEvent(messageID, messageType, fromActor, toActor, event string, fields Fields) {
	entry := l.WithMessage(messageID, messageType, fromActor, toActor).WithField("event", event)
	if fields != nil {
		entry = entry.WithFields(logrus.Fields(fields))
	}
	entry.Info("Message event")
}

// LogHTTPRequest logs an HTTP request
func (l *Logger) LogHTTPRequest(method, path, userAgent, requestID string, statusCode int, duration int64, fields Fields) {
	entry := l.WithRequest(method, path, userAgent, requestID).
		WithField("status_code", statusCode).
		WithField("duration", duration)
	if fields != nil {
		entry = entry.WithFields(logrus.Fields(fields))
	}

	if statusCode >= 400 {
		entry.Warn("HTTP request completed with error")
	} else {
		entry.Info("HTTP request completed")
	}
}

// LogDatabaseOperation logs a database operation
func (l *Logger) LogDatabaseOperation(operation, table string, duration int64, err error, fields Fields) {
	entry := l.WithDatabase(operation, table, duration)
	if fields != nil {
		entry = entry.WithFields(logrus.Fields(fields))
	}

	if err != nil {
		entry.WithError(err).Error("Database operation failed")
	} else {
		entry.Debug("Database operation completed")
	}
}

// LogMetricCollection logs a metric collection event
func (l *Logger) LogMetricCollection(metricType, metricName string, value interface{}, fields Fields) {
	entry := l.WithMetrics(metricType, metricName, value)
	if fields != nil {
		entry = entry.WithFields(logrus.Fields(fields))
	}
	entry.Debug("Metric collected")
}

// LogSystemEvent logs a system-level event
func (l *Logger) LogSystemEvent(event, component string, fields Fields) {
	entry := l.WithComponent(component).WithField("event", event)
	if fields != nil {
		entry = entry.WithFields(logrus.Fields(fields))
	}
	entry.Info("System event")
}

// LogError logs an error with context
func (l *Logger) LogError(err error, component, operation string, fields Fields) {
	entry := l.WithError(err).
		WithField("component", component).
		WithField("operation", operation)
	if fields != nil {
		entry = entry.WithFields(logrus.Fields(fields))
	}
	entry.Error("Operation failed")
}

// LogPanic logs a panic with context
func (l *Logger) LogPanic(panicValue interface{}, component, operation string, fields Fields) {
	entry := l.WithField("panic", panicValue).
		WithField("component", component).
		WithField("operation", operation)
	if fields != nil {
		entry = entry.WithFields(logrus.Fields(fields))
	}
	entry.Fatal("Panic occurred")
}

// Close closes the logger and any associated resources
func (l *Logger) Close() error {
	if l.config.Output == "file" {
		if closer, ok := l.Logger.Out.(io.Closer); ok {
			return closer.Close()
		}
	}
	return nil
}

// Global logger instance
var defaultLogger *Logger

// InitGlobalLogger initializes the global logger
func InitGlobalLogger(cfg *config.LoggingConfig) error {
	logger, err := NewLogger(cfg)
	if err != nil {
		return err
	}
	defaultLogger = logger
	return nil
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() *Logger {
	if defaultLogger == nil {
		// Fallback to a basic logger if not initialized
		logger, _ := NewLogger(&config.LoggingConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		})
		return logger
	}
	return defaultLogger
}

// Convenience functions using the global logger

// Debug logs a debug message
func Debug(args ...interface{}) {
	GetGlobalLogger().Debug(args...)
}

// Info logs an info message
func Info(args ...interface{}) {
	GetGlobalLogger().Info(args...)
}

// Warn logs a warning message
func Warn(args ...interface{}) {
	GetGlobalLogger().Warn(args...)
}

// Error logs an error message
func Error(args ...interface{}) {
	GetGlobalLogger().Error(args...)
}

// Fatal logs a fatal message and exits
func Fatal(args ...interface{}) {
	GetGlobalLogger().Fatal(args...)
}

// WithFields creates a new logger entry with fields using the global logger
func WithFields(fields Fields) *logrus.Entry {
	return GetGlobalLogger().WithFields(fields)
}

// WithField creates a new logger entry with a field using the global logger
func WithField(key string, value interface{}) *logrus.Entry {
	return GetGlobalLogger().WithField(key, value)
}

// WithError creates a new logger entry with an error using the global logger
func WithError(err error) *logrus.Entry {
	return GetGlobalLogger().WithError(err)
}

// WithComponent creates a new logger entry with a component using the global logger
func WithComponent(component string) *logrus.Entry {
	return GetGlobalLogger().WithComponent(component)
}