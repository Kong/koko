package util

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/kong/koko/internal/config"
	"github.com/kong/koko/internal/db"
	"github.com/kong/koko/internal/log"
	"github.com/kong/koko/internal/persistence"
	"github.com/kong/koko/internal/persistence/postgres"
	"github.com/kong/koko/internal/persistence/sqlite"
)

const (
	queryTimeout = 3 * time.Second
)

var testConfig = db.Config{
	SQLite: sqlite.Opts{
		InMemory: true,
	},
	Postgres: postgres.Opts{
		Hostname: "localhost",
		Port:     postgres.DefaultPort,
		User:     "koko",
		Password: "koko",
		DBName:   "koko",
	},
	Logger:       log.Logger,
	QueryTimeout: queryTimeout,
}

func CleanDB(t *testing.T) error {
	_, err := GetPersister(t)
	return err
}

func GetPersister(t *testing.T) (persistence.Persister, error) {
	var appConfig config.Config
	if err := cleanenv.ReadEnv(&appConfig); err != nil {
		return nil, fmt.Errorf("unable to read config from environment: %w", err)
	}

	if err := setDBConfig(&appConfig.Database); err != nil {
		return nil, fmt.Errorf("unable to set DB config: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbConfig, err := GetDatabaseConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to gather test DB config: %w", err)
	}

	dbClient, err := db.NewSQLDBFromConfig(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve SQL DB instance from the given config: %w", err)
	}

	// store may not exist always so ignore the error
	// TODO(hbagdi): add "IF exists" clause
	switch dbConfig.Dialect {
	case db.DialectMariaDB, db.DialectMySQL:
		_, _ = dbClient.ExecContext(ctx, "truncate table store;")
	case db.DialectSQLite3:
		_, _ = dbClient.ExecContext(ctx, "delete from store;")
	case db.DialectPostgres:
		_, _ = dbClient.ExecContext(ctx, "truncate table store;")
	}

	if err := runMigrations(dbConfig); err != nil {
		return nil, err
	}
	persister, err := db.NewPersister(dbConfig)
	if err != nil {
		return nil, err
	}
	t.Cleanup(func() {
		persister.Close()
	})
	return persister, nil
}

func runMigrations(config db.Config) error {
	m, err := db.NewMigrator(config)
	if err != nil {
		return err
	}
	defer m.Close()

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
