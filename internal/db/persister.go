package db

import (
	"fmt"

	"github.com/kong/koko/internal/persistence"
	"github.com/kong/koko/internal/persistence/mysql"
	"github.com/kong/koko/internal/persistence/postgres"
	"github.com/kong/koko/internal/persistence/sqlite"
)

func NewPersister(config Config) (persistence.Persister, error) {
	var (
		persister persistence.Persister
		err       error
	)
	switch config.Dialect {
	case DialectMariaDB:
		// See mysql.MySQL on why MariaDB is not supported.
		err = mysql.ErrMariaDBUnsupported
	case DialectMySQL:
		persister, err = mysql.New(config.MySQL, config.QueryTimeout, config.Logger)
	case DialectSQLite3:
		persister, err = sqlite.New(config.SQLite, config.QueryTimeout, config.Logger)
	case DialectPostgres:
		persister, err = postgres.New(config.Postgres, config.QueryTimeout, config.Logger)
	default:
		err = fmt.Errorf("unsupported database: %v", config.Dialect)
	}

	return persister, err
}
