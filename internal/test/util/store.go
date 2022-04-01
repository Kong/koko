package util

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
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
	dbConfig := testConfig
	dialect := os.Getenv("KOKO_TEST_DB")
	if dialect == "" {
		dialect = "sqlite3"
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	switch dialect {
	case "sqlite3":
		dbClient, err := sqlite.NewSQLClient(dbConfig.SQLite)
		if err != nil {
			t.Fatal(err)
		}
		// store may not exist always so ignore the error
		// TODO(hbagdi): add "IF exists" clause
		_, _ = dbClient.ExecContext(ctx, "delete from store;")

		dbConfig.Dialect = db.DialectSQLite3
	case "postgres":
		dbConfig.Postgres.Logger = dbConfig.Logger
		dbClient, err := postgres.NewSQLClient(dbConfig.Postgres)
		if err != nil {
			t.Fatal(err)
		}
		// store may not exist always so ignore the error
		// TODO(hbagdi): add "IF exists" clause
		_, _ = dbClient.ExecContext(ctx, "truncate table store;")
		dbConfig.Dialect = db.DialectPostgres
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
