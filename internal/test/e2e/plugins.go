package e2e

// FieldUpdateCheck represents the validation or expected value for a given field in a plugin
// configuration.
type FieldUpdateCheck struct {
	// Field is the field to check
	Field string
	// Value is the value of the field
	Value interface{}
}

// VersionCompatibilityPlugins represents a plugin configuration to test and validate.
type VersionCompatibilityPlugins struct {
	// Name is the name of the plugin
	Name string
	// Config is the JSON configuration for the plugin
	Config string
	// Protocols is an array of strings with the protocol names (default: {"http", "https"})
	Protocols []string
	// VersionRange is used to determine when a plugin is not to be expected on a data plane
	VersionRange string
	// FieldUpdateChecks are the values to validate
	FieldUpdateChecks map[string][]FieldUpdateCheck
	// ConfigureForService toggles whether the plugin should be configured for a service
	ConfigureForService bool
	// ConfigureForService toggles whether the plugin should be configured for a route
	ConfigureForRoute bool
	// ExpectedConfig is the expected plugin configuration
	ExpectedConfig string
}

// VersionCompatibilityOSSPluginConfigurationTests are the OSS plugins schemas to test in order to
// validate version compatibility layer.
var VersionCompatibilityOSSPluginConfigurationTests = []VersionCompatibilityPlugins{
	{
		Name: "acl",
		Config: `{
			"allow": [
				"kongers"
			]
		}`,
		ConfigureForService: true,
		ConfigureForRoute:   true,
	},
	// DP < 2.6
	//   - remove 'preferred_chain', 'storage_config.vault.auth_method',
	//     'storage_config.vault.auth_path', 'storage_config.vault.auth_role',
	//     'storage_config.vault.jwt_path'
	//
	// DP < 3.0:
	//   - remove 'allow_any_domain'
	// DP < 3.1:
	//   - remove 'storage_config.redis.ssl*'
	{
		Name: "acme",
		Config: `{
			"account_email": "example@example.com",
			"allow_any_domain": true,
			"storage_config": {
				"redis": {
					"ssl": true,
					"ssl_verify": true,
					"ssl_server_name": "test.com"
				}
			}
		}`,
	},
	// DP < 2.6
	//   - remove 'base64_encode_body'
	//
	// DP < 3.0
	//   - if both 'aws_region' and 'host' are set
	//     just drop 'host' and keep 'aws_region'
	//     since these used to be  mutually exclusive
	//   - remove 'aws_assume_role_arn' and 'aws_role_session_name'
	{
		Name: "aws-lambda",
		Config: `{
			"aws_region": "AWS_REGION",
			"host": "192.168.1.1",
			"function_name": "FUNCTION_NAME",
			"aws_assume_role_arn": "foo",
			"aws_role_session_name": "kong"
		}`,
		ConfigureForService: true,
		ConfigureForRoute:   true,
	},
	{
		Name: "azure-functions",
		Config: `{
			"functionname": "FUNCTIONNAME",
			"appname": "APPNAME"
		}`,
		ConfigureForService: true,
		ConfigureForRoute:   true,
	},
	{
		Name:                "basic-auth",
		ConfigureForService: true,
		ConfigureForRoute:   true,
	},
	{
		Name:                "bot-detection",
		ConfigureForService: true,
		ConfigureForRoute:   true,
	},
	{
		Name:                "correlation-id",
		ConfigureForService: true,
		ConfigureForRoute:   true,
	},
	{
		Name:                "cors",
		ConfigureForService: true,
		ConfigureForRoute:   true,
	},
	{
		Name: "datadog",
		Config: `{
			"service_name_tag": "SERVICE_NAME_TAG",
			"status_tag": "STATUS_TAG",
			"consumer_tag": "CONSUMER_TAG",
			"metrics": [
				{
					"name": "latency",
					"stat_type": "distribution",
					"sample_rate": 1
				}
			]
		}`,
		ConfigureForService: true,
		ConfigureForRoute:   true,
	},
	{
		Name: "file-log",
		Config: `{
			"path": "path/to/file.log"
		}`,
		ConfigureForService: true,
		ConfigureForRoute:   true,
	},
	{
		Name: "grpc-gateway",
		Config: `{
			"proto": "path/to/file.proto"
		}`,
		ConfigureForService: true,
		ConfigureForRoute:   true,
	},
	{
		Name: "grpc-web",
		Config: `{
			"proto": "path/to/file.proto"
		}`,
		ConfigureForService: true,
		ConfigureForRoute:   true,
	},
	{
		Name:                "hmac-auth",
		ConfigureForService: true,
		ConfigureForRoute:   true,
	},
	// DP <= 2.8
	//   - Convert header values from `string` to `[]string`.
	//     e.g.: `"value-1"` -> `[]string{"value-1"}`
	//
	// DP >= 3.0
	//   - Default behavior, use `string` header values as-is.
	{
		Name: "http-log",
		Config: `{
			"http_endpoint": "http://example.com/logs",
			"headers": {
				"header-1": "value-1",
				"header-2": "value-2"
			}
		}`,
		ConfigureForService: true,
		ConfigureForRoute:   true,
	},
	{
		Name: "ip-restriction",
		Config: `{
			"allow": [
				"1.2.3.4"
			],
			"status": 200,
			"message": "MESSAGE"
		}`,
		ConfigureForService: true,
		ConfigureForRoute:   true,
	},
	{
		Name:                "jwt",
		ConfigureForService: true,
		ConfigureForRoute:   true,
	},
	{
		Name:                "key-auth",
		ConfigureForService: true,
		ConfigureForRoute:   true,
	},
	{
		Name: "ldap-auth",
		Config: `{
			"ldap_host": "example.com",
			"ldap_port": 389,
			"base_dn": "dc=example,dc=com",
			"attribute": "cn"
		}`,
		ConfigureForService: true,
		ConfigureForRoute:   true,
	},
	{
		Name: "loggly",
		Config: `{
			"key": "KEY"
		}`,
		ConfigureForService: true,
		ConfigureForRoute:   true,
	},
	{
		Name: "opentelemetry",
		Config: `{
			"endpoint": "http://example.dev"
		}`,
		VersionRange: ">= 3.0.0",
	},
	{
		Name: "post-function",
		Config: `{
			"functions": [
				"kong.log.err('Goodbye Koko!')"
			]
		}`,
		FieldUpdateChecks: map[string][]FieldUpdateCheck{
			">= 3.0.0": {
				{
					Field: "access",
					Value: []string{
						"kong.log.err('Goodbye Koko!')",
					},
				},
			},
		},
		ConfigureForService: true,
		ConfigureForRoute:   true,
	},
	{
		Name: "pre-function",
		Config: `{
			"functions": [
				"kong.log.err('Hello Koko!')"
			]
		}`,
		FieldUpdateChecks: map[string][]FieldUpdateCheck{
			">= 3.0.0": {
				{
					Field: "access",
					Value: []string{
						"kong.log.err('Hello Koko!')",
					},
				},
			},
		},
		ConfigureForService: true,
		ConfigureForRoute:   true,
	},
	// DP < 2.4
	//   - remove 'per_consumer' field (default: false)
	//
	// DP < 3.0
	//   - remove 'status_code_metrics', 'latency_metrics'
	//     'bandwidth_metrics', 'upstream_health_metrics'
	{
		Name: "prometheus",
		Config: `{
			"status_code_metrics": true,
			"latency_metrics": true,
			"bandwidth_metrics": true,
			"upstream_health_metrics": true
		}`,
		ConfigureForService: true,
		ConfigureForRoute:   true,
	},
	{
		Name: "proxy-cache",
		Config: `{
			"strategy": "memory"
		}`,
		ConfigureForService: true,
		ConfigureForRoute:   true,
	},
	// DP < 2.7:
	//   - remove 'redis_ssl', 'redis_ssl_verify', 'redis_server_name' (P108)
	// DP < 2.8:
	//   - remove  'redis_username' (P112)
	// DP < 3.1:
	//   - remove 'error_code', 'error_message' (P137)
	{
		Name: "rate-limiting",
		Config: `{
			"hour": 1,
			"redis_ssl": true,
			"redis_ssl_verify": true,
			"redis_server_name": "redis.example.com",
			"redis_username": "REDIS_USERNAME",
			"redis_password": "REDIS_PASSWORD",
			"error_code": 429,
			"error_message": "API rate limit exceeded"
		}`,
		ConfigureForService: true,
		ConfigureForRoute:   true,
	},
	{
		Name:                "request-size-limiting",
		ConfigureForService: true,
		ConfigureForRoute:   true,
	},
	{
		Name:                "request-termination",
		ConfigureForService: true,
		ConfigureForRoute:   true,
	},
	{
		Name:                "request-transformer",
		ConfigureForService: true,
		ConfigureForRoute:   true,
	},
	// DP < 2.8:
	//   - remove 'redis_username'
	// DP < 3.1:
	//   - remove 'redis_ssl', 'redis_ssl_verify', 'redis_server_name'
	{
		Name: "response-ratelimiting",
		Config: `{
			"limits": {
				"sms": {
					"minute": 20
				}
			},
			"redis_username": "REDIS_USERNAME",
			"redis_ssl": true,
			"redis_ssl_verify": true,
			"redis_server_name": "test.com"
		}`,
		ConfigureForService: true,
		ConfigureForRoute:   true,
	},
	{
		Name:                "response-transformer",
		ConfigureForService: true,
		ConfigureForRoute:   true,
	},
	{
		Name:                "session",
		ConfigureForService: true,
		ConfigureForRoute:   true,
	},
	// DP < 3.0:
	//   - remove 'allow_status_codes', 'udp_packet_size', 'use_tcp',
	//     'hostname_in_prefix', 'consumer_identifier_default',
	//     'service_identifier_default', 'workspace_identifier_default' fields
	//   - remove 'status_count_per_workspace', 'status_count_per_user_per_route',
	//     'shdict_usage' metrics.
	//   - remove 'service_identifier' and 'workspace_identifier' identifiers from metrics
	{
		Name: "statsd",
		Config: `{
			"metrics": [
				{
					"name": "unique_users",
					"stat_type": "set",
					"service_identifier": null,
					"workspace_identifier": null
				},
				{
					"name": "status_count_per_workspace",
					"sample_rate": 1,
					"stat_type": "counter"
				},
				{
					"name": "status_count_per_user_per_route",
					"sample_rate": 1,
					"stat_type": "counter"
				},
				{
					"name": "shdict_usage",
					"sample_rate": 1,
					"stat_type": "gauge"
				}
			],
			"allow_status_codes": ["200-204"],
			"udp_packet_size": 1000,
			"use_tcp": true,
			"hostname_in_prefix": true,
			"consumer_identifier_default": "custom_id",
			"service_identifier_default": "service_name_or_host",
			"workspace_identifier_default": "workspace_id"
		}`,
		ConfigureForService: true,
		ConfigureForRoute:   true,
	},
	{
		Name:                "syslog",
		ConfigureForService: true,
		ConfigureForRoute:   true,
	},
	{
		Name: "tcp-log",
		Config: `{
			"host": "localhost",
			"port": 1234
		}`,
		ConfigureForService: true,
		ConfigureForRoute:   true,
	},
	{
		Name: "udp-log",
		Config: `{
			"host": "localhost",
			"port": 1234
		}`,
		ConfigureForService: true,
		ConfigureForRoute:   true,
	},
	// DP < 2.4:
	//   - remove 'tags_header' field (default: Zipkin-Tags)
	//
	// DP < 2.7:
	//   - change 'header_type' from 'ignore' to 'preserve'
	//
	// DP < 3.0:
	//   - remove 'http_span_name', 'connect_timeout'
	//     'send_timeout', 'read_timeout',
	// DP < 3.1:
	//   - remove 'http_response_header_for_traceid'
	{
		Name: "zipkin",
		Config: `{
			"local_service_name": "LOCAL_SERVICE_NAME",
			"header_type": "ignore",
			"http_span_name": "method_path",
			"connect_timeout": 2001,
			"send_timeout": 2001,
			"read_timeout": 2001,
			"http_response_header_for_traceid": "X-B3-TraceId"
		}`,
		FieldUpdateChecks: map[string][]FieldUpdateCheck{
			"< 2.7.0": {
				{
					Field: "header_type",
					Value: "preserve",
				},
			},
		},
		ConfigureForService: true,
		ConfigureForRoute:   true,
	},
}
