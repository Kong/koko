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
	Load(ctx context.Context, clusterID string) (State, error)
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
) (State, error) {
	var configTable DataPlaneConfig = map[string]interface{}{}
	for _, m := range l.mutators {
		err := m.Mutate(ctx, MutatorOpts{ClusterID: clusterID},
			configTable)
		if err != nil {
			return State{}, err
		}
	}

	return ReconfigurePayload(configTable)
}

func ReconfigurePayload(c DataPlaneConfig) (State, error) {
	hash := configHash(c)
	payload := Map{
		"type":         "reconfigure",
		"config_table": c,
		"config_hash":  hash,
	}

	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	defer writer.Close()

	err := json.Marshaller.NewEncoder(writer).Encode(payload)
	if err != nil {
		return State{}, fmt.Errorf("json marshal: %v", err)
	}
	err = writer.Close()
	if err != nil {
		return State{}, fmt.Errorf("gzip failure: %v", err)
	}
	return State{
		Payload: buf.Bytes(),
		Hash:    hash,
	}, nil
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
