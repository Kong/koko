package db

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/kong/koko/internal/persistence/mysql"
	"github.com/kong/koko/internal/persistence/postgres"
	"github.com/kong/koko/internal/persistence/sqlite"
)

// All various supported DB dialects to be used in Koko's database config.
const (
	DialectMariaDB  = "mariadb"
	DialectMySQL    = "mysql"
	DialectPostgres = "postgres"
	DialectSQLite3  = "sqlite3"
)

// Dialects defines all supported DB dialects.
//
// This is internally used in unit tests to ensure support for all dialects have been implemented.
var Dialects = []string{
	DialectMariaDB,
	DialectMySQL,
	DialectPostgres,
	DialectSQLite3,
}

// NewSQLDBFromConfig returns the relevant *sql.DB instance based on the given dialect set on the config.
func NewSQLDBFromConfig(config Config) (*sql.DB, error) {
	var db *sql.DB
	var err error

	switch config.Dialect {
	case DialectMariaDB:
		// See mysql.MySQL on why MariaDB is not supported.
		err = errors.New("MariaDB is currently unsupported")
	case DialectMySQL:
		db, err = mysql.NewSQLClient(config.MySQL, config.Logger)
	case DialectPostgres:
		db, err = postgres.NewSQLClient(config.Postgres, config.Logger)
	case DialectSQLite3:
		db, err = sqlite.NewSQLClient(config.SQLite, config.Logger)
	default:
		err = fmt.Errorf("unsupported database '%v'", config.Dialect)
	}

	return db, err
}
