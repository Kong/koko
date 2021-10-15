package typedefs

import (
	"github.com/kong/koko/internal/model/json"
)

const (
	maxNameLength = 128
	maxPort       = 65535
	maxTimeout    = (1 << 31) - 2 //nolint:gomnd
	maxTags       = 8
)

var ID = &json.Schema{
	Type:    "string",
	Pattern: "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$",
}

var Name = &json.Schema{
	Type:      "string",
	Pattern:   "^[0-9a-zA-Z.-_~]*$",
	MinLength: 1,
	MaxLength: maxNameLength,
}

var Timeout = &json.Schema{
	Type:    "integer",
	Minimum: 1,
	Maximum: maxTimeout,
}

var Protocol = &json.Schema{
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

var Host = &json.Schema{
	Type:    "string",
	Pattern: "^[a-z0-9]([-a-z0-9]*[a-z0-9])?(.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$",
}

var Port = &json.Schema{
	Type:    "integer",
	Minimum: 1,
	Maximum: maxPort,
}

var Tags = &json.Schema{
	Type:      "array",
	Items:     Name,
	MaxLength: maxTags,
}

var Path = &json.Schema{
	Type: "string",
	AllOf: []*json.Schema{
		{
			Pattern: "^/.*",
		},
		{
			Not: &json.Schema{
				Pattern: "//",
			},
		},
	},
}

var CIDRPort = &json.Schema{
	Type: "object",
	Properties: map[string]*json.Schema{
		"ip": {
			Type: "string",
			OneOf: []*json.Schema{
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
	OneOf: []*json.Schema{
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
