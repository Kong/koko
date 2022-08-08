package config

import (
	"context"
)

type VersionLoader struct{}

func (l VersionLoader) Name() string {
	return "version"
}

func (l *VersionLoader) Mutate(_ context.Context,
	_ MutatorOpts, config DataPlaneConfig,
) error {
	config["_format_version"] = "1.1"
	config["_transform"] = false
	return nil
}
