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

	StoreLoader StoreLoader
}

type CommonOpts struct {
	logger *zap.Logger

	storeLoader StoreLoader
}

func (c CommonOpts) getDB(ctx context.Context,
	cluster *model.RequestCluster) (store.Store, error) {
	store, err := c.storeLoader.Load(ctx, cluster)
	if err != nil {
		if storeLoadErr, ok := err.(StoreLoadErr); ok {
			return nil, status.Error(storeLoadErr.Code, storeLoadErr.Message)
		}
		return nil, err
	}
	return store, nil
}

type services struct {
	service v1.ServiceServiceServer
	route   v1.RouteServiceServer
	node    v1.NodeServiceServer
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
		node: &NodeService{
			CommonOpts: CommonOpts{
				storeLoader: opts.StoreLoader,
				logger: opts.Logger.With(zap.String("admin-service",
					"node")),
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

	err = v1.RegisterNodeServiceHandlerServer(context.Background(),
		mux, services.node)
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
	v1.RegisterNodeServiceServer(server, services.node)
	return server
}
