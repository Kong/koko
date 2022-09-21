package generator

import (
	"fmt"
)

const SchemaVersion = "https://json-schema.org/draft/2020-12/schema"

type SchemaRegistry struct {
	GlobalSchema *Schema
}

func NewSchemaRegistry() *SchemaRegistry {
	return &SchemaRegistry{
		GlobalSchema: &Schema{
			Version:     SchemaVersion,
			Definitions: map[string]*Schema{},
		},
	}
}

func (r *SchemaRegistry) Register(name string, schema *Schema) error {
	if _, ok := r.GlobalSchema.Definitions[name]; ok {
		return fmt.Errorf("type already registered: '%v'", name)
	}
	r.GlobalSchema.Definitions[name] = schema
	return nil
}

func (r *SchemaRegistry) Unregister(name string) (*Schema, error) {
	schema, ok := r.GlobalSchema.Definitions[name]
	if !ok {
		return nil, fmt.Errorf("type not registered yet: '%v'", name)
	}
	delete(r.GlobalSchema.Definitions, name)
	return schema, nil
}

var Registry = NewSchemaRegistry()
