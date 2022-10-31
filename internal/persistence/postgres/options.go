package postgres

import (
	"fmt"
	"strings"

	"github.com/kong/koko/internal/persistence"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

var (
	defaultSSLMode            = "verify-full"
	unsupportedParamErrPrefix = "the '%s' parameter is unsupported, "
	unsupportedParamErrs      = map[string]string{
		"host":        unsupportedParamErrPrefix + " please use 'postgres.hostname'",
		"port":        unsupportedParamErrPrefix + " please use 'postgres.port'",
		"dbname":      unsupportedParamErrPrefix + " please use 'postgres.db_name'",
		"user":        unsupportedParamErrPrefix + " please use 'postgres.user'",
		"password":    unsupportedParamErrPrefix + " please use 'postgres.password'",
		"sslrootcert": unsupportedParamErrPrefix + " please use 'postgres.ca_bundle_path'",
	}
)

// Opts defines various options when creating a Postgres DB instance.
type Opts struct {
	// Primary (read/write) DB connection settings
	DBName   string
	Hostname string
	Port     int
	User     string
	Password string

	// Optional hostname for a read-only replica.
	// Connection to this DB shares the same options as the primary (Opts.Hostname).
	ReadOnlyHostname string

	// TLS options.
	EnableTLS      bool
	CABundleFSPath string

	// Optional function for defining the connection to the DB.
	// When not provided, defaults to persistence.DefaultSQLOpenFunc.
	SQLOpen persistence.SQLOpenFunc

	// Parameters passed to the Postgres DB driver (pgx).
	//
	// This is here to allow the enablement of additional parameters, like `sslmode=verify-ca`.
	// A subset of the connection parameters supported by libpq are supported by pgx. Additionally,
	// pgx also lets you specify runtime parameters directly in the connection string.
	//
	// See: http://www.postgresql.org/docs/current/static/libpq-connect.html#LIBPQ-CONNSTRING for more
	// information about connection parameters.
	// See: https://www.postgresql.org/docs/current/runtime-config.html for more information about runtime parameters.
	Params map[string]string
}

// Validate ensures the provided Postgres options are a valid configuration.
func (opts *Opts) Validate() error {
	for key := range opts.Params {
		if msg, ok := unsupportedParamErrs[key]; ok {
			return fmt.Errorf(msg, key)
		}
	}

	return nil
}

func (opts *Opts) DSN(logger *zap.Logger) (string, error) {
	var dsn string
	sslMode := defaultSSLMode

	if opts.Hostname != "" {
		dsn += fmt.Sprintf("host=%s ", opts.Hostname)
	}
	if opts.Port != 0 {
		dsn += fmt.Sprintf("port=%d ", opts.Port)
	}
	if opts.User != "" {
		dsn += fmt.Sprintf("user=%s ", opts.User)
	}
	if opts.Password != "" {
		dsn += fmt.Sprintf("password=%s ", opts.Password)
	}
	if opts.DBName != "" {
		dsn += fmt.Sprintf("dbname=%s", opts.DBName)
	}

	if !opts.EnableTLS {
		logger.Info("using non-TLS Postgres connection")
		sslMode = "disable"
	} else {
		logger.Info("using TLS Postgres connection")
		logger.Info("ca_bundle_fs_path:" + opts.CABundleFSPath)
		if opts.CABundleFSPath == "" {
			return "", fmt.Errorf("postgres connection requires TLS but ca_bundle_fs_path is empty")
		}
		dsn += fmt.Sprintf(" sslrootcert=%s", opts.CABundleFSPath)
	}

	params := lo.Assign[string, string](
		map[string]string{"sslmode": sslMode},
		opts.Params)
	return dsn + formatDSNParams(params), nil
}

func formatDSNParams(params map[string]string) string {
	var dsn string

	for param, value := range params {
		if strings.Contains(value, "'") {
			value = strings.ReplaceAll(value, "'", "\\'")
		}
		if strings.Contains(value, " ") {
			value = fmt.Sprintf("'%s'", value)
		}
		dsn += fmt.Sprintf(" %s=%s", param, value)
	}

	return dsn
}
