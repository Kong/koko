package mysql

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/kong/koko/internal/persistence"
	"go.uber.org/zap"
)

var defaultCollation = "utf8mb4_unicode_520_ci"

var (
	unsupportedParamErrPrefix = "the '%s' parameter is unsupported, "
	unsupportedParamErrs      = map[string]string{
		"clientFoundRows":  unsupportedParamErrPrefix + " as it will break functionality",
		"collation":        unsupportedParamErrPrefix + " as it is defaulted to " + defaultCollation,
		"columnsWithAlias": unsupportedParamErrPrefix + " as it can break functionality",
		"multiStatements":  unsupportedParamErrPrefix + " to prevent against SQL injection attacks",
		"parseTime":        unsupportedParamErrPrefix + " as it is forced on by default",
		"tls":              unsupportedParamErrPrefix + " please instead use 'tls.enable' & its accompanying fields",
	}
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
	EnableTLS                bool
	RootCA                   string
	Certificate              string
	Key                      string
	SkipHostnameVerification bool

	// Defaults to UTC when not provided.
	Location *time.Location

	// Optional function for defining the connection to the DB.
	// When not provided, defaults to persistence.DefaultSQLOpenFunc.
	SQLOpen persistence.SQLOpenFunc

	// Parameters passed to the MySQL DB driver.
	//
	// This is here to allow the enablement of useful parameters, like `checkConnLiveness=true`.
	//
	// Read more: https://github.com/go-sql-driver/mysql#parameters
	Params map[string]string

	logger *zap.Logger
}

// Validate ensures the provided MySQL options are a valid configuration.
func (o *Opts) Validate() error {
	// Ensure no parameters are set that could hinder functionality or cause potential issues.
	for key := range o.Params {
		if msg, ok := unsupportedParamErrs[key]; ok {
			return fmt.Errorf(msg, key)
		}
	}

	return nil
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
		Collation:            defaultCollation,
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

	if o.EnableTLS {
		o.logger.Info("using TLS MySQL connection")
		if err := o.setTLSConfig(&config); err != nil {
			return "", err
		}
	} else {
		o.logger.Info("using non-TLS MySQL connection")
	}

	return config.FormatDSN(), nil
}

func (o *Opts) setTLSConfig(dbConfig *mysql.Config) error {
	dbConfig.TLSConfig = "custom"

	tlsConfig := &tls.Config{} //nolint:gosec

	if o.RootCA != "" {
		tlsConfig.RootCAs = x509.NewCertPool()
		if ok := tlsConfig.RootCAs.AppendCertsFromPEM([]byte(o.RootCA)); !ok {
			return errors.New("unable to append root cert to pool")
		}
	}

	// Whenever hostname verification needs to be skipped or the cert/key was not provided,
	// we'll need to override the default peer certificate validation logic.
	if o.SkipHostnameVerification || o.Certificate == "" && o.Key == "" {
		tlsConfig.InsecureSkipVerify = true
		tlsConfig.VerifyPeerCertificate = o.verifyPeerCertFunc(tlsConfig.RootCAs)
	}

	if o.Certificate != "" && o.Key != "" {
		cert, err := tls.X509KeyPair([]byte(o.Certificate), []byte(o.Key))
		if err != nil {
			return err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	} else if tlsConfig.RootCAs == nil {
		// Not providing a root CA or a cert/key is very insecure. As such, we'll let them
		// proceed, but emit a warning when peer certificate verification is skipped.
		o.logger.Warn(
			"⚠️ Using an insecure TLS MySQL connection with no peer certificate validation! " +
				"Are you sure you meant to use an insecure TLS connection? ⚠️",
		)
	}

	return mysql.RegisterTLSConfig(dbConfig.TLSConfig, tlsConfig)
}

// verifyPeerCertFunc handles X509 certificate validation in the event we're not doing full TLS validation.
func (o *Opts) verifyPeerCertFunc(pool *x509.CertPool) func([][]byte, [][]*x509.Certificate) error {
	return func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
		// This functions as insecure TLS.
		if pool == nil && o.SkipHostnameVerification {
			return nil
		}

		if len(rawCerts) == 0 {
			return errors.New("no certificates available to verify")
		}

		cert, err := x509.ParseCertificate(rawCerts[0])
		if err != nil {
			return err
		}

		// When no CA certificate provided, we still need to validate the hostname. Otherwise,
		// we'll want to verify the peer certificate without checking the hostname.
		if pool == nil {
			if err := cert.VerifyHostname(o.Hostname); err != nil {
				return err
			}
		} else if _, err = cert.Verify(x509.VerifyOptions{Roots: pool}); err != nil {
			return err
		}

		return nil
	}
}
