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

type PluginService struct {
	v1.UnimplementedPluginServiceServer
	CommonOpts
}

func (s *PluginService) GetPlugin(ctx context.Context,
	req *v1.GetPluginRequest) (*v1.GetPluginResponse, error) {
	if req.Id == "" {
		return nil, s.err(util.ErrClient{Message: "required ID is missing"})
	}
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	result := resource.NewPlugin()
	s.logger.With(zap.String("id", req.Id)).Debug("reading plugin by id")
	err = db.Read(ctx, result, store.GetByID(req.Id))
	if err != nil {
		return nil, s.err(err)
	}
	return &v1.GetPluginResponse{
		Item: result.Plugin,
	}, nil
}

func (s *PluginService) CreatePlugin(ctx context.Context,
	req *v1.CreatePluginRequest) (*v1.CreatePluginResponse, error) {
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	res := resource.NewPlugin()
	res.Plugin = req.Item
	if err := db.Create(ctx, res); err != nil {
		return nil, s.err(err)
	}
	util.SetHeader(ctx, http.StatusCreated)
	return &v1.CreatePluginResponse{
		Item: res.Plugin,
	}, nil
}

func (s *PluginService) UpsertPlugin(ctx context.Context,
	req *v1.UpsertPluginRequest) (*v1.UpsertPluginResponse, error) {
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	res := resource.NewPlugin()
	res.Plugin = req.Item
	if err := db.Upsert(ctx, res); err != nil {
		return nil, s.err(err)
	}
	return &v1.UpsertPluginResponse{
		Item: res.Plugin,
	}, nil
}

func (s *PluginService) DeletePlugin(ctx context.Context,
	req *v1.DeletePluginRequest) (*v1.DeletePluginResponse, error) {
	if req.Id == "" {
		return nil, s.err(util.ErrClient{Message: "required ID is missing"})
	}
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	err = db.Delete(ctx, store.DeleteByID(req.Id),
		store.DeleteByType(resource.TypePlugin))
	if err != nil {
		return nil, s.err(err)
	}
	util.SetHeader(ctx, http.StatusNoContent)
	return &v1.DeletePluginResponse{}, nil
}

func (s *PluginService) ListPlugins(ctx context.Context,
	req *v1.ListPluginsRequest) (*v1.ListPluginsResponse, error) {
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}

	serviceID := strings.TrimSpace(req.GetServiceId())
	routeID := strings.TrimSpace(req.GetRouteId())
	listFn := []store.ListOptsFunc{}
	if len(serviceID) > 0 && len(routeID) > 0 {
		return nil, s.err(util.ErrClient{Message: "service_id and route_id are mutually exclusive"})
	}

	if len(serviceID) > 0 {
		if _, err := uuid.Parse(serviceID); err != nil {
			return nil, s.err(util.ErrClient{
				Message: fmt.Sprintf("service_id '%s' is not a UUID", req.GetServiceId()),
			})
		}
		listFn = append(listFn, store.ListFor(resource.TypeService, serviceID))
	} else if len(routeID) > 0 {
		if _, err := uuid.Parse(routeID); err != nil {
			return nil, s.err(util.ErrClient{
				Message: fmt.Sprintf("route_id '%s' is not a UUID", req.GetRouteId()),
			})
		}
		listFn = append(listFn, store.ListFor(resource.TypeRoute, routeID))
	}

	list := resource.NewList(resource.TypePlugin)
	listOptFns, err := listOptsFromReq(req.Pagination)
	if err != nil {
		return nil, s.err(util.ErrClient{Message: err.Error()})
	}
	listFn = append(listFn, listOptFns...) // combine all the list options

	if err := db.List(ctx, list, listFn...); err != nil {
		return nil, s.err(err)
	}

	return &v1.ListPluginsResponse{
		Items:      pluginsFromObjects(list.GetAll()),
		Pagination: getPaginationResponse(list.GetTotalCount(), list.GetNextPage()),
	}, nil
}

func (s *PluginService) err(err error) error {
	return util.HandleErr(s.logger, err)
}

func pluginsFromObjects(objects []model.Object) []*pbModel.Plugin {
	res := make([]*pbModel.Plugin, 0, len(objects))
	for _, object := range objects {
		plugin, ok := object.Resource().(*pbModel.Plugin)
		if !ok {
			panic(fmt.Sprintf("expected type '%T' but got '%T'",
				&pbModel.Plugin{}, object.Resource()))
		}
		res = append(res, plugin)
	}
	return res
}
