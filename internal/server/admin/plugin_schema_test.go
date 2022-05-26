package admin

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/gavv/httpexpect/v2"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/log"
	"github.com/kong/koko/internal/plugin"
	"github.com/kong/koko/internal/plugin/validators"
	"github.com/kong/koko/internal/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/structpb"
)

const pluginSchemaFormat = `return {
	name = "%s",
	fields = {
		{ config = {
				type = "record",
				fields = {
					{ field = { type = "%s" } }
				}
			}
		}
	}
}`

func init() {
	validator, err := validators.NewLuaValidator(validators.Opts{Logger: log.Logger})
	if err != nil {
		panic(err)
	}

	err = validator.LoadSchemasFromEmbed(plugin.Schemas, "schemas")
	if err != nil {
		panic(err)
	}
	resource.SetValidator(validator)
}

func goodPluginSchema(name, fieldType string) *v1.PluginSchema {
	return &v1.PluginSchema{
		LuaSchema: fmt.Sprintf(pluginSchemaFormat, name, fieldType),
	}
}

func validatePluginSchema(name, fieldType string, body *httpexpect.Object) {
	body.ValueEqual("name", name)
	body.ValueEqual("lua_schema", fmt.Sprintf(pluginSchemaFormat, name, fieldType))
	body.ContainsKey("created_at")
	body.ContainsKey("updated_at")
}

func TestPluginSchema_Create(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	t.Run("create a Lua plugin using the plugin-schemas endpoint", func(t *testing.T) {
		pluginSchemaBytes, err := json.ProtoJSONMarshal(goodPluginSchema("new-lua-plugin", "string"))
		assert.NoError(t, err)
		res := c.POST("/v1/plugin-schemas").WithBytes(pluginSchemaBytes).Expect()
		res.Status(http.StatusCreated)
		res.Header("grpc-metadata-koko-status-code").Empty()
		body := res.JSON().Path("$.item").Object()
		validatePluginSchema("new-lua-plugin", "string", body)
	})

	t.Run("recreating a Lua plugin using the plugin-schemas endpoint fails", func(t *testing.T) {
		// Create first instance
		pluginSchemaBytes, err := json.ProtoJSONMarshal(goodPluginSchema("recreate-new-lua-plugin", "string"))
		assert.NoError(t, err)
		res := c.POST("/v1/plugin-schemas").WithBytes(pluginSchemaBytes).Expect()
		res.Status(http.StatusCreated)
		body := res.JSON().Path("$.item").Object()
		validatePluginSchema("recreate-new-lua-plugin", "string", body)

		// Recreate plugin schema instance
		res = c.POST("/v1/plugin-schemas").WithBytes(pluginSchemaBytes).Expect()
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
		res := c.POST("/v1/plugin-schemas").Expect()
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
			res := c.POST("/v1/plugin-schemas").WithBytes(pluginSchemaBytes).Expect()
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

func TestPluginSchema_Get(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)

	pluginSchemaBytes, err := json.ProtoJSONMarshal(goodPluginSchema("new-lua-plugin", "string"))
	assert.NoError(t, err)
	res := c.POST("/v1/plugin-schemas").WithBytes(pluginSchemaBytes).Expect()
	res.Status(http.StatusCreated)
	name := res.JSON().Path("$.item.name").String().Raw()

	t.Run("valid plugin schema ID returns 200", func(t *testing.T) {
		res := c.GET("/v1/plugin-schemas/" + name).Expect().Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		validatePluginSchema("new-lua-plugin", "string", body)
	})

	t.Run("valid plugin schema name returns 200", func(t *testing.T) {
		res := c.GET("/v1/plugin-schemas/new-lua-plugin").Expect().Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		validatePluginSchema("new-lua-plugin", "string", body)
	})

	t.Run("non-existent plugin schema returns 404", func(t *testing.T) {
		nonExistentName := "/non-existent-plugin-schema"
		c.GET("/v1/plugin-schemas/" + nonExistentName).Expect().Status(http.StatusNotFound)
	})

	t.Run("get request without a name returns 400", func(t *testing.T) {
		res := c.GET("/v1/plugin-schemas/").Expect().Status(http.StatusBadRequest)
		body := res.JSON().Object()
		gotErr := body.Value("details").Array().Element(0)
		gotErr.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_ENTITY.
			String())
		gotErr.Object().ValueEqual("messages", []string{
			"required name is missing",
		})
	})

	t.Run("get request with invalid characters in name returns 400", func(t *testing.T) {
		pluginNames := []string{
			"lua plugin",
			"lua\\plugin",
			"lua+plugin",
			"lua!plugin",
		}
		expectedErrMsg := fmt.Sprintf("must match pattern: '%s'", namePattern)
		for _, pluginName := range pluginNames {
			fmt.Fprintf(os.Stderr, "\n\n%s\n", pluginName)
			res := c.GET(fmt.Sprintf("/v1/plugin-schemas/%s", pluginName)).Expect().Status(http.StatusBadRequest)
			body := res.JSON().Object()
			gotErr := body.Value("details").Array().Element(0)
			gotErr.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_ENTITY.
				String())
			gotErr.Object().ValueEqual("messages", []string{expectedErrMsg})
		}
	})
}

func TestPluginSchema_List(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)

	pluginSchemaNames := make([]string, 0, 6)
	for i := 1; i <= 6; i++ {
		pluginName := fmt.Sprintf("plugin-schema-%d", i)
		pluginSchemaBytes, err := json.ProtoJSONMarshal(goodPluginSchema(pluginName, "string"))
		assert.NoError(t, err)
		res := c.POST("/v1/plugin-schemas").WithBytes(pluginSchemaBytes).Expect()
		res.Status(http.StatusCreated)
		pluginSchemaNames = append(pluginSchemaNames, res.JSON().Path("$.item.name").String().Raw())
	}

	t.Run("list all plugin schemas", func(t *testing.T) {
		body := c.GET("/v1/plugin-schemas").Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(6)
		var gotPluginSchemaNames []string
		for _, item := range items.Iter() {
			gotPluginSchemaNames = append(gotPluginSchemaNames, item.Object().Value("name").String().Raw())
		}
		require.ElementsMatch(t, pluginSchemaNames, gotPluginSchemaNames)
	})

	t.Run("list returns multiple plugin schemas with paging", func(t *testing.T) {
		body := c.GET("/v1/plugin-schemas").
			WithQuery("page.size", "4").
			WithQuery("page.number", "1").
			Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(4)
		var gotPluginSchemaNames []string
		for _, item := range items.Iter() {
			gotPluginSchemaNames = append(gotPluginSchemaNames, item.Object().Value("name").String().Raw())
		}
		body.Value("page").Object().Value("total_count").Number().Equal(6)
		body.Value("page").Object().Value("next_page_num").Number().Equal(2)

		body = c.GET("/v1/plugin-schemas").
			WithQuery("page.size", "4").
			WithQuery("page.number", "2").
			Expect().Status(http.StatusOK).JSON().Object()
		items = body.Value("items").Array()
		items.Length().Equal(2)
		for _, item := range items.Iter() {
			gotPluginSchemaNames = append(gotPluginSchemaNames, item.Object().Value("name").String().Raw())
		}
		body.Value("page").Object().Value("total_count").Number().Equal(6)
		body.Value("page").Object().NotContainsKey("next_page_num")
		require.ElementsMatch(t, pluginSchemaNames, gotPluginSchemaNames)
	})
}

func TestPluginSchema_Put(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)

	t.Run("creating a new schema using PUT succeeds", func(t *testing.T) {
		pluginSchemaBytes, err := json.ProtoJSONMarshal(goodPluginSchema("put-new-lua-plugin", "string"))
		assert.NoError(t, err)

		res := c.PUT("/v1/plugin-schemas/" + "put-new-lua-plugin").WithBytes(pluginSchemaBytes).Expect()
		res.Status(http.StatusOK)

		res.Header("grpc-metadata-koko-status-code").Empty()
		body := res.JSON().Path("$.item").Object()
		validatePluginSchema("put-new-lua-plugin", "string", body)
	})

	t.Run("creating a Lua plugin schema with a missing schema fails", func(t *testing.T) {
		res := c.PUT("/v1/plugin-schemas/" + "put-missing-schema-lua-plugin").Expect()
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
			res := c.PUT("/v1/plugin-schemas/" + pluginName).WithBytes(pluginSchemaBytes).Expect()
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

	t.Run("updating an existing schema with a valid schema succeeds", func(t *testing.T) {
		// Create an instance
		pluginSchemaBytes, err := json.ProtoJSONMarshal(goodPluginSchema("put-again-lua-plugin", "string"))
		assert.NoError(t, err)

		res := c.PUT("/v1/plugin-schemas/" + "put-again-lua-plugin").WithBytes(pluginSchemaBytes).Expect()
		res.Status(http.StatusOK)

		body := res.JSON().Path("$.item").Object()
		validatePluginSchema("put-again-lua-plugin", "string", body)

		// PUT it again, but different type in the schema
		changedSchemaBytes, err := json.ProtoJSONMarshal(goodPluginSchema("put-again-lua-plugin", "number"))
		assert.NoError(t, err)
		res = c.PUT("/v1/plugin-schemas/" + "put-again-lua-plugin").WithBytes(changedSchemaBytes).Expect()
		res.Status(http.StatusOK)

		res.Header("grpc-metadata-koko-status-code").Empty()
		body = res.JSON().Path("$.item").Object()
		validatePluginSchema("put-again-lua-plugin", "number", body)
	})

	t.Run("updating a valid schema with an empty one fails", func(t *testing.T) {
		const name = "replace-with-empty-schema-lua-plugin"
		// Create an instance
		pluginSchemaBytes, err := json.ProtoJSONMarshal(goodPluginSchema(name, "string"))
		assert.NoError(t, err)

		res := c.PUT("/v1/plugin-schemas/" + name).WithBytes(pluginSchemaBytes).Expect()
		res.Status(http.StatusOK)

		body := res.JSON().Path("$.item").Object()
		validatePluginSchema(name, "string", body)

		// put again, without schema
		res = c.PUT("/v1/plugin-schemas/" + name).Expect()
		res.Status(http.StatusBadRequest)
		body = res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		resErr := body.Value("details").Array().Element(0)
		resErr.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_ENTITY.String())
		resErr.Object().ValueEqual("messages", []string{
			"missing properties: 'lua_schema'",
		})
	})
}

func TestPluginSchema_Delete(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)

	// create a plugin schema
	pluginSchemaName := "new-lua-plugin"
	res := c.POST("/v1/plugin-schemas").WithJSON(
		goodPluginSchema(pluginSchemaName, "string"),
	).Expect()
	res.Status(http.StatusCreated)
	res.Header("grpc-metadata-koko-status-code").Empty()
	body := res.JSON().Path("$.item").Object()
	validatePluginSchema(pluginSchemaName, "string", body)

	t.Run("delete an unused plugin-schema successfully", func(t *testing.T) {
		res := c.DELETE("/v1/plugin-schemas/" + pluginSchemaName).Expect()
		res.Status(http.StatusNoContent)
	})
	t.Run("delete request without a name returns 400", func(t *testing.T) {
		res := c.DELETE("/v1/plugin-schemas/").Expect().Status(http.StatusBadRequest)
		body := res.JSON().Object()
		gotErr := body.Value("details").Array().Element(0)
		gotErr.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_ENTITY.
			String())
		gotErr.Object().ValueEqual("messages", []string{
			"required name is missing",
		})
	})
	t.Run("deleting a non-existent plugin-schema returns 404", func(t *testing.T) {
		res := c.DELETE("/v1/plugin-schemas/" + pluginSchemaName).Expect()
		res.Status(http.StatusNotFound)
	})
	t.Run("deleting a plugin-schema currently in use fails", func(t *testing.T) {
		name := "valid"
		// add plugin-schema
		res := c.POST("/v1/plugin-schemas").WithJSON(
			goodPluginSchema(name, "string"),
		).Expect()
		res.Status(http.StatusCreated)

		var config structpb.Struct
		configString := `{"field": "non-bundled-plugin-configuration"}`
		require.Nil(t, json.ProtoJSONUnmarshal([]byte(configString), &config))
		plugin := &v1.Plugin{
			Name:      name,
			Protocols: []string{"http", "https"},
			Config:    &config,
		}
		// add plugin from non-bundled schema
		res = c.POST("/v1/plugins").WithJSON(plugin).Expect()
		res.Status(http.StatusCreated)
		body := res.JSON().Path("$.item").Object()
		body.Value("name").Equal(name)
		cfg := body.Path("$.config").Object()
		cfg.Value("field").Equal("non-bundled-plugin-configuration")

		// attempt to delete the plugin-schema
		res = c.DELETE("/v1/plugin-schemas/" + name).Expect()
		res.Status(http.StatusBadRequest)
		body = res.JSON().Object()
		body.ValueEqual("message", "plugin schema is currently in use, "+
			"please delete existing plugins using the schema and try again")
		body.ValueEqual("code", codes.InvalidArgument)
	})
}
