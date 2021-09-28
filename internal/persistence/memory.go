package persistence

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

type Memory struct {
	store sync.Map
}

func (m *Memory) Get(ctx context.Context, key string) ([]byte, error) {
	v, ok := m.store.Load(key)
	if !ok {
		return nil, ErrNotFound{key: key}
	}
	return v.([]byte), nil
}

func (m *Memory) Put(ctx context.Context, key string, value []byte) error {
	m.store.Store(key, value)
	return nil
}

func (m *Memory) Delete(ctx context.Context, key string) error {
	_, present := m.store.LoadAndDelete(key)
	if !present {
		return ErrNotFound{key: key}
	}
	return nil
}

func (m *Memory) List(ctx context.Context, prefix string) ([][]byte, error) {
	var res [][]byte
	m.store.Range(func(k, v interface{}) bool {
		key, ok := k.(string)
		if !ok {
			panic(fmt.Sprintf("expected string but found %T", k))
		}
		if strings.HasPrefix(key, prefix) {
			res = append(res, v.([]byte))
		}
		return true
	})
	return res, nil
}
