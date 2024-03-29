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

type RouteService struct {
	v1.UnimplementedRouteServiceServer
	CommonOpts
}

func (s *RouteService) GetRoute(ctx context.Context,
	req *v1.GetRouteRequest,
) (*v1.GetRouteResponse, error) {
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	result := resource.NewRoute()
	err = getEntityByIDOrName(ctx, req.Id, result, store.GetByName(req.Id), db, s.logger(ctx))
	if err != nil {
		return nil, s.err(ctx, err)
	}
	return &v1.GetRouteResponse{
		Item: result.Route,
	}, nil
}

func (s *RouteService) CreateRoute(ctx context.Context,
	req *v1.CreateRouteRequest,
) (*v1.CreateRouteResponse, error) {
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	res := resource.NewRoute()
	res.Route = req.Item
	if err := db.Create(ctx, res); err != nil {
		return nil, s.err(ctx, err)
	}
	util.SetHeader(ctx, http.StatusCreated)
	return &v1.CreateRouteResponse{
		Item: res.Route,
	}, nil
}

func (s *RouteService) UpsertRoute(ctx context.Context,
	req *v1.UpsertRouteRequest,
) (*v1.UpsertRouteResponse, error) {
	if err := validUUID(req.Item.Id); err != nil {
		return nil, s.err(ctx, err)
	}
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	res := resource.NewRoute()
	res.Route = req.Item
	if err := db.Upsert(ctx, res); err != nil {
		return nil, s.err(ctx, err)
	}
	return &v1.UpsertRouteResponse{
		Item: res.Route,
	}, nil
}

func (s *RouteService) DeleteRoute(ctx context.Context,
	req *v1.DeleteRouteRequest,
) (*v1.DeleteRouteResponse, error) {
	if err := validUUID(req.Id); err != nil {
		return nil, s.err(ctx, err)
	}
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	err = db.Delete(ctx, store.DeleteByID(req.Id),
		store.DeleteByType(resource.TypeRoute))
	if err != nil {
		return nil, s.err(ctx, err)
	}
	util.SetHeader(ctx, http.StatusNoContent)
	return &v1.DeleteRouteResponse{}, nil
}

func (s *RouteService) ListRoutes(ctx context.Context,
	req *v1.ListRoutesRequest,
) (*v1.ListRoutesResponse, error) {
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}

	serviceID := strings.TrimSpace(req.ServiceId)
	listFn := []store.ListOptsFunc{}
	if len(serviceID) > 0 {
		if _, err := uuid.Parse(serviceID); err != nil {
			return nil, s.err(ctx, util.ErrClient{
				Message: fmt.Sprintf("service_id '%s' is not a UUID", req.ServiceId),
			})
		}
		listFn = append(listFn, store.ListFor(resource.TypeService, serviceID))
	}

	list := resource.NewList(resource.TypeRoute)
	listOptFns, err := ListOptsFromReq(req.Page)
	if err != nil {
		return nil, s.err(ctx, err)
	}

	listFn = append(listFn, listOptFns...)

	if err := db.List(ctx, list, listFn...); err != nil {
		return nil, s.err(ctx, err)
	}

	return &v1.ListRoutesResponse{
		Items: routesFromObjects(list.GetAll()),
		Page:  getPaginationResponse(list.GetTotalCount(), list.GetNextPage()),
	}, nil
}

func (s *RouteService) err(ctx context.Context, err error) error {
	return util.HandleErr(ctx, s.logger(ctx), err)
}

func (s *RouteService) logger(ctx context.Context) *zap.Logger {
	return util.LoggerFromContext(ctx).With(s.loggerFields...)
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
