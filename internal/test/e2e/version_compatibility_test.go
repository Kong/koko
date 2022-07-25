//go:build integration

package e2e

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/gavv/httpexpect/v2"
	"github.com/google/uuid"
	kongClient "github.com/kong/go-kong/kong"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/test/kong"
	"github.com/kong/koko/internal/test/run"
	"github.com/kong/koko/internal/test/util"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type update struct {
	field string
	value interface{}
}

type vcPlugins struct {
	name              string
	id                string
	config            string
	versionRange      string
	fieldUpdateChecks map[string][]update
}

func TestVersionCompatibility(t *testing.T) {
	cleanup := run.Koko(t)
	defer cleanup()

	dpCleanup := run.KongDP(kong.GetKongConfForShared())
	defer dpCleanup()
	util.WaitForKong(t)
	util.WaitForKongAdminAPI(t)

	admin := httpexpect.New(t, "http://localhost:3000")

	tests := []vcPlugins{
		{
			name: "acl",
			id:   uuid.NewString(),
			config: `{
				"allow": [
					"kongers"
				]
			}`,
		},
		{
			name: "acme",
			id:   uuid.NewString(),
			config: `{
				"account_email": "example@example.com"
			}`,
		},
		{
			name: "aws-lambda",
			id:   uuid.NewString(),
			config: `{
				"aws_region": "AWS_REGION",
				"function_name": "FUNCTION_NAME"
			}`,
		},
		{
			name: "azure-functions",
			id:   uuid.NewString(),
			config: `{
				"functionname": "FUNCTIONNAME",
				"appname": "APPNAME"
			}`,
		},
		{
			name: "basic-auth",
			id:   uuid.NewString(),
		},
		{
			name: "bot-detection",
			id:   uuid.NewString(),
		},
		{
			name: "correlation-id",
			id:   uuid.NewString(),
		},
		{
			name: "cors",
			id:   uuid.NewString(),
		},
		{
			name: "datadog",
			id:   uuid.NewString(),
			config: `{
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
		},
		{
			name: "file-log",
			id:   uuid.NewString(),
			config: `{
				"path": "path/to/file.log"
			}`,
		},
		{
			name: "grpc-gateway",
			id:   uuid.NewString(),
			config: `{
				"proto": "path/to/file.proto"
			}`,
		},
		{
			name: "grpc-web",
			id:   uuid.NewString(),
			config: `{
				"proto": "path/to/file.proto"
			}`,
		},
		{
			name: "hmac-auth",
			id:   uuid.NewString(),
		},
		{
			name: "http-log",
			id:   uuid.NewString(),
			config: `{
				"http_endpoint": "http://example.com/logs"
			}`,
		},
		{
			name: "ip-restriction",
			id:   uuid.NewString(),
			config: `{
				"allow": [
					"1.2.3.4"
				],
				"status": 200,
				"message": "MESSAGE"
			}`,
		},
		{
			name: "jwt",
			id:   uuid.NewString(),
		},
		{
			name: "key-auth",
			id:   uuid.NewString(),
		},
		{
			name: "ldap-auth",
			id:   uuid.NewString(),
			config: `{
				"ldap_host": "example.com",
				"ldap_port": 389,
				"base_dn": "dc=example,dc=com",
				"attribute": "cn"
			}`,
		},
		{
			name: "loggly",
			id:   uuid.NewString(),
			config: `{
				"key": "KEY"
			}`,
		},
		{
			name: "post-function",
			id:   uuid.NewString(),
			config: `{
				"access": [
					"kong.log.err('Goodbye Koko!')"
				]
			}`,
		},
		{
			name: "pre-function",
			id:   uuid.NewString(),
			config: `{
				"access": [
					"kong.log.err('Hello Koko!')"
				]
			}`,
		},
		{
			name: "prometheus",
			id:   uuid.NewString(),
		},
		{
			name: "proxy-cache",
			id:   uuid.NewString(),
			config: `{
				"strategy": "memory"
			}`,
		},
		{
			name: "rate-limiting",
			id:   uuid.NewString(),
			config: `{
				"hour": 1,
				"redis_ssl": true,
				"redis_ssl_verify": true,
				"redis_server_name": "redis.example.com",
				"redis_username": "REDIS_USERNAME",
				"redis_password": "REDIS_PASSWORD"
			}`,
		},
		{
			name: "request-size-limiting",
			id:   uuid.NewString(),
		},
		{
			name: "request-termination",
			id:   uuid.NewString(),
		},
		{
			name: "request-transformer",
			id:   uuid.NewString(),
		},
		{
			name: "response-ratelimiting",
			id:   uuid.NewString(),
			config: `{
				"limits": {
					"sms": {
						"minute": 20
					}
				},
				"redis_username": "REDIS_USERNAME"
			}`,
		},
		{
			name: "response-transformer",
			id:   uuid.NewString(),
		},
		{
			name: "session",
			id:   uuid.NewString(),
		},
		{
			name: "statsd",
			id:   uuid.NewString(),
		},
		{
			name: "syslog",
			id:   uuid.NewString(),
		},
		{
			name: "tcp-log",
			id:   uuid.NewString(),
			config: `{
				"host": "localhost",
				"port": 1234
			}`,
		},
		{
			name: "udp-log",
			id:   uuid.NewString(),
			config: `{
				"host": "localhost",
				"port": 1234
			}`,
		},
		{
			name: "zipkin",
			id:   uuid.NewString(),
			config: `{
				"local_service_name": "LOCAL_SERVICE_NAME",
				"header_type": "ignore",
				"http_span_name": "method_path",
				"connect_timeout": 2001,
				"send_timeout": 2001,
				"read_timeout": 2001
			}`,
			fieldUpdateChecks: map[string][]update{
				"< 2.7.0": {
					{
						field: "header_type",
						value: "preserve",
					},
				},
			},
		},
		{
			name: "opentelemetry",
			id:   uuid.NewString(),
			config: `{
				"endpoint": "http://example.dev"
			}`,
			versionRange: ">= 3.0.0",
		},
	}
	for _, test := range tests {
		var config structpb.Struct
		if len(test.config) > 0 {
			require.Nil(t, json.ProtoJSONUnmarshal([]byte(test.config), &config))
		}

		plugin := &v1.Plugin{
			Id:        test.id,
			Name:      test.name,
			Config:    &config,
			Enabled:   wrapperspb.Bool(true),
			Protocols: []string{"http", "https"},
		}
		pluginBytes, err := json.ProtoJSONMarshal(plugin)
		require.Nil(t, err)
		res := admin.POST("/v1/plugins").WithBytes(pluginBytes).Expect()
		res.Status(http.StatusCreated)
	}

	util.WaitFunc(t, func() error {
		err := ensurePlugins(tests)
		t.Log("plugin validation failed", err)
		return err
	})
}

func ensurePlugins(plugins []vcPlugins) error {
	kongAdmin, err := kongClient.NewClient(util.BasedKongAdminAPIAddr, nil)
	if err != nil {
		return fmt.Errorf("create go client for kong: %v", err)
	}
	ctx := context.Background()
	info, err := kongAdmin.Root(ctx)
	if err != nil {
		return fmt.Errorf("fetching Kong Gateway info: %v", err)
	}
	dataPlaneVersion, err := kongClient.ParseSemanticVersion(kongClient.VersionFromInfo(info))
	if err != nil {
		return fmt.Errorf("parsing Kong Gateway version: %v", err)
	}
	dataPlanePlugins, err := kongAdmin.Plugins.ListAll(ctx)
	if err != nil {
		return fmt.Errorf("fetching plugins: %v", err)
	}

	// Remove plugins that may not be expected due to data plane version
	var expectedPlugins []vcPlugins
	for _, plugin := range plugins {
		addPlugin := true
		if len(plugin.versionRange) > 0 {
			version := semver.MustParseRange(plugin.versionRange)
			if !version(dataPlaneVersion) {
				addPlugin = false
			}
		}
		if addPlugin {
			expectedPlugins = append(expectedPlugins, plugin)
		}
	}

	// Because configurations may vary validation occurs via the name and ID for
	// removal items and a special test for updates will be performed which
	// verifies update occurred properly based on versions
	if len(expectedPlugins) != len(dataPlanePlugins) {
		return fmt.Errorf("plugins configured count does not match [%d != %d]", len(expectedPlugins), len(dataPlanePlugins))
	}
	var failedPlugins []string
	var missingPlugins []string
	for _, plugin := range expectedPlugins {
		found := false
		for _, dataPlanePlugin := range dataPlanePlugins {
			if plugin.name == *dataPlanePlugin.Name && plugin.id == *dataPlanePlugin.ID {
				// Ensure field updates occurred and validate
				if len(plugin.fieldUpdateChecks) > 0 {
					config, err := json.ProtoJSONMarshal(dataPlanePlugin.Config)
					if err != nil {
						return fmt.Errorf("marshal %s plugin config: %v", plugin.name, err)
					}
					configStr := string(config)

					for version, updates := range plugin.fieldUpdateChecks {
						version := semver.MustParseRange(version)
						if version(dataPlaneVersion) {
							for _, update := range updates {
								res := gjson.Get(configStr, update.field)
								if !res.Exists() || res.Value() != update.value {
									failedPlugins = append(failedPlugins, plugin.name)
									break
								}
							}
						}
					}
				}

				found = true
				break
			}
		}
		if !found {
			missingPlugins = append(missingPlugins, plugin.name)
		}
	}

	if len(missingPlugins) > 0 {
		return fmt.Errorf("failed to discover plugins %s", strings.Join(missingPlugins, ","))
	}
	if len(failedPlugins) > 0 {
		return fmt.Errorf("failed to validate plugin updates %s", strings.Join(failedPlugins, ","))
	}

	return nil
}
