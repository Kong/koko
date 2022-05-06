package log

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is setup on startup by cmd package.
var Logger *zap.Logger
var zapConfig zap.Config

func init() {
	// initialized only for testing purposes.
	Logger, _ = zap.NewDevelopment()
}

// SetupLogging configure parent logger with logLevel.
func SetupLogging(logLevel string) (*zap.Logger, error) {
	zapConfig = zap.NewProductionConfig()
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	level, err := zapcore.ParseLevel(logLevel)
	if err != nil {
		return nil, err
	}
	zapConfig.Level.SetLevel(level)
	logger, err := zapConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("create logger: %w", err)
	}
	Logger = logger
	return logger, nil
}

// SetLevel updates the level for the global logger config.
// All child loggers generated with the config are updated.
func SetLevel(level string) error {
	parsedLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		return fmt.Errorf("set log level: %w", err)
	}
	zapConfig.Level.SetLevel(parsedLevel)
	return nil
}
