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

type Payload struct {
	content Content
	cache   map[string][]byte
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

	if len(p.cache[versionStr]) == 0 {
		updatedPayload, err := p.vc.ProcessConfigTableUpdates(versionStr, content.CompressedPayload)
		if err != nil {
			return Content{}, fmt.Errorf("downgrade config: %v", err)
		}
		p.cache[versionStr] = updatedPayload
	}

	return Content{
		CompressedPayload: p.cache[versionStr],
		Hash:              content.Hash,
	}, nil
}

func (p *Payload) UpdateBinary(c Content) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.content = c
	p.cache = make(map[string][]byte)
	return nil
}
