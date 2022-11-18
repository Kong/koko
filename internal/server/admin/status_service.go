package admin

import (
	"context"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/server/util"
	"github.com/kong/koko/internal/store"
	"go.uber.org/zap"
)

type StatusService struct {
	v1.UnimplementedStatusServiceServer
	CommonOpts
}

func (s *StatusService) GetHash(ctx context.Context,
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
		CreatedAt:    result.Hash.CreatedAt,
		UpdatedAt:    result.Hash.UpdatedAt,
	}, nil
}

func (s *StatusService) err(ctx context.Context, err error) error {
	return util.HandleErr(ctx, s.logger(ctx), err)
}

func (s *StatusService) logger(ctx context.Context) *zap.Logger {
	return util.LoggerFromContext(ctx).With(s.loggerFields...)
}
