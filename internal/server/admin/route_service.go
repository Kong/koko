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
	result := resource.NewRoute()
	s.logger.With(zap.String("id", req.Id)).Debug("reading route by id")
	err := s.store.Read(ctx, result, store.GetByID(req.Id))
	if err != nil {
		return nil, s.err(err)
	}
	return &v1.GetRouteResponse{
		Item: result.Route,
	}, nil
}

func (s *RouteService) CreateRoute(ctx context.Context,
	req *v1.CreateRouteRequest) (*v1.CreateRouteResponse, error) {
	res := resource.NewRoute()
	res.Route = req.Item
	err := s.store.Create(ctx, res)
	if err != nil {
		return nil, s.err(err)
	}
	setHeader(ctx, http.StatusCreated)
	return &v1.CreateRouteResponse{
		Item: res.Route,
	}, nil
}

func (s *RouteService) DeleteRoute(ctx context.Context,
	request *v1.DeleteRouteRequest) (*v1.DeleteRouteResponse, error) {
	err := s.store.Delete(ctx, store.DeleteByID(request.Id),
		store.DeleteByType(resource.TypeRoute))
	if err != nil {
		return nil, s.err(err)
	}
	setHeader(ctx, http.StatusNoContent)
	return &v1.DeleteRouteResponse{}, nil
}

func (s *RouteService) ListRoutes(ctx context.Context,
	_ *v1.ListRoutesRequest) (*v1.ListRoutesResponse, error) {
	list := resource.NewList(resource.TypeRoute)
	if err := s.store.List(ctx, list); err != nil {
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
