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
	content := p.content
	p.mu.RUnlock()

	// TODO(fero): perf create cache; version aware
	updatedPayload, err := p.vc.ProcessConfigTableUpdates(versionStr, content.CompressedPayload)
	if err != nil {
		return Content{}, fmt.Errorf("downgrade config: %v", err)
	}
	return Content{
		CompressedPayload: updatedPayload,
		Hash:              content.Hash,
	}, nil
}

func (p *Payload) UpdateBinary(c Content) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.content = c
	return nil
}
