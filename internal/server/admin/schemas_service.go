package admin

import (
	"context"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
	"github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/model/json/schema"
	"github.com/kong/koko/internal/server/util"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

type SchemasService struct {
	v1.UnimplementedSchemasServiceServer
	CommonOpts
}

func (s *SchemasService) GetSchemas(ctx context.Context,
	req *v1.GetSchemasRequest) (*v1.GetSchemasResponse, error) {
	if req.Name == "" {
		return nil, s.err(util.ErrClient{Message: "required name is missing"})
	}
	s.logger.With(zap.String("name", req.Name)).Debug("reading schemas by name")
	rawJSONSchema, err := schema.GetRawJSONSchema(req.Name)
	if err != nil {
		return nil, s.err(err)
	}
	if err != nil {
		return nil, s.err(err)
	}
	jsonSchema := &structpb.Struct{}
	err = json.Unmarshal(rawJSONSchema, jsonSchema)
	if err != nil {
		return nil, s.err(err)
	}
	return &v1.GetSchemasResponse{
		Schema: jsonSchema,
	}, nil
}

func (s *SchemasService) err(err error) error {
	return util.HandleErr(s.logger, err)
}
