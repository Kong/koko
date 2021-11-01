package util

import (
	"database/sql"
	"os"
	"testing"

	"github.com/kong/koko/internal/persistence"
	"github.com/kong/koko/internal/persistence/postgres"
	"github.com/kong/koko/internal/persistence/sqlite"
	"github.com/stretchr/testify/require"
)

func GetPersister(t *testing.T) persistence.Persister {
	var (
		res persistence.Persister
		err error
	)
	db := os.Getenv("KOKO_TEST_DB")
	if db == "" {
		db = "sqlite3"
	}
	switch db {
	case "sqlite3":
		res, err = sqlite.NewMemory()
	case "postgres":
		require.Nil(t, cleanPostgres())
		res, err = postgres.New(postgres.Opts{
			Hostname: "localhost",
			Port:     postgres.DefaultPort,
			User:     "koko",
			Password: "koko",
		})
	}
	require.Nil(t, err)

	return res
}

func cleanPostgres() error {
	db, err := sql.Open("postgres",
		"host=localhost user=koko password=koko port=5432 sslmode=disable")
	if err != nil {
		return err
	}
	// store may not exist always so ignore the error
	// TODO(hbagdi): add "IF exists" clause
	_, _ = db.Exec("truncate table store")
	return nil
}
