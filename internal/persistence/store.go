package persistence

import (
	"context"
	"fmt"
)

// Persister is an interface to a KV store.
// A database must support the following operations:
// - Set key to a value
// - Get value based on a key
// - Transactions: Read and Write key-values in a transaction
// - List key-values based on a key prefix.
type Persister interface {
	CRUD
	Tx(context.Context) (Tx, error)
}

type CRUD interface {
	// Get retrieves the value from store.
	// It returns ErrNotFound if key is not found.
	Get(ctx context.Context, key string) ([]byte, error)
	// Put sets key to value in the store.
	// If key is already present, it is overwritten.
	Put(ctx context.Context, key string, value []byte) error
	// Delete deletes the key and its associated value from store.
	// If key is not found, an ErrNotFound error is returned.
	Delete(ctx context.Context, key string) error
	// List returns all keys with prefix.
	List(ctx context.Context, prefix string, opts *ListOpts) ([]*KVResult, error)
}

type KVResult struct {
	Key        []byte
	Value      []byte
	TotalCount int
}

const (
	DefaultPage     = 1
	DefaultPageSize = 1000
)

func NewDefaultListOpts() *ListOpts {
	return &ListOpts{Page: DefaultPage, PageSize: DefaultPageSize}
}

type ListOpts struct {
	PageSize int
	Page     int
}

type Tx interface {
	Commit() error
	Rollback() error
	CRUD
}

type ErrNotFound struct {
	Key string
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("%v not found", e.Key)
}

func ToOffset(opts *ListOpts) int {
	if opts.Page == 1 || opts.Page == 0 {
		return 0
	}
	return opts.PageSize * (opts.Page - 1)
}

func ToLastPage(pageSize int, count int) int {
	if count%2 == 0 {
		return (count / pageSize) - 1
	}
	return count / pageSize
}
