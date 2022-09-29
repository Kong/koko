package generator

import (
	"fmt"
)

const SchemaVersion = "https://json-schema.org/draft/2020-12/schema"

// SchemaRegistry handles resource schema registration.
// A resource has to register its schema to get its json schema file generated.
type SchemaRegistry struct {
	Schema *Schema // the global schema that contains all registered definitions
}

func NewSchemaRegistry() *SchemaRegistry {
	return &SchemaRegistry{
		Schema: &Schema{
			Version:     SchemaVersion,
			Definitions: map[string]*Schema{},
		},
	}
}

func (r *SchemaRegistry) Register(name string, schema *Schema) error {
	if _, ok := r.Schema.Definitions[name]; ok {
		return fmt.Errorf("type already registered: '%v'", name)
	}
	r.Schema.Definitions[name] = schema
	return nil
}

func (r *SchemaRegistry) Unregister(name string) (*Schema, error) {
	schema, ok := r.Schema.Definitions[name]
	if !ok {
		return nil, fmt.Errorf("type not registered yet: '%v'", name)
	}
	delete(r.Schema.Definitions, name)
	return schema, nil
}

var DefaultRegistry = NewSchemaRegistry()
