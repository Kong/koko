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

var (
	schemas     = map[string]*jsonschema.Schema{}
	schemasJSON = map[string]string{}
	once        sync.Once
)

// initSchemas reads and compiles schemas.
// This is not done in a traditional init() to avoid circular dependency on the
// generated JSON schemas.
func initSchemas() {
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
			panic(fmt.Sprintf("expected a JSON file but got: %v", name))
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
		schemas[schemaName] = compiler.MustCompile("internal://" + schemaName)

		// Store the JSON schema in a compact format
		buffer := new(bytes.Buffer)
		if err := json.Compact(buffer, schema); err != nil {
			panic(err)
		}
		schemasJSON[schemaName] = buffer.String()
	}
}

func Get(name string) (*jsonschema.Schema, error) {
	once.Do(initSchemas)
	schema, ok := schemas[name]
	if !ok {
		return nil, fmt.Errorf("schema not found: '%s'", name)
	}
	return schema, nil
}

func GetJSONFields(name string) (string, error) {
	once.Do(initSchemas)
	schemaJSON, ok := schemasJSON[name]
	if !ok {
		return "", fmt.Errorf("JSON schema not found: '%s'", name)
	}
	return schemaJSON, nil
}
