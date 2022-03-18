package compat

import "github.com/kong/koko/internal/server/kong/ws/config"

var PluginConfigTableUpdates = map[uint64][]config.ConfigTableUpdates{
	2003003003: {
		{
			Name: "file-log",
			Type: config.Plugin,
			Fields: []string{
				"custom_fields_by_lua",
			},
		},
		{
			Name: "http-log",
			Type: config.Plugin,
			Fields: []string{
				"custom_fields_by_lua",
			},
		},
		{
			Name: "loggly",
			Type: config.Plugin,
			Fields: []string{
				"custom_fields_by_lua",
			},
		},
		{
			Name: "syslog",
			Type: config.Plugin,
			Fields: []string{
				"custom_fields_by_lua",
			},
		},
		{
			Name: "tcp-log",
			Type: config.Plugin,
			Fields: []string{
				"custom_fields_by_lua",
			},
		},
		{
			Name: "udp-log",
			Type: config.Plugin,
			Fields: []string{
				"custom_fields_by_lua",
			},
		},
	},
	2003999999: {
		{
			Name: "prometheus",
			Type: config.Plugin,
			Fields: []string{
				"per_consumer",
			},
		},
		{
			Name: "zipkin",
			Type: config.Plugin,
			Fields: []string{
				"tags_header",
			},
		},
	},
	2004001002: {
		{
			Name: "syslog",
			Type: config.Plugin,
			Fields: []string{
				"facility",
			},
		},
	},
	2005999999: {
		{
			Name: "acme",
			Type: config.Plugin,
			Fields: []string{
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
			Fields: []string{
				"base64_encode_body",
			},
		},
		{
			Name: "grpc-web",
			Type: config.Plugin,
			Fields: []string{
				"allow_origin_header",
			},
		},
		{
			Name: "request-termination",
			Type: config.Plugin,
			Fields: []string{
				"echo",
				"trigger",
			},
		},
	},
	2006999999: {
		{
			Name: "datadog",
			Type: config.Plugin,
			Fields: []string{
				"service_name_tag",
				"status_tag",
				"consumer_tag",
			},
		},
		{
			Name: "ip-restriction",
			Type: config.Plugin,
			Fields: []string{
				"status",
				"message",
			},
		},
		{
			Name: "rate-limiting",
			Type: config.Plugin,
			Fields: []string{
				"redis_ssl",
				"redis_ssl_verify",
				"redis_server_name",
			},
		},
		{
			Name: "zipkin",
			Type: config.Plugin,
			Fields: []string{
				"local_service_name",
			},
		},
	},
	2007999999: {
		{
			Name: "acme",
			Type: config.Plugin,
			Fields: []string{
				"rsa_key_size",
			},
		},
		{
			Name: "rate-limiting",
			Type: config.Plugin,
			Fields: []string{
				"redis_username",
			},
		},
		{
			Name: "response-ratelimiting",
			Type: config.Plugin,
			Fields: []string{
				"redis_username",
			},
		},
	},
}
