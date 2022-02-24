package admin

import (
	"context"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
	"github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/server"
	"github.com/kong/koko/internal/server/util"
	"github.com/kong/koko/internal/store"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// HandlerWrapper is used to wrap a http.Handler with another http.Handler.
type HandlerWrapper interface {
	Wrap(http.Handler) http.Handler
	server.GrpcInterceptorInjector
}

// ContextKey type must be used to manipulate the context of a request.
type ContextKey struct{}

type HandlerOpts struct {
	Logger *zap.Logger

	StoreLoader util.StoreLoader

	GetRawLuaSchema func(name string) ([]byte, error)
}

type CommonOpts struct {
	logger *zap.Logger

	storeLoader util.StoreLoader
}

func (c CommonOpts) getDB(ctx context.Context,
	cluster *model.RequestCluster) (store.Store, error) {
	store, err := c.storeLoader.Load(ctx, cluster)
	if err != nil {
		if storeLoadErr, ok := err.(util.StoreLoadErr); ok {
			return nil, status.Error(storeLoadErr.Code, storeLoadErr.Message)
		}
		return nil, err
	}
	return store, nil
}

type services struct {
	service  v1.ServiceServiceServer
	route    v1.RouteServiceServer
	plugin   v1.PluginServiceServer
	upstream v1.UpstreamServiceServer
	target   v1.TargetServiceServer
	schemas  v1.SchemasServiceServer
	consumer v1.ConsumerServiceServer

	status v1.StatusServiceServer
	node   v1.NodeServiceServer
}

func buildServices(opts HandlerOpts) services {
	return services{
		service: &ServiceService{
			CommonOpts: CommonOpts{
				storeLoader: opts.StoreLoader,
				logger: opts.Logger.With(zap.String("admin-service",
					"service")),
			},
		},
		route: &RouteService{
			CommonOpts: CommonOpts{
				storeLoader: opts.StoreLoader,
				logger: opts.Logger.With(zap.String("admin-service",
					"route")),
			},
		},
		plugin: &PluginService{
			CommonOpts: CommonOpts{
				storeLoader: opts.StoreLoader,
				logger: opts.Logger.With(zap.String("admin-service",
					"plugin")),
			},
		},
		upstream: &UpstreamService{
			CommonOpts: CommonOpts{
				storeLoader: opts.StoreLoader,
				logger: opts.Logger.With(zap.String("admin-service",
					"upstream")),
			},
		},
		target: &TargetService{
			CommonOpts: CommonOpts{
				storeLoader: opts.StoreLoader,
				logger: opts.Logger.With(zap.String("admin-service",
					"target")),
			},
		},
		schemas: &SchemasService{
			logger:          opts.Logger.With(zap.String("admin-service", "schemas")),
			getRawLuaSchema: opts.GetRawLuaSchema,
		},
		node: &NodeService{
			CommonOpts: CommonOpts{
				storeLoader: opts.StoreLoader,
				logger: opts.Logger.With(zap.String("admin-service",
					"node")),
			},
		},
		status: &StatusService{
			CommonOpts: CommonOpts{
				storeLoader: opts.StoreLoader,
				logger: opts.Logger.With(zap.String("admin-service",
					"node")),
			},
		},
		consumer: &ConsumerService{
			CommonOpts: CommonOpts{
				storeLoader: opts.StoreLoader,
				logger:      opts.Logger.With(zap.String("admin-service", "consumer")),
			},
		},
	}
}

func NewHandler(opts HandlerOpts) (http.Handler, error) {
	err := validateOpts(opts)
	if err != nil {
		return nil, err
	}
	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, json.Marshaller),
		runtime.WithForwardResponseOption(util.SetHTTPStatus),
	)

	err = v1.RegisterMetaServiceHandlerServer(context.Background(),
		mux, &MetaService{})
	if err != nil {
		return nil, err
	}

	services := buildServices(opts)
	err = v1.RegisterServiceServiceHandlerServer(context.Background(),
		mux, services.service)
	if err != nil {
		return nil, err
	}

	err = v1.RegisterRouteServiceHandlerServer(context.Background(),
		mux, services.route)
	if err != nil {
		return nil, err
	}

	err = v1.RegisterPluginServiceHandlerServer(context.Background(),
		mux, services.plugin)
	if err != nil {
		return nil, err
	}

	err = v1.RegisterUpstreamServiceHandlerServer(context.Background(),
		mux, services.upstream)
	if err != nil {
		return nil, err
	}

	err = v1.RegisterTargetServiceHandlerServer(context.Background(),
		mux, services.target)
	if err != nil {
		return nil, err
	}

	err = v1.RegisterSchemasServiceHandlerServer(context.Background(),
		mux, services.schemas)
	if err != nil {
		return nil, err
	}

	err = v1.RegisterNodeServiceHandlerServer(context.Background(),
		mux, services.node)
	if err != nil {
		return nil, err
	}
	err = v1.RegisterStatusServiceHandlerServer(context.Background(),
		mux, services.status)
	if err != nil {
		return nil, err
	}
	err = v1.RegisterConsumerServiceHandlerServer(context.Background(),
		mux, services.consumer)
	if err != nil {
		return nil, err
	}
	return mux, nil
}

func validateOpts(opts HandlerOpts) error {
	if opts.StoreLoader == nil {
		return fmt.Errorf("opts.StoreLoader is required")
	}
	if opts.Logger == nil {
		return fmt.Errorf("opts.Logger is required")
	}
	return nil
}

func NewGRPC(opts HandlerOpts) *grpc.Server {
	server := grpc.NewServer()
	services := buildServices(opts)
	v1.RegisterMetaServiceServer(server, &MetaService{})
	v1.RegisterServiceServiceServer(server, services.service)
	v1.RegisterRouteServiceServer(server, services.route)
	v1.RegisterPluginServiceServer(server, services.plugin)
	v1.RegisterUpstreamServiceServer(server, services.upstream)
	v1.RegisterTargetServiceServer(server, services.target)
	v1.RegisterSchemasServiceServer(server, services.schemas)
	v1.RegisterNodeServiceServer(server, services.node)
	v1.RegisterStatusServiceServer(server, services.status)

	v1.RegisterConsumerServiceServer(server, services.consumer)
	return server
}
