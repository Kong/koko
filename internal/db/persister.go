package db

import (
	"fmt"

	"github.com/kong/koko/internal/persistence"
	"github.com/kong/koko/internal/persistence/postgres"
	"github.com/kong/koko/internal/persistence/sqlite"
)

func NewPersister(config Config) (persistence.Persister, error) {
	var (
		persister persistence.Persister
		err       error
	)
	switch config.Dialect {
	case DialectSQLite3:
		persister, err = sqlite.New(config.SQLite)
		if err != nil {
			return nil, err
		}
	case DialectPostgres:
		persister, err = postgres.New(config.Postgres)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown database: %v", config.Dialect)
	}
	return persister, nil
}
