package logging

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"actor-model-observability/internal/config"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger wraps slog.Logger with additional functionality
type Logger struct {
	*slog.Logger
	config *config.LoggingConfig
}

// Fields type for structured logging
type Fields map[string]interface{}

// NewLogger creates a new logger instance based on configuration
func NewLogger(cfg *config.LoggingConfig) (*Logger, error) {
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

	// Configure handler options
	var handler slog.Handler
	handlerOpts := &slog.HandlerOptions{
		Level:     parseLevel(cfg.Level),
		AddSource: cfg.Level == "debug", // Only add source info for debug level
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Customize attribute names to match previous format
			switch a.Key {
			case slog.TimeKey:
				a.Key = "timestamp"
			case slog.LevelKey:
				a.Key = "level"
			case slog.MessageKey:
				a.Key = "message"
			}
			return a
		},
	}

	// Set formatter based on configuration
	switch cfg.Format {
	case "json":
		handler = slog.NewJSONHandler(output, handlerOpts)
	case "text":
		handler = slog.NewTextHandler(output, handlerOpts)
	default:
		handler = slog.NewJSONHandler(output, handlerOpts)
	}

	logger := slog.New(handler)

	return &Logger{
		Logger: logger,
		config: cfg,
	}, nil
}

// parseLevel converts string level to slog.Level
func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// fieldsToAttrs converts Fields to slog.Attr slice
func fieldsToAttrs(fields Fields) []slog.Attr {
	if fields == nil {
		return nil
	}
	attrs := make([]slog.Attr, 0, len(fields))
	for k, v := range fields {
		attrs = append(attrs, slog.Any(k, v))
	}
	return attrs
}

// WithFields creates a new logger with the specified fields
func (l *Logger) WithFields(fields Fields) *Logger {
	if fields == nil {
		return l
	}
	attrs := fieldsToAttrs(fields)
	args := make([]any, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}
	logger := l.Logger.With(args...)
	return &Logger{Logger: logger, config: l.config}
}

// WithField creates a new logger with a single field
func (l *Logger) WithField(key string, value interface{}) *Logger {
	logger := l.Logger.With(slog.Any(key, value))
	return &Logger{Logger: logger, config: l.config}
}

// WithError creates a new logger with an error field
func (l *Logger) WithError(err error) *Logger {
	logger := l.Logger.With(slog.Any("error", err))
	return &Logger{Logger: logger, config: l.config}
}

// WithComponent creates a new logger with a component field
func (l *Logger) WithComponent(component string) *Logger {
	return l.WithField("component", component)
}

// WithActor creates a new logger with actor-related fields
func (l *Logger) WithActor(actorID, actorType string) *Logger {
	return l.WithFields(Fields{
		"actor_id":   actorID,
		"actor_type": actorType,
		"component":  "actor",
	})
}

// WithMessage creates a new logger with message-related fields
func (l *Logger) WithMessage(messageID, messageType, fromActor, toActor string) *Logger {
	return l.WithFields(Fields{
		"message_id":   messageID,
		"message_type": messageType,
		"from_actor":   fromActor,
		"to_actor":     toActor,
		"component":    "message",
	})
}

// WithRequest creates a new logger with HTTP request fields
func (l *Logger) WithRequest(method, path, userAgent, requestID string) *Logger {
	return l.WithFields(Fields{
		"method":     method,
		"path":       path,
		"user_agent": userAgent,
		"request_id": requestID,
		"component":  "http",
	})
}

// WithDatabase creates a new logger with database operation fields
func (l *Logger) WithDatabase(operation, table string, duration int64) *Logger {
	return l.WithFields(Fields{
		"operation": operation,
		"table":     table,
		"duration":  duration,
		"component": "database",
	})
}

// WithMetrics creates a new logger with metrics fields
func (l *Logger) WithMetrics(metricType, metricName string, value interface{}) *Logger {
	return l.WithFields(Fields{
		"metric_type": metricType,
		"metric_name": metricName,
		"value":       value,
		"component":   "metrics",
	})
}

// WithTrace creates a new logger with distributed tracing fields
func (l *Logger) WithTrace(traceID, spanID, operation string) *Logger {
	return l.WithFields(Fields{
		"trace_id":  traceID,
		"span_id":   spanID,
		"operation": operation,
		"component": "trace",
	})
}

// LogActorEvent logs an actor lifecycle event
func (l *Logger) LogActorEvent(actorID, actorType, event string, fields Fields) {
	logger := l.WithActor(actorID, actorType).WithField("event", event)
	if fields != nil {
		logger = logger.WithFields(fields)
	}
	logger.Info("Actor event")
}

// LogMessageEvent logs a message processing event
func (l *Logger) LogMessageEvent(messageID, messageType, fromActor, toActor, event string, fields Fields) {
	logger := l.WithMessage(messageID, messageType, fromActor, toActor).WithField("event", event)
	if fields != nil {
		logger = logger.WithFields(fields)
	}
	logger.Info("Message event")
}

// LogHTTPRequest logs HTTP request information
func (l *Logger) LogHTTPRequest(method, path, userAgent, requestID string, statusCode int, latencyMs int64, fields Fields) {
	logFields := Fields{
		"method":     method,
		"path":       path,
		"user_agent": userAgent,
		"request_id": requestID,
		"status":     statusCode,
		"latency_ms": latencyMs,
	}

	// Merge additional fields
	for k, v := range fields {
		logFields[k] = v
	}

	l.WithFields(logFields).Info("HTTP request")
}

// LogDatabaseOperation logs a database operation
func (l *Logger) LogDatabaseOperation(operation, table string, duration int64, err error, fields Fields) {
	logger := l.WithDatabase(operation, table, duration)
	if fields != nil {
		logger = logger.WithFields(fields)
	}

	if err != nil {
		logger.WithError(err).Error("Database operation failed")
	} else {
		logger.Debug("Database operation completed")
	}
}

// LogMetricCollection logs a metric collection event
func (l *Logger) LogMetricCollection(metricType, metricName string, value interface{}, fields Fields) {
	logger := l.WithMetrics(metricType, metricName, value)
	if fields != nil {
		logger = logger.WithFields(fields)
	}
	logger.Debug("Metric collected")
}

// LogSystemEvent logs a system-level event
func (l *Logger) LogSystemEvent(event, component string, fields Fields) {
	logger := l.WithComponent(component).WithField("event", event)
	if fields != nil {
		logger = logger.WithFields(fields)
	}
	logger.Info("System event")
}

// LogError logs an error with context
func (l *Logger) LogError(err error, component, operation string, fields Fields) {
	logger := l.WithError(err).
		WithField("component", component).
		WithField("operation", operation)
	if fields != nil {
		logger = logger.WithFields(fields)
	}
	logger.Error("Operation failed")
}

// LogPanic logs a panic with context
func (l *Logger) LogPanic(panicValue interface{}, component, operation string, fields Fields) {
	logger := l.WithField("panic", panicValue).
		WithField("component", component).
		WithField("operation", operation)
	if fields != nil {
		logger = logger.WithFields(fields)
	}
	logger.Error("Panic occurred") // Use Error instead of Fatal to avoid os.Exit
}

// Close closes the logger and any associated resources
func (l *Logger) Close() error {
	if l.config.Output == "file" {
		if closer, ok := l.Logger.Handler().(io.Closer); ok {
			return closer.Close()
		}
	}
	return nil
}

// Convenience methods that match slog interface
func (l *Logger) Debug(msg string, args ...any) {
	l.Logger.Debug(msg, args...)
}

func (l *Logger) Info(msg string, args ...any) {
	l.Logger.Info(msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.Logger.Warn(msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.Logger.Error(msg, args...)
}

func (l *Logger) Fatal(msg string, args ...any) {
	l.Logger.Error(msg, args...)
	os.Exit(1)
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
func Debug(msg string, args ...any) {
	GetGlobalLogger().Debug(msg, args...)
}

// Info logs an info message
func Info(msg string, args ...any) {
	GetGlobalLogger().Info(msg, args...)
}

// Warn logs a warning message
func Warn(msg string, args ...any) {
	GetGlobalLogger().Warn(msg, args...)
}

// Error logs an error message
func Error(msg string, args ...any) {
	GetGlobalLogger().Error(msg, args...)
}

// WithFields creates a new logger with fields using the global logger
func WithFields(fields Fields) *Logger {
	return GetGlobalLogger().WithFields(fields)
}

// WithField creates a new logger with a field using the global logger
func WithField(key string, value interface{}) *Logger {
	return GetGlobalLogger().WithField(key, value)
}

// WithError creates a new logger with an error using the global logger
func WithError(err error) *Logger {
	return GetGlobalLogger().WithError(err)
}

// WithComponent creates a new logger with a component using the global logger
func WithComponent(component string) *Logger {
	return GetGlobalLogger().WithComponent(component)
}
