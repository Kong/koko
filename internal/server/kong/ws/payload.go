package ws

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/kong/koko/internal/server/kong/ws/config"
	"go.uber.org/zap"
)

const unversionedConfigKey = "unversioned"

type Payload struct {
	// configCache is a cache of configuration. It holds the originally fetched
	// configuration as well as massaged configuration for each DP version.
	configCache     configCache
	configCacheLock sync.Mutex
	vc              config.VersionCompatibility
	logger          *zap.Logger
}

type PayloadOpts struct {
	VersionCompatibilityProcessor config.VersionCompatibility
	Logger                        *zap.Logger
}

func NewPayload(opts PayloadOpts) (*Payload, error) {
	if opts.VersionCompatibilityProcessor == nil {
		return nil, fmt.Errorf("opts.VersionCompatibilityProcessor required")
	}

	return &Payload{
		vc:          opts.VersionCompatibilityProcessor,
		configCache: configCache{},
		logger:      opts.Logger,
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
		updatedPayload, err := p.vc.ProcessConfigTableUpdates(version, unversionedConfig.CompressedPayload)
		entry := cacheEntry{
			Content: config.Content{
				CompressedPayload: updatedPayload,
				// Hash must remain stable across version.
				Hash: unversionedConfig.Hash,
			},
			Error: err,
		}
		if err != nil {
			p.logger.Error("failed to process config table update",
				zap.Error(err),
				zap.String("kong-dp-version", version),
			)
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

func (p *Payload) UpdateBinary(_ context.Context, c config.Content) error {
	p.configCacheLock.Lock()
	defer p.configCacheLock.Unlock()
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
