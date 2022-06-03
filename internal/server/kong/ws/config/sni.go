package config

import (
	"context"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	admin "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
)

type KongSNILoader struct {
	Client admin.SNIServiceClient
}

func (l KongSNILoader) Name() string {
	return "sni"
}

func (l *KongSNILoader) Mutate(ctx context.Context,
	opts MutatorOpts, config DataPlaneConfig,
) error {
	ctx, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()

	var pageNum int32 = 1
	var allSNIs []*v1.SNI
	for {
		resp, err := l.Client.ListSNIs(ctx, &admin.ListSNIsRequest{
			Cluster: &v1.RequestCluster{Id: opts.ClusterID},
			Page: &v1.PaginationRequest{
				Size:   pageSize,
				Number: pageNum,
			},
		})
		if err != nil {
			return err
		}
		allSNIs = append(allSNIs, resp.Items...)
		if resp.Page == nil || resp.Page.NextPageNum == 0 {
			break
		}
		pageNum = resp.Page.NextPageNum
	}
	res := make([]Map, 0, len(allSNIs))
	for _, r := range allSNIs {
		m, err := convert(r)
		if err != nil {
			return err
		}
		flattenForeign(m, "certificate")
		delete(m, "updated_at")
		res = append(res, m)
	}
	config["snis"] = res
	return nil
}
