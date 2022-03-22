package config

import (
	"fmt"
	"sync"
)

type Map map[string]interface{}

type State struct {
	Payload []byte
	Hash    string
}

type Payload struct {
	content State
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

func (p *Payload) Payload(versionStr string) (State, error) {
	p.mu.RLock()
	content := p.content
	p.mu.RUnlock()

	// TODO(fero): perf create cache; version aware
	updatedPayload, err := p.vc.ProcessConfigTableUpdates(versionStr, content.Payload)
	if err != nil {
		return State{}, fmt.Errorf("downgrade config: %v", err)
	}
	return State{
		Payload: updatedPayload,
		Hash:    content.Hash,
	}, nil
}

func (p *Payload) UpdateBinary(c State) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.content = c
	return nil
}
