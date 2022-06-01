package admin

import (
	"context"
	"fmt"
	"net/http"

	pbModel "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/server/util"
	"github.com/kong/koko/internal/store"
	"go.uber.org/zap"
)

type UpstreamService struct {
	v1.UnimplementedUpstreamServiceServer
	CommonOpts
}

func (s *UpstreamService) GetUpstream(ctx context.Context,
	req *v1.GetUpstreamRequest,
) (*v1.GetUpstreamResponse, error) {
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	result := resource.NewUpstream()
	err = getEntityByIDOrName(ctx, req.Id, result, store.GetByName(req.Id), db, s.logger(ctx))
	if err != nil {
		return nil, s.err(ctx, err)
	}
	return &v1.GetUpstreamResponse{
		Item: result.Upstream,
	}, nil
}

func (s *UpstreamService) CreateUpstream(ctx context.Context,
	req *v1.CreateUpstreamRequest,
) (*v1.CreateUpstreamResponse, error) {
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	res := resource.NewUpstream()
	res.Upstream = req.Item
	if err := db.Create(ctx, res); err != nil {
		return nil, s.err(ctx, err)
	}
	util.SetHeader(ctx, http.StatusCreated)
	return &v1.CreateUpstreamResponse{
		Item: res.Upstream,
	}, nil
}

func (s *UpstreamService) UpsertUpstream(ctx context.Context,
	req *v1.UpsertUpstreamRequest,
) (*v1.UpsertUpstreamResponse, error) {
	if err := validUUID(req.Item.Id); err != nil {
		return nil, s.err(ctx, err)
	}
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	res := resource.NewUpstream()
	res.Upstream = req.Item
	if err := db.Upsert(ctx, res); err != nil {
		return nil, s.err(ctx, err)
	}
	return &v1.UpsertUpstreamResponse{
		Item: res.Upstream,
	}, nil
}

func (s *UpstreamService) DeleteUpstream(ctx context.Context,
	req *v1.DeleteUpstreamRequest,
) (*v1.DeleteUpstreamResponse, error) {
	if err := validUUID(req.Id); err != nil {
		return nil, s.err(ctx, err)
	}
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	err = db.Delete(ctx, store.DeleteByID(req.Id),
		store.DeleteByType(resource.TypeUpstream))
	if err != nil {
		return nil, s.err(ctx, err)
	}
	util.SetHeader(ctx, http.StatusNoContent)
	return &v1.DeleteUpstreamResponse{}, nil
}

func (s *UpstreamService) ListUpstreams(ctx context.Context,
	req *v1.ListUpstreamsRequest,
) (*v1.ListUpstreamsResponse, error) {
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	list := resource.NewList(resource.TypeUpstream)
	listOptFns, err := listOptsFromReq(req.Page)
	if err != nil {
		return nil, s.err(ctx, err)
	}

	if err := db.List(ctx, list, listOptFns...); err != nil {
		return nil, s.err(ctx, err)
	}

	return &v1.ListUpstreamsResponse{
		Items: upstreamsFromObjects(list.GetAll()),
		Page:  getPaginationResponse(list.GetTotalCount(), list.GetNextPage()),
	}, nil
}

func (s *UpstreamService) err(ctx context.Context, err error) error {
	return util.HandleErr(ctx, s.logger(ctx), err)
}

func (s *UpstreamService) logger(ctx context.Context) *zap.Logger {
	return util.LoggerFromContext(ctx).With(s.loggerFields...)
}

func upstreamsFromObjects(objects []model.Object) []*pbModel.Upstream {
	res := make([]*pbModel.Upstream, 0, len(objects))
	for _, object := range objects {
		upstream, ok := object.Resource().(*pbModel.Upstream)
		if !ok {
			panic(fmt.Sprintf("expected type '%T' but got '%T'",
				&pbModel.Upstream{}, object.Resource()))
		}
		res = append(res, upstream)
	}
	return res
}
