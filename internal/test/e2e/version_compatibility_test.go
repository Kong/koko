//go:build integration

package e2e

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/gavv/httpexpect/v2"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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
	expectedConfig    string
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
		// DP < 2.6
		//   - remove 'preferred_chain', 'storage_config.vault.auth_method',
		//     'storage_config.vault.auth_path', 'storage_config.vault.auth_role',
		//     'storage_config.vault.jwt_path'
		//
		// DP < 3.0:
		//   - remove 'allow_any_domain'
		{
			name: "acme",
			id:   uuid.NewString(),
			config: `{
				"account_email": "example@example.com",
				"allow_any_domain": true
			}`,
		},
		// DP < 2.6
		//   - remove 'base64_encode_body'
		//
		// DP < 3.0
		//   - if both 'aws_region' and 'host' are set
		//     just drop 'host' and keep 'aws_region'
		//     since these used to be  mutually exclusive
		{
			name: "aws-lambda",
			id:   uuid.NewString(),
			config: `{
				"aws_region": "AWS_REGION",
				"host": "192.168.1.1",
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
		// DP < 2.4
		//   - remove 'per_consumer' field (default: false)
		//
		// DP < 3.0
		//   - remove 'status_code_metrics', 'latency_metrics'
		//     'bandwidth_metrics', 'upstream_health_metrics'
		{
			name: "prometheus",
			id:   uuid.NewString(),
			config: `{
				"status_code_metrics": true,
				"latency_metrics": true,
				"bandwidth_metrics": true,
				"upstream_health_metrics": true
			}`,
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
		// DP < 2.4:
		//   - remove 'tags_header' field (default: Zipkin-Tags)
		//
		// DP < 2.7:
		//   - change 'header_type' from 'ignore' to 'preserve'
		//
		// DP < 3.0:
		//   - remove 'http_span_name', 'connect_timeout'
		//     'send_timeout', 'read_timeout'
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

func TestVersionCompatibilitySyslogFacilityField(t *testing.T) {
	cleanup := run.Koko(t)
	defer cleanup()

	dpCleanup := run.KongDP(kong.GetKongConfForShared())
	defer dpCleanup()
	util.WaitForKong(t)
	util.WaitForKongAdminAPI(t)

	admin := httpexpect.New(t, "http://localhost:3000")

	tests := []vcPlugins{
		{
			// make sure facility is set to 'user' for all DP versions
			name:   "syslog",
			config: `{}`,
			expectedConfig: `{
				"client_errors_severity": "info",
				"custom_fields_by_lua": null,
				"facility": "user",
				"log_level": "info",
				"server_errors_severity": "info",
				"successful_severity": "info"
			}`,
		},
	}
	for _, test := range tests {
		var config structpb.Struct
		require.NoError(t, json.ProtoJSONUnmarshal([]byte(test.config), &config))

		plugin := &v1.Plugin{
			Id:        uuid.NewString(),
			Name:      test.name,
			Config:    &config,
			Enabled:   wrapperspb.Bool(true),
			Protocols: []string{"http", "https"},
		}
		pluginBytes, err := json.ProtoJSONMarshal(plugin)
		require.NoError(t, err)
		res := admin.POST("/v1/plugins").WithBytes(pluginBytes).Expect()
		res.Status(http.StatusCreated)
	}

	util.WaitFunc(t, func() error {
		err := ensurePluginsConfig(t, tests)
		t.Log("plugin validation failed", err)
		return err
	})
}

func ensurePluginsConfig(t *testing.T, plugins []vcPlugins) error {
	kongAdmin, err := kongClient.NewClient(util.BasedKongAdminAPIAddr, nil)
	if err != nil {
		return fmt.Errorf("create go client for kong: %v", err)
	}
	ctx := context.Background()
	dataPlanePlugins, err := kongAdmin.Plugins.ListAll(ctx)
	if err != nil {
		return fmt.Errorf("fetching plugins: %v", err)
	}

	var expectedPlugins []*kongClient.Plugin
	for _, plugin := range plugins {
		var expectedConfig kongClient.Configuration
		require.NoError(t, json.ProtoJSONUnmarshal([]byte(plugin.expectedConfig), &expectedConfig))

		plugin := &kongClient.Plugin{
			Name:    kongClient.String(plugin.name),
			Config:  expectedConfig,
			Enabled: kongClient.Bool(true),
		}
		expectedPlugins = append(expectedPlugins, plugin)
	}

	opt := []cmp.Option{
		cmpopts.IgnoreFields(kongClient.Plugin{}, "ID", "CreatedAt", "Protocols"),
		cmpopts.SortSlices(func(a, b *kongClient.Plugin) bool { return *a.Name < *b.Name }),
		cmpopts.EquateEmpty(),
	}

	if diff := cmp.Diff(dataPlanePlugins, expectedPlugins, opt...); diff != "" {
		return errors.New(diff)
	}

	return nil
}
