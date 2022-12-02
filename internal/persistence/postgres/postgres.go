package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/cel-go/common/operators"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/kong/koko/internal/persistence"
	"go.uber.org/zap"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

const (
	getQuery    = `SELECT * from store where key=$1`
	insertQuery = `insert into store(key,value) values($1,$2) on conflict (key) do update set value=$2;`
	deleteQuery = `delete from store where key=$1`

	DefaultPort = 5432
	DefaultPool = "pgx"
)

type Postgres struct {
	dbPool         Pool
	readOnlyDBPool Pool
	queryTimeout   time.Duration
}

// NewSQLClient creates a standard database/sql dbPool client, used for migrations.
func NewSQLClient(opts Opts, logger *zap.Logger) (*sql.DB, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	dsn, err := opts.DSN(logger)
	if err != nil {
		return nil, err
	}

	openFunc := persistence.DefaultSQLOpenFunc(&Postgres{})
	if opts.SQLOpen != nil {
		openFunc = opts.SQLOpen
	}

	return openFunc(dsn)
}

func New(opts Opts, queryTimeout time.Duration, logger *zap.Logger) (persistence.Persister, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	dbPool, err := newPostgresPool(opts, logger)
	if err != nil {
		return nil, fmt.Errorf("unable to set up DB client: %w", err)
	}
	// by default, fallback to primary host for read operations
	readOnlyDBPool := dbPool
	if opts.ReadOnlyHostname != "" {
		readOnlyOpts := opts
		readOnlyOpts.Hostname = opts.ReadOnlyHostname
		readOnlyOpts.Pool.ReadOnly = true
		readOnlyDBPool, err = newPostgresPool(readOnlyOpts, logger)
		if err != nil {
			return nil, fmt.Errorf("unable to set up read-only DB client: %w", err)
		}
	}
	res := &Postgres{
		dbPool:         dbPool,
		readOnlyDBPool: readOnlyDBPool,
		queryTimeout:   queryTimeout,
	}
	return res, nil
}

// Driver implements the persistence.SQLPersister interface.
func (s *Postgres) Driver() persistence.Driver { return persistence.Postgres }

// SetDefaultSQLOptions implements the persistence.SQLPersister interface.
func (s *Postgres) SetDefaultSQLOptions(db *sql.DB) error {
	// default settings are set here because the standard DB client is only used for migrations.
	db.SetMaxOpenConns(persistence.DefaultMaxConn)
	db.SetMaxIdleConns(persistence.DefaultMaxIdleConn)
	db.SetConnMaxLifetime(persistence.DefaultMaxConnLifetime)
	return nil
}

func (s *Postgres) Get(ctx context.Context, key string) ([]byte, error) {
	q := postgresQuery{query: s.readOnlyDBPool, queryTimeout: s.queryTimeout}
	return q.Get(ctx, key)
}

func (s *Postgres) Put(ctx context.Context, key string, value []byte) error {
	q := postgresQuery{query: s.dbPool, queryTimeout: s.queryTimeout}
	return q.Put(ctx, key, value)
}

func (s *Postgres) Delete(ctx context.Context, key string) error {
	q := postgresQuery{query: s.dbPool, queryTimeout: s.queryTimeout}
	return q.Delete(ctx, key)
}

func (s *Postgres) List(ctx context.Context, prefix string, opts *persistence.ListOpts) (persistence.ListResult,
	error,
) {
	q := postgresQuery{query: s.readOnlyDBPool, queryTimeout: s.queryTimeout}
	return q.List(ctx, prefix, opts)
}

func (s *Postgres) Tx(ctx context.Context) (persistence.Tx, error) {
	tx, err := s.dbPool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	return &postgresTx{
		ctx: ctx,
		tx:  tx,
		query: postgresQuery{
			query:        tx,
			queryTimeout: s.queryTimeout,
		},
	}, nil
}

func (s *Postgres) Close() error {
	s.dbPool.Close()
	return nil
}

// pgxQueryer is the interface that wraps the required query methods of pgx. This is required
// to be able to use a pgxpool.Pool, a pgxpool.Conn or a pgx.Tx to execute a query.
type pgxQueryer interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
}

type postgresQuery struct {
	query        pgxQueryer
	queryTimeout time.Duration
}

func (t *postgresQuery) Get(ctx context.Context, key string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, t.queryTimeout)
	defer cancel()
	row := t.query.QueryRow(ctx, getQuery, key)
	var resKey string
	var value []byte
	err := row.Scan(&resKey, &value)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, persistence.ErrNotFound{Key: key}
		}
		return nil, err
	}
	return value, err
}

func (t *postgresQuery) Put(ctx context.Context, key string, value []byte) error {
	ctx, cancel := context.WithTimeout(ctx, t.queryTimeout)
	defer cancel()
	res, err := t.query.Exec(ctx, insertQuery, key, value)
	if err != nil {
		return err
	}
	rowCount := res.RowsAffected()
	if rowCount != 1 {
		return persistence.ErrInvalidRowsAffected
	}
	return nil
}

func (t *postgresQuery) Delete(ctx context.Context, key string) error {
	ctx, cancel := context.WithTimeout(ctx, t.queryTimeout)
	defer cancel()
	res, err := t.query.Exec(ctx, deleteQuery, key)
	if err != nil {
		return err
	}
	rowCount := res.RowsAffected()
	if rowCount == 0 {
		return persistence.ErrNotFound{Key: key}
	}
	if rowCount != 1 {
		return persistence.ErrInvalidRowsAffected
	}
	return nil
}

func (t *postgresQuery) List(
	ctx context.Context,
	prefix string,
	opts *persistence.ListOpts,
) (persistence.ListResult, error) {
	ctx, cancel := context.WithTimeout(ctx, t.queryTimeout)
	defer cancel()
	kvlist := make([]persistence.KVResult, 0, opts.Limit)

	query := sq.StatementBuilder.
		PlaceholderFormat(sq.Dollar).
		Select("key", "value", "COUNT(*) OVER() AS full_count").
		From("store").
		Where("key LIKE ? || '%%'", prefix).
		OrderBy("key").
		Limit(uint64(opts.Limit)).
		Offset(uint64(opts.Offset))

	// Parse out any provided, pre-validated CEL expression & add the proper clauses to the query.
	var err error
	if query, err = addFilterToQuery(query, opts.Filter); err != nil {
		return persistence.ListResult{}, fmt.Errorf("unable to add filter CEL expression to query: %w", err)
	}

	sql, placeholders, err := query.ToSql()
	if err != nil {
		return persistence.ListResult{}, err
	}
	rows, err := t.query.Query(ctx, sql, placeholders...)
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

	// The double question mark is how the SQL builder handles escaping a literal question mark.
	operator := "??&"
	if exprFunction == operators.LogicalOr {
		operator = "??|"
	}
	placeholders := sq.Placeholders(len(queryArgs))

	// As the only supported field name to filter on are tags, we can simply hard-code the predicate.
	return query.Where(
		fmt.Sprintf("value->'object'->'tags' %s array[%s]", operator, placeholders),
		queryArgs...,
	), nil
}
