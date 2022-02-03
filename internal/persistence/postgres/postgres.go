package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kong/koko/internal/persistence"
	_ "github.com/lib/pq"
)

const (
	getQuery    = `SELECT * from store where key=$1`
	insertQuery = `insert into store(key,value) values($1,$2) on conflict (key) do update set value=$2;`
	deleteQuery = `delete from store where key=$1`

	defaultMaxConn = 50
	DefaultPort    = 5432
)

// var listQuery = func(prefix string) string {
//	return fmt.Sprintf(`SELECT * FROM store WHERE key LIKE '%s%%';`, prefix)
//}

var listQueryPaging = func(prefix string, limit int, offset int) string {
	return fmt.Sprintf("SELECT key, value, count(*) OVER() as full_count FROM store WHERE key "+
		"LIKE '%s%%' order by key limit %d offset %d;", prefix, limit, offset)
}

type Postgres struct {
	db *sql.DB
}

type Opts struct {
	DBName   string
	Hostname string
	Port     int
	User     string
	Password string
}

func getDSN(opts Opts) string {
	var res string
	if opts.Hostname != "" {
		res += fmt.Sprintf("host=%s ", opts.Hostname)
	}
	if opts.Port != 0 {
		res += fmt.Sprintf("port=%d ", opts.Port)
	}
	if opts.User != "" {
		res += fmt.Sprintf("user=%s ", opts.User)
	}
	if opts.Password != "" {
		res += fmt.Sprintf("password=%s ", opts.Password)
	}
	if opts.DBName != "" {
		res += fmt.Sprintf("dbname=%s ", opts.DBName)
	}
	res += "sslmode=disable"
	return res
}

func NewSQLClient(opts Opts) (*sql.DB, error) {
	dsn := getDSN(opts)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(defaultMaxConn)
	return db, err
}

func New(opts Opts) (persistence.Persister, error) {
	db, err := NewSQLClient(opts)
	if err != nil {
		return nil, err
	}
	res := &Postgres{
		db: db,
	}
	return res, nil
}

func (s *Postgres) withinTx(ctx context.Context,
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

func (s *Postgres) Get(ctx context.Context, key string) ([]byte, error) {
	var res []byte
	err := s.withinTx(ctx, func(tx persistence.Tx) error {
		var err error
		res, err = tx.Get(ctx, key)
		return err
	})
	return res, err
}

func (s *Postgres) Put(ctx context.Context, key string, value []byte) error {
	return s.withinTx(ctx, func(tx persistence.Tx) error {
		return tx.Put(ctx, key, value)
	})
}

func (s *Postgres) Delete(ctx context.Context, key string) error {
	return s.withinTx(ctx, func(tx persistence.Tx) error {
		return tx.Delete(ctx, key)
	})
}

func (s *Postgres) List(ctx context.Context, prefix string, opts *persistence.ListOpts) ([]*persistence.KVResult,
	error) {
	var res []*persistence.KVResult
	err := s.withinTx(ctx, func(tx persistence.Tx) error {
		var err error
		res, err = tx.List(ctx, prefix, opts)
		return err
	})
	return res, err
}

func (s *Postgres) Tx(ctx context.Context) (persistence.Tx, error) {
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
