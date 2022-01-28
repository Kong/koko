package admin

import (
	"context"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
	"github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/model/json/schema"
	"github.com/kong/koko/internal/server/util"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

type SchemasService struct {
	v1.UnimplementedSchemasServiceServer
	logger *zap.Logger
}

func (s *SchemasService) GetSchemas(ctx context.Context,
	req *v1.GetSchemasRequest) (*v1.GetSchemasResponse, error) {
	if req.Name == "" {
		return nil, s.err(util.ErrClient{Message: "required name is missing"})
	}

	// Retrieve the raw JSON based on entity name
	s.logger.With(zap.String("name", req.Name)).Debug("reading schemas by name")
	rawJSONSchema, err := schema.GetEntityRawJSON(req.Name)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "no entity named '%s'", req.Name)
	}

	// Convert the raw JSON into a map/struct and return response
	jsonSchema := &structpb.Struct{}
	err = json.Unmarshal(rawJSONSchema, jsonSchema)
	if err != nil {
		return nil, s.err(err)
	}
	return &v1.GetSchemasResponse{
		Schema: jsonSchema,
	}, nil
}

func (s *SchemasService) GetSchemasPlugin(ctx context.Context,
	req *v1.GetSchemasPluginRequest) (*v1.GetSchemasPluginResponse, error) {
	if req.Name == "" {
		return nil, s.err(util.ErrClient{Message: "required name is missing"})
	}

	// Retrieve the raw JSON based on plugin name
	s.logger.With(zap.String("name", req.Name)).Debug("reading schemas by name")
	rawJSONSchema, err := schema.GetPluginRawJSON(req.Name)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "no plugin named '%s'", req.Name)
	}

	// Convert the raw JSON into a map/struct and return response
	jsonSchema := &structpb.Struct{}
	err = json.Unmarshal(rawJSONSchema, jsonSchema)
	if err != nil {
		return nil, s.err(err)
	}
	return &v1.GetSchemasPluginResponse{
		Schema: jsonSchema,
	}, nil
}

func (s *SchemasService) err(err error) error {
	return util.HandleErr(s.logger, err)
}
