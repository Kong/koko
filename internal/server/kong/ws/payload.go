package ws

import (
	"context"
	"errors"
	"fmt"
	"sync"

	config_service "github.com/kong/koko/internal/gen/wrpc/kong/services/config/v1"
	"github.com/kong/koko/internal/server/kong/ws/config"
	"go.uber.org/zap"
)

const unversionedConfigKey = "unversioned"

type Payload struct {
	// configCache is a cache of configuration. It holds the originally fetched
	// configuration as well as massaged configuration for each DP version.
	configVersion   uint64
	configCache     configCache
	configCacheLock sync.Mutex
	wrpcCache       configWRPCCache
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
		wrpcCache:   configWRPCCache{},
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

func (p *Payload) readWRPCCache(versionStr string) (wc CachedWRPCContent, err error) {
	p.configCacheLock.Lock()
	defer p.configCacheLock.Unlock()

	wc, err = p.wrpcCache.load(versionStr)
	return
}

func (p *Payload) WrpcConfigPayload(ctx context.Context, versionStr string) (CachedWRPCContent, error) {
	wc, err := p.readWRPCCache(versionStr)
	if err == nil {
		return wc, nil
	}

	if errors.Is(err, errNotFound) {
		c, err := p.Payload(ctx, versionStr)
		if err != nil {
			return CachedWRPCContent{}, err
		}

		configTable := config_service.SyncConfigRequest{
			Config:     c.CompressedPayload,
			Version:    p.configVersion,
			ConfigHash: c.Hash,
		}

		p.configCacheLock.Lock()
		defer p.configCacheLock.Unlock()

		req, err := config_service.PrepareConfigServiceSyncConfigRequest(&configTable)
		wc := CachedWRPCContent{
			Req:   req,
			Error: err,
			Hash:  c.Hash,
		}
		err = p.wrpcCache.store(versionStr, wc)
		if err != nil {
			p.logger.Error("failed to store wRPC configuration to cache",
				zap.Error(err),
				zap.String("kong-dp-version", versionStr))
			return wc, nil
		}
		return wc, nil
	}

	return wc, err
}

func (p *Payload) UpdateBinary(_ context.Context, c config.Content) error {
	p.configCacheLock.Lock()
	defer p.configCacheLock.Unlock()

	p.configVersion++
	err := p.configCache.reset()
	if err != nil {
		return err
	}

	err = p.wrpcCache.reset()
	if err != nil {
		return err
	}

	err = p.configCache.store(unversionedConfigKey, cacheEntry{Content: c})
	if err != nil {
		return err
	}

	return nil
}
