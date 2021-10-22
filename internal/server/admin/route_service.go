package admin

import (
	"context"
	"fmt"
	"net/http"

	pbModel "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/store"
	"go.uber.org/zap"
)

type RouteService struct {
	v1.UnimplementedRouteServiceServer
	CommonOpts
}

func (s *RouteService) GetRoute(ctx context.Context,
	req *v1.GetRouteRequest) (*v1.GetRouteResponse, error) {
	if req.Id == "" {
		return nil, s.err(ErrClient{"required ID is missing"})
	}
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	result := resource.NewRoute()
	s.logger.With(zap.String("id", req.Id)).Debug("reading route by id")
	ctx, cancel := context.WithTimeout(ctx, dbQueryTimeout)
	defer cancel()
	err = db.Read(ctx, result, store.GetByID(req.Id))
	if err != nil {
		return nil, s.err(err)
	}
	return &v1.GetRouteResponse{
		Item: result.Route,
	}, nil
}

func (s *RouteService) CreateRoute(ctx context.Context,
	req *v1.CreateRouteRequest) (*v1.CreateRouteResponse, error) {
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	res := resource.NewRoute()
	res.Route = req.Item
	ctx, cancel := context.WithTimeout(ctx, dbQueryTimeout)
	defer cancel()
	if err := db.Create(ctx, res); err != nil {
		return nil, s.err(err)
	}
	setHeader(ctx, http.StatusCreated)
	return &v1.CreateRouteResponse{
		Item: res.Route,
	}, nil
}

func (s *RouteService) DeleteRoute(ctx context.Context,
	req *v1.DeleteRouteRequest) (*v1.DeleteRouteResponse, error) {
	if req.Id == "" {
		return nil, s.err(ErrClient{"required ID is missing"})
	}
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, dbQueryTimeout)
	defer cancel()
	err = db.Delete(ctx, store.DeleteByID(req.Id),
		store.DeleteByType(resource.TypeRoute))
	if err != nil {
		return nil, s.err(err)
	}
	setHeader(ctx, http.StatusNoContent)
	return &v1.DeleteRouteResponse{}, nil
}

func (s *RouteService) ListRoutes(ctx context.Context,
	req *v1.ListRoutesRequest) (*v1.ListRoutesResponse, error) {
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	list := resource.NewList(resource.TypeRoute)
	ctx, cancel := context.WithTimeout(ctx, dbQueryTimeout)
	defer cancel()
	if err := db.List(ctx, list); err != nil {
		return nil, s.err(err)
	}
	return &v1.ListRoutesResponse{
		Items: routesFromObjects(list.GetAll()),
	}, nil
}

func (s *RouteService) err(err error) error {
	return handleErr(s.logger, err)
}

func routesFromObjects(objects []model.Object) []*pbModel.Route {
	res := make([]*pbModel.Route, 0, len(objects))
	for _, object := range objects {
		route, ok := object.Resource().(*pbModel.Route)
		if !ok {
			panic(fmt.Sprintf("expected type '%T' but got '%T'",
				&pbModel.Route{}, object.Resource()))
		}
		res = append(res, route)
	}
	return res
}
