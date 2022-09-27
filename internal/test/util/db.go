package util

import (
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/kong/koko/internal/config"
	"github.com/kong/koko/internal/db"
	"github.com/kong/koko/internal/persistence/postgres"
)

// GetAppDatabaseConfig returns the application database configuration, which
// is usually provided by means of a config file and/or environment variables.
//
// This will set defaulted values when no other DB config is set.
func GetAppDatabaseConfig() (config.Database, error) {
	var appConfig config.Config
	if err := cleanenv.ReadEnv(&appConfig); err != nil {
		return config.Database{}, fmt.Errorf("unable to read config from environment: %w", err)
	}

	if err := setDBConfig(&appConfig.Database); err != nil {
		return config.Database{}, fmt.Errorf("unable to set DB config: %w", err)
	}

	return appConfig.Database, nil
}

// GetDatabaseConfig is like GetAppDatabaseConfig(), but it converts the
// app database config to an internal representation of the config.
func GetDatabaseConfig() (db.Config, error) {
	appDBConfig, err := GetAppDatabaseConfig()
	if err != nil {
		return db.Config{}, err
	}

	return config.ToDBConfig(appDBConfig, nil)
}

// setDBConfig determines what DB settings to used based on the environment.
//
// We'll assume that when a hostname is provided via the environment, to not use any defaulted config.
func setDBConfig(conf *config.Database) error {
	if conf.QueryTimeout == "" {
		conf.QueryTimeout = queryTimeout.String()
	}

	// TODO(tjasko): The legacy `KOKO_TEST_DB` environment variable takes precedence over `KOKO_DATABASE_DIALECT`.
	//  However, we do need to update the codebase to use the newer environment variable.
	if testDB := os.Getenv("KOKO_TEST_DB"); testDB != "" {
		conf.Dialect = testDB
	} else if conf.Dialect == "" {
		conf.Dialect = db.DialectSQLite3
	}

	switch conf.Dialect {
	case db.DialectSQLite3:
		conf.SQLite = config.SQLite{InMemory: true}
	case db.DialectPostgres:
		if conf.Postgres.Hostname == "" {
			conf.Postgres = config.Postgres{
				DBName:   "koko",
				Hostname: "localhost",
				Port:     postgres.DefaultPort,
				User:     "koko",
				Password: "koko",
			}
		}
	default:
		return fmt.Errorf("unknown DB dialect: %s", conf.Dialect)
	}

	return nil
}
