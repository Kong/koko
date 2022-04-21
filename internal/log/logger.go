package log

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is setup on startup by cmd package.
var Logger *zap.Logger

func init() {
	// initialized only for testing purposes.
	Logger, _ = zap.NewDevelopment()
}

// SetupLogger configure parent logger with logLevel.
func SetupLogging(logLevel string) (*zap.Logger, error) {
	zapConfig := zap.NewProductionConfig()
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
