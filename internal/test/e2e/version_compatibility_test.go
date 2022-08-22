//go:build integration

package e2e

import (
	"context"
	"net/http"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/google/uuid"
	kongClient "github.com/kong/go-kong/kong"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/test/kong"
	"github.com/kong/koko/internal/test/run"
	"github.com/kong/koko/internal/test/util"
	"github.com/kong/koko/internal/versioning"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// TestVersionCompatibility tests multiple plugins to ensure that the version compatibility layer
// is processing multiple versions of the plugins configured globally, on the service, and on the
// route. Plugin IDs are generated during configuration of the plugin and used to validate when
// ensuring the configuration on the data plane is equivalent.
func TestVersionCompatibility(t *testing.T) {
	cleanup := run.Koko(t)
	defer cleanup()

	dpCleanup := run.KongDP(kong.GetKongConfForShared())
	defer dpCleanup()
	util.WaitForKong(t)
	util.WaitForKongAdminAPI(t)

	// Determine the data plane version for removal of plugins that are not expected to be present
	// in the configuration.
	// Note: These plugins will be configured on the control plane, but will be removed during the
	// version compatibility during payload transmission to the data plane.
	kongAdmin, err := kongClient.NewClient(util.BasedKongAdminAPIAddr, nil)
	require.NoError(t, err)
	ctx := context.Background()
	info, err := kongAdmin.Root(ctx)
	require.NoError(t, err)
	dataPlaneVersion, err := versioning.NewVersion(kongClient.VersionFromInfo(info))
	require.NoError(t, err)

	admin := httpexpect.New(t, "http://localhost:3000")
	expectedConfig := &v1.TestingConfig{
		Services: make([]*v1.Service, 1),
		Routes:   make([]*v1.Route, 1),
		Plugins:  []*v1.Plugin{},
	}

	// Create a service
	service := &v1.Service{
		Id:   uuid.NewString(),
		Name: "foo",
		Host: "httpbin.org",
		Path: "/",
	}
	res := admin.POST("/v1/services").WithJSON(service).Expect()
	res.Status(http.StatusCreated)
	expectedConfig.Services[0] = service

	// Create a route
	route := &v1.Route{
		Id:    uuid.NewString(),
		Name:  "bar",
		Paths: []string{"/"},
		Service: &v1.Service{
			Id: service.Id,
		},
	}
	res = admin.POST("/v1/routes").WithJSON(route).Expect()
	res.Status(http.StatusCreated)
	expectedConfig.Routes[0] = route

	// Handle configuration of the plugins and determine the expected plugin configurations
	// Note: All plugin configurations will be removed from the expected configuration due to
	// version compatibility layer transforming the configuration during transmission of the
	// payload to the data plane.
	for _, test := range VersionCompatibilityOSSPluginConfigurationTests {
		var config structpb.Struct
		if len(test.Config) > 0 {
			require.Nil(t, json.ProtoJSONUnmarshal([]byte(test.Config), &config))
		}

		// Determine if the plugin should be added to the expected plugin configurations
		addExpectedPlugin := true
		if len(test.VersionRange) > 0 {
			version := versioning.MustNewRange(test.VersionRange)
			if !version(dataPlaneVersion) {
				addExpectedPlugin = false
			}
		}

		// Configure plugins globally and add to expected plugins configuration
		plugin := &v1.Plugin{
			Id:        uuid.NewString(),
			Name:      test.Name,
			Config:    &config,
			Enabled:   wrapperspb.Bool(true),
			Protocols: []string{"http", "https"},
		}
		pluginBytes, err := json.ProtoJSONMarshal(plugin)
		require.Nil(t, err)
		res := admin.POST("/v1/plugins").WithBytes(pluginBytes).Expect()
		res.Status(http.StatusCreated)
		if addExpectedPlugin {
			expectedConfig.Plugins = append(expectedConfig.Plugins, &v1.Plugin{
				Id:        plugin.Id,
				Name:      plugin.Name,
				Enabled:   plugin.Enabled,
				Protocols: plugin.Protocols,
			})
		}

		// Configure plugin on service and add to expected plugins configuration
		if test.ConfigureForService {
			// Generate a new plugin ID and associate it with the service
			plugin.Id = uuid.NewString()
			plugin.Service = &v1.Service{Id: service.Id}
			pluginBytes, err = json.ProtoJSONMarshal(plugin)
			require.Nil(t, err)
			res := admin.POST("/v1/plugins").WithBytes(pluginBytes).Expect()
			res.Status(http.StatusCreated)
			if addExpectedPlugin {
				expectedConfig.Plugins = append(expectedConfig.Plugins, &v1.Plugin{
					Id:        plugin.Id,
					Name:      plugin.Name,
					Enabled:   plugin.Enabled,
					Protocols: plugin.Protocols,
				})
			}
		}

		// Configure plugin on route and add to expected plugins configuration
		if test.ConfigureForRoute {
			// Generate a new plugin ID and associate it with the route; resetting the possible associated
			// service
			plugin.Id = uuid.NewString()
			plugin.Service = nil
			plugin.Route = &v1.Route{Id: route.Id}
			pluginBytes, err = json.ProtoJSONMarshal(plugin)
			require.Nil(t, err)
			res := admin.POST("/v1/plugins").WithBytes(pluginBytes).Expect()
			res.Status(http.StatusCreated)
			if addExpectedPlugin {
				expectedConfig.Plugins = append(expectedConfig.Plugins, &v1.Plugin{
					Id:        plugin.Id,
					Name:      plugin.Name,
					Enabled:   plugin.Enabled,
					Protocols: plugin.Protocols,
				})
			}
		}
	}

	// Validate the service, route, and plugin configurations
	util.WaitFunc(t, func() error {
		err := util.EnsureConfig(expectedConfig)
		if err != nil {
			t.Log("config validation failed", err)
		}
		return err
	})
}

func TestVersionCompatibilitySyslogFacilityField(t *testing.T) {
	cleanup := run.Koko(t)
	defer cleanup()

	dpCleanup := run.KongDP(kong.GetKongConfForShared())
	defer dpCleanup()
	util.WaitForKong(t)
	util.WaitForKongAdminAPI(t)

	admin := httpexpect.New(t, "http://localhost:3000")

	tests := []VersionCompatibilityPlugins{
		{
			// make sure facility is set to 'user' for all DP versions
			Name:   "syslog",
			Config: `{}`,
			ExpectedConfig: `{
				"client_errors_severity": "info",
				"custom_fields_by_lua": null,
				"facility": "user",
				"log_level": "info",
				"server_errors_severity": "info",
				"successful_severity": "info"
			}`,
		},
	}

	expectedConfig := &v1.TestingConfig{
		Plugins: make([]*v1.Plugin, 0, len(tests)),
	}

	for _, test := range tests {
		var config structpb.Struct
		require.NoError(t, json.ProtoJSONUnmarshal([]byte(test.Config), &config))

		plugin := &v1.Plugin{
			Id:        uuid.NewString(),
			Name:      test.Name,
			Config:    &config,
			Enabled:   wrapperspb.Bool(true),
			Protocols: []string{"http", "https"},
		}
		pluginBytes, err := json.ProtoJSONMarshal(plugin)
		require.NoError(t, err)
		res := admin.POST("/v1/plugins").WithBytes(pluginBytes).Expect()
		res.Status(http.StatusCreated)

		var expected structpb.Struct
		require.NoError(t, json.ProtoJSONUnmarshal([]byte(test.ExpectedConfig), &expected))
		expectedConfig.Plugins = append(expectedConfig.Plugins, &v1.Plugin{
			Id:        plugin.Id,
			Name:      plugin.Name,
			Config:    &expected,
			Enabled:   plugin.Enabled,
			Protocols: plugin.Protocols,
		})
	}

	util.WaitFunc(t, func() error {
		err := util.EnsureConfig(expectedConfig)
		t.Log("plugin validation failed", err)
		return err
	})
}
