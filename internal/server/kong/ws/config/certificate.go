package config

import (
	"context"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	admin "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
)

type KongCertificateLoader struct {
	Client admin.CertificateServiceClient
}

func (l KongCertificateLoader) Name() string {
	return "certificate"
}

func (l *KongCertificateLoader) Mutate(ctx context.Context,
	opts MutatorOpts, config DataPlaneConfig,
) error {
	ctx, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()

	var pageNum int32
	var allCertificates []*v1.Certificate
	for {
		resp, err := l.Client.ListCertificates(ctx, &admin.ListCertificatesRequest{
			Cluster: &v1.RequestCluster{Id: opts.ClusterID},
			Page: &v1.PaginationRequest{
				Size:   pageSize,
				Number: pageNum,
			},
		})
		if err != nil {
			return err
		}
		allCertificates = append(allCertificates, resp.Items...)
		if resp.Page == nil || resp.Page.NextPageNum == 0 {
			break
		}
	}
	res := make([]Map, 0, len(allCertificates))
	for _, r := range allCertificates {
		m, err := convert(r)
		if err != nil {
			return err
		}
		delete(m, "updated_at")
		res = append(res, m)
	}
	config["certificates"] = res
	return nil
}
