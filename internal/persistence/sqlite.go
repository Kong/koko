package persistence

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

const (
	createTableQuery = `create table if not exists store(
key text PRIMARY KEY, value BLOB);`

	getQuery    = `SELECT * from store where key=$1`
	insertQuery = `replace into store(key,value) values(?,?);`
	deleteQuery = `delete from store where key=?`
)

var listQuery = func(prefix string) string {
	return fmt.Sprintf(`SELECT * from store where key glob '%s*'`, prefix)
}

type SQLite struct {
	db *sql.DB
}

func NewMemory() (Persister, error) {
	return NewSQLite("file::memory:")
}

func NewSQLite(filename string) (Persister, error) {
	db, err := sql.Open("sqlite3",
		filename+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}
	err = migrate(context.TODO(), db)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	res := &SQLite{
		db: db,
	}
	return res, nil
}

func migrate(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, createTableQuery)
	return err
}

func (s *SQLite) withinTx(ctx context.Context,
	fn func(tx Tx) error) error {
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
	err := s.withinTx(ctx, func(tx Tx) error {
		var err error
		res, err = tx.Get(ctx, key)
		return err
	})
	return res, err
}

func (s *SQLite) Put(ctx context.Context, key string, value []byte) error {
	return s.withinTx(ctx, func(tx Tx) error {
		return tx.Put(ctx, key, value)
	})
}

func (s *SQLite) Delete(ctx context.Context, key string) error {
	return s.withinTx(ctx, func(tx Tx) error {
		return tx.Delete(ctx, key)
	})
}

func (s *SQLite) List(ctx context.Context, prefix string) ([][]byte, error) {
	var res [][]byte
	err := s.withinTx(ctx, func(tx Tx) error {
		var err error
		res, err = tx.List(ctx, prefix)
		return err
	})
	return res, err
}

func (s *SQLite) Tx(ctx context.Context) (Tx, error) {
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
			return nil, ErrNotFound{key: key}
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
		return ErrNotFound{key: key}
	}
	if rowCount != 1 {
		return fmt.Errorf("invalid rows affected")
	}
	return nil
}

func (t *sqliteTx) List(ctx context.Context, prefix string) ([][]byte, error) {
	var res [][]byte
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
		var resKey string
		var value []byte
		err := rows.Scan(&resKey, &value)
		if err != nil {
			return nil, err
		}
		res = append(res, value)
	}
	return res, nil
}
