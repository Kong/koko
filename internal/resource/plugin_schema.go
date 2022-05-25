package resource

import (
	"context"
	"errors"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/model/json/extension"
	"github.com/kong/koko/internal/model/json/generator"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/kong/koko/internal/model/json/validation/typedefs"
)

const (
	// TypePluginSchema denotes the Plugin Schema type.
	TypePluginSchema = model.Type("plugin_schema")

	maxPluginSchemaSize = 8192
)

// NewPluginSchema defines a new PluginSchema instance.
func NewPluginSchema() PluginSchema {
	return PluginSchema{
		PluginSchema: &v1.PluginSchema{},
	}
}

// PluginSchema represents the schema attributes for a plugin.
type PluginSchema struct {
	PluginSchema *v1.PluginSchema
}

// ID utilizes the plugin name as the ID.
func (r PluginSchema) ID() string {
	if r.PluginSchema == nil {
		return ""
	}
	// Name should be used as ID for DB store
	return r.PluginSchema.Name
}

func (r PluginSchema) Type() model.Type {
	return TypePluginSchema
}

func (r PluginSchema) Resource() model.Resource {
	return r.PluginSchema
}

// SetResource implements the Object.SetResource interface.
func (r PluginSchema) SetResource(pr model.Resource) error { return model.SetResource(r, pr) }

func (r PluginSchema) Validate(ctx context.Context) error {
	if err := validation.Validate(string(TypePluginSchema), r.PluginSchema); err != nil {
		return err
	}

	// Re-uses goks Lua plugin validator
	pluginName, err := validator.ValidateSchema(ctx, r.PluginSchema.LuaSchema)
	if err != nil {
		return err
	}

	// Validate the plugin name iff set matches the derived name from the schema
	if len(r.PluginSchema.Name) > 0 && r.PluginSchema.Name != pluginName {
		return validation.Error{
			Errs: []*v1.ErrorDetail{
				{
					Type:  v1.ErrorType_ERROR_TYPE_FIELD,
					Field: "name",
					Messages: []string{
						"invalid plugin schema: name is derived from the plugin schema and must match if set",
					},
				},
			},
		}
	}

	// Derived plugin name from schema
	r.PluginSchema.Name = pluginName

	// Re-validate to ensure derived plugin name meets definition
	return validation.Validate(string(TypePluginSchema), r.PluginSchema)
}

func (r PluginSchema) ProcessDefaults(ctx context.Context) error {
	if r.PluginSchema == nil {
		return errors.New("invalid nil resource")
	}
	return nil
}

func (r PluginSchema) Indexes() []model.Index {
	return nil
}

func init() {
	err := model.RegisterType(TypePluginSchema, &v1.PluginSchema{}, func() model.Object {
		return NewPluginSchema()
	})
	if err != nil {
		panic(err)
	}

	pluginSchema := &generator.Schema{
		Type: "object",
		Properties: map[string]*generator.Schema{
			"name": pluginName,
			"lua_schema": {
				Type:      "string",
				MaxLength: maxPluginSchemaSize,
			},
			"created_at": typedefs.UnixEpoch,
			"updated_at": typedefs.UnixEpoch,
		},
		AdditionalProperties: &falsy,
		Required: []string{
			"lua_schema",
		},
		XKokoConfig: &extension.Config{DisableValidateEndpoint: true},
	}
	err = generator.Register(string(TypePluginSchema), pluginSchema)
	if err != nil {
		panic(err)
	}
}
