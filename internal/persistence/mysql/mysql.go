package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/cel-go/common/operators"
	"github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/persistence"
	"go.uber.org/zap"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

// ErrMariaDBUnsupported is the error returned when db.DialectMariaDB is attempted to be used.
var ErrMariaDBUnsupported = errors.New("MariaDB is currently unsupported")

// MySQL defines a persistence store integration for databases that speak the MySQL protocol.
//
// Officially supported databases include:
// - MySQL   >= 8.0.17
// - MariaDB >= 10.9.2
//
// MySQL 8.0.17 is the minimum supported version, as it supports multivalued/inverted indexes,
// which is required to ensure indexes are hit when using `JSON_CONTAINS()` & `JSON_OVERLAPS()`.
// Read more: https://dev.mysql.com/doc/refman/8.0/en/create-index.html#create-index-multi-valued
//
// MariaDB 10.9.2 is the minimum supported version as that is when the `JSON_OVERLAPS()` function
// was introduced.
//
// However, MariaDB is not currently supported, as it will require use of a virtual column to handle row
// filtering, to support things like tag-based listing, as it currently does not support functional indexes
// like MySQL does. While this isn't difficult to support, we're punting this support for the future.
// Read more: https://mariadb.com/resources/blog/json-with-mariadb-10-2
type MySQL struct {
	db, readOnlyDB *sql.DB
	queryTimeout   time.Duration
}

func (s *MySQL) Driver() persistence.Driver { return persistence.MySQL }

func (s *MySQL) SetDefaultSQLOptions(db *sql.DB) error {
	db.SetMaxOpenConns(persistence.DefaultMaxConn)
	db.SetMaxIdleConns(persistence.DefaultMaxIdleConn)
	db.SetConnMaxLifetime(persistence.DefaultMaxConnLifetime)
	return nil
}

func (s *MySQL) Get(ctx context.Context, key string) ([]byte, error) {
	return (&mysqlQuery{s.readOnlyDB, s.queryTimeout}).Get(ctx, key)
}

func (s *MySQL) Put(ctx context.Context, key string, value []byte) error {
	return (&mysqlQuery{s.db, s.queryTimeout}).Put(ctx, key, value)
}

func (s *MySQL) Delete(ctx context.Context, key string) error {
	return (&mysqlQuery{s.db, s.queryTimeout}).Delete(ctx, key)
}

func (s *MySQL) List(
	ctx context.Context,
	prefix string,
	opts *persistence.ListOpts,
) (persistence.ListResult, error) {
	return (&mysqlQuery{s.readOnlyDB, s.queryTimeout}).List(ctx, prefix, opts)
}

func (s *MySQL) Tx(ctx context.Context) (persistence.Tx, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &mysqlTx{
		tx: tx,
		query: mysqlQuery{
			StdSqlCtx:    tx,
			queryTimeout: s.queryTimeout,
		},
	}, nil
}

func (s *MySQL) Close() error {
	return s.db.Close()
}

func NewSQLClient(opts Opts, logger *zap.Logger) (*sql.DB, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	opts.logger = logger
	dsn, err := opts.DSN()
	if err != nil {
		return nil, err
	}

	openFunc := persistence.DefaultSQLOpenFunc(&MySQL{})
	if opts.SQLOpen != nil {
		openFunc = opts.SQLOpen
	}

	return openFunc(dsn)
}

func New(opts Opts, queryTimeout time.Duration, logger *zap.Logger) (persistence.Persister, error) {
	db, err := NewSQLClient(opts, logger)
	if err != nil {
		return nil, fmt.Errorf("unable to set up MySQL DB client: %w", err)
	}

	// By default, fallback to primary host for read operations.
	readOnlyDB := db
	if opts.ReadOnlyHostname != "" {
		readOnlyOpts := opts
		readOnlyOpts.Hostname = opts.ReadOnlyHostname
		if readOnlyDB, err = NewSQLClient(readOnlyOpts, logger); err != nil {
			return nil, fmt.Errorf("unable to set up read-only MySQL DB client: %w", err)
		}
	}

	return &MySQL{
		db:           db,
		readOnlyDB:   readOnlyDB,
		queryTimeout: queryTimeout,
	}, nil
}

type mysqlQuery struct {
	sq.StdSqlCtx
	queryTimeout time.Duration
}

func (t *mysqlQuery) Get(ctx context.Context, key string) ([]byte, error) {
	rawSQL, placeholders, err := sq.StatementBuilder.
		Select("*").
		From("store").
		Where("`key` = ?", key).
		ToSql()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, t.queryTimeout)
	defer cancel()

	var resKey string
	var value []byte
	row := t.QueryRowContext(ctx, rawSQL, placeholders...)
	if err := row.Scan(&resKey, &value); err != nil {
		if err == sql.ErrNoRows {
			return nil, persistence.ErrNotFound{Key: key}
		}
		return nil, err
	}

	return value, nil
}

func (t *mysqlQuery) Put(ctx context.Context, key string, value []byte) error {
	rawSQL, placeholders, err := sq.StatementBuilder.
		Insert("store").
		Columns("`key`", "value").
		Values(key, value).
		Suffix("ON DUPLICATE KEY UPDATE value = ?", value).
		ToSql()
	if err != nil {
		return err
	}

	// We cannot check the affected rows, as per the MySQL docs:
	//
	// "For INSERT ... ON DUPLICATE KEY UPDATE statements, the affected-rows value per row is 1
	// if the row is inserted as a new row, 2 if an existing row is updated, and 0 if an existing
	// row is set to its current values. If you specify the CLIENT_FOUND_ROWS flag, the affected-rows
	// value is 1 (not 0) if an existing row is set to its current values."
	//
	// The `CLIENT_FOUND_ROWS` flag cannot be used, as it impacts UPDATE statements by returning
	// total number of found rows. As such, we are not returning a `persistence.ErrInvalidRowsAffected`
	// error when inserting rows & no rows have been affected.
	//
	// Read more: https://dev.mysql.com/doc/c-api/8.0/en/mysql-affected-rows.html
	ctx, cancel := context.WithTimeout(ctx, t.queryTimeout)
	defer cancel()
	if _, err = t.ExecContext(ctx, rawSQL, placeholders...); err != nil {
		return err
	}

	return nil
}

func (t *mysqlQuery) Delete(ctx context.Context, key string) error {
	rawSQL, placeholders, err := sq.StatementBuilder.
		Delete("store").
		Where("`key` = ?", key).
		ToSql()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, t.queryTimeout)
	defer cancel()
	res, err := t.ExecContext(ctx, rawSQL, placeholders...)
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

func (t *mysqlQuery) List(
	ctx context.Context,
	prefix string,
	opts *persistence.ListOpts,
) (persistence.ListResult, error) {
	ctx, cancel := context.WithTimeout(ctx, t.queryTimeout)
	defer cancel()

	// Replacing any wildcard operators with the proper wildcard operator for MySQL.
	prefix = strings.ReplaceAll(prefix, persistence.WildcardOperator, "%")

	query := sq.StatementBuilder.
		Select("`key`", "value", "COUNT(*) OVER() AS full_count").
		From("store").
		Where("`key` LIKE ?", prefix+"%").
		OrderBy("`key`").
		Limit(uint64(opts.Limit)).
		Offset(uint64(opts.Offset))

	// Parse out any provided, pre-validated CEL expression & add the proper clauses to the query.
	var err error
	if query, err = addFilterToQuery(query, opts.Filter); err != nil {
		return persistence.ListResult{}, fmt.Errorf("unable to add filter CEL expression to query: %w", err)
	}

	rawSQL, placeholders, err := query.ToSql()
	if err != nil {
		return persistence.ListResult{}, err
	}
	rows, err := t.QueryContext(ctx, rawSQL, placeholders...)
	if err != nil {
		return persistence.ListResult{}, err
	}
	defer rows.Close()
	if rows.Err() != nil {
		return persistence.ListResult{}, rows.Err()
	}

	res := persistence.ListResult{KVList: make([]persistence.KVResult, 0, opts.Limit)}
	for rows.Next() {
		var kvr persistence.KVResult
		if err := rows.Scan(&kvr.Key, &kvr.Value, &res.TotalCount); err != nil {
			return persistence.ListResult{}, err
		}
		res.KVList = append(res.KVList, kvr)
	}

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

	tagsJSON, err := json.Marshal(queryArgs)
	if err != nil {
		return query, err
	}

	// As the only supported field name to filter on are tags, we can simply hard-code the predicate.
	//
	// NOTE: The `->` & `->>` operators shall never be used, as they are not supported in MariaDB.
	// Instead, `JSON_EXTRACT()` should be used in its place.
	const tagsSelector = "JSON_EXTRACT(value, '$.object.tags')"
	if exprFunction == operators.LogicalOr {
		query = query.Where("JSON_OVERLAPS("+tagsSelector+", ?)", tagsJSON)
	} else {
		query = query.Where("JSON_CONTAINS("+tagsSelector+", ?, '$')", tagsJSON)
	}

	return query, nil
}
