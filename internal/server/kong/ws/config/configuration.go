package config

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"

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
	clusterID string,
) ([]byte, error) {
	var configTable DataPlaneConfig = map[string]interface{}{}
	for _, m := range l.mutators {
		err := m.Mutate(ctx, MutatorOpts{ClusterID: clusterID},
			configTable)
		if err != nil {
			return nil, err
		}
	}

	return ReconfigurePayload(configTable)
}

func ReconfigurePayload(c DataPlaneConfig) ([]byte, error) {
	payload := Map{
		"type":         "reconfigure",
		"config_table": c,
	}

	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	defer writer.Close()
	enc := json.Marshaller.NewEncoder(writer)

	err := enc.Encode(payload)
	if err != nil {
		return nil, fmt.Errorf("json marshal: %v", err)
	}
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("gzip failure: %v", err)
	}
	return buf.Bytes(), nil
}

func CompressPayload(payload []byte) ([]byte, error) {
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

func UncompressPayload(payload []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	buf, err := ioutil.ReadAll(r)
	if err != nil {
		_ = r.Close()
		return nil, err
	}
	err = r.Close()
	if err != nil {
		return nil, err
	}
	return buf, nil
}
