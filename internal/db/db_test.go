package db

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/kong/koko/internal/persistence"
	"github.com/kong/koko/internal/persistence/mysql"
	"github.com/kong/koko/internal/persistence/postgres"
	"github.com/kong/koko/internal/persistence/sqlite"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewSQLDBFromConfig(t *testing.T) {
	config := Config{Logger: zap.L()}

	noOpOpenFunc := func(dataSourceName string) (*sql.DB, error) {
		return &sql.DB{}, nil
	}

	t.Run("MariaDB", func(t *testing.T) {
		config.Dialect = DialectMariaDB
		_, err := NewSQLDBFromConfig(config)
		assert.EqualError(t, err, "MariaDB is currently unsupported")
	})

	t.Run("MySQL", func(t *testing.T) {
		config.Dialect = DialectMySQL
		config.MySQL = mysql.Opts{SQLOpen: noOpOpenFunc}
		_, err := NewSQLDBFromConfig(config)
		assert.NoError(t, err)
	})

	t.Run("SQLite", func(t *testing.T) {
		config.Dialect = DialectSQLite3
		config.SQLite = sqlite.Opts{SQLOpen: noOpOpenFunc, InMemory: true}
		_, err := NewSQLDBFromConfig(config)
		assert.NoError(t, err)
	})

	t.Run("Postgres", func(t *testing.T) {
		config.Dialect = DialectPostgres
		poolOpts := postgres.PoolOpts{
			MaxConns:          persistence.DefaultMaxConn,
			HealthCheckPeriod: persistence.DefaultHealthCheckPeriod,
		}
		config.Postgres = postgres.Opts{SQLOpen: noOpOpenFunc, Pool: poolOpts}
		_, err := NewSQLDBFromConfig(config)
		assert.NoError(t, err)
	})

	t.Run("Check for unimplemented dialects", func(t *testing.T) {
		for _, dialect := range Dialects {
			config.Dialect = dialect
			if _, err := NewSQLDBFromConfig(config); err != nil {
				if strings.HasPrefix(err.Error(), "unsupported database") {
					assert.NoError(t, err)
				}
			}
		}
	})
}
