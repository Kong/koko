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
	req *v1.GetUpstreamRequest) (*v1.GetUpstreamResponse, error) {
	if req.Id == "" {
		return nil, s.err(util.ErrClient{Message: "required ID is missing"})
	}
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	result := resource.NewUpstream()
	s.logger.With(zap.String("id", req.Id)).Debug("reading upstream by id")
	err = db.Read(ctx, result, store.GetByID(req.Id))
	if err != nil {
		return nil, s.err(err)
	}
	return &v1.GetUpstreamResponse{
		Item: result.Upstream,
	}, nil
}

func (s *UpstreamService) CreateUpstream(ctx context.Context,
	req *v1.CreateUpstreamRequest) (*v1.CreateUpstreamResponse, error) {
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	res := resource.NewUpstream()
	res.Upstream = req.Item
	if err := db.Create(ctx, res); err != nil {
		return nil, s.err(err)
	}
	util.SetHeader(ctx, http.StatusCreated)
	return &v1.CreateUpstreamResponse{
		Item: res.Upstream,
	}, nil
}

func (s *UpstreamService) UpsertUpstream(ctx context.Context,
	req *v1.UpsertUpstreamRequest) (*v1.UpsertUpstreamResponse, error) {
	if req.Item.Id == "" {
		return nil, s.err(util.ErrClient{Message: "required ID is missing"})
	}
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	res := resource.NewUpstream()
	res.Upstream = req.Item
	if err := db.Upsert(ctx, res); err != nil {
		return nil, s.err(err)
	}
	return &v1.UpsertUpstreamResponse{
		Item: res.Upstream,
	}, nil
}

func (s *UpstreamService) DeleteUpstream(ctx context.Context,
	req *v1.DeleteUpstreamRequest) (*v1.DeleteUpstreamResponse, error) {
	if req.Id == "" {
		return nil, s.err(util.ErrClient{Message: "required ID is missing"})
	}
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	err = db.Delete(ctx, store.DeleteByID(req.Id),
		store.DeleteByType(resource.TypeUpstream))
	if err != nil {
		return nil, s.err(err)
	}
	util.SetHeader(ctx, http.StatusNoContent)
	return &v1.DeleteUpstreamResponse{}, nil
}

func (s *UpstreamService) ListUpstreams(ctx context.Context,
	req *v1.ListUpstreamsRequest) (*v1.ListUpstreamsResponse, error) {
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	list := resource.NewList(resource.TypeUpstream)
	listOptFns, err := listOptsFromReq(req.Page)
	if err != nil {
		return nil, s.err(util.ErrClient{Message: err.Error()})
	}

	if err := db.List(ctx, list, listOptFns...); err != nil {
		return nil, s.err(err)
	}

	return &v1.ListUpstreamsResponse{
		Items: upstreamsFromObjects(list.GetAll()),
		Page:  getPaginationResponse(list.GetTotalCount(), list.GetNextPage()),
	}, nil
}

func (s *UpstreamService) err(err error) error {
	return util.HandleErr(s.logger, err)
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
