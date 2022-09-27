package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	postgres2 "github.com/kong/koko/internal/persistence/postgres"
	"github.com/kong/koko/internal/persistence/sqlite"
	"go.uber.org/zap"
)

type SQLite struct {
	Filename string
	InMemory bool
}

type Config struct {
	Dialect  string
	SQLite   sqlite.Opts
	Postgres postgres2.Opts

	Logger       *zap.Logger
	QueryTimeout time.Duration
}

type Migrator struct {
	m      *migrate.Migrate
	logger *zap.Logger
	config Config
}

func sourceForDialect(dialect string) (source.Driver, error) {
	source, err := iofs.New(migrations, "sql/"+dialect)
	if err != nil {
		return nil, fmt.Errorf("create migrations source: %v", err)
	}
	return source, nil
}

func driverForDialect(dialect string, db *sql.DB) (database.Driver, error) {
	var dbDriver database.Driver
	var err error

	switch dialect {
	case DialectSQLite3:
		dbDriver, err = sqlite3.WithInstance(db, &sqlite3.Config{
			MigrationsTable: sqlite3.DefaultMigrationsTable,
		})
	case DialectPostgres:
		dbDriver, err = postgres.WithInstance(db, &postgres.Config{
			MigrationsTable: postgres.DefaultMigrationsTable,
		})
	default:
		return nil, fmt.Errorf("unsupported database '%v'", dialect)
	}

	return dbDriver, err
}

func NewMigrator(config Config) (*Migrator, error) {
	sqlDB, err := NewSQLDBFromConfig(config)
	if err != nil {
		return nil, err
	}
	dbDriver, err := driverForDialect(config.Dialect, sqlDB)
	if err != nil {
		return nil, err
	}
	source, err := sourceForDialect(config.Dialect)
	if err != nil {
		return nil, err
	}
	migrate, err := migrate.NewWithInstance("iofs", source, config.Dialect, dbDriver)
	if err != nil {
		return nil, fmt.Errorf("create db instance: %v", err)
	}
	res := &Migrator{
		m:      migrate,
		config: config,
		logger: config.Logger,
	}
	migrate.Log = res
	return res, nil
}

func (m *Migrator) Printf(format string, v ...interface{}) {
	m.logger.Sugar().Infof(format, v...)
}

func (m *Migrator) Verbose() bool {
	return false
}

func (m *Migrator) NeedsMigration() (bool, error) {
	current, latest, err := m.Status()
	if err != nil {
		return false, err
	}
	return current != latest, nil
}

func (m *Migrator) Status() (current uint, latest uint, err error) {
	current, _, err = m.m.Version()
	if err != nil {
		if err != migrate.ErrNilVersion {
			return 0, 0, err
		}
	}

	source, err := sourceForDialect(m.config.Dialect)
	if err != nil {
		return 0, 0, err
	}
	latest, err = latestVersion(source)
	if err != nil {
		return 0, 0, fmt.Errorf("retrieve latest db version: %v", err)
	}

	return current, latest, nil
}

func latestVersion(source source.Driver) (uint, error) {
	var nextVersion uint
	v, err := source.First()
	if err != nil {
		return 0, err
	}
	for {
		nextVersion, err = source.Next(v)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return v, nil
			}
			return 0, err
		}
		v = nextVersion
	}
}

func (m *Migrator) Up() error {
	return m.m.Up()
}

func (m *Migrator) Reset() error {
	return m.m.Down()
}

func (m *Migrator) Close() (error, error) {
	return m.m.Close()
}
