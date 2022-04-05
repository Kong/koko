package postgres

import (
	"testing"

	"github.com/kong/koko/internal/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDSN(t *testing.T) {
	t.Run("Non-TLS", func(t *testing.T) {
		opt := Opts{
			DBName:    "koko",
			Hostname:  "localhost",
			Port:      5432,
			User:      "koko",
			Password:  "koko",
			EnableTLS: false,
		}
		logger := log.Logger
		expectedDSN := "host=localhost port=5432 user=koko password=koko dbname=koko sslmode=disable"
		dsn := getDSN(opt, logger)
		require.Equal(t, expectedDSN, dsn)
	})
	t.Run("TLS DSN", func(t *testing.T) {
		opt := Opts{
			DBName:         "koko",
			Hostname:       "localhost",
			Port:           5432,
			User:           "koko",
			Password:       "koko",
			EnableTLS:      true,
			CABundleFSPath: "/koko-postgres-cabundle/global-bundle.pem",
		}
		logger := log.Logger
		expectedDSN := "host=localhost port=5432 user=koko password=koko dbname=koko sslmode=verify-full " +
			"sslrootcert=/koko-postgres-cabundle/global-bundle.pem"
		dsn := getDSN(opt, logger)
		require.Equal(t, expectedDSN, dsn)
	})
	t.Run("TLS DSN No CABundlePath", func(t *testing.T) {
		opt := Opts{
			DBName:    "koko",
			Hostname:  "localhost",
			Port:      5432,
			User:      "koko",
			Password:  "koko",
			EnableTLS: true,
		}
		logger := log.Logger
		assert.Panics(t, func() { getDSN(opt, logger) }, "getDSN must panic")
	})
}
