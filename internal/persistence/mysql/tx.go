package mysql

import (
	"context"
	"database/sql"

	"github.com/kong/koko/internal/persistence"
)

type mysqlTx struct {
	tx    *sql.Tx
	query mysqlQuery
}

func (t *mysqlTx) Commit() error { return t.tx.Commit() }

func (t *mysqlTx) Rollback() error { return t.tx.Rollback() }

func (t *mysqlTx) Get(ctx context.Context, k string) ([]byte, error) { return t.query.Get(ctx, k) }

func (t *mysqlTx) Insert(ctx context.Context, k string, v []byte) error {
	return t.query.Insert(ctx, k, v)
}

func (t *mysqlTx) Put(ctx context.Context, k string, v []byte) error { return t.query.Put(ctx, k, v) }

func (t *mysqlTx) Delete(ctx context.Context, k string) error { return t.query.Delete(ctx, k) }

func (t *mysqlTx) List(ctx context.Context, prefix string, opts *persistence.ListOpts) (persistence.ListResult, error) {
	return t.query.List(ctx, prefix, opts)
}
