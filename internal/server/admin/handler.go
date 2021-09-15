package admin

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
	"go.uber.org/zap"
)

type HandlerOpts struct {
	Logger *zap.Logger
}

func NewHandler(opts HandlerOpts) (http.Handler, error) {
	mux := runtime.NewServeMux()
	err := v1.RegisterMetaServiceHandlerServer(context.Background(),
		mux, &MetaService{})
	if err != nil {
		return nil, err
	}
	return mux, nil
}
