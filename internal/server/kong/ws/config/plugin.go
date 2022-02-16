package config

import (
	"context"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	admin "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
)

type KongPluginLoader struct {
	Client admin.PluginServiceClient
}

func (l KongPluginLoader) Name() string {
	return "plugin"
}

func (l *KongPluginLoader) Mutate(ctx context.Context,
	opts MutatorOpts, config DataPlaneConfig) error {
	ctx, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()
	var pageNum int32
	var allPlugins []*v1.Plugin
	for {
		resp, err := l.Client.ListPlugins(ctx, &admin.ListPluginsRequest{
			Cluster: &v1.RequestCluster{Id: opts.ClusterID},
			Page: &v1.PaginationRequest{
				Size:   pageSize,
				Number: pageNum,
			},
		})
		if err != nil {
			return err
		}
		allPlugins = append(allPlugins, resp.Items...)
		if resp.Page == nil || resp.Page.NextPageNum == 0 {
			break
		}
	}
	res := make([]Map, 0, len(allPlugins))
	for _, r := range allPlugins {
		m, err := convert(r)
		if err != nil {
			return err
		}
		flattenForeign(m, "service")
		flattenForeign(m, "route")
		delete(m, "updated_at")
		res = append(res, m)
	}
	config["plugins"] = res
	return nil
}
