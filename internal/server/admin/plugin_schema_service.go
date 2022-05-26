package admin

import (
	"context"
	"fmt"
	"net/http"

	pb "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
	"github.com/kong/koko/internal/model"
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
	if req.Name == "" {
		return nil, s.err(ctx, util.ErrClient{Message: "required name is missing"})
	}
	if !nameRegex.MatchString(req.Name) {
		return nil, s.err(ctx, util.ErrClient{Message: "required name is invalid"})
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
	listOptFns, err := listOptsFromReq(req.Page)
	if err != nil {
		return nil, s.err(ctx, util.ErrClient{Message: err.Error()})
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
	if req.Name == "" {
		return nil, s.err(ctx, util.ErrClient{Message: "required name is missing"})
	}
	if !nameRegex.MatchString(req.Name) {
		return nil, s.err(ctx, util.ErrClient{Message: "required name is invalid"})
	}
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
