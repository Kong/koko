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

var listQuery = func(prefix string) string {
	return fmt.Sprintf(`SELECT * from store where key glob '%s*'`, prefix)
}

var listQueryPaging = func(prefix string, limit int, offset int) string {
	return fmt.Sprintf(`SELECT * FROM store where key glob '%s*' order by key limit %d offset %d;`,
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

func (s *SQLite) List(ctx context.Context, prefix string) ([][2][]byte, error) {
	var res [][2][]byte
	err := s.withinTx(ctx, func(tx persistence.Tx) error {
		var err error
		res, err = tx.List(ctx, prefix)
		return err
	})
	return res, err
}

func (s *SQLite) ListWithPaging(ctx context.Context, prefix string, limit int, offset int) ([][2][]byte, error) {
	var res [][2][]byte
	err := s.withinTx(ctx, func(tx persistence.Tx) error {
		var err error
		res, err = tx.ListWithPaging(ctx, prefix, limit, offset)
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

func (t *sqliteTx) List(ctx context.Context, prefix string) ([][2][]byte,
	error) {
	var res [][2][]byte
	rows, err := t.tx.QueryContext(ctx, listQuery(prefix))
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
		var (
			resKey []byte
			value  []byte
		)
		err := rows.Scan(&resKey, &value)
		if err != nil {
			return nil, err
		}
		res = append(res, [2][]byte{resKey, value})
	}
	return res, nil
}

func (t *sqliteTx) ListWithPaging(ctx context.Context, prefix string, limit int, offset int) ([][2][]byte, error) {
	var res [][2][]byte
	rows, err := t.tx.QueryContext(ctx, listQueryPaging(prefix, limit, offset))
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
		var (
			resKey []byte
			value  []byte
		)
		err := rows.Scan(&resKey, &value)
		if err != nil {
			return nil, err
		}
		res = append(res, [2][]byte{resKey, value})
	}
	return res, nil
}
