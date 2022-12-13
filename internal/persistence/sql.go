package persistence

import (
	"database/sql"
)

// WildcardOperator is used to represent a SQL wildcard (`%`) within a key.
const WildcardOperator = "\u0000"

// SQLPersister should be implemented for all persistence stores that implement an underlining sql.DB driver.
type SQLPersister interface {
	// Driver returns the relevant Golang SQL driver used to connect to the database.
	Driver() Driver

	// SetDefaultSQLOptions is used to set default connection options.
	// Such options can be overridden by DefaultSQLOpenFunc.
	SetDefaultSQLOptions(*sql.DB) error
}
