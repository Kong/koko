package compat

import (
	"testing"

	"github.com/kong/koko/internal/log"
	"github.com/kong/koko/internal/server/kong/ws/config"
	"github.com/stretchr/testify/require"
)

func TestDisableChangeTracking(t *testing.T) {
	tests := []struct {
		name                string
		uncompressedPayload string
		dataPlaneVersion    string
		expectedPayload     string
		expectedChanges     config.TrackedChanges
	}{
		{
			name: "[request-termination] change is not emitted with default" +
				" values",
			uncompressedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "request-termination",
				"config": {
					"echo": false
				}
			}
		]
	}
}
`,
			dataPlaneVersion: "2.5.0",
			expectedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "request-termination",
				"config": {
				}
			}
		]
	}
}
`,
			expectedChanges: config.TrackedChanges{},
		},
		{
			name: "[request-termination] change is emitted with non-default" +
				" value of 'config.echo'",
			uncompressedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "request-termination",
				"config": {
					"echo": true
				}
			}
		]
	}
}
`,
			dataPlaneVersion: "2.5.0",
			expectedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "request-termination",
				"config": {
				}
			}
		]
	}
}
`,
			expectedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: "P104",
						Resources: []config.ResourceInfo{
							{
								Type: "plugin",
								ID:   "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							},
						},
					},
				},
			},
		},
		{
			name: "[request-termination] change is emitted with non-default" +
				" value of 'config.trigger'",
			uncompressedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "request-termination",
				"config": {
					"trigger": "foo-header-trigger"
				}
			}
		]
	}
}
`,
			dataPlaneVersion: "2.5.0",
			expectedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "request-termination",
				"config": {
				}
			}
		]
	}
}
`,
			expectedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: "P104",
						Resources: []config.ResourceInfo{
							{
								Type: "plugin",
								ID:   "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							},
						},
					},
				},
			},
		},
		{
			name: "[aws-lambda] change is emitted with non-default" +
				" value of 'config.base64_encode_body'",
			uncompressedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "aws-lambda",
				"config": {
					"base64_encode_body": false
				}
			}
		]
	}
}
`,
			dataPlaneVersion: "2.5.0",
			expectedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "aws-lambda",
				"config": {
				}
			}
		]
	}
}
`,
			expectedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: "P102",
						Resources: []config.ResourceInfo{
							{
								Type: "plugin",
								ID:   "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							},
						},
					},
				},
			},
		},
		{
			name: "[aws-lambda] change is not emitted with default" +
				" value of 'config.base64_encode_body'",
			uncompressedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "aws-lambda",
				"config": {
					"base64_encode_body": true
				}
			}
		]
	}
}
`,
			dataPlaneVersion: "2.5.0",
			expectedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "aws-lambda",
				"config": {
				}
			}
		]
	}
}
`,
			expectedChanges: config.TrackedChanges{},
		},
		{
			name: "[aws-lambda] change is not emitted with no value " +
				"for 'config.base64_encode_body'",
			uncompressedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "aws-lambda",
				"config": {
				}
			}
		]
	}
}
`,
			dataPlaneVersion: "2.5.0",
			expectedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "aws-lambda",
				"config": {
				}
			}
		]
	}
}
`,
			expectedChanges: config.TrackedChanges{},
		},
		{
			name: "[grpc-web] change is not emitted with default" +
				"value of '*' for 'config.allow_origin_header'",
			uncompressedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "grpc-web",
				"config": {
					"allow_origin_header": "*"
				}
			}
		]
	}
}
`,
			dataPlaneVersion: "2.5.0",
			expectedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "grpc-web",
				"config": {
				}
			}
		]
	}
}
`,
			expectedChanges: config.TrackedChanges{},
		},
		{
			name: "[grpc-web] change is emitted with non-default" +
				"value for 'config.allow_origin_header'",
			uncompressedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "grpc-web",
				"config": {
					"allow_origin_header": "foo.com"
				}
			}
		]
	}
}
`,
			dataPlaneVersion: "2.5.0",
			expectedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "grpc-web",
				"config": {
				}
			}
		]
	}
}
`,
			expectedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: "P103",
						Resources: []config.ResourceInfo{
							{
								Type: "plugin",
								ID:   "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							},
						},
					},
				},
			},
		},
		{
			name: "[datadog] change is not emitted with default" +
				"value for all tags",
			uncompressedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "datadog",
				"config": {
					"service_name_tag": "name",
					"consumer_tag": "consumer",
					"status_tag": "status"
				}
			}
		]
	}
}
`,
			dataPlaneVersion: "2.6.0",
			expectedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "datadog",
				"config": {
				}
			}
		]
	}
}
`,
			expectedChanges: config.TrackedChanges{},
		},
		{
			name: "[datadog] change is not emitted with no tags",
			uncompressedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "datadog",
				"config": {
				}
			}
		]
	}
}
`,
			dataPlaneVersion: "2.6.0",
			expectedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "datadog",
				"config": {
				}
			}
		]
	}
}
`,
			expectedChanges: config.TrackedChanges{},
		},
		{
			name: "[datadog] change is emitted if service_tag is not default",
			uncompressedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "datadog",
				"config": {
					"service_name_tag": "service_name",
					"consumer_tag": "consumer",
					"status_tag": "status"
				}
			}
		]
	}
}
`,
			dataPlaneVersion: "2.6.0",
			expectedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "datadog",
				"config": {
				}
			}
		]
	}
}
`,
			expectedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: "P105",
						Resources: []config.ResourceInfo{
							{
								Type: "plugin",
								ID:   "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							},
						},
					},
				},
			},
		},
		{
			name: "[datadog] change is emitted if consumer_tag is not default",
			uncompressedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "datadog",
				"config": {
					"service_name_tag": "name",
					"consumer_tag": "kong_consumer",
					"status_tag": "status"
				}
			}
		]
	}
}
`,
			dataPlaneVersion: "2.6.0",
			expectedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "datadog",
				"config": {
				}
			}
		]
	}
}
`,
			expectedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: "P105",
						Resources: []config.ResourceInfo{
							{
								Type: "plugin",
								ID:   "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							},
						},
					},
				},
			},
		},
		{
			name: "[datadog] change is emitted if status_tag is not default",
			uncompressedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "datadog",
				"config": {
					"service_name_tag": "name",
					"consumer_tag": "consumer",
					"status_tag": "kong_http_status"
				}
			}
		]
	}
}
`,
			dataPlaneVersion: "2.6.0",
			expectedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "datadog",
				"config": {
				}
			}
		]
	}
}
`,
			expectedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: "P105",
						Resources: []config.ResourceInfo{
							{
								Type: "plugin",
								ID:   "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							},
						},
					},
				},
			},
		},
		{
			name: "[rate-limiting] change is not emitted when new fields are" +
				" set to default ",
			uncompressedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "rate-limiting",
				"config": {
					"second": 10,
					"redis_ssl": false,
					"redis_ssl_verify": false,
					"redis_server_name": null
				}
			}
		]
	}
}
`,
			dataPlaneVersion: "2.6.0",
			expectedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "rate-limiting",
				"config": {
					"second": 10
				}
			}
		]
	}
}
`,
			expectedChanges: config.TrackedChanges{},
		},
		{
			name: "[rate-limiting] change is emitted with non-default value" +
				" for redis_ssl",
			uncompressedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "rate-limiting",
				"config": {
					"second": 10,
					"redis_ssl": true,
					"redis_ssl_verify": false,
					"redis_server_name": null
				}
			}
		]
	}
}
`,
			dataPlaneVersion: "2.6.0",
			expectedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "rate-limiting",
				"config": {
					"second": 10
				}
			}
		]
	}
}
`,
			expectedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: "P108",
						Resources: []config.ResourceInfo{
							{
								Type: "plugin",
								ID:   "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							},
						},
					},
				},
			},
		},
		{
			name: "[rate-limiting] change is emitted with non-default value" +
				" for redis_ssl_verify",
			uncompressedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "rate-limiting",
				"config": {
					"second": 10,
					"redis_ssl": false,
					"redis_ssl_verify": true,
					"redis_server_name": null
				}
			}
		]
	}
}
`,
			dataPlaneVersion: "2.6.0",
			expectedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "rate-limiting",
				"config": {
					"second": 10
				}
			}
		]
	}
}
`,
			expectedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: "P108",
						Resources: []config.ResourceInfo{
							{
								Type: "plugin",
								ID:   "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							},
						},
					},
				},
			},
		},
		{
			name: "[rate-limiting] change is emitted with non-default value" +
				" for redis_server_name",
			uncompressedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "rate-limiting",
				"config": {
					"second": 10,
					"redis_server_name": "redis.example.com"
				}
			}
		]
	}
}
`,
			dataPlaneVersion: "2.6.0",
			expectedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "rate-limiting",
				"config": {
					"second": 10
				}
			}
		]
	}
}
`,
			expectedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: "P108",
						Resources: []config.ResourceInfo{
							{
								Type: "plugin",
								ID:   "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							},
						},
					},
				},
			},
		},
		{
			name: "[zipkin] change is emitted with non-default value" +
				" for local_service_name",
			uncompressedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "zipkin",
				"config": {
					"local_service_name": "api-gateway"
				}
			}
		]
	}
}
`,
			dataPlaneVersion: "2.6.0",
			expectedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "zipkin",
				"config": {
				}
			}
		]
	}
}
`,
			expectedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: "P110",
						Resources: []config.ResourceInfo{
							{
								Type: "plugin",
								ID:   "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							},
						},
					},
				},
			},
		},
		{
			name: "[zipkin] change is not emitted with default value" +
				" for local_service_name",
			uncompressedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "zipkin",
				"config": {
					"local_service_name": "kong"
				}
			}
		]
	}
}
`,
			dataPlaneVersion: "2.6.0",
			expectedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "zipkin",
				"config": {
				}
			}
		]
	}
}
`,
			expectedChanges: config.TrackedChanges{},
		},
		{
			name: "[acme] change is not emitted with default value" +
				" for rsa_key_size",
			uncompressedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "acme",
				"config": {
					"rsa_key_size": 4096
				}
			}
		]
	}
}
`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "acme",
				"config": {
				}
			}
		]
	}
}
`,
			expectedChanges: config.TrackedChanges{},
		},
		{
			name: "[acme] change is emitted with non-default value" +
				" for rsa_key_size",
			uncompressedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "acme",
				"config": {
					"rsa_key_size": 2048
				}
			}
		]
	}
}
`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "acme",
				"config": {
				}
			}
		]
	}
}
`,
			expectedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: "P111",
						Resources: []config.ResourceInfo{
							{
								Type: "plugin",
								ID:   "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							},
						},
					},
				},
			},
		},
		{
			name: "[zipkin] change is not emitted with default values",
			uncompressedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "zipkin",
				"config": {
					"http_span_name": "method",
					"connect_timeout": 2000,
					"read_timeout": 5000,
					"send_timeout": 5000,
					"http_response_header_for_traceid": null
				}
			}
		]
	}
}
`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "zipkin",
				"config": {
				}
			}
		]
	}
}
`,
			expectedChanges: config.TrackedChanges{},
		},
		{
			name: "[zipkin] change is emitted with non-default value for" +
				" http_span_name",
			uncompressedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "zipkin",
				"config": {
					"http_span_name": "method_path",
					"connect_timeout": 2000,
					"read_timeout": 5000,
					"send_timeout": 5000,
				}
			}
		]
	}
}
`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "zipkin",
				"config": {
				}
			}
		]
	}
}
`,
			expectedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: "P116",
						Resources: []config.ResourceInfo{
							{
								Type: "plugin",
								ID:   "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							},
						},
					},
				},
			},
		},
		{
			name: "[zipkin] change is emitted with non-default value for" +
				" send_timeout",
			uncompressedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "zipkin",
				"config": {
					"http_span_name": "method",
					"connect_timeout": 2000,
					"read_timeout": 5000,
					"send_timeout": 5001,
				}
			}
		]
	}
}
`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "zipkin",
				"config": {
				}
			}
		]
	}
}
`,
			expectedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: "P116",
						Resources: []config.ResourceInfo{
							{
								Type: "plugin",
								ID:   "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							},
						},
					},
				},
			},
		},
		{
			name: "[zipkin] change is emitted with non-default value for" +
				" read_timeout",
			uncompressedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "zipkin",
				"config": {
					"http_span_name": "method",
					"connect_timeout": 2000,
					"read_timeout": 5001,
					"send_timeout": 5000,
				}
			}
		]
	}
}
`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "zipkin",
				"config": {
				}
			}
		]
	}
}
`,
			expectedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: "P116",
						Resources: []config.ResourceInfo{
							{
								Type: "plugin",
								ID:   "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							},
						},
					},
				},
			},
		},
		{
			name: "[zipkin] change is emitted with non-default value for" +
				" connect_timeout",
			uncompressedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "zipkin",
				"config": {
					"http_span_name": "method",
					"connect_timeout": 200,
					"read_timeout": 5000,
					"send_timeout": 5000,
				}
			}
		]
	}
}
`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "zipkin",
				"config": {
				}
			}
		]
	}
}
`,
			expectedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: "P116",
						Resources: []config.ResourceInfo{
							{
								Type: "plugin",
								ID:   "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							},
						},
					},
				},
			},
		},
		{
			name: "[zipkin] change is emitted with non-default value for" +
				" http_response_header_for_traceid",
			uncompressedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "zipkin",
				"config": {
					"http_response_header_for_traceid": "X-B3-TraceId"
				}
			}
		]
	}
}
`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "zipkin",
				"config": {
				}
			}
		]
	}
}
`,
			expectedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: "P136",
						Resources: []config.ResourceInfo{
							{
								Type: "plugin",
								ID:   "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							},
						},
					},
				},
			},
		},
		{
			name: "[acme] change is emitted with non-default value for" +
				" allow_any_domain",
			uncompressedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "acme",
				"config": {
					"allow_any_domain": true
				}
			}
		]
	}
}
`,
			dataPlaneVersion: "2.8.0",
			expectedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "acme",
				"config": {
				}
			}
		]
	}
}
`,
			expectedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: "P118",
						Resources: []config.ResourceInfo{
							{
								Type: "plugin",
								ID:   "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							},
						},
					},
				},
			},
		},
		{
			name: "[acme] change is not emitted with default value for" +
				" allow_any_domain",
			uncompressedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "acme",
				"config": {
					"allow_any_domain": false
				}
			}
		]
	}
}
`,
			dataPlaneVersion: "2.8.0",
			expectedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "acme",
				"config": {
				}
			}
		]
	}
}
`,
			expectedChanges: config.TrackedChanges{},
		},
		{
			name: "[prometheus] change is emitted with default value for new" +
				" fields in prometheus 3.x",
			uncompressedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "prometheus",
				"config": {
					"status_code_metrics": false,
					"latency_metrics": false,
					"bandwidth_metrics": false,
					"upstream_health_metrics": false
				}
			}
		]
	}
}
`,
			dataPlaneVersion: "2.8.0",
			expectedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "prometheus",
				"config": {
				}
			}
		]
	}
}
`,
			expectedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: "P117",
						Resources: []config.ResourceInfo{
							{
								Type: "plugin",
								ID:   "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							},
						},
					},
				},
			},
		},
		{
			name: "[prometheus] change is emitted with non-default value for" +
				" new fields in prometheus 3.x",
			uncompressedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "prometheus",
				"config": {
					"status_code_metrics": true,
					"latency_metrics": true,
					"bandwidth_metrics": true,
					"upstream_health_metrics": true
				}
			}
		]
	}
}
`,
			dataPlaneVersion: "2.8.0",
			expectedPayload: `
{
	"config_table": {
		"plugins": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "prometheus",
				"config": {
				}
			}
		]
	}
}
`,
			expectedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: "P117",
						Resources: []config.ResourceInfo{
							{
								Type: "plugin",
								ID:   "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							},
						},
					},
				},
			},
		},
		{
			name: "[service] change is not emitted with default value for" +
				" enabled",
			uncompressedPayload: `
{
	"config_table": {
		"services": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "foo",
				"host": "foo.example.org",
				"enabled": true
			}
		]
	}
}
`,
			dataPlaneVersion: "2.6.0",
			expectedPayload: `
{
	"config_table": {
		"services": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "foo",
				"host": "foo.example.org"
			}
		]
	}
}
`,
			expectedChanges: config.TrackedChanges{},
		},
		{
			name: "[service] change is emitted with non-default value for" +
				" enabled",
			uncompressedPayload: `
{
	"config_table": {
		"services": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "foo",
				"host": "foo.example.org",
				"enabled": false
			}
		]
	}
}
`,
			dataPlaneVersion: "2.6.0",
			expectedPayload: `
{
	"config_table": {
		"services": [
			{
				"id": "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "foo",
				"host": "foo.example.org"
			}
		]
	}
}
`,
			expectedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: "P119",
						Resources: []config.ResourceInfo{
							{
								Type: "service",
								ID:   "759c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							},
						},
					},
				},
			},
		},
		{
			name: "[vault] configurations are removed for old versions",
			uncompressedPayload: `
{
	"config_table": {
		"vaults": [
			{
				"id": "462c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "env",
				"prefix": "test-env-vault",
				"config": {
					"PREFIX": "TEST_"
				}
			}
		]
	}
}
`,
			dataPlaneVersion: "2.5.0",
			expectedPayload: `
{
	"config_table": {
	}
}
`,
			expectedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: "P135",
						Resources: []config.ResourceInfo{
							{
								Type: "vault",
								ID:   "462c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							},
						},
					},
				},
			},
		},
		{
			name: "[vault] configurations are removed for old versions",
			uncompressedPayload: `
{
	"config_table": {
		"vaults": [
			{
				"id": "462c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "env",
				"prefix": "test-env-vault",
				"config": {
					"PREFIX": "TEST_"
				}
			}
		]
	}
}
`,
			dataPlaneVersion: "2.6.0",
			expectedPayload: `
{
	"config_table": {
	}
}
`,
			expectedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: "P135",
						Resources: []config.ResourceInfo{
							{
								Type: "vault",
								ID:   "462c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							},
						},
					},
				},
			},
		},
		{
			name: "[vault] configurations are removed for old versions",
			uncompressedPayload: `
{
	"config_table": {
		"vaults": [
			{
				"id": "462c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "env",
				"prefix": "test-env-vault",
				"config": {
					"PREFIX": "TEST_"
				}
			}
		]
	}
}
`,
			dataPlaneVersion: "2.7.0",
			expectedPayload: `
{
	"config_table": {
	}
}
`,
			expectedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: "P135",
						Resources: []config.ResourceInfo{
							{
								Type: "vault",
								ID:   "462c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							},
						},
					},
				},
			},
		},
		{
			name: "[vault] configurations are removed for old versions",
			uncompressedPayload: `
{
	"config_table": {
		"vaults": [
			{
				"id": "462c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "env",
				"prefix": "test-env-vault",
				"config": {
					"PREFIX": "TEST_"
				}
			}
		]
	}
}
`,
			dataPlaneVersion: "2.8.0",
			expectedPayload: `
{
	"config_table": {
	}
}
`,
			expectedChanges: config.TrackedChanges{
				ChangeDetails: []config.ChangeDetail{
					{
						ID: "P135",
						Resources: []config.ResourceInfo{
							{
								Type: "vault",
								ID:   "462c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
							},
						},
					},
				},
			},
		},
		{
			name: "[vault] configurations are removed for old versions",
			uncompressedPayload: `
{
	"config_table": {
		"vaults": [
			{
				"id": "462c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "env",
				"prefix": "test-env-vault",
				"config": {
					"PREFIX": "TEST_"
				}
			}
		]
	}
}
`,
			dataPlaneVersion: "3.0.0",
			expectedPayload: `
{
	"config_table": {
		"vaults": [
			{
				"id": "462c0d3a-bc3d-4ccc-8d4d-f92de95c1f1a",
				"name": "env",
				"prefix": "test-env-vault",
				"config": {
					"PREFIX": "TEST_"
				}
			}
		]
	}
}
`,
			expectedChanges: config.TrackedChanges{},
		},
	}
	vc, err := config.NewVersionCompatibilityProcessor(config.VersionCompatibilityOpts{
		Logger:        log.Logger,
		KongCPVersion: config.KongGatewayCompatibilityVersion,
	})
	require.NoError(t, err)
	require.NotNil(t, vc)
	err = vc.AddConfigTableUpdates(config.ChangeRegistry.GetUpdates())
	require.NoError(t, err)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			compressedPayload, err := config.CompressPayload([]byte(test.uncompressedPayload))
			require.NoError(t, err)
			require.NotEmpty(t, compressedPayload)
			processedPayload, trackedChanges, err := vc.ProcessConfigTableUpdates(test.dataPlaneVersion, compressedPayload)
			require.NoError(t, err)
			require.Equal(t, test.expectedChanges, trackedChanges)
			uncompressedCompatiblePayload, err := config.UncompressPayload(processedPayload)
			require.NoError(t, err)
			require.JSONEq(t, test.expectedPayload, string(uncompressedCompatiblePayload))
		})
	}
}
