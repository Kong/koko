package log

import "go.uber.org/zap"

// Logger is setup on startup by cmd package.
var Logger *zap.Logger

func init() {
	// initialized only for testing purposes.
	Logger, _ = zap.NewDevelopment()
}
