//go:build integration

package e2e

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/google/uuid"
	kongClient "github.com/kong/go-kong/kong"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/test/kong"
	"github.com/kong/koko/internal/test/run"
	"github.com/kong/koko/internal/test/util"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type vcPlugins struct {
	name   string
	id     string
	config string
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
	}
	for _, test := range tests {
		var config structpb.Struct
		if len(test.config) > 0 {
			require.Nil(t, json.Unmarshal([]byte(test.config), &config))
		}

		plugin := &v1.Plugin{
			Id:        test.id,
			Name:      test.name,
			Config:    &config,
			Enabled:   wrapperspb.Bool(true),
			Protocols: []string{"http", "https"},
		}
		pluginBytes, err := json.Marshal(plugin)
		require.Nil(t, err)
		res := admin.POST("/v1/plugins").WithBytes(pluginBytes).Expect()
		res.Status(201)
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
	dataPlanePlugins, err := kongAdmin.Plugins.ListAll(ctx)
	if err != nil {
		return fmt.Errorf("fetching plugins: %v", err)
	}

	// Because configurations may vary validation occurs via the name and ID
	if len(plugins) != len(dataPlanePlugins) {
		return fmt.Errorf("plugins configured count does not match [%d != %d]", len(plugins), len(dataPlanePlugins))
	}
	var failedPlugins []string
	for _, plugin := range plugins {
		found := false
		for _, dataPlanePlugin := range dataPlanePlugins {
			if plugin.name == *dataPlanePlugin.Name && plugin.id == *dataPlanePlugin.ID {
				found = true
				break
			}
		}
		if !found {
			failedPlugins = append(failedPlugins, plugin.name)
		}
	}

	if len(failedPlugins) > 0 {
		return fmt.Errorf("failed to match plugins %s", strings.Join(failedPlugins, ","))
	}
	return nil
}
