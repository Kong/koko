package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/cel-go/common/operators"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/kong/koko/internal/persistence"
	"go.uber.org/zap"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

const (
	getQuery    = `SELECT * from store where key=$1`
	insertQuery = `insert into store(key,value) values($1,$2) on conflict (key) do update set value=$2;`
	deleteQuery = `delete from store where key=$1`

	DefaultPort       = 5432
	defaultTLSSslMode = "verify-full"
)

type Postgres struct {
	db           *sql.DB
	readOnlyDB   *sql.DB
	queryTimeout time.Duration
}

type Opts struct {
	DBName           string
	Hostname         string
	ReadOnlyHostname string
	Port             int
	User             string
	Password         string
	EnableTLS        bool
	CABundleFSPath   string
	SslMode          string
	SQLOpen          persistence.SQLOpenFunc
}

func getDSN(opts Opts, logger *zap.Logger) (string, error) {
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
		logger.Info("using non-TLS Postgres connection")
		res += "sslmode=disable"
		return res, nil
	}
	logger.Info("using TLS Postgres connection")
	logger.Info("ca_bundle_fs_path:" + opts.CABundleFSPath)
	if opts.CABundleFSPath == "" {
		return "", fmt.Errorf("postgres connection requires TLS but ca_bundle_fs_path is empty")
	}
	if opts.SslMode == "" {
		opts.SslMode = defaultTLSSslMode
	}
	res += fmt.Sprintf("sslmode=%s sslrootcert=%s", opts.SslMode, opts.CABundleFSPath)

	return res, nil
}

func NewSQLClient(opts Opts, logger *zap.Logger) (*sql.DB, error) {
	dsn, err := getDSN(opts, logger)
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
	db, err := NewSQLClient(opts, logger)
	if err != nil {
		return nil, fmt.Errorf("unable to set up DB client: %w", err)
	}
	// by default, fallback to primary host for read operations
	readOnlyDB := db
	if opts.ReadOnlyHostname != "" {
		readOnlyOpts := opts
		readOnlyOpts.Hostname = opts.ReadOnlyHostname
		readOnlyDB, err = NewSQLClient(readOnlyOpts, logger)
		if err != nil {
			return nil, fmt.Errorf("unable to set up read-only DB client: %w", err)
		}
	}
	res := &Postgres{
		db:           db,
		readOnlyDB:   readOnlyDB,
		queryTimeout: queryTimeout,
	}
	return res, nil
}

// Driver implements the persistence.SQLPersister interface.
func (s *Postgres) Driver() persistence.Driver { return persistence.Postgres }

// SetDefaultSQLOptions implements the persistence.SQLPersister interface.
func (s *Postgres) SetDefaultSQLOptions(db *sql.DB) error {
	db.SetMaxOpenConns(persistence.DefaultMaxConn)
	db.SetMaxIdleConns(persistence.DefaultMaxIdleConn)
	db.SetConnMaxLifetime(persistence.DefaultMaxConnLifetime)
	return nil
}

func (s *Postgres) Get(ctx context.Context, key string) ([]byte, error) {
	q := postgresQuery{query: s.readOnlyDB, queryTimeout: s.queryTimeout}
	return q.Get(ctx, key)
}

func (s *Postgres) Put(ctx context.Context, key string, value []byte) error {
	q := postgresQuery{query: s.db, queryTimeout: s.queryTimeout}
	return q.Put(ctx, key, value)
}

func (s *Postgres) Delete(ctx context.Context, key string) error {
	q := postgresQuery{query: s.db, queryTimeout: s.queryTimeout}
	return q.Delete(ctx, key)
}

func (s *Postgres) List(ctx context.Context, prefix string, opts *persistence.ListOpts) (persistence.ListResult,
	error,
) {
	q := postgresQuery{query: s.readOnlyDB, queryTimeout: s.queryTimeout}
	return q.List(ctx, prefix, opts)
}

func (s *Postgres) Tx(ctx context.Context) (persistence.Tx, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &postgresTx{
		tx: tx,
		query: postgresQuery{
			query:        tx,
			queryTimeout: s.queryTimeout,
		},
	}, nil
}

func (s *Postgres) Close() error {
	return s.db.Close()
}

type postgresTx struct {
	tx    *sql.Tx
	query postgresQuery
}

func (t *postgresTx) Commit() error {
	return t.tx.Commit()
}

func (t *postgresTx) Rollback() error {
	return t.tx.Rollback()
}

func (t *postgresTx) Get(ctx context.Context, key string) ([]byte, error) {
	return t.query.Get(ctx, key)
}

func (t *postgresTx) Put(ctx context.Context, key string, value []byte) error {
	return t.query.Put(ctx, key, value)
}

func (t *postgresTx) Delete(ctx context.Context, key string) error {
	return t.query.Delete(ctx, key)
}

func (t *postgresTx) List(
	ctx context.Context,
	prefix string,
	opts *persistence.ListOpts,
) (persistence.ListResult, error) {
	return t.query.List(ctx, prefix, opts)
}

type postgresQuery struct {
	query        sq.StdSqlCtx
	queryTimeout time.Duration
}

func (t *postgresQuery) Get(ctx context.Context, key string) ([]byte, error) {
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

func (t *postgresQuery) Put(ctx context.Context, key string, value []byte) error {
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
		return persistence.ErrInvalidRowsAffected
	}
	return nil
}

func (t *postgresQuery) Delete(ctx context.Context, key string) error {
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
