package admin

import (
	"context"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
	"go.uber.org/zap"
)

type MetaService struct {
	v1.UnimplementedMetaServiceServer
	Logger *zap.Logger
}

func (m *MetaService) GetVersion(_ context.Context,
	_ *v1.GetVersionRequest,
) (*v1.GetVersionResponse, error) {
	return &v1.GetVersionResponse{
		Version: "dev",
	}, nil
}
