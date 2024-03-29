package config

import (
	"context"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	admin "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
)

type KongKeySetLoader struct {
	Client admin.KeySetServiceClient
}

func (l KongKeySetLoader) Name() string {
	return "keyset"
}

func (l *KongKeySetLoader) Mutate(
	ctx context.Context,
	opts MutatorOpts,
	config DataPlaneConfig,
) error {
	ctx, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()

	var pageNum int32 = 1
	var allKeySets []*v1.KeySet
	for {
		keysets, err := l.Client.ListKeySets(ctx, &admin.ListKeySetsRequest{
			Cluster: &v1.RequestCluster{Id: opts.ClusterID},
			Page: &v1.PaginationRequest{
				Size:   pageSize,
				Number: pageNum,
			},
		})
		if err != nil {
			return err
		}
		allKeySets = append(allKeySets, keysets.Items...)
		if keysets.Page == nil || keysets.Page.NextPageNum == 0 {
			break
		}
		pageNum = keysets.Page.NextPageNum
	}

	res := make([]Map, len(allKeySets))
	for i, r := range allKeySets {
		m, err := convert(r)
		if err != nil {
			return err
		}
		res[i] = m
	}

	config["key_sets"] = res
	return nil
}
