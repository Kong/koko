package postgres

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/kong/koko/internal/persistence"
)

type postgresTx struct {
	ctx   context.Context
	tx    pgx.Tx
	query postgresQuery
}

func (t *postgresTx) Commit() error {
	return t.tx.Commit(t.ctx)
}

func (t *postgresTx) Rollback() error {
	return t.tx.Rollback(t.ctx)
}

func (t *postgresTx) Get(ctx context.Context, key string) ([]byte, error) {
	return t.query.Get(ctx, key)
}

func (t *postgresTx) Insert(ctx context.Context, key string, value []byte) error {
	return t.query.Insert(ctx, key, value)
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
