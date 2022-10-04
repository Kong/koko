package config

import (
	"context"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	admin "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
)

type KongVaultLoader struct {
	Client admin.VaultServiceClient
}

func (l KongVaultLoader) Name() string {
	return "vault"
}

func (l *KongVaultLoader) Mutate(ctx context.Context,
	opts MutatorOpts, config DataPlaneConfig,
) error {
	ctx, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()

	var pageNum int32 = 1
	var allVaults []*v1.Vault
	for {
		resp, err := l.Client.ListVaults(ctx, &admin.ListVaultsRequest{
			Cluster: &v1.RequestCluster{Id: opts.ClusterID},
			Page: &v1.PaginationRequest{
				Size:   pageSize,
				Number: pageNum,
			},
		})
		if err != nil {
			return err
		}
		allVaults = append(allVaults, resp.Items...)
		if resp.Page == nil || resp.Page.NextPageNum == 0 {
			break
		}
		pageNum = resp.Page.NextPageNum
	}
	res := make([]Map, 0, len(allVaults))
	for _, r := range allVaults {
		m, err := convert(r)
		if err != nil {
			return err
		}
		delete(m, "updated_at")
		res = append(res, m)
	}
	config["vaults"] = res
	return nil
}
