package plugin

import (
	"embed"
	"fmt"
	"io/fs"
	"testing"
	"time"

	"github.com/google/uuid"
	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/log"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/kong/koko/internal/plugin/testdata"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

//go:embed testdata/schemas/*
var badSchemaFS embed.FS

// goodValidator is loaded at init.
// This is an optimization to speed up tests.
var goodValidator *LuaValidator

func init() {
	var err error
	goodValidator, err = NewLuaValidator(Opts{Logger: log.Logger})
	if err != nil {
		panic(err)
	}
	err = goodValidator.LoadSchemasFromEmbed(Schemas, "schemas")
	if err != nil {
		panic(err)
	}
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

func TestLoadSchemasFromEmbed(t *testing.T) {
	validator, err := NewLuaValidator(Opts{Logger: log.Logger})
	require.Nil(t, err)
	t.Run("errors when dir doesn't exist", func(t *testing.T) {
		err = validator.LoadSchemasFromEmbed(Schemas, "does-not-exist")
		require.NotNil(t, err)
		require.IsType(t, &fs.PathError{}, err)
	})
	t.Run("errors when embed.Fs is nil", func(t *testing.T) {
		err = validator.LoadSchemasFromEmbed(embed.FS{}, "schemas")
		require.NotNil(t, err)
	})
	t.Run("loads a value schemas directory", func(t *testing.T) {
		err = validator.LoadSchemasFromEmbed(Schemas, "schemas")
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
		plugin := &model.Plugin{
			Name:   "rate-limiting",
			Config: config,
		}
		err = validator.ProcessDefaults(plugin)
		require.Nil(t, err)
		require.NotPanics(t, func() {
			uuid.MustParse(plugin.Id)
		})
		require.LessOrEqual(t, int32(time.Now().Unix()), plugin.CreatedAt)
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
		}
		require.Equal(t, expectedConfig, processConfig)
	})
}

func TestValidate(t *testing.T) {
	validator := goodValidator
	t.Run("test with entity errors", func(t *testing.T) {
		config, err := structpb.NewStruct(map[string]interface{}{
			"policy": "redis",
		})
		require.Nil(t, err)
		err = validator.Validate(&model.Plugin{
			Name:      "rate-limiting",
			Config:    config,
			Protocols: []string{"http", "https"},
			Enabled:   wrapperspb.Bool(true),
		})
		validationErr, ok := err.(validation.Error)
		require.True(t, ok)
		expected := []*model.ErrorDetail{
			{
				Type: model.ErrorType_ERROR_TYPE_ENTITY,
				Messages: []string{
					"at least one of these fields must be non-empty: 'config.second'" +
						", 'config.minute', 'config.hour', 'config.day', 'config.month', 'config.year'",
					"failed conditional validation given value of field 'config.policy'",
				},
			},
			{
				Type:     model.ErrorType_ERROR_TYPE_FIELD,
				Field:    "config.redis_host",
				Messages: []string{"required field missing"},
			},
		}
		require.ElementsMatch(t, expected, validationErr.Errs)
	})
	t.Run("plugin does not exist", func(t *testing.T) {
		err := validator.Validate(&model.Plugin{
			Name: "no-auth",
		})
		require.NotNil(t, err)
		validationErr, ok := err.(validation.Error)
		require.True(t, ok)
		expected := []*model.ErrorDetail{
			{
				Type:     model.ErrorType_ERROR_TYPE_FIELD,
				Field:    "name",
				Messages: []string{"plugin(no-auth) does not exist"},
			},
		}
		require.Equal(t, expected, validationErr.Errs)
	})
	t.Run("validates nested config structs", func(t *testing.T) {
		var config structpb.Struct
		configString := `{"add":{"headers":["nokey"]}}`
		require.Nil(t, json.Unmarshal([]byte(configString), &config))
		err := validator.Validate(&model.Plugin{
			Name:      "request-transformer",
			Config:    &config,
			Protocols: []string{"http", "https"},
			Enabled:   wrapperspb.Bool(true),
		})
		require.NotNil(t, err)
		validationErr, ok := err.(validation.Error)
		require.True(t, ok)
		expected := []*model.ErrorDetail{
			{
				Type:     model.ErrorType_ERROR_TYPE_FIELD,
				Field:    "config.add.headers[0]",
				Messages: []string{"invalid value: nokey"},
			},
		}
		require.Equal(t, expected, validationErr.Errs)
	})
	t.Run("prometheus plugin schema", func(t *testing.T) {
		p := &model.Plugin{
			Name:      "prometheus",
			Protocols: []string{"http", "https"},
			Enabled:   wrapperspb.Bool(true),
		}
		require.Nil(t, validator.ProcessDefaults(p))
		err := validator.Validate(p)
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
		p := &model.Plugin{
			Name:      "pre-function",
			Protocols: []string{"http", "https"},
			Enabled:   wrapperspb.Bool(true),
			Config:    config,
		}
		require.Nil(t, validator.ProcessDefaults(p))
		err = validator.Validate(p)
		require.Nil(t, err)

		p = &model.Plugin{
			Name:      "post-function",
			Protocols: []string{"http", "https"},
			Enabled:   wrapperspb.Bool(true),
			Config:    config,
		}
		require.Nil(t, validator.ProcessDefaults(p))
		err = validator.Validate(p)
		require.Nil(t, err)
	})
	t.Run("serverless plugin schema accepts invalid lua-code", func(t *testing.T) {
		config, err := structpb.NewStruct(map[string]interface{}{
			"log": []interface{}{
				"this is not valid lua !",
			},
		})
		require.Nil(t, err)
		p := &model.Plugin{
			Name:      "pre-function",
			Protocols: []string{"http", "https"},
			Enabled:   wrapperspb.Bool(true),
			Config:    config,
		}
		require.Nil(t, validator.ProcessDefaults(p))
		err = validator.Validate(p)
		require.Nil(t, err)
	})
	t.Run("aws-lambda plugin errors out when custom_entity_check fails", func(t *testing.T) {
		var config structpb.Struct
		configString := `{"proxy_url": "https://my-proxy-server:3128"}`
		require.Nil(t, json.Unmarshal([]byte(configString), &config))
		p := &model.Plugin{
			Name:      "aws-lambda",
			Protocols: []string{"http", "https"},
			Enabled:   wrapperspb.Bool(true),
			Config:    &config,
		}
		require.Nil(t, validator.ProcessDefaults(p))
		err := validator.Validate(p)
		validationErr, ok := err.(validation.Error)
		require.True(t, ok)
		expected := []*model.ErrorDetail{
			{
				Type:     model.ErrorType_ERROR_TYPE_ENTITY,
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
			p := &model.Plugin{
				Name:      "prometheus",
				Protocols: []string{protocol},
				Enabled:   wrapperspb.Bool(true),
			}
			err := validator.Validate(p)
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
			p := &model.Plugin{
				Name:      "rate-limiting",
				Protocols: []string{protocol},
				Enabled:   wrapperspb.Bool(true),
				Config:    config,
			}
			err := validator.Validate(p)
			require.NotNil(t, err)
			validationErr, ok := err.(validation.Error)
			require.True(t, ok)
			expected := []*model.ErrorDetail{
				{
					Type:     model.ErrorType_ERROR_TYPE_FIELD,
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
			expectedErrs []*model.ErrorDetail
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
				expectedErrs: []*model.ErrorDetail{
					{
						Type:  model.ErrorType_ERROR_TYPE_FIELD,
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
			require.Nil(t, json.Unmarshal([]byte(policy.config), &config))
			p := &model.Plugin{
				Name:      "rate-limiting",
				Protocols: []string{"http", "https"},
				Enabled:   wrapperspb.Bool(true),
				Config:    &config,
			}
			err := validator.Validate(p)
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
			expectedErrs []*model.ErrorDetail
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
				expectedErrs: []*model.ErrorDetail{
					{
						Type:  model.ErrorType_ERROR_TYPE_FIELD,
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
			require.Nil(t, json.Unmarshal([]byte(policy.config), &config))
			p := &model.Plugin{
				Name:      "response-ratelimiting",
				Protocols: []string{"http", "https"},
				Enabled:   wrapperspb.Bool(true),
				Config:    &config,
			}
			err := validator.Validate(p)
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
	t.Run("ensure proxy-cache advanced validation for shared dict checks", func(t *testing.T) {
		config, err := structpb.NewStruct(map[string]interface{}{
			"strategy": "memory",
		})
		require.Nil(t, err)
		p := &model.Plugin{
			Name:      "proxy-cache",
			Protocols: []string{"http", "https"},
			Enabled:   wrapperspb.Bool(true),
			Config:    config,
		}
		err = validator.Validate(p)
		require.Nil(t, err)
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
		err := addLuaSchema(pluginName, jsonSchmea, validator.rawLuaSchemas)
		require.Nil(t, err)
	}

	t.Run("ensure error adding the same plugin name", func(t *testing.T) {
		err := addLuaSchema("two", "{}", validator.rawLuaSchemas)
		require.EqualError(t, err, "schema for plugin 'two' already exists")
	})

	t.Run("ensure error adding an empty schema", func(t *testing.T) {
		err := addLuaSchema("empty", "", validator.rawLuaSchemas)
		require.EqualError(t, err, "schema cannot be empty")
		err = addLuaSchema("empty", "       ", validator.rawLuaSchemas)
		require.EqualError(t, err, "schema cannot be empty")
	})

	t.Run("validate plugin JSON schema", func(t *testing.T) {
		for _, pluginName := range pluginNames {
			var pluginSchema testPluginSchema
			rawJSONSchmea, err := validator.GetRawLuaSchema(pluginName)
			require.Nil(t, err)
			require.Nil(t, json.Unmarshal(rawJSONSchmea, &pluginSchema))
			require.EqualValues(t, pluginName, pluginSchema.Name)
		}
	})

	t.Run("ensure error retrieving unknown plugin JSON schema", func(t *testing.T) {
		rawJSONSchema, err := validator.GetRawLuaSchema("invalid-plugin")
		require.Empty(t, rawJSONSchema)
		require.Errorf(t, err, "raw JSON schema not found for plugin: 'invalid-plugin'")
	})
}
