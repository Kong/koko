package ws

import (
	"fmt"
	"sync"

	"github.com/kong/koko/internal/server/kong/ws/config"
)

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
	content config.Content
	cache   map[string]CachedContent
	mu      sync.Mutex
	vc      config.VersionCompatibility
}

type PayloadOpts struct {
	VersionCompatibilityProcessor config.VersionCompatibility
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

func (p *Payload) Payload(versionStr string) (config.Content, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, found := p.cache[versionStr]; !found {
		updatedPayload, err := p.vc.ProcessConfigTableUpdates(versionStr, p.content.CompressedPayload)
		p.cache[versionStr] = CachedContent{
			CompressedPayload: updatedPayload,
			Error:             err,
			Hash:              p.content.Hash,
		}
	}

	if p.cache[versionStr].Error != nil {
		return config.Content{}, fmt.Errorf("downgrade config: %v", p.cache[versionStr].Error)
	}
	return config.Content{
		CompressedPayload: p.cache[versionStr].CompressedPayload,
		Hash:              p.cache[versionStr].Hash,
	}, nil
}

func (p *Payload) UpdateBinary(c config.Content) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.content = c
	p.cache = make(map[string]CachedContent)
	return nil
}
