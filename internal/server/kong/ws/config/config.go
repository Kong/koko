package config

import (
	"fmt"
	"sync"
)

type Map map[string]interface{}

type Content struct {
	CompressedPayload []byte
	Hash              string
}

// CachedContent holds the processed payload for a particular data plane
// version. This content may be different than the actual Conent container as
// the CompressedPayload could be updated for version compatibility.
type CachedContent struct {
	CompressedPayload []byte
	// Error is stored to return the original error when accessing the cache
	Error error
	Hash  string
}

type Payload struct {
	content Content
	cache   map[string]CachedContent
	mu      sync.RWMutex
	vc      VersionCompatibility
}

type PayloadOpts struct {
	VersionCompatibilityProcessor VersionCompatibility
}

func NewPayload(opts PayloadOpts) (*Payload, error) {
	if opts.VersionCompatibilityProcessor == nil {
		return nil, fmt.Errorf("opts.VersionCompatibilityProcessor required")
	}

	return &Payload{
		vc:    opts.VersionCompatibilityProcessor,
		cache: map[string]CachedContent{},
	}, nil
}

func (p *Payload) Payload(versionStr string) (Content, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if _, found := p.cache[versionStr]; !found {
		updatedPayload, err := p.vc.ProcessConfigTableUpdates(versionStr, p.content.CompressedPayload)
		p.cache[versionStr] = CachedContent{
			CompressedPayload: updatedPayload,
			Error:             err,
			Hash:              p.content.Hash,
		}
	}

	if p.cache[versionStr].Error != nil {
		return Content{}, fmt.Errorf("downgrade config: %v", p.cache[versionStr].Error)
	}
	return Content{
		CompressedPayload: p.cache[versionStr].CompressedPayload,
		Hash:              p.cache[versionStr].Hash,
	}, nil
}

func (p *Payload) UpdateBinary(c Content) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.content = c
	p.cache = make(map[string]CachedContent)
	return nil
}
