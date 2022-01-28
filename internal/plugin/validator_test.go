package plugin

import (
	"embed"
	"io/fs"
	"testing"
	"time"

	"github.com/google/uuid"
	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/log"
	"github.com/kong/koko/internal/model/json/schema"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

//go:embed testdata/schemas/*
var badSchemaFS embed.FS

func TestLoadSchemasFromEmbed(t *testing.T) {
	defer schema.ClearPluginJSONSchema()
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
	validator, err := NewLuaValidator(Opts{Logger: log.Logger})
	require.Nil(t, err)

	err = validator.LoadSchemasFromEmbed(Schemas, "schemas")
	defer schema.ClearPluginJSONSchema()
	require.Nil(t, err)
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
			"policy":              "cluster",
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
	validator, err := NewLuaValidator(Opts{Logger: log.Logger})
	require.Nil(t, err)

	err = validator.LoadSchemasFromEmbed(Schemas, "schemas")
	defer schema.ClearPluginJSONSchema()
	require.Nil(t, err)
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
					"failed conditional validation given value of field 'config.policy'",
					"failed conditional validation given value of field 'config.policy'",
				},
			},
			{
				Type:     model.ErrorType_ERROR_TYPE_FIELD,
				Field:    "config.redis_host",
				Messages: []string{"required field missing"},
			},
			{
				Type:     model.ErrorType_ERROR_TYPE_FIELD,
				Field:    "config.redis_port",
				Messages: []string{"required field missing"},
			},
			{
				Type:     model.ErrorType_ERROR_TYPE_FIELD,
				Field:    "config.redis_timeout",
				Messages: []string{"required field missing"},
			},
		}
		require.ElementsMatch(t, expected, validationErr.Errs)
	})
	t.Run("plugin does not exist", func(t *testing.T) {
		err = validator.Validate(&model.Plugin{
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
		err = validator.Validate(&model.Plugin{
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
		err = validator.Validate(p)
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
}
