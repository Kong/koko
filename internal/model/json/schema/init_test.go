package schema

import (
	"embed"
	"testing"

	genJSONSchema "github.com/kong/koko/internal/gen/jsonschema"
	"github.com/kong/koko/internal/model/json/schema/testdata"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/stretchr/testify/require"
)

func resetSchemaFS() {
	schemaFS = []*embed.FS{&genJSONSchema.KongSchemas}
	initSchemas()
}

func TestRegisterSchemaFS(t *testing.T) {
	resetSchemaFS()

	schema, err := Get("test_entity")
	require.ErrorContains(t, err, "schema not found")
	require.Nil(t, schema)

	RegisterSchemasFromFS(&testdata.TestKongSchemas)
	// make sure to reload schemas
	initSchemas()

	schema, err = Get("test_entity")
	require.IsType(t, schema, &jsonschema.Schema{})
	require.NoError(t, err)
}

func TestSchemaLoadingOverride(t *testing.T) {
	resetSchemaFS()

	// make sure the oss service schema is loaded by default
	schema, _ := Get("service")
	require.Greater(t, len(schema.Properties), 1) // default has 19 properties
	RegisterSchemasFromFS(&testdata.TestKongSchemas)
	initSchemas()

	// check that the new service schema overrides the default one
	schema, _ = Get("service")
	require.Len(t, schema.Properties, 1)
}
