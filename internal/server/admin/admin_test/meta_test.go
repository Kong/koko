package admin_test

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect/v2"
)

func TestMeta(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	c.GET("/v1/meta/version").Expect().Status(http.StatusOK).JSON().
		Object().Value("version").Equal("dev")
}
