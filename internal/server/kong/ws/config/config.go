package config

import (
	"sync"
)

type Map map[string]interface{}

type Payload struct {
	compressed []byte
	mu         sync.RWMutex
}

func (p *Payload) Payload() []byte {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.compressed
}

func (p *Payload) UpdateBinary(c []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.compressed = c
	return nil
}
