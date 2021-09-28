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
	handler, err := admin.NewHandler(admin.HandlerOpts{
		Logger: log.Logger,
		Store:  store.New(&persistence.Memory{}),
	})
	if err != nil {
		t.Fatalf("creating httptest.Server: %v", err)
	}

	s := httptest.NewServer(handler)
	return s, func() {
		s.Close()
	}
}
