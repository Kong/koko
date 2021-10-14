package admin

import (
	"context"
	"fmt"

	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/store"
	"google.golang.org/grpc/codes"
)

type StoreLoadErr struct {
	Code    codes.Code
	Message string
}

func (s StoreLoadErr) Error() string {
	return fmt.Sprintf("%s (grpc-code: %d)", s.Message, s.Code)
}

type StoreLoader interface {
	// Load returns the store to use for the request and cluster.
	// Cluster is derived from request and maybe nil.
	// Ctx is specific to request. It may be expanded in future to include
	// HTTP metadata as needed.
	// If err is of type StoreLoadErr,
	// corresponding GRPC status code and message are returned to the client.
	// For any other error, an internal error is returned to the client
	Load(ctx context.Context, cluster *model.RequestCluster) (store.Store, error)
}

type DefaultStoreLoader struct {
	Store store.Store
}

func (d DefaultStoreLoader) Load(_ context.Context,
	_ *model.RequestCluster) (store.Store, error) {
	return d.Store, nil
}
