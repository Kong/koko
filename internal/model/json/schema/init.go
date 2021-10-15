package schema

import (
	"bytes"
	"fmt"
	"strings"

	genJSONSchema "github.com/kong/koko/internal/gen/jsonschema"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

var schemas = map[string]*jsonschema.Schema{}

func init() {
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
	}
}

func Get(name string) (*jsonschema.Schema, error) {
	schema, ok := schemas[name]
	if !ok {
		return nil, fmt.Errorf("schema not found: '%v'", name)
	}
	return schema, nil
}
