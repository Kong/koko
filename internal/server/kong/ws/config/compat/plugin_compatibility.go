package compat

import "github.com/kong/koko/internal/server/kong/ws/config"

var PluginConfigTableUpdates = map[string][]config.ConfigTableUpdates{
	"< 2.6.0": {
		{
			Name: "acme",
			Type: config.Plugin,
			RemoveFields: []string{
				"preferred_chain",
				"storage_config.vault.auth_method",
				"storage_config.vault.auth_path",
				"storage_config.vault.auth_role",
				"storage_config.vault.jwt_path",
			},
		},
		{
			Name: "aws-lambda",
			Type: config.Plugin,
			RemoveFields: []string{
				"base64_encode_body",
			},
		},
		{
			Name: "grpc-web",
			Type: config.Plugin,
			RemoveFields: []string{
				"allow_origin_header",
			},
		},
		{
			Name: "request-termination",
			Type: config.Plugin,
			RemoveFields: []string{
				"echo",
				"trigger",
			},
		},
	},
	"< 2.7.0": {
		{
			Name: "datadog",
			Type: config.Plugin,
			RemoveFields: []string{
				"service_name_tag",
				"status_tag",
				"consumer_tag",
			},
			RemoveElementsFromArray: []config.ConfigTableFieldCondition{
				{
					Field:     "metrics",
					Condition: "stat_type=distribution",
				},
			},
		},
		{
			Name: "ip-restriction",
			Type: config.Plugin,
			RemoveFields: []string{
				"status",
				"message",
			},
		},
		{
			Name: "rate-limiting",
			Type: config.Plugin,
			RemoveFields: []string{
				"redis_ssl",
				"redis_ssl_verify",
				"redis_server_name",
			},
		},
		{
			Name: "zipkin",
			Type: config.Plugin,
			RemoveFields: []string{
				"local_service_name",
			},
			FieldUpdates: []config.ConfigTableFieldCondition{
				{
					Field:     "header_type",
					Condition: "header_type=ignore",
					Updates: []config.ConfigTableFieldUpdate{
						{
							Field: "header_type",
							Value: "preserve",
						},
					},
				},
			},
		},
	},
	"< 2.8.0": {
		{
			Name: "acme",
			Type: config.Plugin,
			RemoveFields: []string{
				"rsa_key_size",
			},
		},
		{
			Name: "rate-limiting",
			Type: config.Plugin,
			RemoveFields: []string{
				"redis_username",
			},
		},
		{
			Name: "response-ratelimiting",
			Type: config.Plugin,
			RemoveFields: []string{
				"redis_username",
			},
		},
	},
	"< 3.0.0": {
		{
			Name:   "opentelemetry",
			Type:   config.Plugin,
			Remove: true,
		},
		{
			Name: "zipkin",
			Type: config.Plugin,
			RemoveFields: []string{
				"http_span_name",
				"connect_timeout",
				"send_timeout",
				"read_timeout",
			},
		},
		{
			Name: "prometheus",
			Type: config.Plugin,
			RemoveFields: []string{
				"status_code_metrics",
				"latency_metrics",
				"bandwidth_metrics",
				"upstream_health_metrics",
			},
		},
		{
			Name: "acme",
			Type: config.Plugin,
			RemoveFields: []string{
				"allow_any_domain",
			},
		},
	},
	">= 3.0.0": {
		{
			Name: "aws-lambda",
			Type: config.Plugin,
			RemoveFields: []string{
				"proxy_scheme",
			},
		},
	},
}
