package admin

import (
	"context"
	"net/http"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
	"github.com/kong/koko/internal/plugin"
	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/server/util"
	"go.uber.org/zap"
)

type PluginSchemaService struct {
	v1.UnimplementedPluginSchemaServiceServer
	CommonOpts
	validator plugin.Validator
}

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
		return nil, s.err(ctx, err)
	}
	util.SetHeader(ctx, http.StatusCreated)
	return &v1.CreateLuaPluginSchemaResponse{
		Item: res.PluginSchema,
	}, nil
}

func (s *PluginSchemaService) err(ctx context.Context, err error) error {
	return util.HandleErr(ctx, s.logger(ctx), err)
}

func (s *PluginSchemaService) logger(ctx context.Context) *zap.Logger {
	return util.LoggerFromContext(ctx).With(s.loggerFields...)
}
