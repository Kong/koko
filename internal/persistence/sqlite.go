package persistence

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type SQLite struct {
	db *sql.DB
}

func NewSQLite(filename string) (Persister, error) {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}
	err = migrate(context.TODO(), db)
	if err != nil {
		return nil, err
	}

	res := &SQLite{
		db: db,
	}
	return res, nil
}

const (
	createTableQuery = `create table if not exists store(
key text PRIMARY KEY, 
value BLOB);`

	getQuery    = `SELECT * from store where key=$1`
	insertQuery = `replace into store(key,value) values(?,?);`
	deleteQuery = `delete from store where key=?`
	ListQuery   = `SELECT * from store where key glob '?*'`
)

func migrate(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, createTableQuery)
	return err
}

func (s *SQLite) Get(ctx context.Context, key string) ([]byte, error) {
	row := s.db.QueryRowContext(ctx, getQuery, key)
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

func (s *SQLite) Put(ctx context.Context, key string, value []byte) error {
	res, err := s.db.ExecContext(ctx, insertQuery, key, value)
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

func (s *SQLite) Delete(ctx context.Context, key string) error {
	res, err := s.db.ExecContext(ctx, deleteQuery, key)
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

func (s *SQLite) List(ctx context.Context, prefix string) ([][]byte, error) {
	var res [][]byte
	rows, err := s.db.QueryContext(ctx, ListQuery, prefix)
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
