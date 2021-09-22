package persistence

import (
	"context"
	"fmt"
)

type Persister interface {
	Get(context.Context, string) ([]byte, error)
	Put(context.Context, string, []byte) error
	Delete(context.Context, string) error
	List(context.Context, string) ([][]byte, error)
	// TODO(hbagdi): are transactions required?
	// Tx(func(s Persister) error) error
}

type ErrNotFound struct {
	key string
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("%v not found", e.key)
}

func (e ErrNotFound) Key() string {
	return e.key
}
