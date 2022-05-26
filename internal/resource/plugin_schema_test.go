package resource_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/kong/koko/internal/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPluginSchema(t *testing.T) {
	s := resource.NewPluginSchema()
	require.NotNil(t, s)
	require.NotNil(t, s.PluginSchema)
}

func TestPluginSchema_ID(t *testing.T) {
	var s resource.PluginSchema
	require.Empty(t, s.ID())

	s = resource.NewPluginSchema()
	require.Equal(t, "", s.ID())

	s = resource.NewPluginSchema()
	s.PluginSchema.Name = "plugin-schema-name"
	require.Equal(t, "plugin-schema-name", s.ID())
}

func TestPluginSchema_Type(t *testing.T) {
	require.Equal(t, resource.TypePluginSchema, resource.NewPluginSchema().Type())
}

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

func goodPluginSchema(name string) string {
	return fmt.Sprintf(pluginSchemaFormat, name)
}

func TestPluginSchema_ProcessDefaults(t *testing.T) {
	t.Run("no errors occur when defaults are processed", func(t *testing.T) {
		r := resource.PluginSchema{
			PluginSchema: &model.PluginSchema{},
		}
		err := r.ProcessDefaults(context.Background())
		assert.NoError(t, err)
		require.LessOrEqual(t, r.PluginSchema.CreatedAt, int32(time.Now().Unix()))
		require.LessOrEqual(t, r.PluginSchema.UpdatedAt, int32(time.Now().Unix()))
	})

	t.Run("error occurs with nil schema when processed", func(t *testing.T) {
		r := resource.PluginSchema{}
		err := r.ProcessDefaults(context.Background())
		require.NotNil(t, err)
		require.EqualError(t, err, "invalid nil resource")
	})
}

func TestPluginSchema_Validate(t *testing.T) {
	setupLuaValidator(t)
	tests := []struct {
		name               string
		pluginSchema       func() resource.PluginSchema
		wantErr            bool
		expectedPluginName string
		expectedErrs       []*model.ErrorDetail
	}{
		{
			name: "missing plugin schema throws an error",
			pluginSchema: func() resource.PluginSchema {
				return resource.NewPluginSchema()
			},
			wantErr: true,
			expectedErrs: []*model.ErrorDetail{
				{
					Type:     model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{"missing properties: 'lua_schema'"},
				},
			},
		},
		{
			name: "empty plugin schema throws an error",
			pluginSchema: func() resource.PluginSchema {
				r := resource.NewPluginSchema()
				r.PluginSchema.LuaSchema = ""
				_ = r.ProcessDefaults(context.Background())
				return r
			},
			wantErr: true,
			expectedErrs: []*model.ErrorDetail{
				{
					Type:     model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{"missing properties: 'lua_schema'"},
				},
			},
		},
		{
			name: "valid plugin schema doesn't throw any error",
			pluginSchema: func() resource.PluginSchema {
				r := resource.NewPluginSchema()
				r.PluginSchema.LuaSchema = goodPluginSchema("valid-plugin-schema")
				_ = r.ProcessDefaults(context.Background())
				return r
			},
			wantErr:            false,
			expectedPluginName: "valid-plugin-schema",
		},
		{
			name: "error occurs when plugin name provided doesn't match expected",
			pluginSchema: func() resource.PluginSchema {
				r := resource.NewPluginSchema()
				r.PluginSchema.Name = "mismatch-plugin-name"
				r.PluginSchema.LuaSchema = goodPluginSchema("valid-plugin-schema")
				_ = r.ProcessDefaults(context.Background())
				return r
			},
			wantErr: true,
			expectedErrs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "name",
					Messages: []string{
						"invalid plugin schema: name is derived from the plugin schema and must match if set",
					},
				},
			},
		},
		{
			name: "error occurs when invalid schema is validated",
			pluginSchema: func() resource.PluginSchema {
				r := resource.NewPluginSchema()
				r.PluginSchema.LuaSchema = "return {}"
				_ = r.ProcessDefaults(context.Background())
				return r
			},
			wantErr: true,
			expectedErrs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "lua_schema",
					Messages: []string{
						"[goks] 2 schema violations (fields: field required for entity check; name: field required for entity check)",
					},
				},
			},
		},
		{
			name: "error occurs when invalid plugin name is derived from schema",
			pluginSchema: func() resource.PluginSchema {
				r := resource.NewPluginSchema()
				r.PluginSchema.LuaSchema = goodPluginSchema("invalid!name")
				_ = r.ProcessDefaults(context.Background())
				return r
			},
			wantErr: true,
			expectedErrs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "name",
					Messages: []string{
						"must match pattern '^[0-9a-zA-Z.\\-_~]*$'",
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.pluginSchema().Validate(context.Background())
			if (err != nil) != test.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, test.wantErr)
			}
			if test.expectedErrs != nil {
				verr, _ := err.(validation.Error)
				require.ElementsMatch(t, test.expectedErrs, verr.Errs)
			}
		})
	}
}
