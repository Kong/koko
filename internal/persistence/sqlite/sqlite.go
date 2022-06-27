package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kong/koko/internal/persistence"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

const (
	getQuery    = `SELECT * from store where key=$1`
	insertQuery = `replace into store(key,value) values($1,$2);`
	deleteQuery = `delete from store where key=$1`
)

var listQueryPaging = `SELECT key, value, COUNT(*) OVER() AS full_count FROM 
                        store WHERE key GLOB $1 || '*' ORDER BY key LIMIT $2 OFFSET $3;`

type SQLite struct {
	db           *sql.DB
	queryTimeout time.Duration
}

type Opts struct {
	Filename string
	InMemory bool
	SQLOpen  func(driver persistence.Driver, dataSourceName string) (*sql.DB, error)
}

const (
	sqliteParams = "?_journal_mode=WAL&_busy_timeout=5000"
)

func getDSN(opts Opts, logger *zap.Logger) (string, error) {
	logger.Info("using SQLite Database")
	if opts.InMemory {
		return "file::memory:?cache=shared", nil
	}
	if opts.Filename == "" {
		return "", fmt.Errorf("sqlite: no database file name")
	}
	return opts.Filename + sqliteParams, nil
}

func NewSQLClient(opts Opts, logger *zap.Logger) (*sql.DB, error) {
	dsn, err := getDSN(opts, logger)
	if err != nil {
		return nil, err
	}

	open := func(driver persistence.Driver, dsn string) (*sql.DB, error) {
		return sql.Open(driver.String(), dsn)
	}
	if opts.SQLOpen != nil {
		open = opts.SQLOpen
	}

	db, err := open(persistence.SQLite3, dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	return db, nil
}

func New(opts Opts, queryTimeout time.Duration, logger *zap.Logger) (persistence.Persister, error) {
	db, err := NewSQLClient(opts, logger)
	if err != nil {
		return nil, err
	}

	res := &SQLite{
		db:           db,
		queryTimeout: queryTimeout,
	}
	return res, nil
}

func (s *SQLite) Get(ctx context.Context, key string) ([]byte, error) {
	q := sqliteQuery{query: s.db, queryTimeout: s.queryTimeout}
	return q.Get(ctx, key)
}

func (s *SQLite) Put(ctx context.Context, key string, value []byte) error {
	q := sqliteQuery{query: s.db, queryTimeout: s.queryTimeout}
	return q.Put(ctx, key, value)
}

func (s *SQLite) Delete(ctx context.Context, key string) error {
	q := sqliteQuery{query: s.db, queryTimeout: s.queryTimeout}
	return q.Delete(ctx, key)
}

func (s *SQLite) List(ctx context.Context, prefix string, opts *persistence.ListOpts) (persistence.ListResult, error) {
	q := sqliteQuery{query: s.db, queryTimeout: s.queryTimeout}
	return q.List(ctx, prefix, opts)
}

func (s *SQLite) Tx(ctx context.Context) (persistence.Tx, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &sqliteTx{
		tx: tx,
		query: sqliteQuery{
			query:        tx,
			queryTimeout: s.queryTimeout,
		},
	}, nil
}

func (s *SQLite) Close() error {
	return s.db.Close()
}

type sqliteTx struct {
	tx    *sql.Tx
	query sqliteQuery
}

func (t *sqliteTx) Commit() error {
	return t.tx.Commit()
}

func (t *sqliteTx) Rollback() error {
	return t.tx.Rollback()
}

func (t *sqliteTx) Get(ctx context.Context, key string) ([]byte, error) {
	return t.query.Get(ctx, key)
}

func (t *sqliteTx) Put(ctx context.Context, key string, value []byte) error {
	return t.query.Put(ctx, key, value)
}

func (t *sqliteTx) Delete(ctx context.Context, key string) error {
	return t.query.Delete(ctx, key)
}

func (t *sqliteTx) List(
	ctx context.Context,
	prefix string,
	opts *persistence.ListOpts,
) (persistence.ListResult, error) {
	return t.query.List(ctx, prefix, opts)
}

type query interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type sqliteQuery struct {
	query        query
	queryTimeout time.Duration
}

func (t *sqliteQuery) Get(ctx context.Context, key string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, t.queryTimeout)
	defer cancel()
	row := t.query.QueryRowContext(ctx, getQuery, key)
	var resKey string
	var value []byte
	err := row.Scan(&resKey, &value)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, persistence.ErrNotFound{Key: key}
		}
		return nil, err
	}
	return value, err
}

func (t *sqliteQuery) Put(ctx context.Context, key string, value []byte) error {
	ctx, cancel := context.WithTimeout(ctx, t.queryTimeout)
	defer cancel()
	res, err := t.query.ExecContext(ctx, insertQuery, key, value)
	if err != nil {
		return err
	}
	rowCount, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowCount != 1 {
		return fmt.Errorf("invalid rows affected")
	}
	return nil
}

func (t *sqliteQuery) Delete(ctx context.Context, key string) error {
	ctx, cancel := context.WithTimeout(ctx, t.queryTimeout)
	defer cancel()
	res, err := t.query.ExecContext(ctx, deleteQuery, key)
	if err != nil {
		return err
	}
	rowCount, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowCount == 0 {
		return persistence.ErrNotFound{Key: key}
	}
	if rowCount != 1 {
		return fmt.Errorf("invalid rows affected")
	}
	return nil
}

func (t *sqliteQuery) List(ctx context.Context, prefix string,
	opts *persistence.ListOpts,
) (persistence.ListResult, error) {
	ctx, cancel := context.WithTimeout(ctx, t.queryTimeout)
	defer cancel()
	rows, err := t.query.QueryContext(ctx, listQueryPaging, prefix, opts.Limit, opts.Offset)
	if err != nil {
		return persistence.ListResult{}, err
	}
	defer rows.Close()
	if rows.Err() != nil {
		return persistence.ListResult{}, rows.Err()
	}
	if err != nil {
		return persistence.ListResult{}, err
	}
	var res persistence.ListResult
	kvlist := make([]persistence.KVResult, 0, opts.Limit)
	for rows.Next() {
		var kvr persistence.KVResult
		err := rows.Scan(&kvr.Key, &kvr.Value, &res.TotalCount)
		if err != nil {
			return persistence.ListResult{}, err
		}
		kvlist = append(kvlist, kvr)
	}
	res.KVList = kvlist
	return res, nil
}
