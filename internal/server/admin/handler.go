package admin

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

// HandlerWrapper is used to wrap a http.Handler with another http.Handler.
type HandlerWrapper interface {
	Wrap(http.Handler) http.Handler
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

type StoreInjector struct {
	store store.Store
	next  http.Handler
}

type contextKey struct{}

var storeCtxKey = &contextKey{}

func (s StoreInjector) Handler(next http.Handler) http.Handler {
	s.next = next
	return s
}

func (s StoreInjector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(r.Context(), storeCtxKey, s.store)
	s.next.ServeHTTP(w, r.WithContext(ctx))
>>>>>>> feat: dynamic store injection
}
