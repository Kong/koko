package admin

import (
	"context"
	"fmt"
	"net/http"

	pb "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/kong/koko/internal/plugin"
	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/server/util"
	"github.com/kong/koko/internal/store"
	"go.uber.org/zap"
)

type PluginSchemaService struct {
	v1.UnimplementedPluginSchemaServiceServer
	CommonOpts
	validator plugin.Validator
}

const nameFieldName = "name"

func (s *PluginSchemaService) CreateLuaPluginSchema(ctx context.Context,
	req *v1.CreateLuaPluginSchemaRequest,
) (*v1.CreateLuaPluginSchemaResponse, error) {
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	res := resource.NewPluginSchema()
	res.PluginSchema = req.Item
	if err := db.Create(ctx, res); err != nil {
		errConstraint, ok := err.(store.ErrConstraint)
		if ok {
			errConstraint.Index.FieldName = nameFieldName
			errConstraint.Index.Name = nameFieldName
			err = errConstraint
		}
		return nil, s.err(ctx, err)
	}
	util.SetHeader(ctx, http.StatusCreated)
	return &v1.CreateLuaPluginSchemaResponse{
		Item: res.PluginSchema,
	}, nil
}

func (s *PluginSchemaService) GetLuaPluginSchema(ctx context.Context,
	req *v1.GetLuaPluginSchemaRequest,
) (*v1.GetLuaPluginSchemaResponse, error) {
	if err := checkSchemaName(req.Name); err != nil {
		return nil, s.err(ctx, err)
	}
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	res := resource.NewPluginSchema()
	if err := db.Read(ctx, res, store.GetByID(req.Name)); err != nil {
		return nil, s.err(ctx, err)
	}
	return &v1.GetLuaPluginSchemaResponse{
		Item: res.PluginSchema,
	}, nil
}

func (s *PluginSchemaService) ListLuaPluginSchemas(ctx context.Context,
	req *v1.ListLuaPluginSchemasRequest,
) (*v1.ListLuaPluginSchemasResponse, error) {
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}

	listFn := []store.ListOptsFunc{}
	list := resource.NewList(resource.TypePluginSchema)
	listOptFns, err := ListOptsFromReq(req.Page)
	if err != nil {
		return nil, s.err(ctx, err)
	}

	listFn = append(listFn, listOptFns...)

	if err := db.List(ctx, list, listFn...); err != nil {
		return nil, s.err(ctx, err)
	}

	return &v1.ListLuaPluginSchemasResponse{
		Items: pluginSchemasFromObjects(list.GetAll()),
		Page:  getPaginationResponse(list.GetTotalCount(), list.GetNextPage()),
	}, nil
}

func (s *PluginSchemaService) DeleteLuaPluginSchema(ctx context.Context,
	req *v1.DeleteLuaPluginSchemaRequest,
) (*v1.DeleteLuaPluginSchemaResponse, error) {
	if err := checkSchemaName(req.Name); err != nil {
		return nil, s.err(ctx, err)
	}
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}

	inUse, err := s.isPluginSchemaInUse(ctx, req.Name, db)
	if err != nil {
		return nil, err
	}
	if inUse {
		errMsg := "plugin schema is currently in use, " +
			"please delete existing plugins using the schema and try again"
		return nil, s.err(ctx, util.ErrClient{Message: errMsg})
	}

	err = db.Delete(ctx, store.DeleteByID(req.Name),
		store.DeleteByType(resource.TypePluginSchema))
	if err != nil {
		return nil, s.err(ctx, err)
	}
	util.SetHeader(ctx, http.StatusNoContent)
	return &v1.DeleteLuaPluginSchemaResponse{}, nil
}

func (s *PluginSchemaService) isPluginSchemaInUse(
	ctx context.Context, name string, db store.Store,
) (bool, error) {
	var page int32 = 1
	for {
		plugins, page, err := s.getPluginsPage(ctx, page, db)
		if err != nil {
			return false, err
		}
		for _, plugin := range plugins {
			if plugin.Name == name {
				return true, nil
			}
		}
		if page == 0 {
			break
		}
	}
	return false, nil
}

func (s *PluginSchemaService) getPluginsPage(
	ctx context.Context, page int32, db store.Store,
) ([]*pb.Plugin, int, error) {
	list := resource.NewList(resource.TypePlugin)
	listOptFns, err := ListOptsFromReq(&pb.PaginationRequest{
		Number: page,
		Size:   store.MaxPageSize,
	})
	if err != nil {
		return nil, 0, err
	}

	if err := db.List(ctx, list, listOptFns...); err != nil {
		return nil, 0, err
	}

	return pluginsFromObjects(list.GetAll()), list.GetNextPage(), nil
}

func (s *PluginSchemaService) err(ctx context.Context, err error) error {
	return util.HandleErr(ctx, s.logger(ctx), err)
}

func (s *PluginSchemaService) logger(ctx context.Context) *zap.Logger {
	return util.LoggerFromContext(ctx).With(s.loggerFields...)
}

func pluginSchemasFromObjects(objects []model.Object) []*pb.PluginSchema {
	res := make([]*pb.PluginSchema, len(objects))
	for i, object := range objects {
		var ok bool
		if res[i], ok = object.Resource().(*pb.PluginSchema); !ok {
			panic(fmt.Sprintf("expected type '%T' but got '%T'",
				&pb.PluginSchema{}, object.Resource()))
		}
	}
	return res
}

func (s *PluginSchemaService) UpsertLuaPluginSchema(ctx context.Context,
	req *v1.UpsertLuaPluginSchemaRequest,
) (*v1.UpsertLuaPluginSchemaResponse, error) {
	if err := checkSchemaName(req.Name); err != nil {
		return nil, s.err(ctx, err)
	}
	// TODO(hbagdi): validate the ne plugin schema again all existing plugin instances
	// in the database before allowing an update
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	res := resource.NewPluginSchema()
	res.PluginSchema = req.Item
	if err := db.Upsert(ctx, res); err != nil {
		errConstraint, ok := err.(store.ErrConstraint)
		if ok {
			errConstraint.Index.FieldName = nameFieldName
			errConstraint.Index.Name = nameFieldName
			err = errConstraint
		}
		return nil, s.err(ctx, err)
	}

	return &v1.UpsertLuaPluginSchemaResponse{
		Item: res.PluginSchema,
	}, nil
}

func checkSchemaName(name string) error {
	if name == "" {
		return validation.Error{
			Errs: []*pb.ErrorDetail{
				{
					Type:     pb.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{"required name is missing"},
				},
			},
		}
	}
	if !nameRegex.MatchString(name) {
		return validation.Error{
			Errs: []*pb.ErrorDetail{
				{
					Type:     pb.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{fmt.Sprintf("must match pattern: '%s'", nameRegex.String())},
				},
			},
		}
	}
	return nil
}
