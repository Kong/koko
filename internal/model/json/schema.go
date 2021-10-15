package json

// Schema represents a JSON Schema object type.
type Schema struct {
	// RFC draft-wright-json-schema-00
	Version string `json:"$schema,omitempty"` // section 6.1
	Ref     string `json:"$ref,omitempty"`    // section 7
	// RFC draft-wright-json-schema-validation-00, section 5
	MultipleOf           int                `json:"multipleOf,omitempty"`           // section 5.1
	Maximum              int                `json:"maximum,omitempty"`              // section 5.2
	ExclusiveMaximum     bool               `json:"exclusiveMaximum,omitempty"`     // section 5.3
	Minimum              int                `json:"minimum,omitempty"`              // section 5.4
	ExclusiveMinimum     bool               `json:"exclusiveMinimum,omitempty"`     // section 5.5
	MaxLength            int                `json:"maxLength,omitempty"`            // section 5.6
	MinLength            int                `json:"minLength,omitempty"`            // section 5.7
	Pattern              string             `json:"pattern,omitempty"`              // section 5.8
	AdditionalItems      *Schema            `json:"additionalItems,omitempty"`      // section 5.9
	Items                *Schema            `json:"items,omitempty"`                // section 5.9
	MaxItems             int                `json:"maxItems,omitempty"`             // section 5.10
	MinItems             int                `json:"minItems,omitempty"`             // section 5.11
	UniqueItems          bool               `json:"uniqueItems,omitempty"`          // section 5.12
	MaxProperties        int                `json:"maxProperties,omitempty"`        // section 5.13
	MinProperties        int                `json:"minProperties,omitempty"`        // section 5.14
	Required             []string           `json:"required,omitempty"`             // section 5.15
	Properties           map[string]*Schema `json:"properties,omitempty"`           // section 5.16
	PatternProperties    map[string]*Schema `json:"patternProperties,omitempty"`    // section 5.17
	AdditionalProperties bool               `json:"additionalProperties,omitempty"` // section 5.18
	Dependencies         map[string]*Schema `json:"dependencies,omitempty"`         // section 5.19
	Enum                 []interface{}      `json:"enum,omitempty"`                 // section 5.20
	Type                 string             `json:"type,omitempty"`                 // section 5.21
	AllOf                []*Schema          `json:"allOf,omitempty"`                // section 5.22
	AnyOf                []*Schema          `json:"anyOf,omitempty"`                // section 5.23
	OneOf                []*Schema          `json:"oneOf,omitempty"`                // section 5.24
	Not                  *Schema            `json:"not,omitempty"`                  // section 5.25
	Definitions          map[string]*Schema `json:"definitions,omitempty"`          // section 5.26
	// RFC draft-wright-json-schema-validation-00, section 6, 7
	Title       string        `json:"title,omitempty"`       // section 6.1
	Description string        `json:"description,omitempty"` // section 6.1
	Default     interface{}   `json:"default,omitempty"`     // section 6.2
	Format      string        `json:"format,omitempty"`      // section 7
	Examples    []interface{} `json:"examples,omitempty"`    // section 7.4
	// RFC draft-wright-json-schema-hyperschema-00, section 4
	Media          *Schema `json:"media,omitempty"`          // section 4.3
	BinaryEncoding string  `json:"binaryEncoding,omitempty"` // section 4.3

	Extras map[string]interface{} `json:"-"`

	// Added by hbagdi
	If       *Schema     `json:"if,omitempty"`
	Then     *Schema     `json:"then,omitempty"`
	Else     *Schema     `json:"else,omitempty"`
	Const    interface{} `json:"const,omitempty"`
	Contains *Schema     `json:"contains,omitempty"`
}
