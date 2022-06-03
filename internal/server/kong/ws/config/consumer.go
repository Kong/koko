package config

import (
	"context"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	admin "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
)

type KongConsumerLoader struct {
	Client admin.ConsumerServiceClient
}

func (l KongConsumerLoader) Name() string {
	return "consumer"
}

// Mutate reads the Consumer data from CP persistence store and
// populates the read data into config.
func (l *KongConsumerLoader) Mutate(ctx context.Context,
	opts MutatorOpts, config DataPlaneConfig,
) error {
	ctx, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()

	var pageNum int32 = 1
	var allConsumers []*v1.Consumer
	for {
		resp, err := l.Client.ListConsumers(ctx, &admin.ListConsumersRequest{
			Cluster: &v1.RequestCluster{Id: opts.ClusterID},
			Page: &v1.PaginationRequest{
				Size:   pageSize,
				Number: pageNum,
			},
		})
		if err != nil {
			return err
		}
		allConsumers = append(allConsumers, resp.Items...)
		if resp.Page == nil || resp.Page.NextPageNum == 0 {
			break
		}
		pageNum = resp.Page.NextPageNum
	}
	res := make([]Map, 0, len(allConsumers))
	for _, r := range allConsumers {
		m, err := convert(r)
		if err != nil {
			return err
		}
		delete(m, "updated_at")
		res = append(res, m)
	}
	config["consumers"] = res
	return nil
}
