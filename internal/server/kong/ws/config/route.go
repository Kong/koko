package config

import (
	"context"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	admin "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
)

type KongRouteLoader struct {
	Client admin.RouteServiceClient
}

func (l KongRouteLoader) Name() string {
	return "route"
}

func (l *KongRouteLoader) Mutate(ctx context.Context,
	opts MutatorOpts, config DataPlaneConfig,
) error {
	ctx, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()
	var pageNum int32 = 1
	var allRoutes []*v1.Route
	for {
		resp, err := l.Client.ListRoutes(ctx, &admin.ListRoutesRequest{
			Cluster: &v1.RequestCluster{Id: opts.ClusterID},
			Page: &v1.PaginationRequest{
				Size:   pageSize,
				Number: pageNum,
			},
		})
		if err != nil {
			return err
		}
		allRoutes = append(allRoutes, resp.Items...)
		if resp.Page == nil || resp.Page.NextPageNum == 0 {
			break
		}
		pageNum = resp.Page.NextPageNum
	}
	res := make([]Map, 0, len(allRoutes))
	for _, r := range allRoutes {
		m, err := convert(r)
		if err != nil {
			return err
		}
		translateRouteHeaders(r, m)
		flattenForeign(m, "service")
		res = append(res, m)
	}
	config["routes"] = res
	return nil
}

func translateRouteHeaders(route *v1.Route, m Map) {
	if route.Headers == nil {
		return
	}
	res := map[string][]string{}
	for k, v := range route.Headers {
		res[k] = v.Values
	}
	m["headers"] = res
}
