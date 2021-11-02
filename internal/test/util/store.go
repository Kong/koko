package util

import (
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/kong/koko/internal/db"
	"github.com/kong/koko/internal/log"
	"github.com/kong/koko/internal/persistence"
	"github.com/kong/koko/internal/persistence/postgres"
	"github.com/kong/koko/internal/persistence/sqlite"
	"github.com/stretchr/testify/require"
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
	Logger: log.Logger,
}

func GetPersister(t *testing.T) persistence.Persister {
	dbConfig := testConfig
	dialect := os.Getenv("KOKO_TEST_DB")
	if dialect == "" {
		dialect = "sqlite3"
	}
	switch dialect {
	case "sqlite3":
		dbClient, err := sqlite.NewSQLClient(dbConfig.SQLite)
		require.Nil(t, err)
		// store may not exist always so ignore the error
		// TODO(hbagdi): add "IF exists" clause
		_, _ = dbClient.Exec("delete from store;")

		dbConfig.Dialect = db.DialectSQLite3
	case "postgres":
		dbClient, err := postgres.NewSQLClient(dbConfig.Postgres)
		require.Nil(t, err)
		// store may not exist always so ignore the error
		// TODO(hbagdi): add "IF exists" clause
		_, _ = dbClient.Exec("truncate table store;")

		dbConfig.Dialect = db.DialectPostgres
	}

	require.Nil(t, runMigrations(dbConfig))
	persister, err := db.NewPersister(dbConfig)
	require.Nil(t, err)
	return persister
}

func runMigrations(config db.Config) error {
	m, err := db.NewMigrator(config)
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
