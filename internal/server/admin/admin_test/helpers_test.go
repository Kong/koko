package admin_test

import (
	"net/http/httptest"
	"testing"

	"github.com/kong/koko/internal/log"
	"github.com/kong/koko/internal/server/admin"
	"github.com/kong/koko/internal/store"
	"github.com/kong/koko/internal/test/util"
	"github.com/stretchr/testify/require"
)

func setup(t *testing.T) (*httptest.Server, func()) {
	p, err := util.GetPersister()
	require.Nil(t, err)
	objectStore := store.New(p, log.Logger)

	handler, err := admin.NewHandler(admin.HandlerOpts{
		Logger: log.Logger,
		StoreLoader: admin.DefaultStoreLoader{
			Store: objectStore.ForCluster("default"),
		},
	})
	if err != nil {
		t.Fatalf("creating httptest.Server: %v", err)
	}

	s := httptest.NewServer(handler)
	return s, func() {
		s.Close()
	}
}
