package ws

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/bluele/gcache"
	"github.com/kong/koko/internal/metrics"
	"github.com/kong/koko/internal/server/kong/ws/config"
	"go.uber.org/zap"
)

const (
	unversionedConfigKey = "unversioned"
	// configHashToChangesCacheSize is set to 1000 to ensure that 1000 unique
	// hash+node-version can be tracked.
	// The number of unique Kong Node versions for a Manager are expected to be
	// less than 5 in most cases and usually one or two.
	configHashToChangesCacheSize = 1000
)

type Payload struct {
	// configCache is a cache of configuration. It holds the originally fetched
	// configuration as well as massaged configuration for each DP version.
	configCache     configCache
	configCacheLock sync.Mutex
	// configVersion is a counter incremented every config update.  Sent in the
	// config wRPC service, only compared by DP within the lifetime of the connection.
	configVersion uint64
	vc            config.VersionCompatibility
	logger        *zap.Logger
	// configHashToChanges is map of config-hash+version to tracked changes.
	configHashToChanges gcache.Cache
}

type PayloadOpts struct {
	VersionCompatibilityProcessor config.VersionCompatibility
	Logger                        *zap.Logger
}

func NewPayload(opts PayloadOpts) (*Payload, error) {
	if opts.VersionCompatibilityProcessor == nil {
		return nil, fmt.Errorf("opts.VersionCompatibilityProcessor required")
	}
	cache := gcache.New(configHashToChangesCacheSize).LRU().Build()

	return &Payload{
		vc:                  opts.VersionCompatibilityProcessor,
		configCache:         configCache{},
		logger:              opts.Logger,
		configHashToChanges: cache,
	}, nil
}

func (p *Payload) Payload(_ context.Context, version string) (config.Content, error) {
	p.configCacheLock.Lock()
	defer p.configCacheLock.Unlock()

	entry, err := p.configForVersion(version)
	if err != nil {
		return config.Content{}, err
	}

	if entry.Error != nil {
		return config.Content{}, fmt.Errorf("downgrade config: %v", entry.Error)
	}

	return config.Content{
		CompressedPayload: entry.CompressedPayload,
		Hash:              entry.Hash,
		GranularHashes:    entry.GranularHashes,
	}, nil
}

func (p *Payload) configForVersion(version string) (cacheEntry, error) {
	contentCacheEntry, err := p.configCache.load(version)
	if err == nil {
		// fast path
		return contentCacheEntry, nil
	}

	// cache-miss, slow path
	if errors.Is(err, errNotFound) {
		unversionedConfig, err := p.configCache.load(unversionedConfigKey)
		if err != nil {
			return cacheEntry{}, err
		}
		// build the config for version
		updatedPayload, changes, err := p.vc.ProcessConfigTableUpdates(
			version,
			unversionedConfig.CompressedPayload,
		)
		entry := cacheEntry{
			Content: config.Content{
				CompressedPayload: updatedPayload,
				// Hash must remain stable across version.
				Hash:           unversionedConfig.Hash,
				GranularHashes: unversionedConfig.GranularHashes,
			},
			Error: err,
		}
		if err != nil {
			p.logger.Error("failed to process config table update",
				zap.Error(err),
				zap.String("kong-dp-version", version),
			)
		} else {
			resourcesUpdated := 0
			changeCount := len(changes.ChangeDetails)
			for _, detail := range changes.ChangeDetails {
				resourcesUpdated += len(detail.Resources)
			}
			p.logger.Warn("version compatibility changes detected",
				zap.Int("change-count", changeCount),
				zap.Int("total-resource-changes", resourcesUpdated))
			metrics.Gauge("version_compatibility_configuration_change_count",
				int64(changeCount), metrics.Tag{
					Key:   "kong-dp-version",
					Value: version,
				})
			metrics.Gauge("version_compatibility_configuration_resource_change_count",
				int64(resourcesUpdated), metrics.Tag{
					Key:   "kong-dp-version",
					Value: version,
				})
		}
		// cache changes
		err = p.configHashToChanges.Set(unversionedConfig.Hash+version, changes)
		if err != nil {
			p.logger.Error("failed to track config changes in cache",
				zap.Error(err))
			// in case of cache error,
			// continue on with configuration caching
		}
		// cache it
		err = p.configCache.store(version, entry)
		if err != nil {
			p.logger.Error("failed to store configuration from cache",
				zap.Error(err),
				zap.String("kong-dp-version", version),
			)
			// on cache store failures, still serve the config
			return entry, nil
		}
		return entry, nil
	}

	// other errors
	return cacheEntry{}, err
}

func (p *Payload) ChangesFor(hash, version string) (config.TrackedChanges, error) {
	value, err := p.configHashToChanges.Get(hash + version)
	if err != nil {
		if !errors.Is(err, gcache.KeyNotFoundError) {
			p.logger.Error("failed to fetch tracked changes from cache",
				zap.String("hash", hash),
				zap.String("node-version", version),
				zap.Error(err))
		}
		return config.TrackedChanges{}, errNotFound
	}

	trackedChanges, ok := value.(config.TrackedChanges)
	if !ok {
		panic(fmt.Sprintf("invalid type: expected %T, got %T", config.TrackedChanges{}, value))
	}
	return trackedChanges, nil
}

func (p *Payload) UpdateBinary(_ context.Context, c config.Content) error {
	p.configCacheLock.Lock()
	defer p.configCacheLock.Unlock()

	p.configVersion++
	err := p.configCache.reset()
	if err != nil {
		return err
	}

	err = p.configCache.store(unversionedConfigKey, cacheEntry{Content: c})
	if err != nil {
		return err
	}

	return nil
}
