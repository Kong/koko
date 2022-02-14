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

	HTTPHeaderNamePattern = "^[A-Za-z0-9!#$%&'*+-.^_|~]{1,64}$"
	hostnamePattern       = "^[a-z0-9]([-a-z0-9]*[a-z0-9])?(.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$"
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
	},
}

var Host = &generator.Schema{
	Description: "must be a valid hostname",
	Type:        "string",
	MaxLength:   maxHostnameLength,
	Pattern:     hostnamePattern,
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

var Tags = &generator.Schema{
	Type:     "array",
	Items:    Name,
	MaxItems: maxTags,
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
	OneOf: []*generator.Schema{
		{
			Description: "either one of 'ip' or 'port' is required",
			Required:    []string{"ip"},
		},
		{
			Description: "either one of 'ip' or 'port' is required",
			Required:    []string{"port"},
		},
	},
}

func intP(i int) *int { return &i }
