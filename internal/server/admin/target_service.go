package admin

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	pbModel "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/server/util"
	"github.com/kong/koko/internal/store"
	"go.uber.org/zap"
)

type TargetService struct {
	v1.UnimplementedTargetServiceServer
	CommonOpts
}

func (s *TargetService) GetTarget(ctx context.Context,
	req *v1.GetTargetRequest) (*v1.GetTargetResponse, error) {
	if err := validUUID(req.Id); err != nil {
		return nil, s.err(err)
	}
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	result := resource.NewTarget()
	s.logger.With(zap.String("id", req.Id)).Debug("reading target by id")
	err = db.Read(ctx, result, store.GetByID(req.Id))
	if err != nil {
		return nil, s.err(err)
	}
	return &v1.GetTargetResponse{
		Item: result.Target,
	}, nil
}

func (s *TargetService) CreateTarget(ctx context.Context,
	req *v1.CreateTargetRequest) (*v1.CreateTargetResponse, error) {
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	res := resource.NewTarget()
	res.Target = req.Item
	if err := db.Create(ctx, res); err != nil {
		return nil, s.err(err)
	}
	util.SetHeader(ctx, http.StatusCreated)
	return &v1.CreateTargetResponse{
		Item: res.Target,
	}, nil
}

func (s *TargetService) UpsertTarget(ctx context.Context,
	req *v1.UpsertTargetRequest) (*v1.UpsertTargetResponse, error) {
	if err := validUUID(req.Item.Id); err != nil {
		return nil, s.err(err)
	}
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	res := resource.NewTarget()
	res.Target = req.Item
	if err := db.Upsert(ctx, res); err != nil {
		return nil, s.err(err)
	}
	return &v1.UpsertTargetResponse{
		Item: res.Target,
	}, nil
}

func (s *TargetService) DeleteTarget(ctx context.Context,
	req *v1.DeleteTargetRequest) (*v1.DeleteTargetResponse, error) {
	if err := validUUID(req.Id); err != nil {
		return nil, s.err(err)
	}
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	err = db.Delete(ctx, store.DeleteByID(req.Id),
		store.DeleteByType(resource.TypeTarget))
	if err != nil {
		return nil, s.err(err)
	}
	util.SetHeader(ctx, http.StatusNoContent)
	return &v1.DeleteTargetResponse{}, nil
}

func (s *TargetService) ListTargets(ctx context.Context,
	req *v1.ListTargetsRequest) (*v1.ListTargetsResponse, error) {
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}

	upstreamID := strings.TrimSpace(req.UpstreamId)
	var listFn []store.ListOptsFunc
	if len(upstreamID) > 0 {
		if _, err := uuid.Parse(upstreamID); err != nil {
			return nil, s.err(util.ErrClient{
				Message: fmt.Sprintf("upstream_id '%s' is not a UUID",
					req.UpstreamId),
			})
		}
		listFn = append(listFn, store.ListFor(resource.TypeUpstream, upstreamID))
	}

	list := resource.NewList(resource.TypeTarget)

	listOptFns, err := listOptsFromReq(req.Page)
	if err != nil {
		return nil, s.err(util.ErrClient{Message: err.Error()})
	}

	listFn = append(listFn, listOptFns...)

	if err := db.List(ctx, list, listFn...); err != nil {
		return nil, s.err(err)
	}
	return &v1.ListTargetsResponse{
		Items: targetsFromObjects(list.GetAll()),
		Page:  getPaginationResponse(list.GetTotalCount(), list.GetNextPage()),
	}, nil
}

func (s *TargetService) err(err error) error {
	return util.HandleErr(s.logger, err)
}

func targetsFromObjects(objects []model.Object) []*pbModel.Target {
	res := make([]*pbModel.Target, 0, len(objects))
	for _, object := range objects {
		target, ok := object.Resource().(*pbModel.Target)
		if !ok {
			panic(fmt.Sprintf("expected type '%T' but got '%T'",
				&pbModel.Target{}, object.Resource()))
		}
		res = append(res, target)
	}
	return res
}
