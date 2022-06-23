package log

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestLogger(t *testing.T) {
	logger, err := SetupLogging("debug", "json")
	require.Nil(t, err)
	require.True(t, zapConfig.Level.Enabled(zap.DebugLevel))
	require.True(t, logger.Core().Enabled(zapcore.DebugLevel))

	childLogger := logger.With(zap.String("component", "child-logger"))
	require.True(t, childLogger.Core().Enabled(zapcore.DebugLevel))

	t.Run("SetLevel sets level with valid input", func(t *testing.T) {
		err := SetLevel("info")
		require.NoError(t, err)
		require.True(t, zapConfig.Level.Enabled(zap.InfoLevel))
		require.False(t, logger.Core().Enabled(zapcore.DebugLevel))
		require.True(t, logger.Core().Enabled(zapcore.InfoLevel))
	})

	t.Run("SetLevel is case insensitive", func(t *testing.T) {
		err := SetLevel("WARN")
		require.NoError(t, err)
		require.False(t, zapConfig.Level.Enabled(zap.InfoLevel))
		require.False(t, logger.Core().Enabled(zapcore.InfoLevel))
		require.True(t, logger.Core().Enabled(zapcore.WarnLevel))
	})

	t.Run("SetLevel returns error with invalid input", func(t *testing.T) {
		err := SetLevel("banana")
		require.Error(t, err)
		require.False(t, zapConfig.Level.Enabled(zap.DebugLevel))
		require.False(t, logger.Core().Enabled(zapcore.InfoLevel))
	})

	t.Run("SetLevel defaults to 'info' with empty string", func(t *testing.T) {
		err := SetLevel("")
		require.NoError(t, err)
		require.False(t, logger.Core().Enabled(zapcore.DebugLevel))
		require.True(t, logger.Core().Enabled(zapcore.InfoLevel))
		require.True(t, logger.Core().Enabled(zapcore.WarnLevel))
	})

	t.Run("SetLevel effects child loggers", func(t *testing.T) {
		err := SetLevel("info")
		require.NoError(t, err)
		require.False(t, zapConfig.Level.Enabled(zap.DebugLevel))
		require.True(t, zapConfig.Level.Enabled(zap.InfoLevel))
		require.False(t, childLogger.Core().Enabled(zapcore.DebugLevel))
		require.True(t, childLogger.Core().Enabled(zapcore.InfoLevel))

		err = SetLevel("warn")
		require.NoError(t, err)
		require.False(t, zapConfig.Level.Enabled(zap.InfoLevel))
		require.True(t, zapConfig.Level.Enabled(zap.WarnLevel))
		require.False(t, childLogger.Core().Enabled(zapcore.InfoLevel))
		require.True(t, childLogger.Core().Enabled(zapcore.WarnLevel))
	})
}
