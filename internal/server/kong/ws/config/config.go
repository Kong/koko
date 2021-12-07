package config

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"sync"
)

type Map map[string]interface{}

type Content struct {
	FormatVersion string `json:"_format_version"`
	Services      []Map  `json:"services,omitempty"`
	Routes        []Map  `json:"routes,omitempty"`
}

type Payload struct {
	compressed []byte
	mu         sync.RWMutex
}

func (p *Payload) Update(config Content) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	payload := map[string]interface{}{
		"type":         "reconfigure",
		"config_table": config,
	}
	jsonMessage, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("json marshal: %v", err)
	}
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	_, err = w.Write(jsonMessage)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	p.compressed = buf.Bytes()
	return nil
}

func (p *Payload) Payload() []byte {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.compressed
}
