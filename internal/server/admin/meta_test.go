package admin

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect/v2"
)

func TestMeta(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.Default(t, s.URL)
	c.GET("/v1/meta/version").Expect().Status(http.StatusOK).JSON().
		Object().Value("version").Equal("dev")
}
