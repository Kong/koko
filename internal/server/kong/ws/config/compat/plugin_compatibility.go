package compat

import "github.com/kong/koko/internal/server/kong/ws/config"

var PluginConfigTableUpdates = map[uint64][]config.ConfigTableUpdates{
	2003003003: {
		{
			Name: "file-log",
			Type: config.Plugin,
			RemoveFields: []string{
				"custom_fields_by_lua",
			},
		},
		{
			Name: "http-log",
			Type: config.Plugin,
			RemoveFields: []string{
				"custom_fields_by_lua",
			},
		},
		{
			Name: "loggly",
			Type: config.Plugin,
			RemoveFields: []string{
				"custom_fields_by_lua",
			},
		},
		{
			Name: "syslog",
			Type: config.Plugin,
			RemoveFields: []string{
				"custom_fields_by_lua",
			},
		},
		{
			Name: "tcp-log",
			Type: config.Plugin,
			RemoveFields: []string{
				"custom_fields_by_lua",
			},
		},
		{
			Name: "udp-log",
			Type: config.Plugin,
			RemoveFields: []string{
				"custom_fields_by_lua",
			},
		},
	},
	2003999999: {
		{
			Name: "prometheus",
			Type: config.Plugin,
			RemoveFields: []string{
				"per_consumer",
			},
		},
		{
			Name: "zipkin",
			Type: config.Plugin,
			RemoveFields: []string{
				"tags_header",
			},
		},
	},
	2004001002: {
		{
			Name: "syslog",
			Type: config.Plugin,
			RemoveFields: []string{
				"facility",
			},
		},
	},
	2005999999: {
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
	2006999999: {
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
	2007999999: {
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
	3000000000: {
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
				"lantency_metrics",
				"bandwidth_metrics",
				"upstream_health_metrics",
			},
		},
	},
}
