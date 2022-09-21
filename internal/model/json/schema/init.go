package schema

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"sync"

	genJSONSchema "github.com/kong/koko/internal/gen/jsonschema"
	"github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/model/json/extension"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

var (
	schemas        = map[string]*jsonschema.Schema{}
	rawJSONSchemas = map[string][]byte{}
	once           sync.Once
	schemaFS       = []*embed.FS{&genJSONSchema.KongSchemas}
)

// This is not done in a traditional init() to avoid circular dependency on the
// generated JSON schemas.
func initSchemas() {
	for _, fs := range schemaFS {
		loadSchemasFromFS(fs)
	}
}

// loadSchemasFromFS reads and compiles schemas from an embed FS.
func loadSchemasFromFS(fs *embed.FS) {
	const dir = "schemas"
	files, err := fs.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	compiler := jsonschema.NewCompiler()
	compiler.ExtractAnnotations = true
	compiler.AssertFormat = true

	// Register our custom schema config extension.
	registerExtension(compiler, &extension.Config{})

	for _, file := range files {
		name := file.Name()
		if !strings.HasSuffix(name, ".json") {
			panic(fmt.Sprintf("expected a JSON file but got: %v", name))
		}
		schemaName := strings.TrimSuffix(name, ".json")
		schema, err := fs.ReadFile(fmt.Sprintf("%s/%s", dir, name))
		if err != nil {
			panic(err)
		}
		err = compiler.AddResource("internal://"+schemaName, bytes.NewReader(schema))
		if err != nil {
			panic(err)
		}
		schemas[schemaName] = compiler.MustCompile("internal://" + schemaName)

		// Store the raw JSON schema in a compact format
		buffer := new(bytes.Buffer)
		if err := json.Compact(buffer, schema); err != nil {
			panic(err)
		}
		rawJSONSchemas[schemaName] = buffer.Bytes()
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

func GetRawJSONSchema(name string) ([]byte, error) {
	once.Do(initSchemas)
	rawJSONSchema, ok := rawJSONSchemas[name]
	if !ok {
		return []byte{}, fmt.Errorf("raw JSON schema not found for entity: '%s'", name)
	}
	return rawJSONSchema, nil
}

func RegisterSchemaFS(fs *embed.FS) {
	schemaFS = append(schemaFS, fs)
}

func registerExtension(c *jsonschema.Compiler, ext extension.Extension) {
	c.RegisterExtension(ext.Name(), ext.Schema(), ext)
}
