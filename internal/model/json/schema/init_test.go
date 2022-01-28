package schema

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type testPluginSchema struct {
	Name string `json:"plugin_name,omitempty" yaml:"plugin_name,omitempty"`
}

func TestPluginJSONSchema(t *testing.T) {
	pluginNames := []string{
		"one",
		"two",
		"three",
	}
	for _, pluginName := range pluginNames {
		jsonSchmea := fmt.Sprintf("{\"plugin_name\": \"%s\"}", pluginName)
		err := AddPluginJSONSchema(pluginName, jsonSchmea)
		require.Nil(t, err)
	}

	t.Run("ensure error adding the same plugin name", func(t *testing.T) {
		err := AddPluginJSONSchema("two", "{}")
		require.EqualError(t, err, "schema for plugin 'two' already exists")
	})

	t.Run("ensure error adding an empty schema", func(t *testing.T) {
		err := AddPluginJSONSchema("empty", "")
		require.EqualError(t, err, "schema cannot be empty")
		err = AddPluginJSONSchema("empty", "       ")
		require.EqualError(t, err, "schema cannot be empty")
	})

	t.Run("validate plugin JSON schema", func(t *testing.T) {
		for _, pluginName := range pluginNames {
			var pluginSchema testPluginSchema
			rawJSONSchmea, err := GetPluginRawJSON(pluginName)
			require.Nil(t, err)
			require.Nil(t, json.Unmarshal(rawJSONSchmea, &pluginSchema))
			require.EqualValues(t, pluginName, pluginSchema.Name)
		}
	})

	t.Run("ensure error retrieving unknown plugin JSON schema", func(t *testing.T) {
		rawJSONSchema, err := GetPluginRawJSON("invalid-plugin")
		require.Empty(t, rawJSONSchema)
		require.Errorf(t, err, "raw JSON schema not found for plugin: 'invalid-plugin'")
	})
}
