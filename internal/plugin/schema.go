package plugin

import (
	"fmt"
	"strings"
)

var rawLuaSchemas = map[string][]byte{}

// This method should only be called from tests.
func ClearLuaSchemas() {
	rawLuaSchemas = make(map[string][]byte)
}

func AddLuaSchema(name string, schema string) error {
	if _, found := rawLuaSchemas[name]; found {
		return fmt.Errorf("schema for plugin '%s' already exists", name)
	}
	trimmedSchema := strings.TrimSpace(schema)
	if len(trimmedSchema) == 0 {
		return fmt.Errorf("schema cannot be empty")
	}
	rawLuaSchemas[name] = []byte(schema)
	return nil
}

func GetRawLuaSchema(name string) ([]byte, error) {
	rawLuaSchema, ok := rawLuaSchemas[name]
	if !ok {
		return []byte{}, fmt.Errorf("raw Lua schema not found for plugin: '%s'", name)
	}
	return rawLuaSchema, nil
}
