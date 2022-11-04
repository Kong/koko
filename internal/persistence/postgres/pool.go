package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

var pools = map[string]PoolOpenFunc{}

// Pool is the interface of the pgxpool.Pool type, used by the postgres
// persistence layer to query the database. A custom pool type can be
// registered using `RegisterPool` if it implements this interface.
type Pool interface {
	Close()
	Acquire(ctx context.Context) (*pgxpool.Conn, error)
	AcquireFunc(ctx context.Context, f func(*pgxpool.Conn) error) error
	AcquireAllIdle(ctx context.Context) []*pgxpool.Conn
	Config() *pgxpool.Config
	Stat() *pgxpool.Stat
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	QueryFunc(ctx context.Context, sql string, args []interface{}, scans []interface{},
		f func(pgx.QueryFuncRow) error) (pgconn.CommandTag, error)
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	Begin(ctx context.Context) (pgx.Tx, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
	BeginFunc(ctx context.Context, f func(pgx.Tx) error) error
	BeginTxFunc(ctx context.Context, txOptions pgx.TxOptions, f func(pgx.Tx) error) error
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
	Ping(ctx context.Context) error
}

// PoolOpenFunc defines a function that can instantiate a Pool instance with the given options.
type PoolOpenFunc func(opts Opts, logger *zap.Logger) (Pool, error)

// PoolOpts defines configuration for connections pooling.
type PoolOpts struct {
	// The name of the postgres pool implementation to use.
	// By default, only `pgx` is available.
	Name string

	// MaxConns is the maximum size of the pool. Default to `persistence.DefaultMaxConn`.
	MaxConns int

	// MinConns is the minimum size of the pool. Default to `persistence.DefaultMinConn`.
	MinConns int

	// MaxConnLifetime is the duration since creation after which a connection will
	// be automatically closed. Default to `persistence.DefaultMaxConnLifetime`.
	MaxConnLifetime time.Duration

	// MaxConnIdleTime is the duration after which an idle connection will be automatically
	// closed by the health check. Default to `persistence.DefaultMaxConnIdleTime`.
	MaxConnIdleTime time.Duration

	// HealthCheckPeriod is the duration between checks of the health of
	// idle connections. Default to `persistence.DefaultHealthCheckPeriod`.
	HealthCheckPeriod time.Duration

	// ReadOnly indicates the pool will be used for read only operations.
	ReadOnly bool
}

// Validate ensures the provided Postgres.Pool options are a valid configuration.
func (opts *PoolOpts) Validate() error {
	// by validating these values, we avoid panics from pgxpool at connection
	if opts.MaxConns < 1 {
		return fmt.Errorf("invalid Pool.MaxConns value: '%d', should be greater than 0", opts.MaxConns)
	}

	if opts.HealthCheckPeriod < 1 {
		return fmt.Errorf("invalid Pool.HealthCheckPeriod value: '%d', should be a positive duration", opts.HealthCheckPeriod)
	}

	return nil
}

func newPostgresPool(opts Opts, logger *zap.Logger) (Pool, error) {
	if err := opts.Pool.Validate(); err != nil {
		return nil, err
	}

	if poolOpenFunc, present := pools[opts.Pool.Name]; present {
		logger.Info(fmt.Sprintf("using Postgres %s pool", opts.Pool.Name))
		return poolOpenFunc(opts, logger)
	}

	return nil, fmt.Errorf("invalid postgres pool '%s'", opts.Pool.Name)
}

// NewPgxPoolConfig creates and configure a new pgxpool.Config with the given options.
func NewPgxPoolConfig(opts Opts, logger *zap.Logger) (*pgxpool.Config, error) {
	dsn, err := opts.DSN(logger)
	if err != nil {
		return nil, err
	}

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	config.MaxConns = int32(opts.Pool.MaxConns)
	config.MinConns = int32(opts.Pool.MinConns)
	config.MaxConnLifetime = opts.Pool.MaxConnLifetime
	config.MaxConnIdleTime = opts.Pool.MaxConnIdleTime
	config.HealthCheckPeriod = opts.Pool.HealthCheckPeriod

	return config, nil
}

func newPgxPool(opts Opts, logger *zap.Logger) (Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second) //nolint:gomnd
	defer cancel()

	pgxPoolConfig, err := NewPgxPoolConfig(opts, logger)
	if err != nil {
		return nil, err
	}

	return pgxpool.ConnectConfig(ctx, pgxPoolConfig)
}

// RegisterPool adds a `Pool` implementation to the postgres configuration.
// It is not thread-safe.
func RegisterPool(poolName string, poolOpenFunc PoolOpenFunc) error {
	if _, present := pools[poolName]; present {
		return fmt.Errorf("pool name '%s' already exists", poolName)
	}
	pools[poolName] = poolOpenFunc

	return nil
}

func init() {
	if err := RegisterPool(DefaultPool, newPgxPool); err != nil {
		panic(err)
	}
}
