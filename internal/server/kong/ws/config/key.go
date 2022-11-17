package config

import (
	// "context"

	admin "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
)

type KongKeyLoader struct {
	Client admin.KeyServiceClient
}

func (l KongKeyLoader) Name() string {
	return "key"
}

// func (l *KongKeyLoader) Mutate(
// 	ctx context.Context,
// 	opts MutatorOpts,
// 	config DataPlaneConfig,
// ) error {
// 	ctx, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
// 	defer cancel()
//
// 	keys, err := l.Client.ListKeys()
// 	if err != nil {
// 		return err
// 	}
// }
