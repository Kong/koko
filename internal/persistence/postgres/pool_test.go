package postgres

import (
	"fmt"
	"testing"
	"time"

	"github.com/kong/koko/internal/log"
	"github.com/kong/koko/internal/persistence"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestPoolOpts(t *testing.T) {
	t.Run("pool options are applied to the pgxpool.Config", func(t *testing.T) {
		d := 42 * time.Second
		opt := Opts{
			DBName:   "koko",
			Hostname: "localhost",
			Port:     DefaultPort,
			User:     "koko",
			Password: "koko",
			Pool: PoolOpts{
				MaxConns:          42,
				MinConns:          42,
				MaxConnLifetime:   d,
				MaxConnIdleTime:   d,
				HealthCheckPeriod: d,
			},
		}
		logger := log.Logger
		pgxPoolConfig, err := NewPgxPoolConfig(opt, logger)
		require.NoError(t, err)
		require.Equal(t, int32(42), pgxPoolConfig.MaxConns)
		require.Equal(t, int32(42), pgxPoolConfig.MinConns)
		require.Equal(t, d, pgxPoolConfig.MaxConnLifetime)
		require.Equal(t, d, pgxPoolConfig.MaxConnIdleTime)
		require.Equal(t, d, pgxPoolConfig.HealthCheckPeriod)
	})
	t.Run("pool options are validated", func(t *testing.T) {
		d := 10 * time.Second
		poolOpts := []PoolOpts{
			{
				MaxConns:          0,
				HealthCheckPeriod: d,
			},
			{
				MaxConns:          -1,
				HealthCheckPeriod: d,
			},
			{
				MaxConns:          10,
				HealthCheckPeriod: 0,
			},
			{
				MaxConns:          10,
				HealthCheckPeriod: -1,
			},
		}
		for _, poolOpt := range poolOpts {
			err := poolOpt.Validate()
			require.Error(t, err)
		}
	})
}

func TestPoolRegistration(t *testing.T) {
	customPoolOpenFunc := func(opts Opts, logger *zap.Logger) (Pool, error) {
		return nil, fmt.Errorf("custom pool open func")
	}
	t.Run("default pool is registered", func(t *testing.T) {
		defaultPoolOpenFunc, present := pools[DefaultPool]
		require.NotNil(t, defaultPoolOpenFunc)
		require.True(t, present)
	})
	t.Run("custom pool can be registered", func(t *testing.T) {
		err := RegisterPool("customPool", customPoolOpenFunc)
		require.NoError(t, err)
	})
	t.Run("custom pool can be set using PoolOpts.Name", func(t *testing.T) {
		err := RegisterPool("customPool2", customPoolOpenFunc)
		require.NoError(t, err)
		opt := Opts{
			DBName:   "koko",
			Hostname: "localhost",
			Port:     DefaultPort,
			User:     "koko",
			Password: "koko",
			Pool: PoolOpts{
				Name:              "customPool2",
				MaxConns:          persistence.DefaultMaxConn,
				HealthCheckPeriod: persistence.DefaultHealthCheckPeriod,
			},
		}
		logger := log.Logger
		_, err = newPostgresPool(opt, logger)
		require.ErrorContains(t, err, "custom pool open func")
	})
	t.Run("invalid custom pool return error", func(t *testing.T) {
		opt := Opts{
			DBName:   "koko",
			Hostname: "localhost",
			Port:     DefaultPort,
			User:     "koko",
			Password: "koko",
			Pool: PoolOpts{
				Name:              "invalidPool",
				MaxConns:          persistence.DefaultMaxConn,
				HealthCheckPeriod: persistence.DefaultHealthCheckPeriod,
			},
		}
		logger := log.Logger
		_, err := newPostgresPool(opt, logger)
		require.ErrorContains(t, err, "invalid postgres pool")
	})
	t.Run("pool name is unique and cannot be registered twice", func(t *testing.T) {
		err := RegisterPool("myOtherPool", customPoolOpenFunc)
		require.NoError(t, err)
		err = RegisterPool("myOtherPool", customPoolOpenFunc)
		require.Error(t, err)
		require.ErrorContains(t, err, "already exists")
	})
}
