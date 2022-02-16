package config

import (
	"context"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	admin "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
)

type KongUpstreamLoader struct {
	Client admin.UpstreamServiceClient
}

func (l KongUpstreamLoader) Name() string {
	return "upstream"
}

func (l *KongUpstreamLoader) Mutate(ctx context.Context,
	opts MutatorOpts, config DataPlaneConfig) error {
	ctx, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()
	var pageNum int32
	var allUpstreams []*v1.Upstream
	for {
		resp, err := l.Client.ListUpstreams(ctx, &admin.ListUpstreamsRequest{
			Cluster: &v1.RequestCluster{Id: opts.ClusterID},
			Page: &v1.PaginationRequest{
				Size:   pageSize,
				Number: pageNum,
			},
		})
		if err != nil {
			return err
		}
		allUpstreams = append(allUpstreams, resp.Items...)
		if resp.Page == nil || resp.Page.NextPageNum == 0 {
			break
		}
	}
	res := make([]Map, 0, len(allUpstreams))
	for _, r := range allUpstreams {
		m, err := convert(r)
		if err != nil {
			return err
		}
		flattenForeign(m, "service")
		delete(m, "updated_at")
		res = append(res, m)
	}
	config["upstreams"] = res
	return nil
}
