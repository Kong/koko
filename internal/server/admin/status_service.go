package admin

import (
	"context"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
	v2 "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v2"
	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/server/util"
	"github.com/kong/koko/internal/store"
	"go.uber.org/zap"
)

type StatusServiceV1 struct {
	v1.UnimplementedStatusServiceServer
	CommonOpts
}

func (s *StatusServiceV1) GetHash(ctx context.Context,
	req *v1.GetHashRequest,
) (*v1.GetHashResponse, error) {
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}

	result := resource.NewHash()
	err = db.Read(ctx, result, store.GetByID(result.ID()))
	if err != nil {
		return nil, s.err(ctx, err)
	}
	return &v1.GetHashResponse{
		ExpectedHash: result.Hash.ExpectedHash,
	}, nil
}

func (s *StatusServiceV1) err(ctx context.Context, err error) error {
	return util.HandleErr(ctx, s.logger(ctx), err)
}

func (s *StatusServiceV1) logger(ctx context.Context) *zap.Logger {
	return util.LoggerFromContext(ctx).With(s.loggerFields...)
}

type StatusServiceV2 struct {
	v2.UnimplementedStatusServiceServer
	CommonOpts
}

func (s *StatusServiceV2) GetHash(ctx context.Context,
	req *v2.GetHashRequest,
) (*v2.GetHashResponse, error) {
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}

	result := resource.NewHash()
	err = db.Read(ctx, result, store.GetByID(result.ID()))
	if err != nil {
		return nil, s.err(ctx, err)
	}
	return &v2.GetHashResponse{
		Item: result.Hash,
	}, nil
}

func (s *StatusServiceV2) err(ctx context.Context, err error) error {
	return util.HandleErr(ctx, s.logger(ctx), err)
}

func (s *StatusServiceV2) logger(ctx context.Context) *zap.Logger {
	return util.LoggerFromContext(ctx).With(s.loggerFields...)
}
