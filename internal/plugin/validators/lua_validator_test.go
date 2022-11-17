package validators

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	grpcModel "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/log"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/kong/koko/internal/model/json/validation/typedefs"
	"github.com/kong/koko/internal/plugin"
	"github.com/kong/koko/internal/plugin/validators/badtestdata"
	"github.com/kong/koko/internal/plugin/validators/testdata"
	"github.com/kong/koko/internal/resource"
	serverUtil "github.com/kong/koko/internal/server/util"
	"github.com/kong/koko/internal/store"
	"github.com/kong/koko/internal/test/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

//go:embed testdata/schemas/*
var badSchemaFS embed.FS

// goodValidator is loaded at init.
// This is an optimization to speed up tests.
var goodValidator *LuaValidator

const pluginSchemaFormat = `return {
	name = "%s",
	fields = {
		{ config = {
				type = "record",
				fields = {
					{ field = { type = "string", default = "populated" } },
					{ field_2 = { type = "boolean", default = true } },
					{ field_3 = { type = "string" } },
				}
			}
		}
	}
}`

func init() {
	var err error
	if goodValidator, err = getGoodValidatorWithStoreLoader(nil); err != nil {
		panic(err)
	}
}

func getGoodValidatorWithStoreLoader(storeLoader serverUtil.StoreLoader) (*LuaValidator, error) {
	v, err := NewLuaValidator(Opts{Logger: log.Logger, StoreLoader: storeLoader})
	if err != nil {
		return nil, err
	}
	if err := v.LoadSchemasFromEmbed(plugin.Schemas, "schemas"); err != nil {
		return nil, err
	}

	// PluginSchema may already be registered; safe to ignore error
	_ = model.RegisterType("plugin_schema", &grpcModel.PluginSchema{}, func() model.Object {
		return resource.NewPluginSchema()
	})
	resource.SetValidator(v)

	return v, nil
}

func setupStoreLoader(t *testing.T) serverUtil.StoreLoader {
	p, err := util.GetPersister(t)
	require.NoError(t, err)
	objectStore := store.New(p, log.Logger).ForCluster(store.DefaultCluster)
	return serverUtil.DefaultStoreLoader{
		Store: objectStore,
	}
}

func goodPluginSchema(name string) string {
	return fmt.Sprintf(pluginSchemaFormat, name)
}

func insertPluginSchema(t *testing.T, name string, schema string, storeLoader serverUtil.StoreLoader) error {
	pluginSchema := resource.NewPluginSchema()
	pluginSchema.PluginSchema.Name = name
	pluginSchema.PluginSchema.LuaSchema = schema

	db, err := storeLoader.Load(context.Background(), &grpcModel.RequestCluster{Id: store.DefaultCluster})
	require.NoError(t, err)
	return db.Create(context.Background(), pluginSchema)
}

func getValidContext() context.Context {
	return context.WithValue(context.Background(), serverUtil.ContextKeyCluster,
		&grpcModel.RequestCluster{Id: store.DefaultCluster})
}

func getValidContextWithClusterID(cluster string) context.Context {
	return context.WithValue(context.Background(), serverUtil.ContextKeyCluster,
		&grpcModel.RequestCluster{Id: cluster})
}

func TestNewLuaValidator(t *testing.T) {
	t.Run("instantiate validator using inject file system", func(t *testing.T) {
		validator, err := NewLuaValidator(Opts{
			Logger:   log.Logger,
			InjectFS: &testdata.LuaTree,
		})
		require.Nil(t, err)
		require.NotNil(t, validator)
	})
	t.Run("fail to instantiate validator using inject file system", func(t *testing.T) {
		_, err := NewLuaValidator(Opts{
			Logger:   log.Logger,
			InjectFS: &badSchemaFS,
		})
		require.NotNil(t, err)
		require.Contains(t, err.Error(), "file system must contain 'lua-tree/share/lua/5.1'")
	})
}

func TestLuaValidator_LoadPatch(t *testing.T) {
	validator, err := NewLuaValidator(Opts{
		Logger:   log.Logger,
		InjectFS: &testdata.LuaTree,
	})
	require.NoError(t, err)

	// let's use a "standard" module
	v, err := validator.goksV.Execute(`return require "version"`)
	require.NoError(t, err)
	require.Equal(t, "0.0.1", v)

	// now load a patch
	err = validator.LoadPatch("bump_version")
	require.NoError(t, err)

	// and verify that any new use gets the patched content
	v, err = validator.goksV.Execute(`return require "version"`)
	require.NoError(t, err)
	require.Equal(t, "0.1-extra-plus", v)
}

func TestLoadSchemasFromEmbed(t *testing.T) {
	validator, err := NewLuaValidator(Opts{Logger: log.Logger})
	require.Nil(t, err)
	t.Run("errors when dir doesn't exist", func(t *testing.T) {
		err = validator.LoadSchemasFromEmbed(plugin.Schemas, "does-not-exist")
		require.NotNil(t, err)
		require.IsType(t, &fs.PathError{}, err)
	})
	t.Run("errors when embed.Fs is nil", func(t *testing.T) {
		err = validator.LoadSchemasFromEmbed(embed.FS{}, "schemas")
		require.NotNil(t, err)
	})
	t.Run("loads a value schemas directory", func(t *testing.T) {
		err = validator.LoadSchemasFromEmbed(plugin.Schemas, "schemas")
		require.Nil(t, err)
	})
	t.Run("loading bad schema fails", func(t *testing.T) {
		err = validator.LoadSchemasFromEmbed(badSchemaFS, "testdata/schemas")
		require.NotNil(t, err)
		require.IsType(t, &lua.ApiError{}, err)
	})
}

func TestProcessAutoFields(t *testing.T) {
	validator := goodValidator
	t.Run("injects default fields for a plugin", func(t *testing.T) {
		config, err := structpb.NewStruct(map[string]interface{}{
			"second": 42,
		})
		require.Nil(t, err)
		plugin := &grpcModel.Plugin{
			Name:   "rate-limiting",
			Config: config,
		}
		err = validator.ProcessDefaults(context.Background(), plugin)
		require.Nil(t, err)
		require.NotPanics(t, func() {
			uuid.MustParse(plugin.Id)
		})
		require.LessOrEqual(t, plugin.CreatedAt, int32(time.Now().Unix()))
		require.True(t, plugin.Enabled.Value)
		require.ElementsMatch(t, plugin.Protocols,
			[]string{"http", "https", "grpc", "grpcs"},
		)
		processConfig := plugin.Config.AsMap()
		expectedConfig := map[string]interface{}{
			"second":              float64(42),
			"day":                 nil,
			"hour":                nil,
			"fault_tolerant":      true,
			"header_name":         nil,
			"hide_client_headers": false,
			"limit_by":            "consumer",
			"minute":              nil,
			"redis_ssl":           false,
			"redis_host":          nil,
			"path":                nil,
			"month":               nil,
			"policy":              "local",
			"redis_port":          float64(6379),
			"redis_ssl_verify":    false,
			"redis_database":      float64(0),
			"redis_username":      nil,
			"redis_password":      nil,
			"year":                nil,
			"redis_server_name":   nil,
			"redis_timeout":       float64(2000),
			"error_code":          429,
			"error_message":       "API rate limit exceeded",
		}
		require.Equal(t, expectedConfig, processConfig)
	})
	t.Run("injects default fields for a non bundled plugin using a plugin schema ", func(t *testing.T) {
		storeLoader := setupStoreLoader(t)
		require.NotNil(t, storeLoader)

		validator, err := NewLuaValidator(Opts{
			Logger:      log.Logger,
			StoreLoader: storeLoader,
		})
		require.NoError(t, err)
		require.NotNil(t, validator)

		resource.SetValidator(validator)

		err = insertPluginSchema(t, "non-bundled", goodPluginSchema("non-bundled"), storeLoader)
		assert.NoError(t, err)

		plugin := &grpcModel.Plugin{
			Name:      "non-bundled",
			Protocols: []string{"http", "https"},
			Config:    &structpb.Struct{},
		}
		err = validator.ProcessDefaults(getValidContext(), plugin)
		require.NoError(t, err)
		require.NotPanics(t, func() {
			uuid.MustParse(plugin.Id)
		})
		processConfig := plugin.Config.AsMap()
		expectedConfig := map[string]interface{}{
			"field":   "populated",
			"field_2": true,
			"field_3": nil,
		}
		require.Equal(t, expectedConfig, processConfig)
	})
	t.Run("fails to process defaults with plugin schema which has not been loaded", func(t *testing.T) {
		validator, err := NewLuaValidator(Opts{
			Logger: log.Logger,
		})
		assert.NoError(t, err)
		require.NotNil(t, validator)

		plugin := &grpcModel.Plugin{
			Name:      "non-bundled-not-loaded",
			Protocols: []string{"http", "https"},
			Config:    &structpb.Struct{},
		}
		err = validator.ProcessDefaults(getValidContext(), plugin)
		require.NotNil(t, err)
		require.Contains(t, err.Error(), "unmarshal JSON:")
	})
}

func TestValidate(t *testing.T) {
	validator := goodValidator
	t.Run("invalid JSON returned", func(t *testing.T) {
		badValidator, err := NewLuaValidator(Opts{
			Logger:   log.Logger,
			InjectFS: &badtestdata.BadLuaTree,
		})
		require.NoError(t, err)
		require.NoError(t, badValidator.LoadSchemasFromEmbed(plugin.Schemas, "schemas"))
		require.NoError(t, badValidator.LoadPatch("validate-invalid-json"))

		err = badValidator.Validate(context.Background(), &grpcModel.Plugin{
			Name:      "prometheus",
			Protocols: []string{"http", "https"},
			Enabled:   wrapperspb.Bool(true),
		})
		require.IsType(t, validation.Error{}, err)
		vErr := err.(validation.Error)
		require.Len(t, vErr.Errs, 1)
		require.Len(t, vErr.Errs[0].Messages, 1)
		require.Equal(
			t,
			"(prometheus) unknown plugin validation error, please file a bug with Kong Inc",
			vErr.Errs[0].Messages[0],
		)
	})
	t.Run("test with entity errors", func(t *testing.T) {
		config, err := structpb.NewStruct(map[string]interface{}{
			"policy": "redis",
		})
		require.Nil(t, err)
		err = validator.Validate(context.Background(), &grpcModel.Plugin{
			Name:      "rate-limiting",
			Config:    config,
			Protocols: []string{"http", "https"},
			Enabled:   wrapperspb.Bool(true),
		})
		validationErr, ok := err.(validation.Error)
		require.True(t, ok)
		expected := []*grpcModel.ErrorDetail{
			{
				Type: grpcModel.ErrorType_ERROR_TYPE_ENTITY,
				Messages: []string{
					"at least one of these fields must be non-empty: 'config.second'" +
						", 'config.minute', 'config.hour', 'config.day', 'config.month', 'config.year'",
					"failed conditional validation given value of field 'config.policy'",
				},
			},
			{
				Type:     grpcModel.ErrorType_ERROR_TYPE_FIELD,
				Field:    "config.redis_host",
				Messages: []string{"required field missing"},
			},
		}
		require.ElementsMatch(t, expected, validationErr.Errs)
	})
	t.Run("plugin does not exist", func(t *testing.T) {
		err := validator.Validate(context.Background(), &grpcModel.Plugin{
			Name: "no-auth",
		})
		require.NotNil(t, err)
		validationErr, ok := err.(validation.Error)
		require.True(t, ok)
		expected := []*grpcModel.ErrorDetail{
			{
				Type:     grpcModel.ErrorType_ERROR_TYPE_FIELD,
				Field:    "name",
				Messages: []string{"plugin(no-auth) does not exist"},
			},
		}
		require.Equal(t, expected, validationErr.Errs)
	})
	t.Run("validates nested config structs", func(t *testing.T) {
		var config structpb.Struct
		configString := `{"add":{"headers":["nokey"]}}`
		require.Nil(t, json.ProtoJSONUnmarshal([]byte(configString), &config))
		err := validator.Validate(context.Background(), &grpcModel.Plugin{
			Name:      "request-transformer",
			Config:    &config,
			Protocols: []string{"http", "https"},
			Enabled:   wrapperspb.Bool(true),
		})
		require.NotNil(t, err)
		validationErr, ok := err.(validation.Error)
		require.True(t, ok)
		expected := []*grpcModel.ErrorDetail{
			{
				Type:     grpcModel.ErrorType_ERROR_TYPE_FIELD,
				Field:    "config.add.headers[0]",
				Messages: []string{"invalid value: nokey"},
			},
		}
		require.Equal(t, expected, validationErr.Errs)
	})
	t.Run("prometheus plugin schema", func(t *testing.T) {
		p := &grpcModel.Plugin{
			Name:      "prometheus",
			Protocols: []string{"http", "https"},
			Enabled:   wrapperspb.Bool(true),
		}
		require.Nil(t, validator.ProcessDefaults(context.Background(), p))
		err := validator.Validate(context.Background(), p)
		require.Nil(t, err)
	})
	t.Run("serverless plugin schema", func(t *testing.T) {
		config, err := structpb.NewStruct(map[string]interface{}{
			"access": []interface{}{
				`
   -- Get list of request headers
   local custom_auth = kong.request.get_header("x-custom-auth")

   -- Terminate request early if our custom authentication header
   -- does not exist
   if not custom_auth then
     return kong.response.exit(401, "Invalid Credentials")
   end

   -- Remove custom authentication header from request
   kong.service.request.clear_header('x-custom-auth')
`,
				`kong.log.err('Hi there Access!')`,
			},
			"header_filter": []interface{}{
				`kong.log.err('Hi there header filter!')`,
			},
			"log": []interface{}{
				`kong.log.err('Hi there Log!')`,
			},
		})
		require.Nil(t, err)
		p := &grpcModel.Plugin{
			Name:      "pre-function",
			Protocols: []string{"http", "https"},
			Enabled:   wrapperspb.Bool(true),
			Config:    config,
		}
		require.Nil(t, validator.ProcessDefaults(context.Background(), p))
		err = validator.Validate(context.Background(), p)
		require.Nil(t, err)

		p = &grpcModel.Plugin{
			Name:      "post-function",
			Protocols: []string{"http", "https"},
			Enabled:   wrapperspb.Bool(true),
			Config:    config,
		}
		require.Nil(t, validator.ProcessDefaults(context.Background(), p))
		err = validator.Validate(context.Background(), p)
		require.Nil(t, err)
	})
	t.Run("serverless plugin schema accepts invalid lua-code", func(t *testing.T) {
		config, err := structpb.NewStruct(map[string]interface{}{
			"log": []interface{}{
				"this is not valid lua !",
			},
		})
		require.Nil(t, err)
		p := &grpcModel.Plugin{
			Name:      "pre-function",
			Protocols: []string{"http", "https"},
			Enabled:   wrapperspb.Bool(true),
			Config:    config,
		}
		require.Nil(t, validator.ProcessDefaults(context.Background(), p))
		err = validator.Validate(context.Background(), p)
		require.Nil(t, err)
	})
	t.Run("aws-lambda plugin errors out when custom_entity_check fails", func(t *testing.T) {
		var config structpb.Struct
		configString := `{"proxy_url": "https://my-proxy-server:3128"}`
		require.Nil(t, json.ProtoJSONUnmarshal([]byte(configString), &config))
		p := &grpcModel.Plugin{
			Name:      "aws-lambda",
			Protocols: []string{"http", "https"},
			Enabled:   wrapperspb.Bool(true),
			Config:    &config,
		}
		require.Nil(t, validator.ProcessDefaults(context.Background(), p))
		err := validator.Validate(context.Background(), p)
		validationErr, ok := err.(validation.Error)
		require.True(t, ok)
		expected := []*grpcModel.ErrorDetail{
			{
				Type:     grpcModel.ErrorType_ERROR_TYPE_ENTITY,
				Messages: []string{"proxy_url scheme must be http"},
			},
		}
		require.Equal(t, expected, validationErr.Errs)
	})
	t.Run("ensure all protocols can be assigned", func(t *testing.T) {
		protocols := []string{
			// http
			"http",
			"https",
			"grpc",
			"grpcs",

			// stream
			"tcp",
			"tls",
			"udp",
		}
		for _, protocol := range protocols {
			p := &grpcModel.Plugin{
				Name:      "prometheus",
				Protocols: []string{protocol},
				Enabled:   wrapperspb.Bool(true),
			}
			err := validator.Validate(context.Background(), p)
			require.Nil(t, err)
		}
	})
	t.Run("ensure stream protocols fail plugins which only allow for http", func(t *testing.T) {
		protocols := []string{
			"tcp",
			"tls",
			"udp",
		}
		config, err := structpb.NewStruct(map[string]interface{}{
			"second": 42,
		})
		require.Nil(t, err)
		for _, protocol := range protocols {
			p := &grpcModel.Plugin{
				Name:      "rate-limiting",
				Protocols: []string{protocol},
				Enabled:   wrapperspb.Bool(true),
				Config:    config,
			}
			err := validator.Validate(context.Background(), p)
			require.NotNil(t, err)
			validationErr, ok := err.(validation.Error)
			require.True(t, ok)
			expected := []*grpcModel.ErrorDetail{
				{
					Type:     grpcModel.ErrorType_ERROR_TYPE_FIELD,
					Field:    "protocols[0]",
					Messages: []string{"expected one of: grpc, grpcs, http, https"},
				},
			}
			require.Equal(t, expected, validationErr.Errs)
		}
	})
	t.Run("validate different policies for rate-limiting plugin", func(t *testing.T) {
		policies := []struct {
			config       string
			wantsErr     bool
			expectedErrs []*grpcModel.ErrorDetail
		}{
			{
				config: `{
					"second": 42,
					"policy": "local"
				}`,
			},
			{
				config: `{
					"second": 42,
					"policy": "redis",
					"redis_host": "localhost"
				}`,
			},
			{
				config: `{
					"second": 42,
					"policy": "cluster"
				}`,
				wantsErr: true,
				expectedErrs: []*grpcModel.ErrorDetail{
					{
						Type:  grpcModel.ErrorType_ERROR_TYPE_FIELD,
						Field: "config.policy",
						Messages: []string{
							"expected one of: local, redis",
						},
					},
				},
			},
		}

		for _, policy := range policies {
			var config structpb.Struct
			require.Nil(t, json.ProtoJSONUnmarshal([]byte(policy.config), &config))
			p := &grpcModel.Plugin{
				Name:      "rate-limiting",
				Protocols: []string{"http", "https"},
				Enabled:   wrapperspb.Bool(true),
				Config:    &config,
			}
			err := validator.Validate(context.Background(), p)
			if policy.wantsErr {
				require.NotNil(t, err)
				validationErr, ok := err.(validation.Error)
				require.True(t, ok)
				require.Equal(t, policy.expectedErrs, validationErr.Errs)
			} else {
				require.Nil(t, err)
			}
		}
	})
	t.Run("validate different 'path' for rate-limiting plugin", func(t *testing.T) {
		policies := []struct {
			config       string
			wantsErr     bool
			expectedErrs []*grpcModel.ErrorDetail
		}{
			{
				config: `{
					"second": 42,
					"policy": "local",
					"path": "/"
				}`,
			},
			{
				config: `{
					"second": 42,
					"policy": "local",
					"path": "/path"
				}`,
			},
			{
				config: `{
					"second": 42,
					"policy": "local",
					"path": "/path/test"
				}`,
			},
			{
				config: `{
					"second": 42,
					"policy": "local",
					"path": ""
				}`,
				wantsErr: true,
				expectedErrs: []*grpcModel.ErrorDetail{
					{
						Type:  grpcModel.ErrorType_ERROR_TYPE_FIELD,
						Field: "config.path",
						Messages: []string{
							"length must be at least 1",
						},
					},
				},
			},
			{
				config: `{
					"second": 42,
					"policy": "local",
					"path": "path"
				}`,
				wantsErr: true,
				expectedErrs: []*grpcModel.ErrorDetail{
					{
						Type:  grpcModel.ErrorType_ERROR_TYPE_FIELD,
						Field: "config.path",
						Messages: []string{
							"should start with: /",
						},
					},
				},
			},
			{
				config: `{
					"second": 42,
					"policy": "local",
					"path": "/path/200?"
				}`,
				wantsErr: true,
				expectedErrs: []*grpcModel.ErrorDetail{
					{
						Type:  grpcModel.ErrorType_ERROR_TYPE_FIELD,
						Field: "config.path",
						Messages: []string{
							"invalid path: '/path/200?' (characters outside of the reserved list of RFC 3986 found)",
						},
					},
				},
			},
			{
				config: `{
					"second": 42,
					"policy": "local",
					"path": "/some%20words%"
				}`,
				wantsErr: true,
				expectedErrs: []*grpcModel.ErrorDetail{
					{
						Type:  grpcModel.ErrorType_ERROR_TYPE_FIELD,
						Field: "config.path",
						Messages: []string{
							"invalid url-encoded value: '%'",
						},
					},
				},
			},
			{
				config: `{
					"second": 42,
					"policy": "local",
					"path": "/foo/bar//"
				}`,
				wantsErr: true,
				expectedErrs: []*grpcModel.ErrorDetail{
					{
						Type:  grpcModel.ErrorType_ERROR_TYPE_FIELD,
						Field: "config.path",
						Messages: []string{
							"must not have empty segments",
						},
					},
				},
			},
		}

		for _, policy := range policies {
			var config structpb.Struct
			require.Nil(t, json.ProtoJSONUnmarshal([]byte(policy.config), &config))
			p := &grpcModel.Plugin{
				Name:      "rate-limiting",
				Protocols: []string{"http", "https"},
				Enabled:   wrapperspb.Bool(true),
				Config:    &config,
			}
			err := validator.Validate(context.Background(), p)
			if policy.wantsErr {
				require.NotNil(t, err)
				validationErr, ok := err.(validation.Error)
				require.True(t, ok)
				require.Equal(t, policy.expectedErrs, validationErr.Errs)
			} else {
				require.Nil(t, err)
			}
		}
	})
	t.Run("validate different policies for response-ratelimiting plugin", func(t *testing.T) {
		policies := []struct {
			config       string
			wantsErr     bool
			expectedErrs []*grpcModel.ErrorDetail
		}{
			{
				config: `{
					"limits": {
						"sms": {
							"second": 42
						}
					},
					"policy": "local"
				}`,
			},
			{
				config: `{
					"limits": {
						"sms": {
							"second": 42
						}
					},
					"policy": "redis",
					"redis_host": "localhost"
				}`,
			},
			{
				config: `{
					"limits": {
						"sms": {
							"second": 42
						}
					},
					"policy": "cluster"
				}`,
				wantsErr: true,
				expectedErrs: []*grpcModel.ErrorDetail{
					{
						Type:  grpcModel.ErrorType_ERROR_TYPE_FIELD,
						Field: "config.policy",
						Messages: []string{
							"expected one of: local, redis",
						},
					},
				},
			},
		}

		for _, policy := range policies {
			var config structpb.Struct
			require.Nil(t, json.ProtoJSONUnmarshal([]byte(policy.config), &config))
			p := &grpcModel.Plugin{
				Name:      "response-ratelimiting",
				Protocols: []string{"http", "https"},
				Enabled:   wrapperspb.Bool(true),
				Config:    &config,
			}
			err := validator.Validate(context.Background(), p)
			if policy.wantsErr {
				require.NotNil(t, err)
				validationErr, ok := err.(validation.Error)
				require.True(t, ok)
				require.Equal(t, policy.expectedErrs, validationErr.Errs)
			} else {
				require.Nil(t, err)
			}
		}
	})
	t.Run("validate different strategies and shared dict names for proxy-cache", func(t *testing.T) {
		strategies := []struct {
			config       string
			wantsErr     bool
			expectedErrs []*grpcModel.ErrorDetail
		}{
			{
				config: `{
					"strategy": "memory"
				}`,
			},
			{
				config: `{
					"strategy": "memory",
					"memory": {
						"dictionary_name": "not_the_default_shared_dict"
					}
				}`,
			},
			{
				config: `{
					"strategy": "memory",
					"memory": {
						"dictionary_name": ""
					}
				}`,
				wantsErr: true,
				expectedErrs: []*grpcModel.ErrorDetail{
					{
						Type:  grpcModel.ErrorType_ERROR_TYPE_FIELD,
						Field: "config.memory.dictionary_name",
						Messages: []string{
							"length must be at least 1",
						},
					},
				},
			},
		}

		for _, strategy := range strategies {
			var config structpb.Struct
			require.Nil(t, json.ProtoJSONUnmarshal([]byte(strategy.config), &config))
			p := &grpcModel.Plugin{
				Name:      "proxy-cache",
				Protocols: []string{"http", "https"},
				Enabled:   wrapperspb.Bool(true),
				Config:    &config,
			}
			err := validator.Validate(context.Background(), p)
			if strategy.wantsErr {
				require.NotNil(t, err)
				validationErr, ok := err.(validation.Error)
				require.True(t, ok)
				require.Equal(t, strategy.expectedErrs, validationErr.Errs)
			} else {
				require.Nil(t, err)
			}
		}
	})

	t.Run("validate opentelemetry configuration", func(t *testing.T) {
		tt := []struct {
			name         string
			config       string
			wantsErr     bool
			expectedErrs []*grpcModel.ErrorDetail
		}{
			{
				name: "good config validates",
				config: `{
					"endpoint": "http://example.dev",
					"resource_attributes": {
						"service.name": "kong_oss",
						"os.version": "debian"
					}
				}`,
			},
			{
				name: "bad value type doesn't validates",
				config: `{
					"endpoint": "http://example.dev",
					"resource_attributes": {
						"service.name": "kong_oss",
						"os.version": 10
					}
				}`,
				wantsErr: true,
				expectedErrs: []*grpcModel.ErrorDetail{
					{
						Type:  grpcModel.ErrorType_ERROR_TYPE_FIELD,
						Field: "config.resource_attributes",
						Messages: []string{
							"expected a string",
						},
					},
				},
			},
			{
				name: "empty value doesn't validates",
				config: `{
					"endpoint": "http://example.dev",
					"resource_attributes": {
						"service.name": "kong_oss",
						"os.version": ""
					}
				}`,
				wantsErr: true,
				expectedErrs: []*grpcModel.ErrorDetail{
					{
						Type:  grpcModel.ErrorType_ERROR_TYPE_FIELD,
						Field: "config.resource_attributes",
						Messages: []string{
							"length must be at least 1",
						},
					},
				},
			},
			{
				name: "no resource_attributes validates",
				config: `{
					"endpoint": "http://example.dev"
				}`,
				wantsErr: false,
			},
		}
		for _, plugin := range tt {
			var config structpb.Struct
			require.Nil(t, json.ProtoJSONUnmarshal([]byte(plugin.config), &config))
			p := &grpcModel.Plugin{
				Name:      "opentelemetry",
				Protocols: []string{"http", "https"},
				Enabled:   wrapperspb.Bool(true),
				Config:    &config,
			}
			err := validator.Validate(context.Background(), p)
			if plugin.wantsErr {
				require.NotNil(t, err)
				validationErr, ok := err.(validation.Error)
				require.True(t, ok)
				require.ElementsMatch(t, plugin.expectedErrs, validationErr.Errs)
			} else {
				require.Nil(t, err)
			}
		}
	})

	t.Run("test plugins' typedefs validators", func(t *testing.T) {
		tt := []struct {
			name         string
			config       string
			wantsErr     bool
			expectedErrs []*grpcModel.ErrorDetail
		}{
			{
				// test typedefs.sni
				name: "rate-limiting",
				config: `{
					"second": 42,
					"policy": "redis",
					"redis_host": "localhost",
					"redis_server_name": "192.0.2.1"
				}`,
				wantsErr: true,
				expectedErrs: []*grpcModel.ErrorDetail{
					{
						Type:  grpcModel.ErrorType_ERROR_TYPE_FIELD,
						Field: "config.redis_server_name",
						Messages: []string{
							"must not be an IP",
						},
					},
				},
			},
			{
				// test typedefs.port
				name: "response-ratelimiting",
				config: `{
					"limits": {
						"sms": {
							"second": 42
						}
					},
					"policy": "redis",
					"redis_host": "localhost",
					"redis_port": 99999999
				}`,
				wantsErr: true,
				expectedErrs: []*grpcModel.ErrorDetail{
					{
						Type: grpcModel.ErrorType_ERROR_TYPE_ENTITY,
						Messages: []string{
							"failed conditional validation given value of field 'config.policy'",
						},
					},
					{
						Type:  grpcModel.ErrorType_ERROR_TYPE_FIELD,
						Field: "config.redis_port",
						Messages: []string{
							"value should be between 0 and 65535",
						},
					},
				},
			},
			{
				// test typedefs.ip_or_cidr
				name: "ip-restriction",
				config: `{
					"allow": [
						"1.2.3.4",
						"192.168.129.23/17"
					]
				}`,
				wantsErr: false,
			},
			{
				// test typedefs.ip_or_cidr
				name: "ip-restriction",
				config: `{
					"allow": [
						"1.2.3.4.4",
						"1.2.3.4",
						"192.168.129.23/40",
						"asd",
						"::1"
					]
				}`,
				wantsErr: true,
				expectedErrs: []*grpcModel.ErrorDetail{
					{
						Type:  grpcModel.ErrorType_ERROR_TYPE_FIELD,
						Field: "config.allow[0]",
						Messages: []string{
							"invalid ip or cidr range: '1.2.3.4.4'",
						},
					},
					{
						Type:  grpcModel.ErrorType_ERROR_TYPE_FIELD,
						Field: "config.allow[1]",
						Messages: []string{
							"[3] = invalid ip or cidr range: '192.168.129.23/40'",
						},
					},
					{
						Type:  grpcModel.ErrorType_ERROR_TYPE_FIELD,
						Field: "config.allow[2]",
						Messages: []string{
							"[4] = invalid ip or cidr range: 'asd'",
						},
					},
				},
			},
			{
				name: "acme",
				config: `{
					"account_email": "example@example.com"
				}`,
				wantsErr: false,
			},
			{
				// test typedefs.url
				name: "acme",
				config: `{
					"account_email": "example-example",
					"api_uri": "acme"
				}`,
				wantsErr: true,
				expectedErrs: []*grpcModel.ErrorDetail{
					{
						Type:  grpcModel.ErrorType_ERROR_TYPE_FIELD,
						Field: "config.account_email",
						Messages: []string{
							"invalid value: example-example",
						},
					},
					{
						Type:  grpcModel.ErrorType_ERROR_TYPE_FIELD,
						Field: "config.api_uri",
						Messages: []string{
							"missing host in url",
						},
					},
				},
			},
			{
				// test typedefs.header_name
				name: "key-auth",
				config: `{
					"key_names": [
						"header!"
					]
				}`,
				wantsErr: true,
				expectedErrs: []*grpcModel.ErrorDetail{
					{
						Type:  grpcModel.ErrorType_ERROR_TYPE_FIELD,
						Field: "config.key_names[0]",
						Messages: []string{
							"bad header name 'header!', allowed characters are A-Z, a-z, 0-9, '_', and '-'",
						},
					},
				},
			},
			{
				// test valid typedefs.lua_code
				name: "loggly",
				config: `{
					"key": "KEY",
					"custom_fields_by_lua": {
						"header": "return nil"
					}
				}`,
			},
			{
				// test valid but unsafe typedefs.lua_code
				name: "loggly",
				config: `{
					"key": "KEY",
					"custom_fields_by_lua": {
						"header": "os.execute('echo hello')"
					}
				}`,
			},
			{
				// test invalid typedefs.lua_code
				name: "loggly",
				config: `{
					"key": "KEY",
					"custom_fields_by_lua": {
						"header": "hello"
					}
				}`,
				wantsErr: true,
				expectedErrs: []*grpcModel.ErrorDetail{
					{
						Type:  grpcModel.ErrorType_ERROR_TYPE_FIELD,
						Field: "config.custom_fields_by_lua",
						Messages: []string{
							"Error parsing function: lua-tree/share/lua/5.1/kong/tools/kong-lua-sandbox.lua:146: " +
								"<string> at EOF:   parse error\n",
						},
					},
				},
			},
		}
		for _, policy := range tt {
			var config structpb.Struct
			require.Nil(t, json.ProtoJSONUnmarshal([]byte(policy.config), &config))
			p := &grpcModel.Plugin{
				Name:      policy.name,
				Protocols: []string{"http", "https"},
				Enabled:   wrapperspb.Bool(true),
				Config:    &config,
			}
			err := validator.Validate(context.Background(), p)
			if policy.wantsErr {
				require.NotNil(t, err)
				validationErr, ok := err.(validation.Error)
				require.True(t, ok)
				require.ElementsMatch(t, policy.expectedErrs, validationErr.Errs)
			} else {
				require.Nil(t, err)
			}
		}
	})
	// This test uses bundled plugin schemas which have not been loaded into the validator. This
	// allows for multiple plugin configurations to be tested without creating non-bundled schemas and
	// further covers use case for non-bundled plugins.
	t.Run("validate plugin configuration with plugin schemas; embed not loaded", func(t *testing.T) {
		storeLoader := setupStoreLoader(t)
		require.NotNil(t, storeLoader)
		validator, err := NewLuaValidator(Opts{
			Logger:      log.Logger,
			StoreLoader: storeLoader,
		})
		assert.NoError(t, err)
		require.NotNil(t, validator)
		resource.SetValidator(validator)

		tests := []struct {
			name   string
			config string
		}{
			{
				name: "acme",
				config: `{
						"account_email": "example@example.com"
					}`,
			},
			{
				name: "rate-limiting",
				config: `{
					"second": 42,
					"policy": "local",
					"path": "/"
				}`,
			},
			{
				name: "key-auth",
				config: `{
					"key_names": [
						"koko-header"
					]
				}`,
			},
			{
				name: "ip-restriction",
				config: `{
					"allow": [
						"1.2.3.4",
						"::1"
					]
				}`,
			},
		}
		for _, test := range tests {
			var config structpb.Struct
			require.Nil(t, json.ProtoJSONUnmarshal([]byte(test.config), &config))
			p := &grpcModel.Plugin{
				Name:      test.name,
				Protocols: []string{"http", "https"},
				Config:    &config,
			}

			schema, err := plugin.Schemas.ReadFile("schemas" + "/" + test.name + ".lua")
			require.NoError(t, err)
			err = insertPluginSchema(t, test.name, string(schema), storeLoader)
			require.NoError(t, err)

			assert.NoError(t, err)
			err = validator.Validate(getValidContext(), p)
			assert.NoError(t, err)
			require.Empty(t, validator.luaSchemaNames)
			require.Empty(t, validator.rawLuaSchemas)
		}
	})

	t.Run("[statsd] accepts valid identifier_default", func(t *testing.T) {
		var config structpb.Struct
		require.NoError(
			t,
			json.ProtoJSONUnmarshal(
				[]byte(`{
					"consumer_identifier_default": "consumer_id",
					"service_identifier_default": "service_id",
					"workspace_identifier_default": "workspace_id"
				}`),
				&config,
			),
		)

		assert.NoError(t, validator.Validate(
			context.Background(),
			&grpcModel.Plugin{
				Name:      "statsd",
				Config:    &config,
				Protocols: []string{typedefs.ProtocolHTTP, typedefs.ProtocolHTTPS},
				Enabled:   wrapperspb.Bool(true),
			}),
		)
	})

	t.Run("[statsd] rejects invalid identifier_default", func(t *testing.T) {
		var config structpb.Struct
		require.NoError(
			t,
			json.ProtoJSONUnmarshal(
				[]byte(`{
					"consumer_identifier_default": "invalid",
					"service_identifier_default": "invalid",
					"workspace_identifier_default": "invalid"
				}`),
				&config,
			),
		)
		err := validator.Validate(context.Background(), &grpcModel.Plugin{
			Name:      "statsd",
			Config:    &config,
			Protocols: []string{"http", "https"},
			Enabled:   wrapperspb.Bool(true),
		})
		require.Error(t, err)
		require.IsType(t, validation.Error{}, err)
		assert.ElementsMatch(
			t,
			[]*grpcModel.ErrorDetail{
				{
					Type:     grpcModel.ErrorType_ERROR_TYPE_FIELD,
					Field:    "config.consumer_identifier_default",
					Messages: []string{"expected one of: consumer_id, custom_id, username"},
				},
				{
					Type:     grpcModel.ErrorType_ERROR_TYPE_FIELD,
					Field:    "config.service_identifier_default",
					Messages: []string{"expected one of: service_id, service_name, service_host, service_name_or_host"},
				},
				{
					Type:     grpcModel.ErrorType_ERROR_TYPE_FIELD,
					Field:    "config.workspace_identifier_default",
					Messages: []string{"expected one of: workspace_id, workspace_name"},
				},
			},
			err.(validation.Error).Errs,
		)
	})

	t.Run("[statsd] accepts valid allow_status_codes entries as ranges", func(t *testing.T) {
		var config structpb.Struct
		require.NoError(
			t,
			json.ProtoJSONUnmarshal(
				[]byte(`{
					"allow_status_codes": [
						"200-299",
						"300-399"
					]
				}`),
				&config,
			),
		)

		assert.NoError(t, validator.Validate(
			context.Background(),
			&grpcModel.Plugin{
				Name:      "statsd",
				Config:    &config,
				Protocols: []string{typedefs.ProtocolHTTP, typedefs.ProtocolHTTPS},
				Enabled:   wrapperspb.Bool(true),
			}),
		)
	})

	t.Run("[statsd] rejects invalid allow_status_codes formats", func(t *testing.T) {
		var config structpb.Struct
		// invalid formats:
		// - literals
		// - special chars
		// - range format with literals
		// - fixed number without range
		require.NoError(
			t,
			json.ProtoJSONUnmarshal(
				[]byte(`{
					"allow_status_codes": [
						"test",
						"$%%",
						"test-test",
						"200"
					]
				}`),
				&config,
			),
		)
		err := validator.Validate(context.Background(), &grpcModel.Plugin{
			Name:      "statsd",
			Config:    &config,
			Protocols: []string{"http", "https"},
			Enabled:   wrapperspb.Bool(true),
		})
		require.Error(t, err)
		require.IsType(t, validation.Error{}, err)
		assert.ElementsMatch(
			t,
			[]*grpcModel.ErrorDetail{
				{
					Type:     grpcModel.ErrorType_ERROR_TYPE_FIELD,
					Field:    "config.allow_status_codes[0]",
					Messages: []string{"invalid value: test"},
				},
				{
					Type:     grpcModel.ErrorType_ERROR_TYPE_FIELD,
					Field:    "config.allow_status_codes[1]",
					Messages: []string{"invalid value: $%%"},
				},
				{
					Type:     grpcModel.ErrorType_ERROR_TYPE_FIELD,
					Field:    "config.allow_status_codes[2]",
					Messages: []string{"invalid value: test-test"},
				},
				{
					Type:     grpcModel.ErrorType_ERROR_TYPE_FIELD,
					Field:    "config.allow_status_codes[3]",
					Messages: []string{"invalid value: 200"},
				},
			},
			err.(validation.Error).Errs,
		)
	})

	t.Run("[statsd] rejects invalid allow_status_codes with numbers and literals", func(t *testing.T) {
		var config structpb.Struct
		require.NoError(
			t,
			json.ProtoJSONUnmarshal(
				[]byte(`{
					"allow_status_codes": [
						"test-299",
						"300-test"
					]
				}`),
				&config,
			),
		)
		err := validator.Validate(context.Background(), &grpcModel.Plugin{
			Name:      "statsd",
			Config:    &config,
			Protocols: []string{"http", "https"},
			Enabled:   wrapperspb.Bool(true),
		})
		require.Error(t, err)
		require.IsType(t, validation.Error{}, err)
		assert.ElementsMatch(
			t,
			[]*grpcModel.ErrorDetail{
				{
					Type:     grpcModel.ErrorType_ERROR_TYPE_FIELD,
					Field:    "config.allow_status_codes[0]",
					Messages: []string{"invalid value: test-299"},
				},
				{
					Type:     grpcModel.ErrorType_ERROR_TYPE_FIELD,
					Field:    "config.allow_status_codes[1]",
					Messages: []string{"invalid value: 300-test"},
				},
			},
			err.(validation.Error).Errs,
		)
	})

	t.Run("[statsd] rejects allow_status_codes with valid and invalid entries", func(t *testing.T) {
		var config structpb.Struct
		require.NoError(
			t,
			json.ProtoJSONUnmarshal(
				[]byte(`{
					"allow_status_codes": [
						"200-300",
						"test-test"
					]
				}`),
				&config,
			),
		)
		err := validator.Validate(context.Background(), &grpcModel.Plugin{
			Name:      "statsd",
			Config:    &config,
			Protocols: []string{"http", "https"},
			Enabled:   wrapperspb.Bool(true),
		})
		require.Error(t, err)
		require.IsType(t, validation.Error{}, err)
		assert.ElementsMatch(
			t,
			[]*grpcModel.ErrorDetail{
				{
					Type:     grpcModel.ErrorType_ERROR_TYPE_FIELD,
					Field:    "config.allow_status_codes[0]",
					Messages: []string{"[2] = invalid value: test-test"},
				},
			},
			err.(validation.Error).Errs,
		)
	})
}

func TestValidateSchema(t *testing.T) {
	validator := goodValidator
	t.Run("valid plugin schema will properly validate", func(t *testing.T) {
		pluginName, err := validator.ValidateSchema(context.Background(),
			goodPluginSchema("validate-schema-test"))
		assert.NoError(t, err)
		require.Equal(t, "validate-schema-test", pluginName)
	})

	t.Run("validation should fail for bundled plugin schemas", func(t *testing.T) {
		schemaFiles, _ := plugin.Schemas.ReadDir("schemas")
		for _, schemaFile := range schemaFiles {
			name := schemaFile.Name()
			pluginName := strings.TrimSuffix(name, filepath.Ext(name))
			schema, _ := plugin.Schemas.ReadFile("schemas/" + name)
			_, err := validator.ValidateSchema(context.Background(), string(schema))
			require.NotNil(t, err)
			validationErr, ok := err.(validation.Error)
			require.True(t, ok)
			require.ElementsMatch(t, validationErr.Errs, []*grpcModel.ErrorDetail{
				{
					Type: grpcModel.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						fmt.Sprintf("unique constraint failed: schema already exists for plugin '%s'", pluginName),
					},
				},
			})
		}
	})

	t.Run("invalid plugin schema will fail to validate", func(t *testing.T) {
		tests := []struct {
			schema     string
			errMessage string
		}{
			{
				errMessage: "invalid plugin schema: cannot be empty",
			},
			{
				schema:     "    ",
				errMessage: "invalid plugin schema: cannot be empty",
			},
			{
				schema:     "invalid plugin schema",
				errMessage: "stack traceback",
			},
			{
				schema: "return {}",
				errMessage: "[goks] 2 schema violations (fields: field required for entity check; " +
					"name: field required for entity check)",
			},
			{
				schema: `return {
					name = "invalid-schema-test",
					{ field = { type = "string" } }
				}`,
				errMessage: "stack traceback",
			},
		}
		for _, test := range tests {
			pluginName, err := validator.ValidateSchema(context.Background(), test.schema)
			require.NotNil(t, err)
			validationErr, ok := err.(validation.Error)
			require.True(t, ok)
			assert.Len(t, pluginName, 0)
			assert.Len(t, validationErr.Errs, 1)
			assert.Len(t, validationErr.Errs[0].Messages, 1)
			assert.Equal(t, grpcModel.ErrorType_ERROR_TYPE_FIELD, validationErr.Errs[0].Type)
			assert.Equal(t, "lua_schema", validationErr.Errs[0].Field)
			assert.Contains(t, validationErr.Errs[0].Messages[0], test.errMessage)
		}
	})
}

type testPluginSchema struct {
	Name string `json:"plugin_name,omitempty" yaml:"plugin_name,omitempty"`
}

func TestPluginLuaSchema(t *testing.T) {
	validator := goodValidator
	pluginNames := []string{
		"one",
		"two",
		"three",
	}
	for _, pluginName := range pluginNames {
		jsonSchmea := fmt.Sprintf("{\"plugin_name\": \"%s\"}", pluginName)
		err := addLuaSchema(pluginName, jsonSchmea, validator.rawLuaSchemas, &validator.luaSchemaNames)
		require.Nil(t, err)
	}

	t.Run("ensure error adding the same plugin name", func(t *testing.T) {
		err := addLuaSchema("two", "{}", validator.rawLuaSchemas, &validator.luaSchemaNames)
		require.EqualError(t, err, "schema for plugin 'two' already exists")
	})

	t.Run("ensure error adding an empty schema", func(t *testing.T) {
		err := addLuaSchema("empty", "", validator.rawLuaSchemas, &validator.luaSchemaNames)
		require.EqualError(t, err, "schema cannot be empty")
		err = addLuaSchema("empty", "       ", validator.rawLuaSchemas, &validator.luaSchemaNames)
		require.EqualError(t, err, "schema cannot be empty")
	})

	t.Run("validate plugin JSON schema", func(t *testing.T) {
		for _, pluginName := range pluginNames {
			var pluginSchema testPluginSchema
			rawJSONSchmea, err := validator.GetRawLuaSchema(context.Background(), pluginName)
			require.Nil(t, err)
			require.Nil(t, json.ProtoJSONUnmarshal(rawJSONSchmea, &pluginSchema))
			require.EqualValues(t, pluginName, pluginSchema.Name)
		}
	})

	t.Run("ensure error retrieving unknown plugin JSON schema", func(t *testing.T) {
		rawJSONSchema, err := validator.GetRawLuaSchema(context.Background(), "invalid-plugin")
		require.Empty(t, rawJSONSchema)
		require.Errorf(t, err, "raw JSON schema not found for plugin: 'invalid-plugin'")
	})
}

func TestLuaValidator_GetAvailablePluginNames(t *testing.T) {
	validator := &LuaValidator{luaSchemaNames: []string{"a", "b", "c"}}
	assert.Equal(t, validator.luaSchemaNames, validator.GetAvailablePluginNames(context.Background()))
}

func TestLuaValidator_LoadLuaPluginSchemaNoStoreLoader(t *testing.T) {
	validator, err := NewLuaValidator(Opts{
		Logger: log.Logger,
	})
	assert.NoError(t, err)
	require.NotNil(t, validator)

	cleanup := validator.loadLuaPluginSchema(getValidContext(), "")
	assert.NotNil(t, cleanup)
	defer cleanup()
}

func TestLuaValidator_LoadLuaPluginSchema(t *testing.T) {
	storeLoader := setupStoreLoader(t)
	require.NotNil(t, storeLoader)
	goodValidator.storeLoader = storeLoader

	t.Run("bundled plugin schema does not throw error", func(t *testing.T) {
		cleanup := goodValidator.loadLuaPluginSchema(getValidContext(), "acl")
		defer cleanup()
		assert.NotNil(t, cleanup)
	})
	t.Run("non-bundled plugin schema does not throw error", func(t *testing.T) {
		err := insertPluginSchema(t, "non-bundled", goodPluginSchema("non-bundled"), storeLoader)
		assert.NoError(t, err)
		cleanup := goodValidator.loadLuaPluginSchema(getValidContext(), "non-bundled")
		defer cleanup()
		assert.NotNil(t, cleanup)
	})
	t.Run("plugin schema not-found does not throw error", func(t *testing.T) {
		cleanup := goodValidator.loadLuaPluginSchema(getValidContext(), "plugin-name")
		defer cleanup()
		assert.NotNil(t, cleanup)
	})
	t.Run("empty plugin name does not throw error", func(t *testing.T) {
		cleanup := goodValidator.loadLuaPluginSchema(getValidContext(), "")
		defer cleanup()
		assert.NotNil(t, cleanup)
	})
	t.Run("invalid plugin name does not throw error", func(t *testing.T) {
		cleanup := goodValidator.loadLuaPluginSchema(getValidContext(), "!nva!d-plug!n-name")
		defer cleanup()
		assert.NotNil(t, cleanup)
	})
	t.Run("invalid context does not throw error", func(t *testing.T) {
		cleanup := goodValidator.loadLuaPluginSchema(context.Background(), "")
		defer cleanup()
		assert.NotNil(t, cleanup)
	})
}

func TestLuaValidator_GetPluginSchema(t *testing.T) {
	storeLoader := setupStoreLoader(t)
	require.NotNil(t, storeLoader)
	goodValidator.SetStoreLoader(storeLoader)

	t.Run("bundled plugin schema returns empty string", func(t *testing.T) {
		schema := goodValidator.getPluginSchemaFromDB(getValidContext(), "acl")
		assert.Zero(t, schema)
	})
	t.Run("non-bundled plugin schema returns schema", func(t *testing.T) {
		expectedSchema := goodPluginSchema("non-bundled")
		err := insertPluginSchema(t, "non-bundled", expectedSchema, storeLoader)
		assert.NoError(t, err)
		schema := goodValidator.getPluginSchemaFromDB(getValidContext(), "non-bundled")
		assert.Equal(t, expectedSchema, schema)
	})
	t.Run("plugin schema not-found returns empty string", func(t *testing.T) {
		schema := goodValidator.getPluginSchemaFromDB(getValidContext(), "plugin-name")
		assert.Zero(t, schema)
	})
	t.Run("empty plugin name returns empty string", func(t *testing.T) {
		schema := goodValidator.getPluginSchemaFromDB(getValidContext(), "")
		assert.Zero(t, schema)
	})
	t.Run("invalid plugin name returns empty string", func(t *testing.T) {
		schema := goodValidator.getPluginSchemaFromDB(getValidContext(), "!nvalId-plug!n-name")
		assert.Zero(t, schema)
	})
	t.Run("invalid context returns empty string", func(t *testing.T) {
		schema := goodValidator.getPluginSchemaFromDB(context.Background(), "")
		assert.Zero(t, schema)
	})
}

func TestLuaValidator_GetDB(t *testing.T) {
	storeLoader := setupStoreLoader(t)
	require.NotNil(t, storeLoader)
	goodValidator.storeLoader = storeLoader
	t.Run("context not containing RequestCluster returns default store", func(t *testing.T) {
		db, err := goodValidator.getDB(getValidContext())
		require.NoError(t, err)
		assert.NotNil(t, db)
		assert.Equal(t, store.DefaultCluster, db.Cluster())
	})
	t.Run("nil context returns error", func(t *testing.T) {
		// disable staticcheck since nil context is being tested
		db, err := goodValidator.getDB(nil) //nolint:staticcheck
		assert.Nil(t, db)
		assert.Error(t, err, "invalid context: failed to retrieve RequestCluster from context")
	})
	t.Run("store loader not set returns error", func(t *testing.T) {
		validator, err := NewLuaValidator(Opts{Logger: log.Logger})
		require.Nil(t, err)
		require.NotNil(t, validator)
		db, err := validator.getDB(context.Background())
		assert.Nil(t, db)
		assert.Error(t, err, "invalid StoreLoader: store loader cannot be nil")
	})
	t.Run("return proper database", func(t *testing.T) {
		v, err := getGoodValidatorWithStoreLoader(&testLoader{clusterID: "test-cluster"})
		require.NoError(t, err)

		db, err := v.getDB(getValidContextWithClusterID("test-cluster"))
		require.NoError(t, err)

		assert.NotNil(t, db)
		assert.Equal(t, "test-cluster", db.Cluster())
	})
	t.Run("store loader error results in a grpc status error", func(t *testing.T) {
		validator, err := NewLuaValidator(Opts{Logger: log.Logger})
		require.Nil(t, err)
		require.NotNil(t, validator)
		validator.storeLoader = badLoader{}
		db, err := validator.getDB(getValidContext())
		assert.Nil(t, db)
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})
}

type testLoader struct {
	serverUtil.StoreLoader
	clusterID string
}

func (l *testLoader) Load(_ context.Context, cluster *grpcModel.RequestCluster) (store.Store, error) {
	return (&store.ObjectStore{}).ForCluster(cluster.Id), nil
}

type badLoader struct{}

func (b badLoader) Load(ctx context.Context, cluster *grpcModel.RequestCluster) (store.Store, error) {
	return nil, serverUtil.StoreLoadErr{
		Code: 3,
	}
}
