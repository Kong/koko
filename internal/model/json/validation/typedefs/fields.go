package typedefs

import (
	"fmt"

	"github.com/kong/koko/internal/model/json/generator"
)

const (
	maxNameLength     = 128
	maxPort           = 65535
	maxTimeout        = (1 << 31) - 2 //nolint:gomnd
	maxTags           = 8
	namePattern       = `^[0-9a-zA-Z.\-_~]*$`
	maxHostnameLength = 256
	maxPathLength     = 1024
	hostnamePattern   = "[a-zA-Z0-9]([-a-zA-Z0-9]*[a-zA-Z0-9])?(.[a-zA-Z0-9]([-a-zA-Z0-9]*[a-zA-Z0-9])?)*$"

	DefaultTimeout = 60000

	// Allow alphanumerics, dashes, underscores, colons, periods, and tildes anywhere
	// within the tag. Spaces are allowed, but not as the first/last character.
	//
	// This disallows tags like: ` `, `  `, ` tag `, `  tag  `, etc.
	tagPattern = `^(?:[0-9a-zA-Z.\-_~:]+(?: *[0-9a-zA-Z.\-_~:])*)?$`

	HTTPHeaderNamePattern = "^[A-Za-z0-9!#$%&'*+-.^_|~]{1,64}$"
	ReferencePattern      = "^{vault://.*}$"
)

var (
	HostnamePattern        = fmt.Sprintf("^%s", hostnamePattern)
	WilcardHostnamePattern = fmt.Sprintf("^(\\*\\.)?%s", hostnamePattern)
)

var ID = &generator.Schema{
	Description: "must be a valid UUID",
	Type:        "string",
	Pattern:     "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$",
}

var falsy = false

var ReferenceObject = &generator.Schema{
	Type: "object",
	Properties: map[string]*generator.Schema{
		"id": ID,
	},
	Required:             []string{"id"},
	AdditionalProperties: &falsy,
}

var Name = &generator.Schema{
	Type:      "string",
	Pattern:   namePattern,
	MinLength: 1,
	MaxLength: maxNameLength,
}

var Timeout = &generator.Schema{
	Type:    "integer",
	Minimum: intP(1),
	Maximum: maxTimeout,
	Default: DefaultTimeout,
}

var Protocol = &generator.Schema{
	Type: "string",
	Enum: []interface{}{
		ProtocolHTTP,
		ProtocolHTTPS,
		ProtocolGRPC,
		ProtocolGRPCS,
		ProtocolTCP,
		ProtocolUDP,
		ProtocolTLS,
		ProtocolTLSPassthrough,
	},
}

var AllProtocols = &generator.Schema{
	Type: "string",
	Enum: []interface{}{
		ProtocolHTTP,
		ProtocolHTTPS,
		ProtocolGRPC,
		ProtocolGRPCS,
		ProtocolTCP,
		ProtocolUDP,
		ProtocolTLS,
		ProtocolTLSPassthrough,
		ProtocolWS,
		ProtocolWSS,
	},
}

var Host = &generator.Schema{
	Description: "must be a valid hostname",
	Type:        "string",
	MaxLength:   maxHostnameLength,
	Pattern:     HostnamePattern,
}

var WilcardHost = &generator.Schema{
	Description: "must be a valid hostname with a wildcard prefix '*' or without",
	Type:        "string",
	MaxLength:   maxHostnameLength,
	Pattern:     WilcardHostnamePattern,
}

var Port = &generator.Schema{
	Type:    "integer",
	Minimum: intP(1),
	Maximum: maxPort,
}

var UnixEpoch = &generator.Schema{
	Type:    "integer",
	Minimum: intP(1),
}

var tag = &generator.Schema{
	Type:      "string",
	Pattern:   tagPattern,
	MinLength: 1,
	MaxLength: maxNameLength,
}

var Tags = &generator.Schema{
	Type:        "array",
	Items:       tag,
	MaxItems:    maxTags,
	UniqueItems: true,
}

var Header = &generator.Schema{
	Type:    "string",
	Pattern: HTTPHeaderNamePattern,
}

var Path = &generator.Schema{
	Type: "string",
	AllOf: []*generator.Schema{
		{
			Description: "must begin with `/`",
			Pattern:     "^/.*",
		},
		{
			Description: fmt.Sprintf("length must not exceed %d", maxPathLength),
			MaxLength:   maxPathLength,
		},
		{
			Not: &generator.Schema{
				Description: "must not contain `//`",
				Pattern:     "//",
			},
		},
	},
}

var RouterPath = &generator.Schema{
	Type: "string",
	AllOf: []*generator.Schema{
		{
			Description: "must begin with `/` (prefix path) or `~/` (regex path)",
			Pattern:     "^/.*|^~/.*",
		},
		{
			Description: fmt.Sprintf("length must not exceed %d", maxPathLength),
			MaxLength:   maxPathLength,
		},
		{
			Not: &generator.Schema{
				Description: "must not contain `//`",
				Pattern:     "//",
			},
		},
	},
}

var CIDRPort = &generator.Schema{
	Type: "object",
	Properties: map[string]*generator.Schema{
		"ip": {
			Type: "string",
			AnyOf: []*generator.Schema{
				// TODO(hbagdi): add ipv6
				{
					Description: "must be a valid IP or CIDR",
					// TODO(hbagdi): replace with a stricter matcher
					Pattern: "^([0-9]{1,3}[.]{1}){3}[0-9]{1,3}$",
				},
				{
					Description: "must be a valid IP or CIDR",
					// TODO(hbagdi): replace with a stricter matcher
					Pattern: "^([0-9]{1,3}[.]{1}){3}[0-9]{1,3}/[0-9]{1,3}$",
				},
			},
		},
		"port": Port,
	},
	AnyOf: []*generator.Schema{
		{
			Description: "at least one of 'ip' or 'port' is required",
			Required:    []string{"ip"},
		},
		{
			Description: "at least one of 'ip' or 'port' is required",
			Required:    []string{"port"},
		},
	},
}

var Reference = &generator.Schema{
	Type:        "string",
	Description: "referenceable field must contain a valid 'Reference'",
	Pattern:     ReferencePattern,
}

var PEMKey = &generator.Schema{
	Type: "object",
	Properties: map[string]*generator.Schema{
		"private_key": {Format: "pem-encoded-private-key"},
		"public_key":  {Format: "pem-encoded-public-key"},
	},
	Required: []string{"private_key"},
}

var JWKKey = &generator.Schema{
	Type: "string",
}

func intP(i int) *int { return &i }
