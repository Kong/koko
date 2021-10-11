package admin

import (
	"context"
	"net/http"

	"github.com/kong/koko/internal/store"
	"google.golang.org/grpc"
)

// StoreContextKey is used by the Admin handler to retrieve a store.Store
// from ctx of a request.
var StoreContextKey = &ContextKey{}

type DefaultStoreWrapper struct {
	Store store.Store
}

func (s DefaultStoreWrapper) Handle(ctx context.Context,
	req interface{},
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (resp interface{}, err error) {
	ctx = inject(ctx, s.Store)
	return handler(ctx, req)
}

func (s DefaultStoreWrapper) Wrap(handler http.Handler) http.Handler {
	return defaultStoreInjector{store: s.Store, next: handler}
}

type defaultStoreInjector struct {
	store store.Store
	next  http.Handler
}

func (s defaultStoreInjector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := inject(r.Context(), s.store)
	s.next.ServeHTTP(w, r.WithContext(ctx))
}

func inject(ctx context.Context, store store.Store) context.Context {
	return context.WithValue(ctx, StoreContextKey, store)
}
