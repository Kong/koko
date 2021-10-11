package admin

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
	"github.com/kong/koko/internal/server"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

// HandlerWrapper is used to wrap a http.Handler with another http.Handler.
type HandlerWrapper interface {
	Wrap(http.Handler) http.Handler
	server.GrpcInterceptorInjector
}

// ContextKey type must be used to manipulate the context of a request.
type ContextKey struct{}

type HandlerOpts struct {
	Logger        *zap.Logger
	StoreInjector HandlerWrapper
}

type CommonOpts struct {
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
				logger: opts.Logger.With(zap.String("admin-service",
					"route")),
			},
		})
	if err != nil {
		return nil, err
	}

	return opts.StoreInjector.Wrap(mux), nil
}

func NewGRPC(opts HandlerOpts) *grpc.Server {
	server := grpc.NewServer(
		grpc.UnaryInterceptor(opts.StoreInjector.Handle),
	)
	v1.RegisterMetaServiceServer(server, &MetaService{})
	v1.RegisterServiceServiceServer(server, &ServiceService{
		CommonOpts: CommonOpts{
			logger: opts.Logger.With(zap.String("admin-service",
				"service")),
		},
	})
	v1.RegisterRouteServiceServer(server, &RouteService{
		CommonOpts: CommonOpts{
			logger: opts.Logger.With(zap.String("admin-service",
				"route")),
		},
	})
	return server
}
