package ws

import (
	"errors"
	"fmt"
	"time"

	"github.com/bluele/gcache"
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
type configCache struct {
	cache gcache.Cache
}

const (
	// configCacheSize is set to 10 to hold 10 configurations in memory.
	// This is deemed enough since during steady state only a single
	// configuration is required to be cached. In case when there are DP nodes
	// of different versions connected, the limit of 10 configurations should be
	// sufficient.
	configCacheSize = 10
	// configCacheExpiration time is set to 1 hour. This value is deemed as a
	// good starting point to balance (a) pressure on the database to rebuild
	// configurations (b) unbounded memory usage of the control-plane.
	// For rapidly changing configuration, configCacheSize provides an upperbound
	// on the number of cache entries. No upperbound for memory that this cache
	// can consume exists since a single cache entry has no max limit.
	configCacheExpiration = 15 * time.Minute
)

func newConfigCache() configCache {
	return configCache{
		cache: gcache.New(configCacheSize).LRU().Expiration(configCacheExpiration).Build(),
	}
}

func (c configCache) store(key string, value cacheEntry) error {
	err := c.cache.Set(key, value)
	if err != nil {
		return fmt.Errorf("save cache key '%v': %w", key, err)
	}
	return nil
}

func (c configCache) load(key string) (cacheEntry, error) {
	value, err := c.cache.Get(key)
	if err != nil {
		if errors.Is(err, gcache.KeyNotFoundError) {
			return cacheEntry{}, errNotFound
		}
		return cacheEntry{}, fmt.Errorf("failed to load key '%v': %w", key, err)
	}
	entry, ok := value.(cacheEntry)
	if !ok {
		panic(fmt.Sprintf("expected %T but got %T", cacheEntry{}, value))
	}
	return entry, nil
}

func (c configCache) reset() error {
	c.cache.Purge()
	return nil
}
