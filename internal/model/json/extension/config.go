package extension

import (
	"github.com/mitchellh/mapstructure"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

var configSchema = jsonschema.MustCompileString((&Config{}).Name(), `{
	"properties" : {
		"hasValidateEndpoint": {
			"type": "boolean"
		}
	}
}`)

// Config defines an internal-only vendor extension used to power
// various business logic around our JSON schema entities.
//
// End-users do not see this extension when issuing calls to get a
// schema for a specific entity. As such, no properties on this
// extension shall be required.
type Config struct {
	// When this config extension is redacted on the schema or this value is false, the underlining
	// JSON schema entity is expected to have its own validation endpoint. When left true, this hints
	// to our tests to skip asserting that a validation endpoint should exist for that entity.
	//
	// For example, when this value is false, for our `Consumer`
	// entity, it is expected that the below HTTP endpoint exists:
	// POST /v1/schemas/json/consumer/validate
	DisableValidateEndpoint bool `json:"disableValidateEndpoint"`

	// The resource's object name used in the REST API path. For example, for our `Consumer`
	// entity, this should be set to "consumers" as the REST path is `/v1/consumers`.
	//
	// This should be left empty in the event the resource is not exposed in the REST API.
	ResourceAPIPath string `json:"resourceAPIPath,omitempty"`
}

// Name implements the Extension interface.
func (c *Config) Name() string { return "x-koko-config" }

// Schema implements the Extension interface.
func (c *Config) Schema() *jsonschema.Schema { return configSchema }

// Compile implements the jsonschema.ExtCompiler interface.
func (c *Config) Compile(_ jsonschema.CompilerContext, m map[string]interface{}) (jsonschema.ExtSchema, error) {
	configIface, ok := m[c.Name()]
	if !ok {
		// There's nothing to compile, no-op.
		return nil, nil
	}

	// Covert the map[string]interface{} into a config struct.
	var config Config
	return &config, mapstructure.Decode(configIface, &config)
}

// Validate implements the jsonschema.ExtSchema interface.
func (c *Config) Validate(jsonschema.ValidationContext, interface{}) error {
	// No special validation needs to happen right now.
	return nil
}
