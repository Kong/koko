package admin

import (
	"fmt"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/gavv/httpexpect/v2"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/log"
	"github.com/kong/koko/internal/plugin"
	"github.com/kong/koko/internal/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const pluginSchemaFormat = `return {
	name = "%s",
	fields = {
		{ config = {
				type = "record",
				fields = {
					{ field = { type = "string" } }
				}
			}
		}
	}
}`

func init() {
	validator, err := plugin.NewLuaValidator(plugin.Opts{Logger: log.Logger})
	if err != nil {
		panic(err)
	}

	err = validator.LoadSchemasFromEmbed(plugin.Schemas, "schemas")
	if err != nil {
		panic(err)
	}
	resource.SetValidator(validator)
}

func goodPluginSchema(name string) *v1.PluginSchema {
	return &v1.PluginSchema{
		LuaSchema: fmt.Sprintf(pluginSchemaFormat, name),
	}
}

func validatePluginSchema(name string, body *httpexpect.Object) {
	body.ValueEqual("name", name)
	body.ValueEqual("lua_schema", fmt.Sprintf(pluginSchemaFormat, name))
	body.ContainsKey("created_at")
	body.ContainsKey("updated_at")
}

func TestPluginSchema_Create(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	t.Run("create a Lua plugin using the plugin-schemas endpoint", func(t *testing.T) {
		pluginSchemaBytes, err := json.ProtoJSONMarshal(goodPluginSchema("new-lua-plugin"))
		assert.NoError(t, err)
		res := c.POST("/v1/plugin-schemas/lua").WithBytes(pluginSchemaBytes).Expect()
		res.Status(http.StatusCreated)
		res.Header("grpc-metadata-koko-status-code").Empty()
		body := res.JSON().Path("$.item").Object()
		validatePluginSchema("new-lua-plugin", body)
	})

	t.Run("recreating a Lua plugin using the plugin-schemas endpoint fails", func(t *testing.T) {
		// Create first instance
		pluginSchemaBytes, err := json.ProtoJSONMarshal(goodPluginSchema("recreate-new-lua-plugin"))
		assert.NoError(t, err)
		res := c.POST("/v1/plugin-schemas/lua").WithBytes(pluginSchemaBytes).Expect()
		res.Status(http.StatusCreated)
		body := res.JSON().Path("$.item").Object()
		validatePluginSchema("recreate-new-lua-plugin", body)

		// Recreate plugin schema instance
		res = c.POST("/v1/plugin-schemas/lua").WithBytes(pluginSchemaBytes).Expect()
		res.Status(http.StatusBadRequest)
		res.Header("grpc-metadata-koko-status-code").Empty()
		body = res.JSON().Object()
		body.ValueEqual("message", "data constraint error")
		body.Value("details").Array().Length().Equal(1)
		resErr := body.Value("details").Array().Element(0)
		resErr.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_REFERENCE.String())
		resErr.Object().ValueEqual("field", "name")
		resErr.Object().ValueEqual("messages", []string{
			"name (type: unique) constraint failed for value 'recreate-new-lua-plugin': ",
		})
	})

	t.Run("creating a Lua plugin schema with a missing schema fails", func(t *testing.T) {
		res := c.POST("/v1/plugin-schemas/lua").Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		resErr := body.Value("details").Array().Element(0)
		resErr.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_ENTITY.String())
		resErr.Object().ValueEqual("messages", []string{
			"missing properties: 'lua_schema'",
		})
	})

	t.Run("creating Lua plugin schema should fail for bundled plugin schemas", func(t *testing.T) {
		schemaFiles, _ := plugin.Schemas.ReadDir("schemas")
		for _, schemaFile := range schemaFiles {
			name := schemaFile.Name()
			pluginName := name[:len(name)-len(filepath.Ext(name))]
			schema, _ := plugin.Schemas.ReadFile("schemas/" + name)

			pluginSchemaBytes, err := json.ProtoJSONMarshal(&v1.PluginSchema{
				LuaSchema: string(schema),
			})
			assert.NoError(t, err)
			res := c.POST("/v1/plugin-schemas/lua").WithBytes(pluginSchemaBytes).Expect()
			res.Status(http.StatusBadRequest)
			body := res.JSON().Object()
			body.ValueEqual("message", "validation error")
			body.Value("details").Array().Length().Equal(1)
			resErr := body.Value("details").Array().Element(0)
			resErr.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_ENTITY.String())
			resErr.Object().ValueEqual("messages", []string{
				fmt.Sprintf("unique constraint failed: schema already exists for plugin '%s'", pluginName),
			})
		}
	})
}
