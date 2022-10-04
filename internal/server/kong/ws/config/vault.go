package config

import (
	"context"
	"fmt"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	admin "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
	model "github.com/kong/koko/internal/resource"
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
		res = append(res, flattenVaultConfig(m))
	}
	config["vaults"] = res
	return nil
}

func flattenVaultConfig(m Map) Map {
	if _, ok := m["config"]; !ok {
		return m
	}

	config, ok := m["config"].(map[string]interface{})
	if !ok {
		panic(fmt.Sprintf("'%s' key is not a JSON object ("+
			"map[string]interface{}", "config"))
	}

	for _, v := range append(model.VaultTypes, model.EnterpriseVaultTypes...) {
		vt, ok := v.(string)
		if !ok {
			panic(fmt.Sprintf("expected vaultType to be string but got %T", v))
		}
		if _, ok := config[vt]; ok {
			m["config"] = config[vt]
			return m
		}
	}
	return m
}
