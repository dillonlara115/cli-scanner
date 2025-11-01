package utils

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is a global logger instance
var Logger *zap.Logger

// InitLogger initializes the global logger
func InitLogger(debug bool) error {
	var config zap.Config
	if debug {
		config = zap.NewDevelopmentConfig()
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	} else {
		config = zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	logger, err := config.Build(zap.AddCallerSkip(1))
	if err != nil {
		return err
	}

	Logger = logger
	return nil
}

// Info logs an info message
func Info(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Info(msg, fields...)
	}
}

// Debug logs a debug message
func Debug(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Debug(msg, fields...)
	}
}

// Error logs an error message
func Error(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Error(msg, fields...)
	}
}

// Warn logs a warning message
func Warn(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Warn(msg, fields...)
	}
}

// Sync flushes any buffered log entries
func Sync() {
	if Logger != nil {
		Logger.Sync()
	}
}

// NewField creates a zap field
func NewField(key string, value interface{}) zap.Field {
	switch v := value.(type) {
	case string:
		return zap.String(key, v)
	case int:
		return zap.Int(key, v)
	case bool:
		return zap.Bool(key, v)
	case zapcore.Field:
		return v
	default:
		return zap.Any(key, value)
	}
}

