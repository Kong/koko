package extension

import (
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Compile(t *testing.T) {
	ext := &Config{}
	c := jsonschema.NewCompiler()
	c.RegisterExtension(ext.Name(), ext.Schema(), ext)

	require.NoError(t, c.AddResource("schema.json", strings.NewReader(`{
		"x-koko-config": {"disableValidateEndpoint": true}
	}`)))

	sch, err := c.Compile("schema.json")
	require.NoError(t, err)

	assert.Equal(t, map[string]jsonschema.ExtSchema{
		ext.Name(): &Config{DisableValidateEndpoint: true},
	}, sch.Extensions)
}

func TestConfig_Name(t *testing.T) {
	assert.Equal(t, "x-koko-config", (&Config{}).Name())
}

func TestConfig_Schema(t *testing.T) {
	assert.Exactly(t, configSchema, (&Config{}).Schema())
}
