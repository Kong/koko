package admin

import (
	"fmt"
	"testing"

	"github.com/gavv/httpexpect/v2"
)

func TestSchemasGetEntity(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)

	t.Run("get a valid entity", func(t *testing.T) {
		paths := []string{
			"node",
			"plugin",
			"route",
			"service",
			"status",
		}

		for _, path := range paths {
			res := c.GET(fmt.Sprintf("/v1/schemas/json/%s", path)).Expect()
			res.Status(200)
			value := res.JSON().Path("$.type").String()
			value.Equal("object") // all JSON schemas indicate type object
		}
	})

	t.Run("get 404 for invalid entity", func(t *testing.T) {
		paths := []string{
			"invalid",
			"not-available",
			",,,",
			"©¥§",
		}

		for _, path := range paths {
			res := c.GET(fmt.Sprintf("/v1/schemas/json/%s", path)).Expect()
			res.Status(404)
			message := res.JSON().Path("$.message").String()
			message.Equal(fmt.Sprintf("no entity named '%s'", path))
		}
	})

	t.Run("ensure the path/name must be present", func(t *testing.T) {
		res := c.GET("/v1/schemas/json/").Expect()
		res.Status(400)
		message := res.JSON().Path("$.message").String()
		message.Equal("required name is missing")
	})
}
