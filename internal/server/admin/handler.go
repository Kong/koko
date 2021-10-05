package admin

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
	"github.com/kong/koko/internal/store"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

type HandlerOpts struct {
	Logger *zap.Logger
	Store  store.Store
}

type CommonOpts struct {
	store  store.Store
	logger *zap.Logger
}

func NewHandler(opts HandlerOpts) (http.Handler, error) {
	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.
			MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames: true,
			},
		}),
		runtime.WithForwardResponseOption(setHTTPStatus),
	)

	err := v1.RegisterMetaServiceHandlerServer(context.Background(),
		mux, &MetaService{})
	if err != nil {
		return nil, err
	}

	err = v1.RegisterServiceServiceHandlerServer(context.Background(),
		mux, &ServiceService{
			CommonOpts: CommonOpts{
				store: opts.Store,
				logger: opts.Logger.With(zap.String("admin-service",
					"service")),
			},
		})
	if err != nil {
		return nil, err
	}

	err = v1.RegisterRouteServiceHandlerServer(context.Background(),
		mux, &RouteService{
			CommonOpts: CommonOpts{
				store: opts.Store,
				logger: opts.Logger.With(zap.String("admin-service",
					"route")),
			},
		})
	if err != nil {
		return nil, err
	}

	return mux, nil
}
