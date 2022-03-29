package config

import (
	"fmt"
	"sync"
)

type Map map[string]interface{}

type Payload struct {
	compressed []byte
	mu         sync.RWMutex
	vc         VersionCompatibility
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

func (p *Payload) Payload(versionStr string) ([]byte, error) {
	p.mu.RLock()
	payload := p.compressed
	p.mu.RUnlock()

	// TODO(fero): perf create cache; version aware
	updatedPayload, err := p.vc.ProcessConfigTableUpdates(versionStr, payload)
	if err != nil {
		return nil, fmt.Errorf("downgrade config: %v", err)
	}
	return updatedPayload, nil
}

func (p *Payload) UpdateBinary(c []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.compressed = c
	return nil
}
