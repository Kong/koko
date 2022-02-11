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
	List(ctx context.Context, prefix string, opts *ListOpts) (ListResult, error)
}

type KVResult struct {
	Key   []byte
	Value []byte
}

type ListResult struct {
	KVList     []KVResult
	TotalCount int
}

const (
	MaxLimit      = 1000
	DefaultLimit  = 100
	DefaultOffset = 0
)

func NewDefaultListOpts() *ListOpts {
	return &ListOpts{Offset: DefaultOffset, Limit: DefaultLimit}
}

type ListOpts struct {
	Limit  int
	Offset int
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
