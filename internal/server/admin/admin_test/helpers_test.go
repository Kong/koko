package admin_test

import (
	"net/http/httptest"
	"testing"

	"github.com/kong/koko/internal/log"
	"github.com/kong/koko/internal/persistence"
	"github.com/kong/koko/internal/server/admin"
	"github.com/kong/koko/internal/store"
)

func setup(t *testing.T) (*httptest.Server, func()) {
	persister, err := persistence.NewMemory()
	if err != nil {
		t.Fatalf("create persister: %v", err)
	}
	objectStore := store.New(persister, log.Logger)

	handler, err := admin.NewHandler(admin.HandlerOpts{
		Logger: log.Logger,
		StoreInjector: admin.DefaultStoreWrapper{Store: objectStore.
			ForCluster("default")},
	})
	if err != nil {
		t.Fatalf("creating httptest.Server: %v", err)
	}

	s := httptest.NewServer(handler)
	return s, func() {
		s.Close()
	}
}
