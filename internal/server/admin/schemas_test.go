package admin

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gavv/httpexpect/v2"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/json"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
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
			res.Status(http.StatusOK)
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
			res.Status(http.StatusNotFound)
			message := res.JSON().Path("$.message").String()
			message.Equal(fmt.Sprintf("no entity named '%s'", path))
		}
	})

	t.Run("ensure the path/name is present", func(t *testing.T) {
		res := c.GET("/v1/schemas/json/").Expect()
		res.Status(http.StatusBadRequest)
		message := res.JSON().Path("$.message").String()
		message.Equal("required name is missing")
	})
}

func TestSchemasGetPlugin(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)

	t.Run("get a valid plugin schema", func(t *testing.T) {
		paths := []string{
			"acl",
			"http-log",
			"jwt",
			"loggly",
			"rate-limiting",
		}

		for _, path := range paths {
			res := c.GET(fmt.Sprintf("/v1/schemas/lua/plugins/%s", path)).Expect()
			res.Status(http.StatusOK)
			value := res.JSON().Path("$..protocols").Array()
			value.NotEmpty()
			value = res.JSON().Path("$..config.required").Array()
			value.Length().Equal(1)
			value.ContainsOnly(true) // all config objects are required for plugins
		}
	})

	t.Run("get 404 for invalid plugin name", func(t *testing.T) {
		paths := []string{
			"invalid-plugin",
			"not-available",
			"---",
			"ÅÊÏÕÜÝ",
		}

		for _, path := range paths {
			res := c.GET(fmt.Sprintf("/v1/schemas/lua/plugins/%s", path)).Expect()
			res.Status(http.StatusNotFound)
			message := res.JSON().Path("$.message").String()
			message.Equal(fmt.Sprintf("no plugin named '%s'", path))
		}
	})

	t.Run("ensure the path/name is present", func(t *testing.T) {
		res := c.GET("/v1/schemas/lua/plugins/").Expect()
		res.Status(http.StatusBadRequest)
		message := res.JSON().Path("$.message").String()
		message.Equal("required name is missing")
	})
}

func TestPluginValidate(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()

	invalidConfig := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"foo": structpb.NewStringValue("bar"),
		},
	}

	c, path := httpexpect.New(t, s.URL), "/v1/schemas/lua/plugins/validate"

	t.Run("validate a global plugin with valid config", func(t *testing.T) {
		pluginBytes, err := json.Marshal(goodKeyAuthPlugin())
		require.NoError(t, err)
		res := c.POST(path).WithBytes(pluginBytes).Expect()
		res.Status(http.StatusOK)
	})

	t.Run("validate a global plugin with improper type", func(t *testing.T) {
		plugin := goodKeyAuthPlugin()
		plugin.Config = &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"key_names": structpb.NewStringValue("apikey"),
			},
		}
		pluginBytes, err := json.Marshal(plugin)
		require.NoError(t, err)
		res := c.POST(path).WithBytes(pluginBytes).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		errRes := body.Value("details").Array().Element(0)
		errRes.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_FIELD.String())
		errRes.Object().ValueEqual("field", "config.key_names")
		errRes.Object().ValueEqual("messages", []string{"expected an array"})
	})

	t.Run("validate a global plugin with invalid config", func(t *testing.T) {
		plugin := goodKeyAuthPlugin()
		plugin.Config = invalidConfig
		pluginBytes, err := json.Marshal(plugin)
		require.NoError(t, err)
		res := c.POST(path).WithBytes(pluginBytes).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		errRes := body.Value("details").Array().Element(0)
		errRes.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_FIELD.String())
		errRes.Object().ValueEqual("field", "config.foo")
		errRes.Object().ValueEqual("messages", []string{"unknown field"})
	})

	t.Run("validate an unknown plugin with valid config", func(t *testing.T) {
		pluginBytes, err := json.Marshal(&v1.Plugin{Name: "no-auth"})
		require.NoError(t, err)
		res := c.POST(path).WithBytes(pluginBytes).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		errRes := body.Value("details").Array().Element(0)
		errRes.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_FIELD.String())
		errRes.Object().ValueEqual("field", "name")
		errRes.Object().ValueEqual("messages", []string{"plugin(no-auth) does not exist"})
	})

	t.Run("validate an unknown plugin with invalid config", func(t *testing.T) {
		pluginBytes, err := json.Marshal(&v1.Plugin{Name: "no-auth", Config: invalidConfig})
		require.NoError(t, err)
		res := c.POST(path).WithBytes(pluginBytes).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		errRes := body.Value("details").Array().Element(0)
		errRes.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_FIELD.String())
		errRes.Object().ValueEqual("field", "name")
		errRes.Object().ValueEqual("messages", []string{"plugin(no-auth) does not exist"})
	})

	t.Run("validate a plugin with empty request", func(t *testing.T) {
		res := c.POST(path).WithBytes([]byte("{}")).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		errRes := body.Value("details").Array().Element(0)
		errRes.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_ENTITY.String())
		errRes.Object().ValueEqual("messages", []string{"missing properties: 'name'"})
	})
}
