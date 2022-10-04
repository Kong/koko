package admin

import (
	"context"
	"errors"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
	"github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/model/json/extension"
	"github.com/kong/koko/internal/model/json/schema"
	"github.com/kong/koko/internal/plugin"
	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/server/util"
	"github.com/kong/koko/internal/store"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type SchemasService struct {
	v1.UnimplementedSchemasServiceServer
	validator plugin.Validator

	CommonOpts
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
	err = json.ProtoJSONUnmarshal(rawJSONSchema, jsonSchema)
	if err != nil {
		return nil, s.err(ctx, err)
	}

	// Remove our custom JSON schema extension used for internal-only reasons.
	if f := jsonSchema.Fields; f != nil {
		delete(f, (&extension.Config{}).Name())
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
	s.logger(ctx).With(zap.String("name", req.Name)).Debug("reading Lua plugin schema by name")
	ctx = context.WithValue(ctx, util.ContextKeyCluster, req.Cluster)

	// Retrieve the raw JSON based on plugin name
	rawLuaSchema, err := s.validator.GetRawLuaSchema(ctx, req.Name)
	if err != nil {
		// if it's not found, return custom error
		if errors.Is(err, store.ErrNotFound) || errors.Is(err, plugin.ErrSchemaNotFound) {
			return nil, status.Errorf(codes.NotFound, "no plugin-schema for '%s'", req.Name)
		}
		return nil, s.err(ctx, err)
	}

	// Convert the raw Lua (JSON) into a map/struct and return response
	luaSchema := &structpb.Struct{}
	err = json.ProtoJSONUnmarshal(rawLuaSchema, luaSchema)
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
	if err := res.ProcessDefaults(ctx); err != nil {
		return nil, s.err(ctx, err)
	} else if err := res.Validate(ctx); err != nil {
		return nil, s.err(ctx, err)
	}
	return &v1.ValidateLuaPluginResponse{}, nil
}

func (s *SchemasService) ValidateCACertificateSchema(
	ctx context.Context,
	req *v1.ValidateCACertificateSchemaRequest,
) (*v1.ValidateCACertificateSchemaResponse, error) {
	return &v1.ValidateCACertificateSchemaResponse{}, s.validateSchema(ctx, req.Item)
}

func (s *SchemasService) ValidateCertificateSchema(
	ctx context.Context,
	req *v1.ValidateCertificateSchemaRequest,
) (*v1.ValidateCertificateSchemaResponse, error) {
	return &v1.ValidateCertificateSchemaResponse{}, s.validateSchema(ctx, req.Item)
}

func (s *SchemasService) ValidateConsumerSchema(
	ctx context.Context,
	req *v1.ValidateConsumerSchemaRequest,
) (*v1.ValidateConsumerSchemaResponse, error) {
	return &v1.ValidateConsumerSchemaResponse{}, s.validateSchema(ctx, req.Item)
}

func (s *SchemasService) ValidatePluginSchema(
	ctx context.Context,
	req *v1.ValidatePluginSchemaRequest,
) (*v1.ValidatePluginSchemaResponse, error) {
	return &v1.ValidatePluginSchemaResponse{}, s.validateSchema(ctx, req.Item)
}

func (s *SchemasService) ValidateRouteSchema(
	ctx context.Context,
	req *v1.ValidateRouteSchemaRequest,
) (*v1.ValidateRouteSchemaResponse, error) {
	return &v1.ValidateRouteSchemaResponse{}, s.validateSchema(ctx, req.Item)
}

func (s *SchemasService) ValidateServiceSchema(
	ctx context.Context,
	req *v1.ValidateServiceSchemaRequest,
) (*v1.ValidateServiceSchemaResponse, error) {
	return &v1.ValidateServiceSchemaResponse{}, s.validateSchema(ctx, req.Item)
}

func (s *SchemasService) ValidateSNISchema(
	ctx context.Context,
	req *v1.ValidateSNISchemaRequest,
) (*v1.ValidateSNISchemaResponse, error) {
	return &v1.ValidateSNISchemaResponse{}, s.validateSchema(ctx, req.Item)
}

func (s *SchemasService) ValidateVaultSchema(
	ctx context.Context,
	req *v1.ValidateVaultSchemaRequest,
) (*v1.ValidateVaultSchemaResponse, error) {
	return &v1.ValidateVaultSchemaResponse{}, s.validateSchema(ctx, req.Item)
}

func (s *SchemasService) ValidateTargetSchema(
	ctx context.Context,
	req *v1.ValidateTargetSchemaRequest,
) (*v1.ValidateTargetSchemaResponse, error) {
	return &v1.ValidateTargetSchemaResponse{}, s.validateSchema(ctx, req.Item)
}

func (s *SchemasService) ValidateUpstreamSchema(
	ctx context.Context,
	req *v1.ValidateUpstreamSchemaRequest,
) (*v1.ValidateUpstreamSchemaResponse, error) {
	return &v1.ValidateUpstreamSchemaResponse{}, s.validateSchema(ctx, req.Item)
}

// validateSchema handles the relevant entity's JSONSchema validation,
// along with any specific validation associated to the entity.
func (s *SchemasService) validateSchema(ctx context.Context, item proto.Message) error {
	obj, err := model.ObjectFromProto(item)
	if err != nil {
		// As long as every model object has had its type & underlining Protobuf definition
		// registered (via `model.RegisterType()`), this would never error.
		return err
	}

	return s.err(ctx, obj.Validate(ctx))
}

func (s *SchemasService) err(ctx context.Context, err error) error {
	return util.HandleErr(ctx, s.logger(ctx), err)
}

func (s *SchemasService) logger(ctx context.Context) *zap.Logger {
	return util.LoggerFromContext(ctx).With(s.loggerFields...)
}
