package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	genJSONSchema "github.com/kong/koko/internal/gen/jsonschema"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

type entitySchema struct {
	once           sync.Once
	rawJSONSchemas map[string][]byte
	schemas        map[string]*jsonschema.Schema
}

type pluginSchema struct {
	rawJSONSchemas map[string][]byte
}

var (
	entity entitySchema
	plugin pluginSchema
)

// initEntitySchemas reads and compiles schemas.
// This is not done in a traditional init() to avoid circular dependency on the
// generated JSON schemas.
func initEntitySchemas() {
	const dir = "schemas"
	schemaFS := genJSONSchema.KongSchemas
	files, err := schemaFS.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	compiler := jsonschema.NewCompiler()
	compiler.ExtractAnnotations = true
	for _, file := range files {
		name := file.Name()
		if !strings.HasSuffix(name, ".json") {
			panic(fmt.Sprintf("expected a JSON file for entity but got: %v", name))
		}
		schemaName := strings.TrimSuffix(name, ".json")
		schema, err := schemaFS.ReadFile(fmt.Sprintf("%s/%s", dir, name))
		if err != nil {
			panic(err)
		}
		err = compiler.AddResource("internal://"+schemaName, bytes.NewReader(schema))
		if err != nil {
			panic(err)
		}
		entity.schemas[schemaName] = compiler.MustCompile("internal://" + schemaName)

		// Store the raw JSON schema in a compact format
		buffer := new(bytes.Buffer)
		if err := json.Compact(buffer, schema); err != nil {
			panic(err)
		}
		entity.rawJSONSchemas[schemaName] = buffer.Bytes()
	}
}

func GetEntity(name string) (*jsonschema.Schema, error) {
	entity.once.Do(initEntitySchemas)
	schema, ok := entity.schemas[name]
	if !ok {
		return nil, fmt.Errorf("schema not found for entity: '%s'", name)
	}
	return schema, nil
}

func GetEntityRawJSON(name string) ([]byte, error) {
	entity.once.Do(initEntitySchemas)
	rawJSONSchema, ok := entity.rawJSONSchemas[name]
	if !ok {
		return []byte{}, fmt.Errorf("raw JSON schema not found for entity: '%s'", name)
	}
	return rawJSONSchema, nil
}

// This method should only be called from tests.
func ClearPluginJSONSchema() {
	plugin.rawJSONSchemas = make(map[string][]byte)
}

func AddPluginJSONSchema(name string, schema string) error {
	if _, found := plugin.rawJSONSchemas[name]; found {
		return fmt.Errorf("schema for plugin '%s' already exists", name)
	}
	trimmedSchema := strings.TrimSpace(schema)
	if len(trimmedSchema) == 0 {
		return fmt.Errorf("schema cannot be empty")
	}
	plugin.rawJSONSchemas[name] = []byte(schema)
	return nil
}

func GetPluginRawJSON(name string) ([]byte, error) {
	rawJSONSchema, ok := plugin.rawJSONSchemas[name]
	if !ok {
		return []byte{}, fmt.Errorf("raw JSON schema not found for plugin: '%s'", name)
	}
	return rawJSONSchema, nil
}

func init() {
	entity.rawJSONSchemas = make(map[string][]byte)
	entity.schemas = make(map[string]*jsonschema.Schema)
	plugin.rawJSONSchemas = make(map[string][]byte)
}
