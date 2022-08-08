package config

import (
	"context"

	"github.com/google/uuid"
)

type ParametersLoader struct {
	ClusterID string
}

// NewParametersLoader creates a parameters configuration loader. It requires a valid UUID.
func NewParametersLoader(clusterID string) (*ParametersLoader, error) {
	// invalid UUID always returned by Parse as 00000000-0000-0000-0000-000000000000
	if _, err := uuid.Parse(clusterID); err != nil {
		return nil, err
	}
	return &ParametersLoader{
		ClusterID: clusterID,
	}, nil
}

// Name returns the name of the loader.
func (l ParametersLoader) Name() string {
	return "parameters"
}

// Mutate updates the config parameters, a map of key:value pairs.
func (l *ParametersLoader) Mutate(_ context.Context,
	_ MutatorOpts, config DataPlaneConfig,
) error {
	config["parameters"] = []Map{{"key": "cluster_id", "value": l.ClusterID}}
	return nil
}
