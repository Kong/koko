package admin

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/google/uuid"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/test/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func goodKeyAuthPlugin() *v1.Plugin {
	return &v1.Plugin{
		Name: "key-auth",
	}
}

func validateKeyAuthPlugin(body *httpexpect.Object) {
	body.ContainsKey("id")
	body.ValueEqual("name", "key-auth")
	body.ContainsKey("created_at")
	body.ContainsKey("updated_at")
	body.Value("protocols").Array().ContainsOnly(
		"http",
		"https",
		"grpc",
		"grpcs")
	body.ValueEqual("enabled", true)
	config := body.Value("config").Object()
	config.ValueEqual("anonymous", nil)
	config.ValueEqual("key_in_body", false)
	config.ValueEqual("hide_credentials", false)
	config.ValueEqual("key_in_query", true)
	config.ValueEqual("key_in_header", true)
	config.ValueEqual("key_names", []string{"apikey"})
	config.ValueEqual("run_on_preflight", true)
}

func TestPluginCreate(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	t.Run("creates a valid global plugin", func(t *testing.T) {
		pluginBytes, err := json.ProtoJSONMarshal(goodKeyAuthPlugin())
		require.Nil(t, err)
		res := c.POST("/v1/plugins").WithBytes(pluginBytes).Expect()
		res.Status(http.StatusCreated)
		res.Header("grpc-metadata-koko-status-code").Empty()
		body := res.JSON().Path("$.item").Object()
		validateKeyAuthPlugin(body)
	})
	t.Run("creating a plugin with 'ordering' set fails", func(t *testing.T) {
		util.SkipForEnterpriseTests(t, true)
		plugin := &v1.Plugin{
			Name: "prometheus",
			Ordering: &v1.Ordering{
				Before: &v1.Order{
					Access: []string{"pre-function"},
				},
			},
		}
		pluginBytes, err := json.ProtoJSONMarshal(plugin)
		require.NoError(t, err)
		res := c.POST("/v1/plugins").WithBytes(pluginBytes).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		resErr := body.Value("details").Array().Element(0)
		resErr.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_ENTITY.String())
		resErr.Object().ValueEqual("messages", []string{
			"'ordering' is a Kong Enterprise-only feature. " +
				"Please upgrade to Kong Enterprise to use this feature.",
		})
	})
	t.Run("recreating the same plugin fails", func(t *testing.T) {
		pluginBytes, err := json.ProtoJSONMarshal(goodKeyAuthPlugin())
		require.Nil(t, err)
		res := c.POST("/v1/plugins").WithBytes(pluginBytes).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "data constraint error")
		body.Value("details").Array().Length().Equal(1)
		resErr := body.Value("details").Array().Element(0)
		resErr.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_REFERENCE.String())
		resErr.Object().ValueEqual("messages", []string{
			"unique-plugin-per-entity (" +
				"type: unique) constraint failed for value 'key-auth...': ",
		})
	})
	t.Run("creating a plugin with a non-existent service fails", func(t *testing.T) {
		plugin := &v1.Plugin{
			Name: "key-auth",
			Service: &v1.Service{
				Id: uuid.NewString(),
			},
			Enabled:   wrapperspb.Bool(true),
			Protocols: []string{"http", "https"},
		}
		pluginBytes, err := json.ProtoJSONMarshal(plugin)
		require.Nil(t, err)
		res := c.POST("/v1/plugins").WithBytes(pluginBytes).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "data constraint error")
		body.Value("details").Array().Length().Equal(1)
		errRes := body.Value("details").Array().Element(0)
		errRes.Object().ValueEqual("type",
			v1.ErrorType_ERROR_TYPE_REFERENCE.String())
		errRes.Object().ValueEqual("field", "service.id")
	})
	t.Run("creating a plugin with a valid service.id succeeds", func(t *testing.T) {
		service := goodService()
		service.Id = uuid.NewString()
		res := c.POST("/v1/services").WithJSON(service).Expect()
		res.Status(http.StatusCreated)
		plugin := &v1.Plugin{
			Name: "key-auth",
			Service: &v1.Service{
				Id: service.Id,
			},
		}
		pluginBytes, err := json.ProtoJSONMarshal(plugin)
		require.Nil(t, err)
		res = c.POST("/v1/plugins").WithBytes(pluginBytes).Expect()
		res.Status(http.StatusCreated)
	})
	t.Run("creating a plugin with a non-existent route fails", func(t *testing.T) {
		plugin := &v1.Plugin{
			Name: "key-auth",
			Route: &v1.Route{
				Id: uuid.NewString(),
			},
			Enabled:   wrapperspb.Bool(true),
			Protocols: []string{"http", "https"},
		}
		pluginBytes, err := json.ProtoJSONMarshal(plugin)
		require.Nil(t, err)
		res := c.POST("/v1/plugins").WithBytes(pluginBytes).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "data constraint error")
		body.Value("details").Array().Length().Equal(1)
		errRes := body.Value("details").Array().Element(0)
		errRes.Object().ValueEqual("type",
			v1.ErrorType_ERROR_TYPE_REFERENCE.String())
		errRes.Object().ValueEqual("field", "route.id")
	})
	t.Run("creating a plugin with a valid route.id succeeds", func(t *testing.T) {
		route := goodRoute()
		route.Id = uuid.NewString()
		res := c.POST("/v1/routes").WithJSON(route).Expect()
		res.Status(http.StatusCreated)
		plugin := &v1.Plugin{
			Name: "key-auth",
			Route: &v1.Route{
				Id: route.Id,
			},
		}
		pluginBytes, err := json.ProtoJSONMarshal(plugin)
		require.Nil(t, err)
		res = c.POST("/v1/plugins").WithBytes(pluginBytes).Expect()
		res.Status(http.StatusCreated)
	})

	t.Run("creating a plugin with a valid consumer.id succeeds", func(t *testing.T) {
		consumer := goodConsumer()
		res := c.POST("/v1/consumers").WithJSON(consumer).Expect()
		res.Status(http.StatusCreated)
		body := res.JSON().Path("$.item").Object()
		consumerID := body.Value("id").String().Raw()
		var config structpb.Struct
		configString := `{"header_name": "Kong-Request-ID", "generator": "uuid#counter", "echo_downstream": true }`
		require.Nil(t, json.ProtoJSONUnmarshal([]byte(configString), &config))
		plugin := &v1.Plugin{
			Name:      "correlation-id",
			Protocols: []string{"http", "https"},
			Consumer: &v1.Consumer{
				Id: consumerID,
			},
			Config: &config,
		}
		pluginBytes, err := json.ProtoJSONMarshal(plugin)
		require.Nil(t, err)
		res = c.POST("/v1/plugins").WithBytes(pluginBytes).Expect()
		res.Status(http.StatusCreated)
		body = res.JSON().Path("$.item").Object()
		body.Value("consumer").Object().Value("id").Equal(consumerID)
	})

	t.Run("creating a plugin with a valid route.id and service.id succeeds",
		func(t *testing.T) {
			service := goodService()
			service.Id = uuid.NewString()
			service.Name = "foo-plugin-service"
			res := c.POST("/v1/services").WithJSON(service).Expect()
			res.Status(http.StatusCreated)

			route := goodRoute()
			route.Id = uuid.NewString()
			route.Name = "foo-plugin-route"
			res = c.POST("/v1/routes").WithJSON(route).Expect()
			res.Status(http.StatusCreated)

			plugin := &v1.Plugin{
				Name: "key-auth",
				Route: &v1.Route{
					Id: route.Id,
				},
				Service: &v1.Service{
					Id: service.Id,
				},
			}
			pluginBytes, err := json.ProtoJSONMarshal(plugin)
			require.Nil(t, err)
			res = c.POST("/v1/plugins").WithBytes(pluginBytes).Expect()
			res.Status(http.StatusCreated)
		})
	t.Run("creating a unknown plugin error",
		func(t *testing.T) {
			plugin := &v1.Plugin{
				Name: "no-auth",
			}
			pluginBytes, err := json.ProtoJSONMarshal(plugin)
			require.Nil(t, err)
			res := c.POST("/v1/plugins").WithBytes(pluginBytes).Expect()
			res.Status(http.StatusBadRequest)
			body := res.JSON().Object()
			body.ValueEqual("message", "validation error")
			body.Value("details").Array().Length().Equal(1)
			errRes := body.Value("details").Array().Element(0)
			errRes.Object().ValueEqual("type",
				v1.ErrorType_ERROR_TYPE_FIELD.String())
			errRes.Object().ValueEqual("field", "name")
			errRes.Object().ValueEqual("messages", []string{
				"plugin(no-auth) does not exist",
			})
		})
	t.Run("creates a valid plugin specifying the ID using POST", func(t *testing.T) {
		plugin := &v1.Plugin{
			Name: "basic-auth",
			Id:   uuid.NewString(),
		}
		res := c.POST("/v1/plugins").WithJSON(plugin).Expect()
		res.Status(http.StatusCreated)
		body := res.JSON().Path("$.item").Object()
		body.Value("id").Equal(plugin.Id)
	})

	t.Run("create a valid plugin from a non bundled schema", func(t *testing.T) {
		res := c.POST("/v1/plugin-schemas").
			WithJSON(goodPluginSchema("valid", "string")).Expect()
		res.Status(http.StatusCreated)

		var config structpb.Struct
		configString := `{"field": "non-bundled-plugin-configuration"}`
		require.Nil(t, json.ProtoJSONUnmarshal([]byte(configString), &config))
		plugin := &v1.Plugin{
			Name:      "valid",
			Protocols: []string{"http", "https"},
			Config:    &config,
		}
		res = c.POST("/v1/plugins").WithJSON(plugin).Expect()
		res.Status(http.StatusCreated)
		body := res.JSON().Path("$.item").Object()
		body.Value("name").Equal("valid")
		cfg := body.Path("$.config").Object()
		cfg.Value("field").Equal("non-bundled-plugin-configuration")
	})

	t.Run("fail to create a plugin from a non bundled schema", func(t *testing.T) {
		res := c.POST("/v1/plugin-schemas").
			WithJSON(goodPluginSchema("fail-validation", "string")).Expect()
		res.Status(http.StatusCreated)

		var config structpb.Struct
		configString := `{"non-existent-field": "non-bundled-plugin-configuration"}`
		require.Nil(t, json.ProtoJSONUnmarshal([]byte(configString), &config))
		plugin := &v1.Plugin{
			Name:      "fail-validation",
			Protocols: []string{"http", "https"},
			Config:    &config,
		}
		res = c.POST("/v1/plugins").WithJSON(plugin).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		errRes := body.Value("details").Array().Element(0)
		errRes.Object().ValueEqual("type",
			v1.ErrorType_ERROR_TYPE_FIELD.String())
		errRes.Object().ValueEqual("field", "config.non-existent-field")
		errRes.Object().ValueEqual("messages", []string{
			"unknown field",
		})
	})
}

func TestPluginUpsert(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	t.Run("upsert a valid plugin", func(t *testing.T) {
		pluginBytes, err := json.ProtoJSONMarshal(goodKeyAuthPlugin())
		require.Nil(t, err)
		res := c.PUT("/v1/plugins/" + uuid.NewString()).
			WithBytes(pluginBytes).
			Expect()
		res.Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		validateKeyAuthPlugin(body)
	})
	t.Run("re-upserting the same plugin with different id fails",
		func(t *testing.T) {
			pluginBytes, err := json.ProtoJSONMarshal(goodKeyAuthPlugin())
			require.Nil(t, err)
			res := c.PUT("/v1/plugins/" + uuid.NewString()).
				WithBytes(pluginBytes).
				Expect()
			res.Status(http.StatusBadRequest)
			body := res.JSON().Object()
			body.ValueEqual("message", "data constraint error")
			body.Value("details").Array().Length().Equal(1)
			resErr := body.Value("details").Array().Element(0)
			resErr.Object().ValueEqual("type",
				v1.ErrorType_ERROR_TYPE_REFERENCE.String())
			resErr.Object().ValueEqual("messages", []string{
				"unique-plugin-per-entity (" +
					"type: unique) constraint failed for value 'key-auth...': ",
			})
		})
	t.Run("upserting a plugin with a non-existent service fails", func(t *testing.T) {
		plugin := &v1.Plugin{
			Name: "key-auth",
			Service: &v1.Service{
				Id: uuid.NewString(),
			},
		}
		pluginBytes, err := json.ProtoJSONMarshal(plugin)
		require.Nil(t, err)
		res := c.PUT("/v1/plugins/" + uuid.NewString()).
			WithBytes(pluginBytes).
			Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "data constraint error")
		body.Value("details").Array().Length().Equal(1)
		errRes := body.Value("details").Array().Element(0)
		errRes.Object().ValueEqual("type",
			v1.ErrorType_ERROR_TYPE_REFERENCE.String())
		errRes.Object().ValueEqual("field", "service.id")
	})
	t.Run("upsert correctly updates a plugin", func(t *testing.T) {
		pid := uuid.NewString()
		config, err := structpb.NewStruct(map[string]interface{}{"second": 42})
		require.Nil(t, err)
		plugin := &v1.Plugin{
			Id:     pid,
			Name:   "rate-limiting",
			Config: config,
		}
		res := c.POST("/v1/plugins").
			WithJSON(plugin).
			Expect()
		res.Status(http.StatusCreated)
		res.JSON().Path("$.item.config.second").Number().Equal(42)
		res.JSON().Path("$.item.config.day").Null()

		config, err = structpb.NewStruct(map[string]interface{}{"day": 42})
		require.Nil(t, err)
		plugin = &v1.Plugin{
			Name:   "rate-limiting",
			Config: config,
		}
		res = c.PUT("/v1/plugins/" + pid).
			WithJSON(plugin).
			Expect()
		res.Status(http.StatusOK)

		res = c.GET("/v1/plugins/" + pid).Expect()
		res.Status(http.StatusOK)
		res.JSON().Path("$.item.config.day").Number().Equal(42)
		res.JSON().Path("$.item.config.second").Null()
	})
	t.Run("upsert plugin without id fails", func(t *testing.T) {
		pluginBytes, err := json.ProtoJSONMarshal(goodKeyAuthPlugin())
		require.Nil(t, err)
		res := c.PUT("/v1/plugins/").
			WithBytes(pluginBytes).
			Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", " '' is not a valid uuid")
	})
	t.Run("upsert a valid plugin from a non bundled schema", func(t *testing.T) {
		res := c.POST("/v1/plugin-schemas").
			WithJSON(goodPluginSchema("valid-upsert", "string")).Expect()
		res.Status(http.StatusCreated)

		var config structpb.Struct
		configString := `{"field": "non-bundled-plugin-configuration"}`
		require.Nil(t, json.ProtoJSONUnmarshal([]byte(configString), &config))
		plugin := &v1.Plugin{
			Name:   "valid-upsert",
			Config: &config,
		}
		res = c.PUT("/v1/plugins/" + uuid.NewString()).WithJSON(plugin).Expect()
		res.Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		body.Value("name").Equal("valid-upsert")
		cfg := body.Path("$.config").Object()
		cfg.Value("field").Equal("non-bundled-plugin-configuration")
	})
	t.Run("update a valid plugin from a non bundled schema", func(t *testing.T) {
		res := c.POST("/v1/plugin-schemas").
			WithJSON(goodPluginSchema("valid-update", "string")).Expect()
		res.Status(http.StatusCreated)

		id := uuid.NewString()
		var config structpb.Struct
		configString := `{"field": "non-bundled-plugin-configuration"}`
		require.Nil(t, json.ProtoJSONUnmarshal([]byte(configString), &config))
		plugin := &v1.Plugin{
			Id:     id,
			Name:   "valid-update",
			Config: &config,
		}
		res = c.POST("/v1/plugins").WithJSON(plugin).Expect()
		res.Status(http.StatusCreated)
		body := res.JSON().Path("$.item").Object()
		body.Value("name").Equal("valid-update")
		cfg := body.Path("$.config").Object()
		cfg.Value("field").Equal("non-bundled-plugin-configuration")

		configString = `{"field": "updated-non-bundled-plugin-configuration"}`
		require.Nil(t, json.ProtoJSONUnmarshal([]byte(configString), &config))
		plugin = &v1.Plugin{
			Id:     id,
			Name:   "valid-update",
			Config: &config,
		}
		res = c.PUT("/v1/plugins/" + id).WithJSON(plugin).Expect()
		res.Status(http.StatusOK)
		body = res.JSON().Path("$.item").Object()
		body.Value("name").Equal("valid-update")
		cfg = body.Path("$.config").Object()
		cfg.Value("field").Equal("updated-non-bundled-plugin-configuration")
	})
}

func TestPluginDelete(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	pluginBytes, err := json.ProtoJSONMarshal(goodKeyAuthPlugin())
	require.Nil(t, err)
	res := c.POST("/v1/plugins").WithBytes(pluginBytes).Expect()
	id := res.JSON().Path("$.item.id").String().Raw()
	res.Status(http.StatusCreated)
	t.Run("deleting a non-existent plugin returns 404", func(t *testing.T) {
		randomID := "071f5040-3e4a-46df-9d98-451e79e318fd"
		c.DELETE("/v1/plugins/" + randomID).Expect().Status(http.StatusNotFound)
	})
	t.Run("deleting a plugin return 204", func(t *testing.T) {
		c.DELETE("/v1/plugins/" + id).Expect().Status(http.StatusNoContent)
	})
	t.Run("delete request without an ID returns 400", func(t *testing.T) {
		res := c.DELETE("/v1/plugins/").Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", " '' is not a valid uuid")
	})
	t.Run("delete request with an invalid ID returns 400", func(t *testing.T) {
		res := c.DELETE("/v1/plugins/" + "Not-Valid").Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", " 'Not-Valid' is not a valid uuid")
	})
}

func TestPluginRead(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	pluginBytes, err := json.ProtoJSONMarshal(goodKeyAuthPlugin())
	require.Nil(t, err)
	res := c.POST("/v1/plugins").WithBytes(pluginBytes).Expect()
	res.Status(http.StatusCreated)
	id := res.JSON().Path("$.item.id").String().Raw()
	t.Run("reading a non-existent plugin returns 404", func(t *testing.T) {
		randomID := "071f5040-3e4a-46df-9d98-451e79e318fd"
		c.GET("/v1/plugins/" + randomID).Expect().Status(http.StatusNotFound)
	})
	t.Run("reading a plugin return 200", func(t *testing.T) {
		res := c.GET("/v1/plugins/" + id).Expect().Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		validateKeyAuthPlugin(body)
	})
	t.Run("read request without an ID returns 400", func(t *testing.T) {
		res := c.GET("/v1/plugins/").Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "required ID is missing")
	})
}

func TestConfiguredPluginsList(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)

	svc := &v1.Service{
		Id:   uuid.NewString(),
		Name: "foo",
		Host: "example.com",
		Path: "/foo",
	}
	c.POST("/v1/services").WithJSON(svc).
		Expect().Status(http.StatusCreated)

	route := &v1.Route{
		Id:    uuid.NewString(),
		Name:  "quux",
		Paths: []string{"/quux"},
	}
	c.POST("/v1/routes").WithJSON(route).
		Expect().Status(http.StatusCreated)

	plugin := &v1.Plugin{
		Name:      "request-transformer",
		Enabled:   wrapperspb.Bool(true),
		Protocols: []string{"http", "https"},
	}
	pluginBytes, err := json.ProtoJSONMarshal(plugin)
	require.Nil(t, err)
	c.POST("/v1/plugins").WithBytes(pluginBytes).
		Expect().Status(http.StatusCreated)

	plugin = &v1.Plugin{
		Name:      "basic-auth",
		Enabled:   wrapperspb.Bool(true),
		Protocols: []string{"http", "https"},
		Service: &v1.Service{
			Id: svc.Id,
		},
	}
	pluginBytes, err = json.ProtoJSONMarshal(plugin)
	require.Nil(t, err)
	c.POST("/v1/plugins").WithBytes(pluginBytes).
		Expect().Status(http.StatusCreated)

	plugin = &v1.Plugin{
		Name:      "basic-auth",
		Enabled:   wrapperspb.Bool(true),
		Protocols: []string{"http", "https"},
		Route: &v1.Route{
			Id: route.Id,
		},
	}
	pluginBytes, err = json.ProtoJSONMarshal(plugin)
	require.Nil(t, err)
	c.POST("/v1/plugins").WithBytes(pluginBytes).
		Expect().Status(http.StatusCreated)

	plugin = &v1.Plugin{
		Name:      "request-size-limiting",
		Enabled:   wrapperspb.Bool(true),
		Protocols: []string{"http", "https"},
		Route: &v1.Route{
			Id: route.Id,
		},
	}
	pluginBytes, err = json.ProtoJSONMarshal(plugin)
	require.Nil(t, err)
	c.POST("/v1/plugins").WithBytes(pluginBytes).
		Expect().Status(http.StatusCreated)

	t.Run("get configured plugins returns unique plugin names", func(t *testing.T) {
		body := c.GET("/v1/configured_plugins").Expect().
			Status(http.StatusOK).JSON()
		names := body.Path("$.names").Array()
		names.Contains("basic-auth")
		names.Contains("request-size-limiting")
		names.Contains("request-transformer")
		names.Length().Equal(3)
	})
	t.Run("get configured plugins using /v1/configured-plugins", func(t *testing.T) {
		body := c.GET("/v1/configured-plugins").Expect().
			Status(http.StatusOK).JSON()
		names := body.Path("$.names").Array()
		names.Contains("basic-auth")
		names.Contains("request-size-limiting")
		names.Contains("request-transformer")
		names.Length().Equal(3)
	})
}

func TestPluginList(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)

	svc := &v1.Service{
		Name: "foo",
		Host: "example.com",
		Path: "/foo",
	}
	res := c.POST("/v1/services").WithJSON(svc).Expect()
	res.Status(http.StatusCreated)
	serviceID1 := res.JSON().Path("$.item.id").String().Raw()
	svc = &v1.Service{
		Name: "bar",
		Host: "example.com",
		Path: "/bar",
	}
	res = c.POST("/v1/services").WithJSON(svc).Expect()
	res.Status(http.StatusCreated)
	serviceID2 := res.JSON().Path("$.item.id").String().Raw()
	svc = &v1.Service{
		Name: "baz",
		Host: "example.com",
		Path: "/baz",
	}
	res = c.POST("/v1/services").WithJSON(svc).Expect()
	res.Status(http.StatusCreated)
	serviceID3 := res.JSON().Path("$.item.id").String().Raw()

	rte := &v1.Route{
		Name:  "qux",
		Paths: []string{"/qux"},
	}
	res = c.POST("/v1/routes").WithJSON(rte).Expect()
	res.Status(http.StatusCreated)
	routeID1 := res.JSON().Path("$.item.id").String().Raw()
	rte = &v1.Route{
		Name:  "quux",
		Paths: []string{"/quux"},
	}
	res = c.POST("/v1/routes").WithJSON(rte).Expect()
	res.Status(http.StatusCreated)
	routeID2 := res.JSON().Path("$.item.id").String().Raw()
	rte = &v1.Route{
		Name:  "quuz",
		Paths: []string{"/quuz"},
	}
	res = c.POST("/v1/routes").WithJSON(rte).Expect()
	res.Status(http.StatusCreated)
	routeID3 := res.JSON().Path("$.item.id").String().Raw()

	pluginBytes, err := json.ProtoJSONMarshal(goodKeyAuthPlugin())
	require.Nil(t, err)
	res = c.POST("/v1/plugins").WithBytes(pluginBytes).Expect()
	res.Status(http.StatusCreated)
	pluginID1 := res.JSON().Path("$.item.id").String().Raw()
	plg := &v1.Plugin{
		Name:      "request-transformer",
		Enabled:   wrapperspb.Bool(true),
		Protocols: []string{"http", "https"},
	}
	pluginBytes, err = json.ProtoJSONMarshal(plg)
	require.Nil(t, err)
	res = c.POST("/v1/plugins").WithBytes(pluginBytes).Expect()
	res.Status(http.StatusCreated)
	pluginID2 := res.JSON().Path("$.item.id").String().Raw()
	plg = &v1.Plugin{
		Name:      "basic-auth",
		Enabled:   wrapperspb.Bool(true),
		Protocols: []string{"http", "https"},
		Service: &v1.Service{
			Id: serviceID1,
		},
	}
	pluginBytes, err = json.ProtoJSONMarshal(plg)
	require.Nil(t, err)
	res = c.POST("/v1/plugins").WithBytes(pluginBytes).Expect()
	res.Status(http.StatusCreated)
	pluginID3 := res.JSON().Path("$.item.id").String().Raw()
	plg = &v1.Plugin{
		Name:      "bot-detection",
		Enabled:   wrapperspb.Bool(true),
		Protocols: []string{"http", "https"},
		Service: &v1.Service{
			Id: serviceID1,
		},
	}
	pluginBytes, err = json.ProtoJSONMarshal(plg)
	require.Nil(t, err)
	res = c.POST("/v1/plugins").WithBytes(pluginBytes).Expect()
	res.Status(http.StatusCreated)
	pluginID4 := res.JSON().Path("$.item.id").String().Raw()
	plg = &v1.Plugin{
		Name:      "cors",
		Enabled:   wrapperspb.Bool(true),
		Protocols: []string{"http", "https"},
		Service: &v1.Service{
			Id: serviceID2,
		},
	}
	pluginBytes, err = json.ProtoJSONMarshal(plg)
	require.Nil(t, err)
	res = c.POST("/v1/plugins").WithBytes(pluginBytes).Expect()
	res.Status(http.StatusCreated)
	pluginID5 := res.JSON().Path("$.item.id").String().Raw()
	plg = &v1.Plugin{
		Name:      "hmac-auth",
		Enabled:   wrapperspb.Bool(true),
		Protocols: []string{"http", "https"},
		Route: &v1.Route{
			Id: routeID1,
		},
	}
	pluginBytes, err = json.ProtoJSONMarshal(plg)
	require.Nil(t, err)
	res = c.POST("/v1/plugins").WithBytes(pluginBytes).Expect()
	res.Status(http.StatusCreated)
	pluginID6 := res.JSON().Path("$.item.id").String().Raw()
	plg = &v1.Plugin{
		Name:      "jwt",
		Enabled:   wrapperspb.Bool(true),
		Protocols: []string{"http", "https"},
		Route: &v1.Route{
			Id: routeID2,
		},
	}
	pluginBytes, err = json.ProtoJSONMarshal(plg)
	require.Nil(t, err)
	res = c.POST("/v1/plugins").WithBytes(pluginBytes).Expect()
	res.Status(http.StatusCreated)
	pluginID7 := res.JSON().Path("$.item.id").String().Raw()
	plg = &v1.Plugin{
		Name:      "request-size-limiting",
		Enabled:   wrapperspb.Bool(true),
		Protocols: []string{"http", "https"},
		Route: &v1.Route{
			Id: routeID2,
		},
	}
	pluginBytes, err = json.ProtoJSONMarshal(plg)
	require.Nil(t, err)
	res = c.POST("/v1/plugins").WithBytes(pluginBytes).Expect()
	res.Status(http.StatusCreated)
	pluginID8 := res.JSON().Path("$.item.id").String().Raw()

	t.Run("list all plugins", func(t *testing.T) {
		body := c.GET("/v1/plugins").Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(8)
		var gotIDs []string
		for _, item := range items.Iter() {
			gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
		}
		require.ElementsMatch(t, []string{
			pluginID1,
			pluginID2,
			pluginID3,
			pluginID4,
			pluginID5,
			pluginID6,
			pluginID7,
			pluginID8,
		}, gotIDs)
	})

	t.Run("list plugins by service", func(t *testing.T) {
		body := c.GET("/v1/plugins").WithQuery("service_id", serviceID1).
			Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(2)
		var gotIDs []string
		for _, item := range items.Iter() {
			gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
		}
		require.ElementsMatch(t, []string{
			pluginID3,
			pluginID4,
		}, gotIDs)

		body = c.GET("/v1/plugins").WithQuery("service_id", serviceID2).
			Expect().Status(http.StatusOK).JSON().Object()
		items = body.Value("items").Array()
		items.Length().Equal(1)
		gotIDs = nil
		for _, item := range items.Iter() {
			gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
		}
		require.ElementsMatch(t, []string{pluginID5}, gotIDs)
	})

	t.Run("list plugins by service - no plugins associated with service", func(t *testing.T) {
		body := c.GET("/v1/plugins").WithQuery("service_id", serviceID3).
			Expect().Status(http.StatusOK).JSON().Object()
		body.Empty()
	})

	t.Run("list plugins by service - invalid service UUID", func(t *testing.T) {
		body := c.GET("/v1/plugins").WithQuery("service_id", "invalid-uuid").
			Expect().Status(http.StatusBadRequest).JSON().Object()
		body.Keys().Length().Equal(2)
		body.ValueEqual("code", 3)
		body.ValueEqual("message", "service_id 'invalid-uuid' is not a UUID")
	})

	t.Run("list plugins by route", func(t *testing.T) {
		body := c.GET("/v1/plugins").WithQuery("route_id", routeID1).
			Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(1)
		var gotIDs []string
		for _, item := range items.Iter() {
			gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
		}
		require.ElementsMatch(t, []string{pluginID6}, gotIDs)

		body = c.GET("/v1/plugins").WithQuery("route_id", routeID2).
			Expect().Status(http.StatusOK).JSON().Object()
		items = body.Value("items").Array()
		items.Length().Equal(2)
		gotIDs = nil
		for _, item := range items.Iter() {
			gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
		}
		require.ElementsMatch(t, []string{
			pluginID7,
			pluginID8,
		}, gotIDs)
	})

	t.Run("list plugins by route - no plugins associated with route", func(t *testing.T) {
		body := c.GET("/v1/plugins").WithQuery("route_id", routeID3).
			Expect().Status(http.StatusOK).JSON().Object()
		body.Empty()
	})

	t.Run("list plugins by route - invalid route UUID", func(t *testing.T) {
		body := c.GET("/v1/plugins").WithQuery("route_id", "invalid-uuid").
			Expect().Status(http.StatusBadRequest).JSON().Object()
		body.Keys().Length().Equal(2)
		body.ValueEqual("code", 3)
		body.ValueEqual("message", "route_id 'invalid-uuid' is not a UUID")
	})

	t.Run("list plugins by route and service - invalid request", func(t *testing.T) {
		body := c.GET("/v1/plugins").
			WithQuery("service_id", serviceID2).
			WithQuery("route_id", routeID1).
			Expect().Status(http.StatusBadRequest).JSON().Object()
		body.Keys().Length().Equal(2)
		body.ValueEqual("code", 3)
		body.ValueEqual("message", "service_id and route_id are mutually exclusive")
	})
	t.Run("list plugins by route and consumer - invalid request", func(t *testing.T) {
		body := c.GET("/v1/plugins").
			WithQuery("consumer_id", uuid.NewString()).
			WithQuery("route_id", routeID1).
			Expect().Status(http.StatusBadRequest).JSON().Object()
		body.Keys().Length().Equal(2)
		body.ValueEqual("code", 3)
		body.ValueEqual("message", "route_id and consumer_id are mutually exclusive")
	})
	t.Run("list plugins by service and consumer - invalid request", func(t *testing.T) {
		body := c.GET("/v1/plugins").
			WithQuery("consumer_id", uuid.NewString()).
			WithQuery("service_id", serviceID2).
			Expect().Status(http.StatusBadRequest).JSON().Object()
		body.Keys().Length().Equal(2)
		body.ValueEqual("code", 3)
		body.ValueEqual("message", "service_id and consumer_id are mutually exclusive")
	})
	t.Run("list returns multiple plugins with paging", func(t *testing.T) {
		// Get First Page
		body := c.GET("/v1/plugins").
			WithQuery("page.size", "4").
			WithQuery("page.number", "1").
			Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(4)
		var gotIDs []string
		for _, item := range items.Iter() {
			gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
		}
		body.Value("page").Object().Value("total_count").Number().Equal(8)
		body.Value("page").Object().Value("next_page_num").Number().Equal(2)
		// Get second page
		body = c.GET("/v1/plugins").
			WithQuery("page.size", "4").
			WithQuery("page.number", "2").
			Expect().Status(http.StatusOK).JSON().Object()
		items = body.Value("items").Array()
		items.Length().Equal(4)
		for _, item := range items.Iter() {
			gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
		}
		body.Value("page").Object().Value("total_count").Number().Equal(8)
		body.Value("page").Object().NotContainsKey("next_page_num")
		require.ElementsMatch(t, []string{
			pluginID1,
			pluginID2,
			pluginID3,
			pluginID4,
			pluginID5,
			pluginID6,
			pluginID7,
			pluginID8,
		}, gotIDs)
	})
}

func TestPluginListByConsumer(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)

	consumer := goodConsumer()
	res := c.POST("/v1/consumers").WithJSON(consumer).Expect()
	res.Status(http.StatusCreated)
	body := res.JSON().Path("$.item").Object()
	consumerID := body.Value("id").String().Raw()
	var config structpb.Struct
	configString := `{"header_name": "Kong-Request-ID", "generator": "uuid#counter", "echo_downstream": true }`
	require.Nil(t, json.ProtoJSONUnmarshal([]byte(configString), &config))
	plugin := &v1.Plugin{
		Name:      "correlation-id",
		Protocols: []string{"http", "https"},
		Consumer: &v1.Consumer{
			Id: consumerID,
		},
		Config: &config,
	}
	pluginBytes, err := json.ProtoJSONMarshal(plugin)
	require.Nil(t, err)
	res = c.POST("/v1/plugins").WithBytes(pluginBytes).Expect()
	res.Status(http.StatusCreated)
	body = res.JSON().Path("$.item").Object()
	PluginIDOne := body.Value("id").String().Raw()

	configString = `{"allow": ["10.10.10.10"]}`
	require.Nil(t, json.ProtoJSONUnmarshal([]byte(configString), &config))

	plugin = &v1.Plugin{
		Name:      "ip-restriction",
		Protocols: []string{"http", "https"},
		Consumer: &v1.Consumer{
			Id: consumerID,
		},
		Config: &config,
	}
	pluginBytes, err = json.ProtoJSONMarshal(plugin)
	require.Nil(t, err)
	res = c.POST("/v1/plugins").WithBytes(pluginBytes).Expect()
	res.Status(http.StatusCreated)
	body = res.JSON().Path("$.item").Object()
	PluginIDTwo := body.Value("id").String().Raw()

	// create one more to ensure filter works
	plugin = &v1.Plugin{
		Name:      "request-size-limiting",
		Enabled:   wrapperspb.Bool(true),
		Protocols: []string{"http", "https"},
	}
	pluginBytes, err = json.ProtoJSONMarshal(plugin)
	require.Nil(t, err)
	res = c.POST("/v1/plugins").WithBytes(pluginBytes).Expect()
	res.Status(http.StatusCreated)

	body = c.GET("/v1/plugins").WithQuery("consumer_id", consumerID).
		Expect().Status(http.StatusOK).JSON().Object()
	items := body.Value("items").Array()
	items.Length().Equal(2)
	itemArr := items.Iter()
	gotIDs := make([]string, 0, len(itemArr))
	for _, item := range itemArr {
		gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
	}
	require.ElementsMatch(t, []string{PluginIDOne, PluginIDTwo}, gotIDs)
}

func TestAvailablePlugins(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()

	c := httpexpect.New(t, s.URL)
	body := c.GET("/v1/available-plugins").Expect().Status(http.StatusOK).JSON().Object()
	names := body.Value("names").Array().Iter()
	actual := make([]string, 0, len(names))
	for _, item := range names {
		actual = append(actual, item.String().Raw())
	}

	assert.Equal(t, []string{
		"acl",
		"acme",
		"aws-lambda",
		"azure-functions",
		"basic-auth",
		"bot-detection",
		"correlation-id",
		"cors",
		"datadog",
		"file-log",
		"grpc-gateway",
		"grpc-web",
		"hmac-auth",
		"http-log",
		"ip-restriction",
		"jwt",
		"key-auth",
		"ldap-auth",
		"loggly",
		"opentelemetry",
		"post-function",
		"pre-function",
		"prometheus",
		"proxy-cache",
		"rate-limiting",
		"request-size-limiting",
		"request-termination",
		"request-transformer",
		"response-ratelimiting",
		"response-transformer",
		"session",
		"statsd",
		"syslog",
		"tcp-log",
		"udp-log",
		"zipkin",
	}, actual)
}

func TestAvailableAndCustomPlugins(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()

	c := httpexpect.New(t, s.URL)
	// Add custom plugin
	pluginSchemaBytes, err := json.ProtoJSONMarshal(goodPluginSchema("abc-lua-plugin", "string"))
	assert.NoError(t, err)
	res := c.POST("/v1/plugin-schemas").WithBytes(pluginSchemaBytes).Expect()
	res.Status(http.StatusCreated)
	res.Header("grpc-metadata-koko-status-code").Empty()
	body := res.JSON().Path("$.item").Object()
	validatePluginSchema("abc-lua-plugin", "string", body)

	// get available-plugins
	body = c.GET("/v1/available-plugins").Expect().Status(http.StatusOK).JSON().Object()
	names := body.Value("names").Array().Iter()
	actual := make([]string, 0, len(names))
	for _, item := range names {
		actual = append(actual, item.String().Raw())
	}

	assert.Equal(t, []string{
		"abc-lua-plugin",
		"acl",
		"acme",
		"aws-lambda",
		"azure-functions",
		"basic-auth",
		"bot-detection",
		"correlation-id",
		"cors",
		"datadog",
		"file-log",
		"grpc-gateway",
		"grpc-web",
		"hmac-auth",
		"http-log",
		"ip-restriction",
		"jwt",
		"key-auth",
		"ldap-auth",
		"loggly",
		"opentelemetry",
		"post-function",
		"pre-function",
		"prometheus",
		"proxy-cache",
		"rate-limiting",
		"request-size-limiting",
		"request-termination",
		"request-transformer",
		"response-ratelimiting",
		"response-transformer",
		"session",
		"statsd",
		"syslog",
		"tcp-log",
		"udp-log",
		"zipkin",
	}, actual)
}
