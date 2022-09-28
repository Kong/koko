package mysql

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"strconv"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/kong/koko/internal/persistence"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

// Opts defines various options when creating a MySQL DB instance.
type Opts struct {
	// Primary (read/write) DB connection settings.
	DBName   string
	Hostname string
	Port     int
	User     string
	Password string

	// Optional hostname for a read-only replica.
	// Connection to this DB shares the same options as the primary (Opts.Hostname).
	ReadOnlyHostname string

	// TLS options.
	EnableTLS    bool
	RootCAs      []string
	Certificates []string

	// Optional function used for tls.Config.VerifyPeerCertificate when TLS is enabled.
	// This may be set to mysql.VerifyPeerCertFunc.
	VerifyPeerCertificateFunc func(*x509.CertPool) func([][]byte, [][]*x509.Certificate) error

	// Defaults to UTC when not provided.
	Location *time.Location

	// Optional function for defining the connection to the DB.
	// When not provided, defaults to persistence.DefaultSQLOpenFunc.
	SQLOpen persistence.SQLOpenFunc

	// Parameters passed to the MySQL DB driver.
	//
	// Here be dragons using this, as utilizing parameters like `clientFoundRows=true` will break things.
	// This is here to allow the enablement of useful parameters, like `checkConnLiveness=true`.
	//
	// Read more: https://github.com/go-sql-driver/mysql#parameters
	Params map[string]string

	logger *zap.Logger
}

// DSN formats the given Opts into a DSN string which can be passed to the driver.
func (o *Opts) DSN() (string, error) {
	if o.Params == nil {
		o.Params = make(map[string]string)
	}

	if o.Location == nil {
		o.Location = time.UTC
	}

	if o.Port != 0 {
		o.Hostname += ":" + strconv.Itoa(o.Port)
	}

	if o.logger == nil {
		o.logger = zap.L()
	}

	config := mysql.Config{
		// Default settings.
		AllowNativePasswords: true,
		Collation:            "utf8mb4_unicode_520_ci",
		Net:                  "tcp",
		ParseTime:            true,
		Timeout:              persistence.DefaultDialTimeout,

		// Dynamic settings.
		Addr:   o.Hostname,
		DBName: o.DBName,
		Loc:    o.Location,
		Params: o.Params,
		User:   o.User,
		Passwd: o.Password,
	}

	if o.tlsEnabled() {
		o.logger.Info("using TLS MySQL connection")
		if err := o.setTLSConfig(&config); err != nil {
			return "", err
		}
	} else {
		o.logger.Info("using non-TLS MySQL connection")
	}

	return config.FormatDSN(), nil
}

// VerifyPeerCertFunc returns a function that verifies the peer certificate is in the provided cert pool.
func VerifyPeerCertFunc(pool *x509.CertPool) func([][]byte, [][]*x509.Certificate) error {
	return func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
		if len(rawCerts) == 0 {
			return errors.New("no certificates available to verify")
		}

		cert, err := x509.ParseCertificate(rawCerts[0])
		if err != nil {
			return err
		}

		if _, err = cert.Verify(x509.VerifyOptions{Roots: pool}); err != nil {
			return err
		}

		return nil
	}
}

func (o *Opts) setTLSConfig(dbConfig *mysql.Config) error {
	dbConfig.TLSConfig = "custom"

	pool := x509.NewCertPool()

	tlsConfig := &tls.Config{
		// TODO(tjasko): Implement me.
		Certificates: nil,
		RootCAs:      nil,

		// This replicates the behavior of the MySQL driver.
		//
		// There are valid use-cases for this, like when using Google Cloud SQL.
		// e.g.: https://cloud.google.com/sql/docs/mysql/samples/cloud-sql-mysql-databasesql-sslcerts
		//
		//nolint:gosec
		InsecureSkipVerify: lo.Contains([]string{"skip-verify", "preferred"}, dbConfig.Params["tls"]),
	}

	if o.VerifyPeerCertificateFunc != nil {
		tlsConfig.VerifyPeerCertificate = o.VerifyPeerCertificateFunc(pool)
	} else if tlsConfig.InsecureSkipVerify {
		// There may be a valid use-case for this, however we'll at least emit a warning
		// when TLS certificate verification is skipped with no VerifyPeerCertificateFunc.
		o.logger.Warn(
			"⚠️ Using an insecure TLS MySQL connection with no peer certificate validation! " +
				"Are you sure you meant to use an insecure TLS connection? ⚠️",
		)
	}

	return mysql.RegisterTLSConfig(dbConfig.TLSConfig, tlsConfig)
}

func (o *Opts) tlsEnabled() bool {
	return o.EnableTLS || !lo.Contains([]string{"false", ""}, o.Params["tls"])
}
