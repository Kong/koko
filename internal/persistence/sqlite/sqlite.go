package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kong/koko/internal/persistence"
	_ "github.com/mattn/go-sqlite3"
)

const (
	getQuery    = `SELECT * from store where key=$1`
	insertQuery = `replace into store(key,value) values($1,$2);`
	deleteQuery = `delete from store where key=$1`
)

var listQueryPaging = func(prefix string, limit int, offset int) string {
	return fmt.Sprintf(`SELECT key, value, COUNT(*) OVER() AS full_count FROM 
                               store WHERE key GLOB '%s*' ORDER BY key LIMIT %d OFFSET %d;`,
		prefix, limit, offset)
}

type SQLite struct {
	db *sql.DB
}

type Opts struct {
	Filename string
	InMemory bool
}

const (
	sqliteParams = "?_journal_mode=WAL&_busy_timeout=5000"
)

func getDSN(opts Opts) (string, error) {
	if opts.InMemory {
		return "file::memory:?cache=shared", nil
	}
	if opts.Filename == "" {
		return "", fmt.Errorf("sqlite: no database file name")
	}
	return opts.Filename + sqliteParams, nil
}

func NewSQLClient(opts Opts) (*sql.DB, error) {
	dsn, err := getDSN(opts)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	return db, nil
}

func New(opts Opts) (persistence.Persister, error) {
	db, err := NewSQLClient(opts)
	if err != nil {
		return nil, err
	}

	res := &SQLite{
		db: db,
	}
	return res, nil
}

func (s *SQLite) withinTx(ctx context.Context,
	fn func(tx persistence.Tx) error) error {
	tx, err := s.Tx(ctx)
	if err != nil {
		return err
	}
	err = fn(tx)
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return err
		}
		return err
	}
	return tx.Commit()
}

func (s *SQLite) Get(ctx context.Context, key string) ([]byte, error) {
	var res []byte
	err := s.withinTx(ctx, func(tx persistence.Tx) error {
		var err error
		res, err = tx.Get(ctx, key)
		return err
	})
	return res, err
}

func (s *SQLite) Put(ctx context.Context, key string, value []byte) error {
	return s.withinTx(ctx, func(tx persistence.Tx) error {
		return tx.Put(ctx, key, value)
	})
}

func (s *SQLite) Delete(ctx context.Context, key string) error {
	return s.withinTx(ctx, func(tx persistence.Tx) error {
		return tx.Delete(ctx, key)
	})
}

func (s *SQLite) List(ctx context.Context, prefix string, opts *persistence.ListOpts) ([]*persistence.KVResult, error) {
	var res []*persistence.KVResult
	err := s.withinTx(ctx, func(tx persistence.Tx) error {
		var err error
		res, err = tx.List(ctx, prefix, opts)
		return err
	})
	return res, err
}

func (s *SQLite) Tx(ctx context.Context) (persistence.Tx, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &sqliteTx{tx: tx}, nil
}

type sqliteTx struct {
	tx *sql.Tx
}

func (t *sqliteTx) Commit() error {
	return t.tx.Commit()
}

func (t *sqliteTx) Rollback() error {
	return t.tx.Rollback()
}

func (t *sqliteTx) Get(ctx context.Context, key string) ([]byte, error) {
	row := t.tx.QueryRowContext(ctx, getQuery, key)
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

func (t *sqliteTx) Put(ctx context.Context, key string, value []byte) error {
	res, err := t.tx.ExecContext(ctx, insertQuery, key, value)
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

func (t *sqliteTx) Delete(ctx context.Context, key string) error {
	res, err := t.tx.ExecContext(ctx, deleteQuery, key)
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

func (t *sqliteTx) List(ctx context.Context, prefix string, opts *persistence.ListOpts) ([]*persistence.KVResult,
	error) {
	res := make([]*persistence.KVResult, 0, opts.PageSize)
	rows, err := t.tx.QueryContext(ctx, listQueryPaging(prefix, opts.PageSize, persistence.ToOffset(opts)))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var kvr persistence.KVResult
		err := rows.Scan(&kvr.Key, &kvr.Value, &kvr.TotalCount)
		if err != nil {
			return nil, err
		}
		res = append(res, &kvr)
	}
	return res, nil
}
