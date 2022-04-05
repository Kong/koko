package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kong/koko/internal/persistence"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

const (
	getQuery    = `SELECT * from store where key=$1`
	insertQuery = `insert into store(key,value) values($1,$2) on conflict (key) do update set value=$2;`
	deleteQuery = `delete from store where key=$1`

	defaultMaxConn = 50
	DefaultPort    = 5432
)

var listQueryPaging = `SELECT key, value, COUNT(*) OVER() AS full_count FROM store WHERE key
                       LIKE $1 || '%%' ORDER BY key LIMIT $2 OFFSET $3;`

type Postgres struct {
	db           *sql.DB
	queryTimeout time.Duration
}

type Opts struct {
	DBName         string
	Hostname       string
	Port           int
	User           string
	Password       string
	EnableTLS      bool
	CABundleFSPath string
}

func getDSN(opts Opts, logger *zap.Logger) string {
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
	if !opts.EnableTLS {
		logger.Info("Using non-TLS Postgres connection")
		res += "sslmode=disable"
		return res
	}
	logger.Info("Using TLS Postgres connection")
	logger.Info("ca_bundle_fs_path:" + opts.CABundleFSPath)
	if opts.CABundleFSPath == "" {
		panic("Postgres connection requires TLS but ca_bundle_fs_path is empty")
	}
	res += "sslmode=verify-full sslrootcert=" + opts.CABundleFSPath

	return res
}

func NewSQLClient(opts Opts, logger *zap.Logger) (*sql.DB, error) {
	dsn := getDSN(opts, logger)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(defaultMaxConn)
	return db, err
}

func New(opts Opts, queryTimeout time.Duration, logger *zap.Logger) (persistence.Persister, error) {
	db, err := NewSQLClient(opts, logger)
	if err != nil {
		return nil, err
	}
	res := &Postgres{
		db:           db,
		queryTimeout: queryTimeout,
	}
	return res, nil
}

func (s *Postgres) withinTx(ctx context.Context,
	fn func(tx persistence.Tx) error,
) error {
	ctx, cancel := context.WithTimeout(ctx, s.queryTimeout)
	defer cancel()
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

func (s *Postgres) List(ctx context.Context, prefix string, opts *persistence.ListOpts) (persistence.ListResult,
	error,
) {
	var res persistence.ListResult
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
	return &sqliteTx{tx: tx, queryTimeout: s.queryTimeout}, nil
}

func (s *Postgres) Close() error {
	return s.db.Close()
}

type sqliteTx struct {
	tx           *sql.Tx
	queryTimeout time.Duration
}

func (t *sqliteTx) Commit() error {
	return t.tx.Commit()
}

func (t *sqliteTx) Rollback() error {
	return t.tx.Rollback()
}

func (t *sqliteTx) Get(ctx context.Context, key string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, t.queryTimeout)
	defer cancel()
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
	ctx, cancel := context.WithTimeout(ctx, t.queryTimeout)
	defer cancel()
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
	ctx, cancel := context.WithTimeout(ctx, t.queryTimeout)
	defer cancel()
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

func (t *sqliteTx) List(ctx context.Context, prefix string, opts *persistence.ListOpts) (persistence.ListResult,
	error,
) {
	ctx, cancel := context.WithTimeout(ctx, t.queryTimeout)
	defer cancel()
	kvlist := make([]persistence.KVResult, 0, opts.Limit)
	rows, err := t.tx.QueryContext(ctx, listQueryPaging, prefix, opts.Limit, opts.Offset)
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
	res := persistence.ListResult{}
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
