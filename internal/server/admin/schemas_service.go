package admin

import (
	"context"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
	"github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/model/json/schema"
	"github.com/kong/koko/internal/plugin"
	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/server/util"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

type SchemasService struct {
	v1.UnimplementedSchemasServiceServer
	loggerFields []zapcore.Field
	validator    plugin.Validator
}

func (s *SchemasService) GetSchemas(ctx context.Context,
	req *v1.GetSchemasRequest,
) (*v1.GetSchemasResponse, error) {
	if req.Name == "" {
		return nil, s.err(ctx, util.ErrClient{Message: "required name is missing"})
	}

	// Retrieve the raw JSON based on entity name
	s.logger(ctx).With(zap.String("name", req.Name)).Debug("reading schemas by name")
	rawJSONSchema, err := schema.GetRawJSONSchema(req.Name)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "no entity named '%s'", req.Name)
	}

	// Convert the raw JSON into a map/struct and return response
	jsonSchema := &structpb.Struct{}
	err = json.Unmarshal(rawJSONSchema, jsonSchema)
	if err != nil {
		return nil, s.err(ctx, err)
	}
	return &v1.GetSchemasResponse{
		Schema: jsonSchema,
	}, nil
}

func (s *SchemasService) GetLuaSchemasPlugin(ctx context.Context,
	req *v1.GetLuaSchemasPluginRequest,
) (*v1.GetLuaSchemasPluginResponse, error) {
	if req.Name == "" {
		return nil, s.err(ctx, util.ErrClient{Message: "required name is missing"})
	}

	// Retrieve the raw JSON based on plugin name
	s.logger(ctx).With(zap.String("name", req.Name)).Debug("reading Lua plugin schema by name")
	rawLuaSchema, err := s.validator.GetRawLuaSchema(req.Name)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "no plugin named '%s'", req.Name)
	}

	// Convert the raw Lua (JSON) into a map/struct and return response
	luaSchema := &structpb.Struct{}
	err = json.Unmarshal(rawLuaSchema, luaSchema)
	if err != nil {
		return nil, s.err(ctx, err)
	}
	return &v1.GetLuaSchemasPluginResponse{
		Schema: luaSchema,
	}, nil
}

func (s *SchemasService) ValidateLuaPlugin(
	ctx context.Context,
	req *v1.ValidateLuaPluginRequest,
) (*v1.ValidateLuaPluginResponse, error) {
	res := resource.NewPlugin()
	res.Plugin = req.Item
	if err := res.ProcessDefaults(); err != nil {
		return nil, s.err(ctx, err)
	} else if err := res.Validate(); err != nil {
		return nil, s.err(ctx, err)
	}
	return &v1.ValidateLuaPluginResponse{}, nil
}

func (s *SchemasService) err(ctx context.Context, err error) error {
	return util.HandleErr(ctx, s.logger(ctx), err)
}

func (s *SchemasService) logger(ctx context.Context) *zap.Logger {
	return util.LoggerFromContext(ctx).With(s.loggerFields...)
}
