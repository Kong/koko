package persistence

import (
	"context"
	"fmt"

	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
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
	Close() error
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

// ListOpts defines various options that affect the results returned by a `CRUD.List()` call.
type ListOpts struct {
	// Limit is used to set the amount of results returned. Must be between zero & MaxLimit.
	Limit int

	// Offset is used for purposes of pagination. Must be a positive
	// number and zero is used to indicate the first page.
	Offset int

	// CEL expression used for filtering.
	//
	// When nil, no filtering of any kind will be done. When provided, the filter is
	// expected to be pre-validated for correctness. More specific validations can
	// occur later, e.g.: such validations that are specific to a particular resource.
	//
	// Read more: https://github.com/google/cel-spec
	Filter *exprpb.Expr
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
