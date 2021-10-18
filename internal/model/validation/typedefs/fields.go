package typedefs

import (
	"github.com/kong/koko/internal/model/json/generator"
)

const (
	maxNameLength = 128
	maxPort       = 65535
	maxTimeout    = (1 << 31) - 2 //nolint:gomnd
	maxTags       = 8
)

var ID = &generator.Schema{
	Description: "must be a valid UUID",
	Type:        "string",
	Pattern:     "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$",
}

var Name = &generator.Schema{
	Type:      "string",
	Pattern:   "^[0-9a-zA-Z.-_~]*$",
	MinLength: 1,
	MaxLength: maxNameLength,
}

var Timeout = &generator.Schema{
	Type:    "integer",
	Minimum: 1,
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
	Pattern:     "^[a-z0-9]([-a-z0-9]*[a-z0-9])?(.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$",
}

var Port = &generator.Schema{
	Type:    "integer",
	Minimum: 1,
	Maximum: maxPort,
}

var Tags = &generator.Schema{
	Type:     "array",
	Items:    Name,
	MaxItems: maxTags,
}

var Path = &generator.Schema{
	Type: "string",
	AllOf: []*generator.Schema{
		{
			Description: "must begin with `/`",
			Pattern:     "^/.*",
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
			OneOf: []*generator.Schema{
				// TODO(hbagdi): add ipv6
				{
					Description: "must be a valid IP or CIDR",
					// TODO(hbagdi): replace with a stricter matcher
					Pattern: "^(?:[0-9]{1,3}.){3}[0-9]{1,3}$",
				},
				{
					Description: "must be a valid IP or CIDR",
					// TODO(hbagdi): replace with a stricter matcher
					Pattern: "^(?:[0-9]{1,3}.){3}[0-9]{1,3}/[0-9]{1,3}$",
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
