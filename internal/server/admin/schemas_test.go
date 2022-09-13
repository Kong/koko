package admin

import (
	"fmt"
	"net/http"
	"path"
	"strings"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/google/uuid"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/model/json/extension"
	"github.com/kong/koko/internal/model/json/schema"
	"github.com/kong/koko/internal/model/json/validation/typedefs"
	"github.com/kong/koko/internal/resource"
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
		}

		for _, p := range paths {
			res := c.GET(fmt.Sprintf("/v1/schemas/json/%s", p)).Expect()
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

		for _, p := range paths {
			res := c.GET(fmt.Sprintf("/v1/schemas/json/%s", p)).Expect()
			res.Status(http.StatusNotFound)
			message := res.JSON().Path("$.message").String()
			message.Equal(fmt.Sprintf("no entity named '%s'", p))
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

		for _, p := range paths {
			res := c.GET(fmt.Sprintf("/v1/schemas/lua/plugins/%s", p)).Expect()
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

		for _, p := range paths {
			res := c.GET(fmt.Sprintf("/v1/schemas/lua/plugins/%s", p)).Expect()
			res.Status(http.StatusNotFound)
			message := res.JSON().Path("$.message").String()
			message.Equal(fmt.Sprintf("no plugin-schema for '%s'", p))
		}
	})

	t.Run("ensure the path/name is present", func(t *testing.T) {
		res := c.GET("/v1/schemas/lua/plugins/").Expect()
		res.Status(http.StatusBadRequest)
		message := res.JSON().Path("$.message").String()
		message.Equal("required name is missing")
	})

	t.Run("get non-bundled plugins schema", func(t *testing.T) {
		nonBundledPlugin := "valid"
		res := c.POST("/v1/plugin-schemas").WithJSON(
			goodPluginSchema(nonBundledPlugin, "string"),
		).Expect()
		res.Status(http.StatusCreated)

		res = c.GET(fmt.Sprintf("/v1/schemas/lua/plugins/%s", nonBundledPlugin)).Expect()
		res.Status(http.StatusOK)
		value := res.JSON().Path("$..config.fields").Array()
		value.Length().Equal(1)
		value = res.JSON().Path("$..config.required").Array()
		value.Length().Equal(1)
		value.ContainsOnly(true)
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

	c, p := httpexpect.New(t, s.URL), "/v1/schemas/lua/plugins/validate"

	t.Run("validate a global plugin with valid config", func(t *testing.T) {
		pluginBytes, err := json.ProtoJSONMarshal(goodKeyAuthPlugin())
		require.NoError(t, err)
		res := c.POST(p).WithBytes(pluginBytes).Expect()
		res.Status(http.StatusOK)
	})

	t.Run("validate a global plugin with improper type", func(t *testing.T) {
		plugin := goodKeyAuthPlugin()
		plugin.Config = &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"key_names": structpb.NewStringValue("apikey"),
			},
		}
		pluginBytes, err := json.ProtoJSONMarshal(plugin)
		require.NoError(t, err)
		res := c.POST(p).WithBytes(pluginBytes).Expect()
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
		pluginBytes, err := json.ProtoJSONMarshal(plugin)
		require.NoError(t, err)
		res := c.POST(p).WithBytes(pluginBytes).Expect()
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
		pluginBytes, err := json.ProtoJSONMarshal(&v1.Plugin{Name: "no-auth"})
		require.NoError(t, err)
		res := c.POST(p).WithBytes(pluginBytes).Expect()
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
		pluginBytes, err := json.ProtoJSONMarshal(&v1.Plugin{Name: "no-auth", Config: invalidConfig})
		require.NoError(t, err)
		res := c.POST(p).WithBytes(pluginBytes).Expect()
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
		res := c.POST(p).WithBytes([]byte("{}")).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		errRes := body.Value("details").Array().Element(0)
		errRes.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_ENTITY.String())
		errRes.Object().ValueEqual("messages", []string{"missing properties: 'name'"})
	})
}

func TestValidateJSONSchema(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)

	// Ensures all registered types have a JSON schema validation endpoint.
	t.Run("check all types have routes", func(t *testing.T) {
		for _, typ := range model.AllTypes() {
			t.Run(string(typ), func(t *testing.T) {
				// Lookup JSON schema to ensure we're supposed to be validating this endpoint.
				sch, err := schema.Get(string(typ))
				require.NoError(t, err)
				skipType, configExtName := false, (&extension.Config{}).Name()
				if e := sch.Extensions; e != nil && e[configExtName] != nil {
					if e[configExtName].(*extension.Config).DisableValidateEndpoint {
						skipType = true
					}
				}
				if skipType {
					t.Skipf("Validating endpoint for type %s skipped, as it is disabled per the JSON schema.", typ)
					return
				}

				p := path.Join("/v1/schemas/json/", typeToRoute(typ), "validate")
				res := c.POST(p).WithBytes([]byte("{}")).Expect()
				// If the route was not registered for the type, an HTTP 404 would be returned. If the route
				// was registered, however the gRPC service wasn't implemented, a 501 would be returned.
				// Otherwise, we're expecting a validation error due to not providing any fields.
				res.Status(http.StatusBadRequest)
			})
		}
	})

	p := path.Join("/v1/schemas/json/", typeToRoute(resource.TypeConsumer), "validate")

	t.Run("invalid JSON schema", func(t *testing.T) {
		res := c.POST(p).WithJSON(&v1.Consumer{
			CreatedAt: -1,
			CustomId:  "invalid!",
			Tags:      []string{"some-tag", "some$tag"},
		}).Expect()
		res.Status(http.StatusBadRequest)

		body := res.JSON().Object()
		body.Value("message").String().Equal("validation error")
		body.Value("details").Array().Length().Equal(4)
		errRes := body.Value("details").Array()

		entityErr := errRes.Element(0).Object()
		entityErr.Value("type").String().Equal(v1.ErrorType_ERROR_TYPE_ENTITY.String())
		messages := entityErr.Value("messages").Array()
		messages.Length().Equal(1)
		messages.First().String().Equal("missing properties: 'id'")

		createdAtErr := errRes.Element(1).Object()
		createdAtErr.Value("type").String().Equal(v1.ErrorType_ERROR_TYPE_FIELD.String())
		createdAtErr.Value("field").String().Equal("created_at")
		messages = createdAtErr.Value("messages").Array()
		messages.Length().Equal(1)
		messages.First().String().Equal("must be >= 1 but found -1")

		customIDErr := errRes.Element(2).Object()
		customIDErr.Value("type").String().Equal(v1.ErrorType_ERROR_TYPE_FIELD.String())
		customIDErr.Value("field").String().Equal("custom_id")
		messages = customIDErr.Value("messages").Array()
		messages.Length().Equal(1)
		messages.First().String().Equal(
			`must match pattern '^[0-9a-zA-Z.\-_~\(\)#%@|+]+(?: [0-9a-zA-Z.\-_~\(\)#%@|+]+)*$'`,
		)

		tagsErr := errRes.Element(3).Object()
		tagsErr.Value("type").String().Equal(v1.ErrorType_ERROR_TYPE_FIELD.String())
		tagsErr.Value("field").String().Equal("tags[1]")
		messages = tagsErr.Value("messages").Array()
		messages.Length().Equal(1)
		messages.First().String().Equal("must match pattern '^(?:[0-9a-zA-Z.\\-_~:]+(?: *[0-9a-zA-Z.\\-_~:])*)?$'")
	})

	t.Run("duplicate tags returns an error", func(t *testing.T) {
		res := c.POST(p).WithJSON(&v1.Consumer{
			Id:       uuid.NewString(),
			Username: "testConsumer",
			Tags:     []string{"duplicate", "duplicate"},
		}).Expect()
		res.Status(http.StatusBadRequest)

		body := res.JSON().Object()
		body.Value("message").String().Equal("validation error")
		body.Value("details").Array().Length().Equal(1)
		errRes := body.Value("details").Array()

		tagsErr := errRes.First().Object()
		tagsErr.Value("type").String().Equal(v1.ErrorType_ERROR_TYPE_FIELD.String())
		tagsErr.Value("field").String().Equal("tags")
		messages := tagsErr.Value("messages").Array()
		messages.Length().Equal(1)
		messages.First().String().Equal("items at index 0 and 1 are equal")
	})

	t.Run("tag with space produces valid JSON schema for", func(t *testing.T) {
		t.Run("consumers", func(t *testing.T) {
			res := c.POST(p).WithJSON(&v1.Consumer{
				Id:       uuid.NewString(),
				Username: "spacesInTagsConsumers",
				Tags:     []string{"some tag", "with multiple spaces"},
			}).Expect()
			res.Status(http.StatusOK)
		})

		t.Run("services", func(t *testing.T) {
			p := path.Join("/v1/schemas/json/", typeToRoute(resource.TypeService), "validate")
			res := c.POST(p).WithJSON(&v1.Service{
				Id:             uuid.NewString(),
				Protocol:       typedefs.ProtocolHTTP,
				Host:           "example.com",
				Path:           "/",
				Port:           80,
				ConnectTimeout: 5000,
				ReadTimeout:    5000,
				WriteTimeout:   5000,
				Tags:           []string{"some tag", "with multiple spaces"},
			}).Expect()
			res.Status(http.StatusOK)
		})
	})

	t.Run("valid JSON schema", func(t *testing.T) {
		res := c.POST(p).WithJSON(&v1.Consumer{
			Id:       uuid.New().String(),
			CustomId: "custom-id",
		}).Expect()
		res.Status(http.StatusOK)
	})

	t.Run("plugin schema", func(t *testing.T) {
		p := path.Join("/v1/schemas/json/", typeToRoute(resource.TypePlugin), "validate")

		t.Run("valid schema", func(t *testing.T) {
			plugin := &v1.Plugin{
				Id:        uuid.NewString(),
				Name:      "acl",
				Enabled:   &wrappers.BoolValue{Value: false},
				Protocols: []string{typedefs.ProtocolHTTP, typedefs.ProtocolGRPC},
				Route:     &v1.Route{Id: uuid.NewString()},
				Service:   &v1.Service{Id: uuid.NewString()},
				Tags:      []string{"1", "2", "3", "4", "5", "6", "7", strings.Repeat("8", 128)},
			}
			var err error
			plugin.Config, err = structpb.NewStruct(map[string]interface{}{
				"deny":               []interface{}{"1.2.3.4", "10.0.0.0/24", "2a42:30c1:1:2:1337:c0de:4:11fe"},
				"hide_groups_header": true,
			})
			require.NoError(t, err)
			requestBody, _ := json.ProtoJSONMarshal(plugin)
			res := c.POST(p).WithBytes(requestBody).Expect()
			res.Status(http.StatusOK)
		})

		t.Run("invalid schema", func(t *testing.T) {
			plugin := &v1.Plugin{
				Id:        "invalid ID",
				Name:      "invalid plugin name",
				Protocols: []string{"foobar"},
				Route:     &v1.Route{Id: "invalid ID"},
				Service:   &v1.Service{Id: "invalid ID"},
				Tags:      []string{"1", "2", "3", "4", "5", "6", "7", "8", strings.Repeat("9", 129)},
			}
			var err error
			plugin.Config, err = structpb.NewStruct(map[string]interface{}{
				"deny":               []interface{}{"1.2.3.4", "10.0.0.0/24", "2a42:30c1:1:2:1337:c0de:4:11fe"},
				"hide_groups_header": true,
			})
			require.NoError(t, err)
			requestBody, _ := json.ProtoJSONMarshal(plugin)
			res := c.POST(p).WithBytes(requestBody).Expect()
			res.Status(http.StatusBadRequest)

			body := res.JSON().Object()
			body.Value("message").String().Equal("validation error")
			body.Value("details").Array().Length().Equal(7)
			errRes := body.Value("details").Array()

			errDetail := errRes.Element(0).Object()
			errDetail.Value("type").String().Equal(v1.ErrorType_ERROR_TYPE_FIELD.String())
			errDetail.Value("field").String().Equal("id")
			messages := errDetail.Value("messages").Array()
			messages.Length().Equal(1)
			messages.First().String().Equal("must be a valid UUID")

			errDetail = errRes.Element(1).Object()
			errDetail.Value("type").String().Equal(v1.ErrorType_ERROR_TYPE_FIELD.String())
			errDetail.Value("field").String().Equal("name")
			messages = errDetail.Value("messages").Array()
			messages.Length().Equal(1)
			messages.First().String().Equal("must match pattern '^[0-9a-zA-Z\\-]*$'")

			errDetail = errRes.Element(2).Object()
			errDetail.Value("type").String().Equal(v1.ErrorType_ERROR_TYPE_FIELD.String())
			errDetail.Value("field").String().Equal("protocols[0]")
			messages = errDetail.Value("messages").Array()
			messages.Length().Equal(1)
			messages.First().String().Contains("value must be one of")

			errDetail = errRes.Element(3).Object()
			errDetail.Value("type").String().Equal(v1.ErrorType_ERROR_TYPE_FIELD.String())
			errDetail.Value("field").String().Equal("route.id")
			messages = errDetail.Value("messages").Array()
			messages.Length().Equal(1)
			messages.First().String().Contains("must be a valid UUID")

			errDetail = errRes.Element(4).Object()
			errDetail.Value("type").String().Equal(v1.ErrorType_ERROR_TYPE_FIELD.String())
			errDetail.Value("field").String().Equal("service.id")
			messages = errDetail.Value("messages").Array()
			messages.Length().Equal(1)
			messages.First().String().Contains("must be a valid UUID")

			errDetail = errRes.Element(5).Object()
			errDetail.Value("type").String().Equal(v1.ErrorType_ERROR_TYPE_FIELD.String())
			errDetail.Value("field").String().Equal("tags")
			messages = errDetail.Value("messages").Array()
			messages.Length().Equal(1)
			messages.First().String().Contains("maximum 8 items required, but found 9 items")

			errDetail = errRes.Element(6).Object()
			errDetail.Value("type").String().Equal(v1.ErrorType_ERROR_TYPE_FIELD.String())
			errDetail.Value("field").String().Equal("tags[8]")
			messages = errDetail.Value("messages").Array()
			messages.Length().Equal(1)
			messages.First().String().Contains("length must be <= 128, but got 129")
		})

		t.Run("invalid Lua config", func(t *testing.T) {
			plugin := &v1.Plugin{Name: "acl"}
			var err error
			plugin.Config, err = structpb.NewStruct(map[string]interface{}{
				"deny":               []interface{}{16909060},
				"hide_groups_header": "true",
			})
			require.NoError(t, err)
			requestBody, _ := json.ProtoJSONMarshal(plugin)
			res := c.POST(p).WithBytes(requestBody).Expect()
			res.Status(http.StatusBadRequest)

			body := res.JSON().Object()
			body.Value("message").String().Equal("validation error")
			body.Value("details").Array().Length().Equal(2)
			errRes := body.Value("details").Array()

			errDetail := errRes.Element(0).Object()
			errDetail.Value("type").String().Equal(v1.ErrorType_ERROR_TYPE_FIELD.String())
			errDetail.Value("field").String().Equal("config.deny[0]")
			messages := errDetail.Value("messages").Array()
			messages.Length().Equal(1)
			messages.First().String().Equal("expected a string")

			errDetail = errRes.Element(1).Object()
			errDetail.Value("type").String().Equal(v1.ErrorType_ERROR_TYPE_FIELD.String())
			errDetail.Value("field").String().Equal("config.hide_groups_header")
			messages = errDetail.Value("messages").Array()
			messages.Length().Equal(1)
			messages.First().String().Equal("expected a boolean")
		})
	})
}

func typeToRoute(typ model.Type) string {
	// Assume a type like, "ca_certificate" has the "ca-certificate" route.
	return strings.ReplaceAll(string(typ), "_", "-")
}
