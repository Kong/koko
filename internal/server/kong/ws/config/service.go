package config

import (
	"context"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	admin "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
)

type KongServiceLoader struct {
	Client admin.ServiceServiceClient
}

func (l KongServiceLoader) Name() string {
	return "service"
}

func (l *KongServiceLoader) Mutate(ctx context.Context,
	opts MutatorOpts, config DataPlaneConfig,
) error {
	ctx, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()
	var pageNum int32 = 1
	var allServices []*v1.Service
	for {
		resp, err := l.Client.ListServices(ctx, &admin.ListServicesRequest{
			Cluster: &v1.RequestCluster{Id: opts.ClusterID},
			Page: &v1.PaginationRequest{
				Size:   pageSize,
				Number: pageNum,
			},
		})
		if err != nil {
			return err
		}
		allServices = append(allServices, resp.Items...)
		if resp.Page == nil || resp.Page.NextPageNum == 0 {
			break
		}
		pageNum = resp.Page.NextPageNum
	}
	res := make([]Map, 0)
	for _, svc := range allServices {
		m, err := convert(svc)
		if err != nil {
			return err
		}
		flattenForeign(m, "client_certificate")
		res = append(res, m)
	}
	config["services"] = res
	return nil
}
