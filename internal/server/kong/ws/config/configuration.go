package config

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"

	"github.com/kong/koko/internal/json"
)

type MutatorOpts struct {
	ClusterID string
}

type Mutator interface {
	Name() string
	Mutate(context.Context, MutatorOpts, DataPlaneConfig) error
}

type Loader interface {
	Load(ctx context.Context, clusterID string) ([]byte, error)
}

type DataPlaneConfig Map

type KongConfigurationLoader struct {
	mutators []Mutator
}

func (l *KongConfigurationLoader) Register(mutator Mutator) error {
	for _, m := range l.mutators {
		if m.Name() == mutator.Name() {
			return fmt.Errorf("mutator '%v' already registered", m.Name())
		}
	}
	l.mutators = append(l.mutators, mutator)
	return nil
}

func (l *KongConfigurationLoader) Load(ctx context.Context,
	clusterID string) ([]byte, error) {
	var configTable DataPlaneConfig = map[string]interface{}{}
	for _, m := range l.mutators {
		err := m.Mutate(ctx, MutatorOpts{ClusterID: clusterID},
			configTable)
		if err != nil {
			return nil, err
		}
	}

	payload := reconfigurePayload(configTable)

	jsonMessage, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("json marshal: %v", err)
	}

	res, err := compressPayload(jsonMessage)
	if err != nil {
		return nil, fmt.Errorf("gzip compression: %v", err)
	}
	return res, nil
}

func reconfigurePayload(c DataPlaneConfig) Map {
	return Map{
		"type":         "reconfigure",
		"config_table": c,
	}
}

func compressPayload(payload []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	_, err := w.Write(payload)
	if err != nil {
		return nil, err
	}
	err = w.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
