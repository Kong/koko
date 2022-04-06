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

type Cache struct {
	CompressedPayload []byte
	Error             error
}

type Payload struct {
	content Content
	cache   map[string]Cache
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
		vc: opts.VersionCompatibilityProcessor,
	}, nil
}

func (p *Payload) Payload(versionStr string) (Content, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	content := p.content

	if _, found := p.cache[versionStr]; !found {
		updatedPayload, err := p.vc.ProcessConfigTableUpdates(versionStr, content.CompressedPayload)
		p.cache[versionStr] = Cache{
			CompressedPayload: updatedPayload,
			Error:             err,
		}
	}

	if p.cache[versionStr].Error != nil {
		return Content{}, fmt.Errorf("downgrade config: %v", p.cache[versionStr].Error)
	}
	return Content{
		CompressedPayload: p.cache[versionStr].CompressedPayload,
		Hash:              content.Hash,
	}, nil
}

func (p *Payload) UpdateBinary(c Content) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.content = c
	p.cache = make(map[string]Cache)
	return nil
}
