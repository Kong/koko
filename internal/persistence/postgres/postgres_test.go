package postgres

import (
	"testing"

	"github.com/kong/koko/internal/log"
	"github.com/stretchr/testify/require"
)

func TestDSN(t *testing.T) {
	t.Run("Non-TLS", func(t *testing.T) {
		opt := Opts{
			DBName:    "koko",
			Hostname:  "localhost",
			Port:      DefaultPort,
			User:      "koko",
			Password:  "koko",
			EnableTLS: false,
		}
		logger := log.Logger
		expectedDSN := "host=localhost port=5432 user=koko password=koko dbname=koko sslmode=disable"
		dsn, err := opt.DSN(logger)
		require.NoError(t, err)
		require.Equal(t, expectedDSN, dsn)
	})
	t.Run("TLS DSN", func(t *testing.T) {
		opt := Opts{
			DBName:         "koko",
			Hostname:       "localhost",
			Port:           DefaultPort,
			User:           "koko",
			Password:       "koko",
			EnableTLS:      true,
			CABundleFSPath: "/koko-postgres-cabundle/global-bundle.pem",
		}
		logger := log.Logger
		expectedDSN := "host=localhost port=5432 user=koko password=koko dbname=koko " +
			"sslrootcert=/koko-postgres-cabundle/global-bundle.pem sslmode=verify-full"
		dsn, err := opt.DSN(logger)
		require.NoError(t, err)
		require.Equal(t, expectedDSN, dsn)
	})
}

func TestDSNParams(t *testing.T) {
	t.Run("params validation", func(t *testing.T) {
		invalidParams := []string{
			"dbname",
			"host",
			"port",
			"user",
			"password",
			"sslrootcert",
		}
		for _, invalidParam := range invalidParams {
			opt := Opts{
				DBName:         "koko",
				Hostname:       "localhost",
				Port:           DefaultPort,
				User:           "koko",
				Password:       "koko",
				EnableTLS:      true,
				CABundleFSPath: "/koko-postgres-cabundle/global-bundle.pem",
				Params: map[string]string{
					invalidParam: "test",
				},
			}
			logger := log.Logger
			client, err := NewSQLClient(opt, logger)
			require.Nil(t, client)
			require.Error(t, err)
		}
	})
	t.Run("params inclusion in DSN", func(t *testing.T) {
		opt := Opts{
			DBName:         "koko",
			Hostname:       "localhost",
			Port:           DefaultPort,
			User:           "koko",
			Password:       "koko",
			EnableTLS:      true,
			CABundleFSPath: "/koko-postgres-cabundle/global-bundle.pem",
			Params: map[string]string{
				"connect_timeout":           "10",
				"fallback_application_name": "testApp",
			},
		}
		logger := log.Logger
		expectedBaseDSN := "host=localhost port=5432 user=koko password=koko dbname=koko " +
			"sslrootcert=/koko-postgres-cabundle/global-bundle.pem"
		dsn, err := opt.DSN(logger)
		require.NoError(t, err)
		require.Contains(t, dsn, expectedBaseDSN)
		require.Contains(t, dsn, "connect_timeout=10")
		require.Contains(t, dsn, "fallback_application_name=testApp")
	})
	t.Run("sslmode param has priority over default", func(t *testing.T) {
		optNoTLS := Opts{
			DBName:   "koko",
			Hostname: "localhost",
			Port:     DefaultPort,
			User:     "koko",
			Password: "koko",
			Params: map[string]string{
				"sslmode": "verify-ca",
			},
		}
		optTLS := Opts{
			DBName:         "koko",
			Hostname:       "localhost",
			Port:           DefaultPort,
			User:           "koko",
			Password:       "koko",
			EnableTLS:      true,
			CABundleFSPath: "/koko-postgres-cabundle/global-bundle.pem",
			Params: map[string]string{
				"sslmode": "verify-ca",
			},
		}
		for _, opt := range []Opts{optNoTLS, optTLS} {
			logger := log.Logger
			dsn, err := opt.DSN(logger)
			require.Contains(t, dsn, "sslmode=verify-ca")
			require.NotContains(t, dsn, "sslmode=disable")
			require.NotContains(t, dsn, "sslmode="+defaultSSLMode)
			require.NoError(t, err)
		}
	})
	t.Run("params are quoted and escaped in DSN", func(t *testing.T) {
		opt := Opts{
			DBName:   "koko",
			Hostname: "localhost",
			Port:     DefaultPort,
			User:     "koko",
			Password: "koko",
			Params: map[string]string{
				"application_name":          "I'm the main app",
				"fallback_application_name": "My Test App",
				"sslkey":                    "/paths'certs/",
			},
		}
		logger := log.Logger
		dsn, err := opt.DSN(logger)
		require.NoError(t, err)
		require.Contains(t, dsn, "application_name='I\\'m the main app'")
		require.Contains(t, dsn, "fallback_application_name='My Test App'")
		require.Contains(t, dsn, "sslkey=/paths\\'certs/")
	})
}
