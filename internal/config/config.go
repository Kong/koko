package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/kong/koko/internal/db"
	"github.com/kong/koko/internal/persistence/postgres"
	"github.com/kong/koko/internal/persistence/sqlite"
)

// Get constructs the Config using the filename, env vars and defaults.
func Get(filename string) (Config, error) {
	var c Config
	if filename != "" {
		if _, err := os.Stat(filename); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				filename = ""
			}
		}
	}
	var err error
	if filename == "" {
		err = cleanenv.ReadEnv(&c)
	} else {
		err = cleanenv.ReadConfig(filename, &c)
	}
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	return c, nil
}

func ToDBConfig(configDB Database) (db.Config, error) {
	queryTimeout, err := time.ParseDuration(configDB.QueryTimeout)
	if err != nil {
		return db.Config{}, fmt.Errorf("failed to parse query timeout: %v", err)
	}
	return db.Config{
		Dialect: configDB.Dialect,
		SQLite: sqlite.Opts{
			InMemory: configDB.SQLite.InMemory,
			Filename: configDB.SQLite.Filename,
		},
		Postgres: postgres.Opts{
			DBName:           configDB.Postgres.DBName,
			Hostname:         configDB.Postgres.Hostname,
			ReadOnlyHostname: configDB.Postgres.ReadReplica.Hostname,
			Port:             configDB.Postgres.Port,
			User:             configDB.Postgres.User,
			Password:         configDB.Postgres.Password,
			EnableTLS:        configDB.Postgres.TLS.Enable,
			CABundleFSPath:   configDB.Postgres.TLS.CABundlePath,
		},
		QueryTimeout: queryTimeout,
	}, nil
}
