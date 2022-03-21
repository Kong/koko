package config

import (
	"context"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	admin "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
)

type KongTargetLoader struct {
	Client admin.TargetServiceClient
}

func (l KongTargetLoader) Name() string {
	return "target"
}

func (l *KongTargetLoader) Mutate(ctx context.Context,
	opts MutatorOpts, config DataPlaneConfig,
) error {
	ctx, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()
	var pageNum int32
	var allTargets []*v1.Target
	for {
		resp, err := l.Client.ListTargets(ctx, &admin.ListTargetsRequest{
			Cluster: &v1.RequestCluster{Id: opts.ClusterID},
			Page: &v1.PaginationRequest{
				Size:   pageSize,
				Number: pageNum,
			},
		})
		if err != nil {
			return err
		}
		allTargets = append(allTargets, resp.Items...)
		if resp.Page == nil || resp.Page.NextPageNum == 0 {
			break
		}
	}
	res := make([]Map, 0, len(allTargets))
	for _, r := range allTargets {
		m, err := convert(r)
		if err != nil {
			return err
		}
		delete(m, "updated_at")
		flattenForeign(m, "upstream")
		res = append(res, m)
	}
	config["targets"] = res
	return nil
}
