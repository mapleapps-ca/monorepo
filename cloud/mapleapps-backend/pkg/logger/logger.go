// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/logger/logger.go
package logger

import (
	"os"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewProduction creates a production-ready logger with appropriate configuration
func NewProduction() (*zap.Logger, error) {
	// Get log level from environment
	logLevel := getLogLevel()

	// Configure encoder for production (JSON format)
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create core
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		logLevel,
	)

	// Create logger with caller information
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	// Add service information
	logger = logger.With(
		zap.String("service", "mapleapps-backend"),
		zap.String("version", getServiceVersion()),
	)

	return logger, nil
}

// NewDevelopment creates a development logger (for backward compatibility)
func NewDevelopment() (*zap.Logger, error) {
	return zap.NewDevelopment()
}

// getLogLevel determines log level from environment
func getLogLevel() zapcore.Level {
	levelStr := os.Getenv("LOG_LEVEL")
	switch levelStr {
	case "debug", "DEBUG":
		return zapcore.DebugLevel
	case "info", "INFO":
		return zapcore.InfoLevel
	case "warn", "WARN", "warning", "WARNING":
		return zapcore.WarnLevel
	case "error", "ERROR":
		return zapcore.ErrorLevel
	case "panic", "PANIC":
		return zapcore.PanicLevel
	case "fatal", "FATAL":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// getServiceVersion gets the service version (could be injected at build time)
func getServiceVersion() string {
	version := os.Getenv("SERVICE_VERSION")
	if version == "" {
		return "1.0.0"
	}
	return version
}

// Module provides the logger for FX dependency injection
func Module() fx.Option {
	return fx.Options(
		fx.Provide(NewProduction),
	)
}
