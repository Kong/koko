package config

import (
	"fmt"
	"time"

	"github.com/kong/koko/internal/db"
	"github.com/kong/koko/internal/persistence/mysql"
	"github.com/kong/koko/internal/persistence/postgres"
	"github.com/kong/koko/internal/persistence/sqlite"
	"go.uber.org/zap"
)

// Database allows for configuration of Koko's datastore.
type Database struct {
	// See db.Dialects for all supported Dialects. This defines which
	// DB configuration is used, even if multiple DBs are provided.
	Dialect string `yaml:"dialect" json:"dialect" env:"DIALECT" env-default:"sqlite3"`

	QueryTimeout string `yaml:"query_timeout" json:"query_timeout" env:"QUERY_TIMEOUT" env-default:"5s"`

	MySQL    MySQL    `yaml:"mysql" json:"mysql" env-prefix:"MYSQL_"`
	SQLite   SQLite   `yaml:"sqlite" json:"sqlite" env-prefix:"SQLITE_"`
	Postgres Postgres `yaml:"postgres" json:"postgres" env-prefix:"POSTGRES_"`
}

// MySQL defines configuration for using MySQL as the persistent store.
type MySQL struct {
	DBName      string           `yaml:"db_name" json:"db_name" env:"DB_NAME"`
	Hostname    string           `yaml:"hostname" json:"hostname" env:"HOSTNAME"`
	ReadReplica MySQLReadReplica `yaml:"read_replica" json:"read_replica" env-prefix:"READ_REPLICA_"`
	Port        int              `yaml:"port" json:"port" env:"PORT"`
	User        string           `yaml:"user" json:"user" env:"USER"`
	Password    string           `yaml:"password" json:"password" env:"PASSWORD"`
	TLS         TLS              `yaml:"tls" json:"tls" env-prefix:"TLS_"`

	// See the `Params` field on mysql.Opts for more info.
	Params map[string]string `yaml:"params" json:"params" env:"PARAMS"`
}

// MySQLReadReplica defines configuration for specifying a single read-replica.
// This configuration overrides fields set in config.MySQL.
type MySQLReadReplica struct {
	Hostname string `yaml:"hostname" json:"hostname" env:"HOSTNAME"`
}

// Postgres defines configuration for using Postgres as the persistent store.
type Postgres struct {
	DBName      string              `yaml:"db_name" json:"db_name" env:"DB_NAME"`
	Hostname    string              `yaml:"hostname" json:"hostname" env:"HOSTNAME"`
	Port        int                 `yaml:"port" json:"port" env:"PORT"`
	ReadReplica PostgresReadReplica `yaml:"read_replica" json:"read_replica" env-prefix:"READ_REPLICA_"`
	TLS         PostgresTLS         `yaml:"tls" json:"tls" env-prefix:"TLS_"`
	User        string              `yaml:"user" json:"user" env:"USER"`
	Password    string              `yaml:"password" json:"password" env:"PASSWORD"`
}

// PostgresTLS defines configuration for using TLS with Postgres.
type PostgresTLS struct {
	CABundlePath string `yaml:"ca_bundle_path" json:"ca_bundle_path" env:"CA_BUNDLE_PATH"`
	Enable       bool   `yaml:"enable" json:"enable" env:"ENABLE"`
}

// PostgresReadReplica allows for using a read replica in addition to the
// primary which shares the same connection settings as the primary DB.
type PostgresReadReplica struct {
	Hostname string `yaml:"hostname" json:"hostname" env:"HOSTNAME"`
}

// SQLite defines configuration for using SQLite as the persistent store.
type SQLite struct {
	Filename string `yaml:"filename" json:"filename" env:"FILENAME"`
	InMemory bool   `yaml:"in_memory" json:"in_memory" env:"IN_MEMORY"`
}

// Opts returns the options required to instantiate a MySQL persistence.Persister.
func (c *MySQL) Opts() mysql.Opts {
	opts := mysql.Opts{
		DBName:           c.DBName,
		EnableTLS:        c.TLS.Enable,
		Hostname:         c.Hostname,
		ReadOnlyHostname: c.ReadReplica.Hostname,
		Params:           c.Params,
		Port:             c.Port,
		User:             c.User,
		Password:         c.Password,

		// TODO(tjasko): Implement me.
		Certificates: nil,
		RootCAs:      nil,
	}

	if c.TLS.VerifyPeerCertificate {
		opts.VerifyPeerCertificateFunc = mysql.VerifyPeerCertFunc
	}

	return opts
}

// Opts returns the options required to instantiate a Postgres persistence.Persister.
func (c *Postgres) Opts() postgres.Opts {
	if c.Port == 0 {
		c.Port = postgres.DefaultPort
	}

	return postgres.Opts{
		CABundleFSPath:   c.TLS.CABundlePath,
		DBName:           c.DBName,
		EnableTLS:        c.TLS.Enable,
		Port:             c.Port,
		Hostname:         c.Hostname,
		ReadOnlyHostname: c.ReadReplica.Hostname,
		User:             c.User,
		Password:         c.Password,
	}
}

// Opts returns the options required to instantiate a SQLite persistence.Persister.
func (c *SQLite) Opts() sqlite.Opts {
	return sqlite.Opts{
		InMemory: c.InMemory,
		Filename: c.Filename,
	}
}

// ToDBConfig maps the provided DB application config to the internal representation of the DB config.
// The resulting config will have its DB config set based on the passed in Database.Dialect.
func ToDBConfig(config Database, logger *zap.Logger) (db.Config, error) {
	if logger == nil {
		logger = zap.L()
	}

	queryTimeout, err := time.ParseDuration(config.QueryTimeout)
	if err != nil {
		return db.Config{}, fmt.Errorf("failed to parse DB query timeout: %w", err)
	}

	return db.Config{
		Dialect:      config.Dialect,
		QueryTimeout: queryTimeout,
		Logger:       logger,
		MySQL:        config.MySQL.Opts(),
		Postgres:     config.Postgres.Opts(),
		SQLite:       config.SQLite.Opts(),
	}, nil
}
