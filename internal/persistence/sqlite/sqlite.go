package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/cel-go/common/operators"
	"github.com/kong/koko/internal/persistence"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

const (
	getQuery    = `SELECT * from store where key=$1`
	insertQuery = `replace into store(key,value) values($1,$2);`
	deleteQuery = `delete from store where key=$1`
)

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
	query := sq.StatementBuilder.
		PlaceholderFormat(sq.Dollar).
		Select("s.key", "s.value", "COUNT(*) OVER() AS full_count").
		From("store s").
		Where("s.key GLOB ? || '*'", prefix).
		OrderBy("s.key").
		Limit(uint64(opts.Limit)).
		Offset(uint64(opts.Offset))

	// Parse out any provided, pre-validated CEL expression & add the proper clauses to the query.
	var err error
	if query, err = addFilterToQuery(query, opts.Filter); err != nil {
		return persistence.ListResult{}, fmt.Errorf("unable to add filter CEL expression to query: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, t.queryTimeout)
	defer cancel()
	sql, placeholders, err := query.ToSql()
	if err != nil {
		return persistence.ListResult{}, err
	}
	rows, err := t.query.QueryContext(ctx, sql, placeholders...)
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

func addFilterToQuery(query sq.SelectBuilder, expr *exprpb.Expr) (sq.SelectBuilder, error) {
	// No-op when no expression is provided.
	if expr == nil {
		return query, nil
	}

	tags, exprFunction, err := persistence.GetTagsFromExpression(expr)
	if err != nil {
		return query, err
	}

	queryArgs, err := persistence.GetQueryArgsFromExprConstants(tags)
	if err != nil {
		return query, err
	}

	// No-op when there are no tags to filter against, to treat it as if there is no filter at all.
	if len(queryArgs) == 0 {
		return query, nil
	}

	// As the only supported field name to filter on are tags, we can simply hard-code the predicate.
	query = query.
		// Output a row for each tag value.
		Join("json_each(s.value, '$.object.tags') t").
		// The `atom` column contains the result of the `json_each()` function.
		// Read more: https://www.sqlite.org/json1.html#the_json_each_and_json_tree_table_valued_functions
		Where(sq.Eq{"t.atom": queryArgs}).
		GroupBy("s.key")

	// In the event we need to assert that the resource has all provided tags, we'll simply check
	// the number of tags returned. This works as we're assuming duplicates in the DB are not
	// allowed, and any duplicate tags in the expression have already been filtered out.
	if exprFunction == operators.LogicalAnd {
		query = query.Having("COUNT(DISTINCT t.atom) = ?", len(queryArgs))
	}

	return query, nil
}
