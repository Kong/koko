package extension

import (
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// Extension defines a simple interface to build out JSON schema extensions.
type Extension interface {
	jsonschema.ExtCompiler

	// Name returns the name of the extension. This should match
	// the property named used on the JSON schema itself.
	Name() string

	// Schema returns the JSON schema for the extension.
	Schema() *jsonschema.Schema
}
