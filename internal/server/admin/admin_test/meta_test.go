package admin_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/kong/koko/internal/log"
	"github.com/kong/koko/internal/server/admin"
)

func TestMeta(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	c.GET("/v1/meta/version").Expect().Status(http.StatusOK).JSON().
		Object().Value("version").Equal("dev")
}

func setup(t *testing.T) (*httptest.Server, func()) {
	handler, err := admin.NewHandler(admin.HandlerOpts{
		Logger: log.Logger,
	})
	if err != nil {
		t.Fatalf("creating httptest.Server: %v", err)
	}

	s := httptest.NewServer(handler)
	return s, func() {
		s.Close()
	}
}
