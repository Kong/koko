package ws

import (
	"fmt"

	"github.com/kong/go-wrpc/wrpc"
	"github.com/kong/koko/internal/server/kong/ws/config"
)

// cacheEntry holds the processed payload or an error.
type cacheEntry struct {
	config.Content
	// Error is stored to return the original error in case of a processing
	// error.
	Error error
}

var errNotFound = fmt.Errorf("not found")

// configCache holds configuration based on keys.
// None of the functions are thread-safe and it is up to the caller to ensure
// thread-safe behavior.
type configCache map[string]cacheEntry

func (c configCache) store(key string, value cacheEntry) error {
	c[key] = value
	return nil
}

func (c configCache) load(key string) (cacheEntry, error) {
	value, found := c[key]
	if !found {
		return cacheEntry{}, errNotFound
	}
	return value, nil
}

func (c configCache) reset() error {
	for k := range c {
		delete(c, k)
	}
	return nil
}

// CachedWRPCContent holds a prepared wRPC sync request.
type CachedWRPCContent struct {
	Req   wrpc.Request
	Error error
	Hash  string
}

// configWRPCCache keeps config requests parallel to configCache.
type configWRPCCache map[string]CachedWRPCContent

func (c configWRPCCache) store(key string, value CachedWRPCContent) error { // nolint: unparam
	c[key] = value
	return nil
}

func (c configWRPCCache) load(key string) (CachedWRPCContent, error) {
	value, found := c[key]
	if !found {
		return CachedWRPCContent{}, errNotFound
	}
	return value, nil
}

func (c configWRPCCache) reset() error {
	for k := range c {
		delete(c, k)
	}
	return nil
}
