package generator

import (
	"fmt"
)

const SchemaVersion = "https://json-schema.org/draft/2020-12/schema"

var globalSchema = &Schema{
	Version:     SchemaVersion,
	Definitions: map[string]*Schema{},
}

func Register(name string, schema *Schema) error {
	if _, ok := globalSchema.Definitions[name]; ok {
		return fmt.Errorf("type already registered: '%v'", name)
	}
	globalSchema.Definitions[name] = schema
	return nil
}

func Unregister(name string) (*Schema, error) {
	schema, ok := globalSchema.Definitions[name]
	if !ok {
		return nil, fmt.Errorf("type not registered yet: '%v'", name)
	}
	delete(globalSchema.Definitions, name)
	return schema, nil
}

func GlobalSchema() *Schema {
	return globalSchema
}
