package plugin

import (
	"context"

	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
)

// Validator handles various needs for plugin validation.
type Validator interface {
	// Validate executes the validate() Lua function for the given plugin.
	Validate(ctx context.Context, plugin *model.Plugin) error

	// ValidateSchema executes the ValidateSchema() Lua function for the given plugin schema
	// and returns the plugin name. In the event the schema is not valid or the plugin name
	// already exists (e.g. bundled plugin), an error is returned.
	ValidateSchema(ctx context.Context, pluginSchema string) (string, error)

	// ProcessDefaults executes the process_auto_fields() Lua function for the given plugin.
	ProcessDefaults(ctx context.Context, plugin *model.Plugin) error

	// GetAvailablePluginNames returns all available plugins, in
	// ascending order. The returned slice must not be modified.
	GetAvailablePluginNames(ctx context.Context) []string

	// GetRawLuaSchema returns the raw Lua schema for the given plugin. In the event the plugin
	// does not exist, an error is returned. The returned slice must not be modified.
	GetRawLuaSchema(ctx context.Context, name string) ([]byte, error)
}
