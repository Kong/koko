package schema

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strings"
	"sync"

	internalCrypto "github.com/kong/koko/internal/crypto"
	genJSONSchema "github.com/kong/koko/internal/gen/jsonschema"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

var (
	schemas        = map[string]*jsonschema.Schema{}
	rawJSONSchemas = map[string][]byte{}
	once           sync.Once
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

	initCustomFormats()

	compiler := jsonschema.NewCompiler()
	compiler.ExtractAnnotations = true
	compiler.AssertFormat = true
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

		// Store the raw JSON schema in a compact format
		buffer := new(bytes.Buffer)
		if err := json.Compact(buffer, schema); err != nil {
			panic(err)
		}
		rawJSONSchemas[schemaName] = buffer.Bytes()
	}
}

func initCustomFormats() {
	jsonschema.Formats["pem-encoded-cert"] = func(v interface{}) bool {
		switch v := v.(type) {
		case string:
			_, err := internalCrypto.ParsePEMCert([]byte(v))
			return err == nil
		default:
			return false
		}
	}
	jsonschema.Formats["pem-encoded-private-key"] = func(v interface{}) bool {
		switch v := v.(type) {
		case string:
			block, _ := pem.Decode([]byte(v))
			if block == nil {
				return false
			}
			_, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err == nil {
				return true
			}
			_, err = x509.ParsePKCS1PrivateKey(block.Bytes)
			if err == nil {
				return true
			}
			_, err = x509.ParseECPrivateKey(block.Bytes)
			if err == nil {
				return true
			}
			return false
		default:
			return false
		}
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
