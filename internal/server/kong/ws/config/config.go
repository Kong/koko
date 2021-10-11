package config

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"

	"github.com/kong/koko/internal/gen/wrpc/kong/model"
)

type Service struct {
	*model.Service
	Routes []*model.Route `json:"routes"`
}

type Content struct {
	FormatVersion string     `json:"_format_version"`
	Services      []*Service `json:"services,omitempty"`
}

type Payload struct {
	compressed []byte
}

func (p *Payload) Update(config Content) error {
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
	return p.compressed
}
