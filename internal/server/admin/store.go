package admin

import (
	"context"
	"net/http"

	"github.com/kong/koko/internal/store"
)

// StoreContextKey is used by the Admin handler to retrieve a store.Store
// from ctx of a request.
var StoreContextKey = &ContextKey{}

type DefaultStoreWrapper struct {
	Store store.Store
}

func (s DefaultStoreWrapper) Wrap(handler http.Handler) http.Handler {
	return defaultStoreInjector{store: s.Store, next: handler}
}

type defaultStoreInjector struct {
	store store.Store
	next  http.Handler
}

func (s defaultStoreInjector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(r.Context(), StoreContextKey, s.store)
	s.next.ServeHTTP(w, r.WithContext(ctx))
}
