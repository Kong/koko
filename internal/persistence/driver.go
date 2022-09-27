package persistence

import (
	"database/sql"
	"time"
)

// Various default settings used for non-SQLite DB drivers.
const (
	DefaultDialTimeout     = 15 * time.Second
	DefaultMaxConn         = 50
	DefaultMaxConnLifetime = time.Hour
	DefaultMaxIdleConn     = 20
)

// SQLOpenFunc defines a function that can instantiate a sql.DB instance from the given DSN.
type SQLOpenFunc func(dataSourceName string) (*sql.DB, error)

// Driver represents a specific SQL driver.
type Driver int

// Supported DB drivers.
const (
	SQLite3 Driver = iota
	Postgres
)

// DefaultSQLOpenFunc generates a default SQLOpenFunc function for the given SQL persister.
var DefaultSQLOpenFunc = func(p SQLPersister) SQLOpenFunc {
	return func(dsn string) (*sql.DB, error) {
		db, err := sql.Open(p.Driver().String(), dsn)
		if err != nil {
			return nil, err
		}

		if err := p.SetDefaultSQLOptions(db); err != nil {
			return nil, err
		}

		return db, nil
	}
}

// String returns the applicable registered DB driver name (set via sql.Register).
func (d Driver) String() string {
	return [...]string{"sqlite3", "pgx"}[d]
}
