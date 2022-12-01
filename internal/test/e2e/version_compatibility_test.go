//go:build integration

package e2e

import (
	"context"
	"errors"
	"fmt"
	"net/http"
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
	"github.com/kong/koko/internal/versioning"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
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
	require.NoError(t, util.WaitForKong(t))
	require.NoError(t, util.WaitForKongAdminAPI(t))

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

	admin := httpexpect.Default(t, "http://localhost:3000")
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
	chunk := 25
	for i := 0; i < len(VersionCompatibilityOSSPluginConfigurationTests); i += chunk {
		min := intMin(i+chunk, len(VersionCompatibilityOSSPluginConfigurationTests))
		batch := VersionCompatibilityOSSPluginConfigurationTests[i:min]
		var names []string
		for _, plugin := range batch {
			names = append(names, plugin.Name)
		}
		t.Run(strings.Join(names, "_"), func(t *testing.T) {
			pluginIDs := []string{}
			for _, test := range batch {
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
				pluginIDs = append(pluginIDs, plugin.Id)
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
					pluginIDs = append(pluginIDs, plugin.Id)
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
					pluginIDs = append(pluginIDs, plugin.Id)
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

			// Remove all batched plugin configurations and reset expected plugins
			for _, pluginID := range pluginIDs {
				res := admin.DELETE(fmt.Sprintf("/v1/plugins/%s", pluginID)).Expect()
				res.Status(http.StatusNoContent)
			}
			expectedConfig.Plugins = []*v1.Plugin{}
		})
	}
}

func TestVersionCompatibility_PluginFieldUpdates(t *testing.T) {
	cleanup := run.Koko(t)
	defer cleanup()

	dpCleanup := run.KongDP(kong.GetKongConfForShared())
	defer dpCleanup()
	require.NoError(t, util.WaitForKong(t))
	require.NoError(t, util.WaitForKongAdminAPI(t))

	kongAdmin, err := kongClient.NewClient(util.BasedKongAdminAPIAddr, nil)
	require.NoError(t, err, "create go client for kong")
	ctx := context.Background()
	info, err := kongAdmin.Root(ctx)
	require.NoError(t, err, "fetching Kong Gateway info")

	dataPlaneVersion, err := versioning.NewVersion(kongClient.VersionFromInfo(info))
	require.NoError(t, err)

	admin := httpexpect.Default(t, "http://localhost:3000")

	// Remove plugins that may not be expected due to data plane version or aren't subject to compatibility updates
	expectedPluginsMap := make(map[string]VersionCompatibilityPlugins, 0)
	for _, plugin := range VersionCompatibilityOSSPluginConfigurationTests {
		if plugin.FieldUpdateChecks == nil {
			continue
		}

		if len(plugin.VersionRange) > 0 {
			version := versioning.MustNewRange(plugin.VersionRange)
			if !version(dataPlaneVersion) {
				continue
			}
		}
		expectedPluginsMap[plugin.Name] = plugin
	}

	// create plugins
	for _, plugin := range expectedPluginsMap {
		var config structpb.Struct
		if len(plugin.Config) > 0 {
			require.NoError(t, json.ProtoJSONUnmarshal([]byte(plugin.Config), &config), plugin)
		}

		p := &v1.Plugin{
			Id:        uuid.NewString(),
			Name:      plugin.Name,
			Config:    &config,
			Enabled:   wrapperspb.Bool(true),
			Protocols: []string{"http", "https"},
		}

		pluginBytes, err := json.ProtoJSONMarshal(p)
		require.NoError(t, err)
		res := admin.POST("/v1/plugins").WithBytes(pluginBytes).Expect()
		res.Status(http.StatusCreated)
	}

	// fetch list of plugins from the DP
	dataPlanePlugins := []*kongClient.Plugin{}
	util.WaitFunc(t, func() error {
		dataPlanePlugins, err = kongAdmin.Plugins.ListAll(ctx)
		if err != nil {
			t.Log("fetch plugin failed", err)
		}
		if len(dataPlanePlugins) == 0 {
			return errors.New("plugins len was 0")
		}
		return err
	})

	require.Equal(t, len(expectedPluginsMap), len(dataPlanePlugins), "plugins configured count does not match")
	dpPluginsMap := make(map[string]string, 0)
	for _, dpPlugin := range dataPlanePlugins {
		b, err := json.ProtoJSONMarshal(dpPlugin.Config)
		require.NoError(t, err)
		dpPluginsMap[*dpPlugin.Name] = string(b)
	}

	for name, expectedPlugin := range expectedPluginsMap {
		t.Run(fmt.Sprintf("%s successfully updates fields", name), func(t *testing.T) {
			require.Contains(t, dpPluginsMap, name, "plugin not present in the DP")
			for rawVersion, updates := range expectedPlugin.FieldUpdateChecks {
				parsedVersion := versioning.MustNewRange(rawVersion)
				if parsedVersion(dataPlaneVersion) {
					for _, update := range updates {
						want := update.Value
						have := gjson.Get(dpPluginsMap[name], update.Field)
						switch want.(type) {
						case string:
							require.Equal(t, want, have.String())
						case []string:
							require.ElementsMatch(t, want, have.Value())
						default:
							require.Equal(t, want, have.Value())
						}
					}
				}
			}
		})
	}
}

func TestVersionCompatibility_EnsureTargetFieldsAreNotOverridden(t *testing.T) {
	cleanup := run.Koko(t)
	defer cleanup()

	dpCleanup := run.KongDP(kong.GetKongConfForShared())
	defer dpCleanup()
	require.NoError(t, util.WaitForKong(t))
	require.NoError(t, util.WaitForKongAdminAPI(t))

	kongClient.RunWhenKong(t, ">=3.0.0")
	kongAdmin, err := kongClient.NewClient(util.BasedKongAdminAPIAddr, nil)
	require.NoError(t, err, "create go client for kong")
	ctx := context.Background()
	info, err := kongAdmin.Root(ctx)
	require.NoError(t, err, "fetching Kong Gateway info")

	dataPlaneVersion, err := versioning.NewVersion(kongClient.VersionFromInfo(info))
	require.NoError(t, err)

	admin := httpexpect.Default(t, "http://localhost:3000")

	pluginTests := []VersionCompatibilityPlugins{
		{
			Name: "pre-function",
			Config: `{
				"access": [
					"kong.log.err('Hello Koko!')"
				],
				"functions": []
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
		{
			Name: "post-function",
			Config: `{
				"access": [
					"kong.log.err('Goodbye Koko!')"
				],
				"functions": []
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
	}

	expectedPluginsMap := make(map[string]VersionCompatibilityPlugins, 0)
	for _, plugin := range pluginTests {
		if plugin.FieldUpdateChecks == nil {
			continue
		}

		if len(plugin.VersionRange) > 0 {
			version := versioning.MustNewRange(plugin.VersionRange)
			if !version(dataPlaneVersion) {
				continue
			}
		}
		expectedPluginsMap[plugin.Name] = plugin
	}

	// create plugins
	for _, plugin := range expectedPluginsMap {
		var config structpb.Struct
		if len(plugin.Config) > 0 {
			require.NoError(t, json.ProtoJSONUnmarshal([]byte(plugin.Config), &config))
		}

		p := &v1.Plugin{
			Id:        uuid.NewString(),
			Name:      plugin.Name,
			Config:    &config,
			Enabled:   wrapperspb.Bool(true),
			Protocols: []string{"http", "https"},
		}

		pluginBytes, err := json.ProtoJSONMarshal(p)
		require.NoError(t, err)
		res := admin.POST("/v1/plugins").WithBytes(pluginBytes).Expect()
		res.Status(http.StatusCreated)
	}

	// fetch list of plugins from the DP
	dataPlanePlugins := []*kongClient.Plugin{}
	util.WaitFunc(t, func() error {
		dataPlanePlugins, err = kongAdmin.Plugins.ListAll(ctx)
		if err != nil {
			t.Log("fetch plugin failed", err)
		}
		if len(dataPlanePlugins) == 0 {
			return errors.New("plugins len was 0")
		}
		return err
	})

	require.Equal(t, len(expectedPluginsMap), len(dataPlanePlugins), "plugins configured count does not match")
	dpPluginsMap := make(map[string]string, 0)
	for _, dpPlugin := range dataPlanePlugins {
		b, err := json.ProtoJSONMarshal(dpPlugin.Config)
		require.NoError(t, err)
		dpPluginsMap[*dpPlugin.Name] = string(b)
	}

	for name, expectedPlugin := range expectedPluginsMap {
		t.Run(fmt.Sprintf("%s successfully updates fields", name), func(t *testing.T) {
			require.Contains(t, dpPluginsMap, name, "plugin not present in the DP")
			for rawVersion, updates := range expectedPlugin.FieldUpdateChecks {
				parsedVersion := versioning.MustNewRange(rawVersion)
				if parsedVersion(dataPlaneVersion) {
					for _, update := range updates {
						want := update.Value
						have := gjson.Get(dpPluginsMap[name], update.Field)
						switch want.(type) {
						case string:
							require.Equal(t, want, have.String())
						case []string:
							require.ElementsMatch(t, want, have.Value())
						default:
							require.Equal(t, want, have.Value())
						}
					}
				}
			}
		})
	}
}

func TestVersionCompatibilitySyslogFacilityField(t *testing.T) {
	cleanup := run.Koko(t)
	defer cleanup()

	dpCleanup := run.KongDP(kong.GetKongConfForShared())
	defer dpCleanup()
	require.NoError(t, util.WaitForKong(t))
	require.NoError(t, util.WaitForKongAdminAPI(t))

	admin := httpexpect.Default(t, "http://localhost:3000")

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

type vcUpstreamsTC struct {
	name              string
	upstream          *v1.Upstream
	versionedExpected map[string]*v1.Upstream
}

func TestUpstreamsVersionCompatibility(t *testing.T) {
	cleanup := run.Koko(t)
	defer cleanup()

	dpCleanup := run.KongDP(kong.GetKongConfForShared())
	defer dpCleanup()
	require.NoError(t, util.WaitForKong(t))
	require.NoError(t, util.WaitForKongAdminAPI(t))

	admin := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  "http://localhost:3000",
		Reporter: httpexpect.NewRequireReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewCompactPrinter(t),
		},
	})

	tests := []vcUpstreamsTC{
		{
			name: "ensure hash_on_query_arg is dropped for DP < 3.0",
			upstream: &v1.Upstream{
				Id:             uuid.NewString(),
				Name:           "foo-with-hash_on_query_arg",
				HashOn:         "ip",
				HashOnQueryArg: "test",
			},
			versionedExpected: map[string]*v1.Upstream{
				"< 3.0.0": {
					Name:           "foo-with-hash_on_query_arg",
					HashOn:         "ip",
					HashOnQueryArg: "",
				},
				">= 3.0.0": {
					Name:           "foo-with-hash_on_query_arg",
					HashOn:         "ip",
					HashOnQueryArg: "test",
				},
			},
		},
		{
			name: "ensure hash_on is reverted to 'none' when configured to incompatible values for DP < 3.0",
			upstream: &v1.Upstream{
				Id:             uuid.NewString(),
				Name:           "foo-with-hash_on",
				HashOn:         "path",
				HashOnQueryArg: "test",
			},
			versionedExpected: map[string]*v1.Upstream{
				"< 3.0.0": {
					Name:           "foo-with-hash_on",
					HashOn:         "none",
					HashOnQueryArg: "",
				},
				">= 3.0.0": {
					Name:           "foo-with-hash_on",
					HashOn:         "path",
					HashOnQueryArg: "test",
				},
			},
		},
		{
			name: "ensure hash_fallback is reverted to 'none' when configured to incompatible values for DP < 3.0",
			upstream: &v1.Upstream{
				Id:             uuid.NewString(),
				Name:           "foo-with-hash_fallback",
				HashFallback:   "path",
				HashOn:         "ip",
				HashOnQueryArg: "test",
			},
			versionedExpected: map[string]*v1.Upstream{
				"< 3.0.0": {
					Name:           "foo-with-hash_fallback",
					HashFallback:   "none",
					HashOn:         "ip",
					HashOnQueryArg: "",
				},
				">= 3.0.0": {
					Name:           "foo-with-hash_fallback",
					HashFallback:   "path",
					HashOn:         "ip",
					HashOnQueryArg: "test",
				},
			},
		},
	}
	for _, test := range tests {
		res := admin.POST("/v1/upstreams").WithJSON(test.upstream).Expect()
		res.Status(http.StatusCreated)
	}

	util.WaitFunc(t, func() error {
		err := ensureUpstreams(tests)
		t.Log("upstreams validation failed", err)
		return err
	})
}

func ensureUpstreams(upstreams []vcUpstreamsTC) error {
	kongAdmin, err := kongClient.NewClient(util.BasedKongAdminAPIAddr, nil)
	if err != nil {
		return fmt.Errorf("create go client for kong: %v", err)
	}
	ctx := context.Background()
	info, err := kongAdmin.Root(ctx)
	if err != nil {
		return fmt.Errorf("fetching Kong Gateway info: %v", err)
	}
	dataPlaneVersion, err := versioning.NewVersion(kongClient.VersionFromInfo(info))
	if err != nil {
		return fmt.Errorf("parsing Kong Gateway version: %v", err)
	}
	dataPlaneUpstreams, err := kongAdmin.Upstreams.ListAll(ctx)
	if err != nil {
		return fmt.Errorf("fetching upstreams: %v", err)
	}

	if len(upstreams) != len(dataPlaneUpstreams) {
		return fmt.Errorf("upstreams configured count does not match [%d != %d]", len(upstreams), len(dataPlaneUpstreams))
	}

	expectedConfig := &v1.TestingConfig{
		Upstreams: []*v1.Upstream{},
	}
	for _, u := range upstreams {
		for _, dataPlaneUpstream := range dataPlaneUpstreams {
			if u.upstream.Name == *dataPlaneUpstream.Name && u.upstream.Id == *dataPlaneUpstream.ID {
				for version, expectedUpstream := range u.versionedExpected {
					version := versioning.MustNewRange(version)
					if version(dataPlaneVersion) {
						expectedConfig.Upstreams = append(expectedConfig.Upstreams, expectedUpstream)
					}
				}
			}
		}
	}

	return util.EnsureConfig(expectedConfig)
}

type vcPlugins struct {
	name           string
	config         string
	expectedConfig string
}

// Ensure that extra-processing logic correctly injects metric identifier default values
// for DP < 3.0, which doesn't support setting default identifiers in the schema.
func TestVersionCompatibilityTransformations_StatsdMetricsDefaults(t *testing.T) {
	cleanup := run.Koko(t)
	defer cleanup()

	admin := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  "http://localhost:3000",
		Reporter: httpexpect.NewRequireReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewCompactPrinter(t),
		},
	})

	tests := []vcPlugins{
		// these 2 test metrics ensure that:
		//
		// 1. a metric with unsupported identifiers gets cleaned up
		// 2. a metric with non-default identifiers gets filled with defaults
		{
			config: `{
				"metrics": [
				  {
					"name": "response_size",
					"stat_type": "timer",
					"service_identifier": "service_name_or_host"
				  },
				  {
					"name": "unique_users",
					"consumer_identifier": null,
					"stat_type": "set"
				  }
				]
			}`,
			expectedConfig: `{
				"host": "localhost",
				"metrics": [
					{
						"name": "response_size",
						"stat_type": "timer"
					},
					{
						"name": "unique_users",
						"consumer_identifier": "custom_id",
						"stat_type": "set"
					  }
				],
				"port": 8125,
				"prefix": "kong"
			}`,
		},
	}
	expectedConfig := &v1.TestingConfig{
		Plugins: make([]*v1.Plugin, 0, len(tests)),
	}

	for _, test := range tests {
		var config structpb.Struct
		require.NoError(t, json.ProtoJSONUnmarshal([]byte(test.config), &config))

		plugin := &v1.Plugin{
			Id:        uuid.NewString(),
			Name:      "statsd",
			Config:    &config,
			Enabled:   wrapperspb.Bool(true),
			Protocols: []string{"http", "https"},
		}
		pluginBytes, err := json.ProtoJSONMarshal(plugin)
		require.NoError(t, err)
		res := admin.POST("/v1/plugins").WithBytes(pluginBytes).Expect()
		res.Status(http.StatusCreated)

		var expected structpb.Struct
		require.NoError(t, json.ProtoJSONUnmarshal([]byte(test.expectedConfig), &expected))
		expectedConfig.Plugins = append(expectedConfig.Plugins, &v1.Plugin{
			Id:        plugin.Id,
			Name:      plugin.Name,
			Config:    &expected,
			Enabled:   plugin.Enabled,
			Protocols: plugin.Protocols,
		})
	}

	dpCleanup := run.KongDP(kong.GetKongConfForShared())
	defer dpCleanup()
	require.NoError(t, util.WaitForKong(t))
	require.NoError(t, util.WaitForKongAdminAPI(t))

	kongClient.RunWhenKong(t, "<3.0.0")

	util.WaitFunc(t, func() error {
		err := util.EnsureConfig(expectedConfig)
		t.Log("plugin validation failed", err)
		return err
	})
}

// Ensure that a plugin configured with exactly the same pre-3.0 default configuration,
// doesn't get changed via extra-processing logic.
func TestVersionCompatibilityTransformations_StatsdDefaultConfig(t *testing.T) {
	cleanup := run.Koko(t)
	defer cleanup()

	admin := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  "http://localhost:3000",
		Reporter: httpexpect.NewRequireReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewCompactPrinter(t),
		},
	})

	tests := []vcPlugins{
		{
			name:   "statsd",
			config: `{}`,
			expectedConfig: `{
				"host": "localhost",
				"metrics": [
					{
					  "name": "request_count",
					  "stat_type": "counter",
					  "sample_rate": 1
					},
					{
					  "name": "latency",
					  "stat_type": "timer"
					},
					{
					  "name": "request_size",
					  "stat_type": "timer"
					},
					{
					  "name": "status_count",
					  "stat_type": "counter",
					  "sample_rate": 1
					},
					{
					  "name": "response_size",
					  "stat_type": "timer"
					},
					{
					  "name": "unique_users",
					  "stat_type": "set",
					  "consumer_identifier": "custom_id"
					},
					{
					  "name": "request_per_user",
					  "stat_type": "counter",
					  "sample_rate": 1,
					  "consumer_identifier": "custom_id"
					},
					{
					  "name": "upstream_latency",
					  "stat_type": "timer"
					},
					{
					  "name": "kong_latency",
					  "stat_type": "timer"
					},
					{
					  "name": "status_count_per_user",
					  "stat_type": "counter",
					  "sample_rate": 1,
					  "consumer_identifier": "custom_id"
					}
				],
				"port": 8125,
				"prefix": "kong"
			}`,
		},
	}
	expectedConfig := &v1.TestingConfig{
		Plugins: make([]*v1.Plugin, 0, len(tests)),
	}

	for _, test := range tests {
		var config structpb.Struct
		require.NoError(t, json.ProtoJSONUnmarshal([]byte(test.config), &config))

		plugin := &v1.Plugin{
			Id:        uuid.NewString(),
			Name:      "statsd",
			Config:    &config,
			Enabled:   wrapperspb.Bool(true),
			Protocols: []string{"http", "https"},
		}
		pluginBytes, err := json.ProtoJSONMarshal(plugin)
		require.NoError(t, err)
		res := admin.POST("/v1/plugins").WithBytes(pluginBytes).Expect()
		res.Status(http.StatusCreated)

		var expected structpb.Struct
		require.NoError(t, json.ProtoJSONUnmarshal([]byte(test.expectedConfig), &expected))
		expectedConfig.Plugins = append(expectedConfig.Plugins, &v1.Plugin{
			Id:        plugin.Id,
			Name:      plugin.Name,
			Config:    &expected,
			Enabled:   plugin.Enabled,
			Protocols: plugin.Protocols,
		})
	}

	dpCleanup := run.KongDP(kong.GetKongConfForShared())
	defer dpCleanup()
	require.NoError(t, util.WaitForKong(t))
	require.NoError(t, util.WaitForKongAdminAPI(t))

	kongClient.RunWhenKong(t, "<3.0.0")

	util.WaitFunc(t, func() error {
		err := util.EnsureConfig(expectedConfig)
		t.Log("plugin validation failed", err)
		return err
	})
}

type vcRoutesTC struct {
	name  string
	route *v1.Route
}

func TestRoutePathVersionCompatibility(t *testing.T) {
	cleanup := run.Koko(t)
	defer cleanup()

	dpCleanup := run.KongDP(kong.GetKongConfForShared())
	defer dpCleanup()
	require.NoError(t, util.WaitForKong(t))
	require.NoError(t, util.WaitForKongAdminAPI(t))

	admin := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  "http://localhost:3000",
		Reporter: httpexpect.NewRequireReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewCompactPrinter(t),
		},
	})

	service := &v1.Service{
		Id:   uuid.NewString(),
		Name: "foo",
		Host: "httpbin.org",
		Path: "/",
	}
	res := admin.POST("/v1/services").WithJSON(service).Expect()
	res.Status(http.StatusCreated)

	tests := []vcRoutesTC{
		{
			name: "plain path",
			route: &v1.Route{
				Id:    uuid.NewString(),
				Name:  "plain-path",
				Paths: []string{"/foo/bar"},
				Service: &v1.Service{
					Id: service.Id,
				},
			},
		},
		{
			name: "plain path, prefixed",
			route: &v1.Route{
				Id:    uuid.NewString(),
				Name:  "plain-path-prefixed",
				Paths: []string{"~/foo/bar"},
				Service: &v1.Service{
					Id: service.Id,
				},
			},
		},
		{
			name: "regex-like",
			route: &v1.Route{
				Id:    uuid.NewString(),
				Name:  "regex-like",
				Paths: []string{"/blog-\\d+"},
				Service: &v1.Service{
					Id: service.Id,
				},
			},
		},
		{
			name: "regex-like, prefixed",
			route: &v1.Route{
				Id:    uuid.NewString(),
				Name:  "regex-like-prefixed",
				Paths: []string{"~/blog-\\d+"},
				Service: &v1.Service{
					Id: service.Id,
				},
			},
		},
		{
			name: "mixed",
			route: &v1.Route{
				Id:    uuid.NewString(),
				Name:  "mixed",
				Paths: []string{"/foo", "~/bar", "/blog-\\d+", "~/bl[ao]go(sphere|web)/"},
			},
		},
	}

	for _, test := range tests {
		res := admin.POST("/v1/routes").WithJSON(test.route).Expect()
		res.Status(http.StatusCreated)
	}

	t.Run("pre 3.0", func(t *testing.T) {
		kongClient.RunWhenKong(t, "< 3.0.0")

		util.WaitFunc(t, func() error {
			return util.EnsureConfig(&v1.TestingConfig{
				Routes: []*v1.Route{
					{
						Name:  "plain-path",
						Paths: []string{"/foo/bar"},
					},
					{
						Name:  "plain-path-prefixed",
						Paths: []string{"/foo/bar"},
					},
					{
						Name:  "regex-like",
						Paths: []string{"/blog-\\d+"},
					},
					{
						Name:  "regex-like-prefixed",
						Paths: []string{"/blog-\\d+"},
					},
					{
						Name:  "mixed",
						Paths: []string{"/foo", "/bar", "/blog-\\d+", "/bl[ao]go(sphere|web)/"},
					},
				},
			})
		})
	})

	t.Run("3.0.0 and above", func(t *testing.T) {
		kongClient.RunWhenKong(t, ">= 3.0.0")

		util.WaitFunc(t, func() error {
			return util.EnsureConfig(&v1.TestingConfig{
				Routes: []*v1.Route{
					{
						Name:  "plain-path",
						Paths: []string{"/foo/bar"},
					},
					{
						Name:  "plain-path-prefixed",
						Paths: []string{"~/foo/bar"},
					},
					{
						Name:  "regex-like",
						Paths: []string{"/blog-\\d+"},
					},
					{
						Name:  "regex-like-prefixed",
						Paths: []string{"~/blog-\\d+"},
					},
					{
						Name:  "mixed",
						Paths: []string{"/foo", "~/bar", "/blog-\\d+", "~/bl[ao]go(sphere|web)/"},
					},
				},
			})
		})
	})
}

func TestTagsVersionCompatibility(t *testing.T) {
	cleanup := run.Koko(t)
	defer cleanup()

	dpCleanup := run.KongDP(kong.GetKongConfForShared())
	defer dpCleanup()
	require.NoError(t, util.WaitForKong(t))
	require.NoError(t, util.WaitForKongAdminAPI(t))

	admin := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  "http://localhost:3000",
		Reporter: httpexpect.NewRequireReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewCompactPrinter(t),
		},
	})

	service := &v1.Service{
		Id:   uuid.NewString(),
		Name: "service-1",
		Host: "example.com",
		Path: "/",
		Tags: []string{"tag-1", "tag-2", "tag 3"},
	}
	res := admin.POST("/v1/services").WithJSON(service).Expect()
	res.Status(http.StatusCreated)

	for _, route := range []*v1.Route{
		{Name: "path-1", Paths: []string{"/path/one"}},
		{Name: "path-2", Paths: []string{"/path/two"}, Tags: []string{"tag-1"}},
		{Name: "path-3", Paths: []string{"/path/three"}, Tags: []string{"tag 2"}},
		{Name: "path-4", Paths: []string{"/path/four"}, Tags: []string{"tag-3", "tag 4"}},
	} {
		route.Service = &v1.Service{Id: service.Id}
		res := admin.POST("/v1/routes").WithJSON(route).Expect()
		res.Status(http.StatusCreated)
	}

	t.Run("pre 3.0", func(t *testing.T) {
		kongClient.RunWhenKong(t, "< 3.0.0")

		util.WaitFunc(t, func() error {
			return util.EnsureConfig(&v1.TestingConfig{
				Services: []*v1.Service{{
					Name: "service-1",
					Tags: []string{"tag-1", "tag-2"},
				}},
				Routes: []*v1.Route{
					{Name: "path-1"},
					{Name: "path-2", Tags: []string{"tag-1"}},
					{Name: "path-3"},
					{Name: "path-4", Tags: []string{"tag-3"}},
				},
			})
		})
	})

	t.Run("3.0.0 and above", func(t *testing.T) {
		kongClient.RunWhenKong(t, ">= 3.0.0")

		util.WaitFunc(t, func() error {
			return util.EnsureConfig(&v1.TestingConfig{
				Services: []*v1.Service{{
					Name: "service-1",
					Tags: []string{"tag-1", "tag-2", "tag 3"},
				}},
				Routes: []*v1.Route{
					{Name: "path-1"},
					{Name: "path-2", Tags: []string{"tag-1"}},
					{Name: "path-3", Tags: []string{"tag 2"}},
					{Name: "path-4", Tags: []string{"tag-3", "tag 4"}},
				},
			})
		})
	})
}

func TestVersionCompatibility_300OrNewer(t *testing.T) {
	cleanup := run.Koko(t)
	defer cleanup()

	admin := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  "http://localhost:3000",
		Reporter: httpexpect.NewRequireReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewCompactPrinter(t),
		},
	})

	tests := []vcPlugins{
		{
			name: "statsd",
			config: `{
				"metrics": [
				  {
					"name": "status_count_per_workspace",
					"stat_type": "counter",
					"sample_rate": 1
				  },
				  {
					"name": "status_count_per_user_per_route",
					"stat_type": "counter",
					"sample_rate": 1,
					"consumer_identifier": "consumer_id"
				  },
				  {
					"name": "shdict_usage",
					"stat_type": "gauge",
					"sample_rate": 1
				  }
				]
			}`,
			expectedConfig: `{
				"allow_status_codes": null,
				"host": "localhost",
				"hostname_in_prefix": false,
				"metrics": [
					{
						"name": "status_count_per_workspace",
						"stat_type": "counter",
						"sample_rate": 1,
						"consumer_identifier": null,
						"service_identifier": null,
						"workspace_identifier": null
					},
					{
						"name": "status_count_per_user_per_route",
						"stat_type": "counter",
						"sample_rate": 1,
						"consumer_identifier": "consumer_id",
						"service_identifier": null,
						"workspace_identifier": null
					},
					{
						"name": "shdict_usage",
						"stat_type": "gauge",
						"sample_rate": 1,
						"consumer_identifier": null,
						"service_identifier": null,
						"workspace_identifier": null
					}
				],
				"port": 8125,
				"prefix": "kong",
				"udp_packet_size": 0,
				"use_tcp": false,
				"consumer_identifier_default": "custom_id",
				"service_identifier_default": "service_name_or_host",
				"workspace_identifier_default": "workspace_id"

			}`,
		},
		{
			name: "aws-lambda",
			config: `{
				"aws_region": "AWS_REGION",
				"function_name": "FUNCTION_NAME",
				"aws_assume_role_arn": "foo",
				"aws_role_session_name": "kong"
			}`,
			expectedConfig: `{
				"aws_assume_role_arn": "foo",
				"aws_key": null,
				"aws_region": "AWS_REGION",
				"aws_role_session_name": "kong",
				"aws_secret": null,
				"awsgateway_compatible": false,
				"base64_encode_body": true,
				"forward_request_body": false,
				"forward_request_headers": false,
				"forward_request_method": false,
				"forward_request_uri": false,
				"function_name": "FUNCTION_NAME",
				"host": null,
				"invocation_type": "RequestResponse",
				"is_proxy_integration": false,
				"keepalive": 60000,
				"log_type": "Tail",
				"port": 443,
				"proxy_url": null,
				"qualifier": null,
				"skip_large_bodies": true,
				"timeout": 60000,
				"unhandled_status": null
			}`,
		},
	}
	expectedConfig := &v1.TestingConfig{
		Plugins: make([]*v1.Plugin, 0, len(tests)),
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

		var expected structpb.Struct
		require.NoError(t, json.ProtoJSONUnmarshal([]byte(test.expectedConfig), &expected))
		expectedConfig.Plugins = append(expectedConfig.Plugins, &v1.Plugin{
			Id:        plugin.Id,
			Name:      plugin.Name,
			Config:    &expected,
			Enabled:   plugin.Enabled,
			Protocols: plugin.Protocols,
		})
	}

	dpCleanup := run.KongDP(kong.GetKongConfForShared())
	defer dpCleanup()
	require.NoError(t, util.WaitForKong(t))
	require.NoError(t, util.WaitForKongAdminAPI(t))

	kongClient.RunWhenKong(t, ">=3.0.0")

	util.WaitFunc(t, func() error {
		err := util.EnsureConfig(expectedConfig)
		t.Log("plugin validation failed", err)
		return err
	})
}

// Ensure valid field configuration for DP >= 3.1.
func TestVersionCompatibility_310OrNewer(t *testing.T) {
	cleanup := run.Koko(t)
	defer cleanup()

	admin := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  "http://localhost:3000",
		Reporter: httpexpect.NewRequireReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewCompactPrinter(t),
		},
	})

	tests := []vcPlugins{
		{
			name: "zipkin",
			config: `{
				"local_service_name": "LOCAL_SERVICE_NAME",
				"header_type": "ignore",
				"http_span_name": "method_path",
				"connect_timeout": 2001,
				"send_timeout": 2001,
				"read_timeout": 2001,
				"http_response_header_for_traceid": "X-B3-TraceId"
			}`,
			expectedConfig: `{
				"local_service_name": "LOCAL_SERVICE_NAME",
				"header_type": "ignore",
				"http_span_name": "method_path",
				"connect_timeout": 2001,
				"send_timeout": 2001,
				"read_timeout": 2001,
				"http_response_header_for_traceid": "X-B3-TraceId"
			}`,
		},
		{
			name: "rate-limiting",
			config: `{
				"hour": 1,
				"redis_ssl": true,
				"redis_ssl_verify": true,
				"redis_server_name": "redis.example.com",
				"redis_username": "REDIS_USERNAME",
				"redis_password": "REDIS_PASSWORD",
				"error_code": 429,
				"error_message": "API rate limit exceeded"
			}`,
			expectedConfig: `{
				"hour": 1,
				"redis_ssl": true,
				"redis_ssl_verify": true,
				"redis_server_name": "redis.example.com",
				"redis_username": "REDIS_USERNAME",
				"redis_password": "REDIS_PASSWORD",
				"error_code": 429,
				"error_message": "API rate limit exceeded"
			}`,
		},
		{
			name: "acme",
			config: `{
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
			expectedConfig: `{
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
		{
			name: "response-ratelimiting",
			config: `{
				"limits": {
					"sms": {
						"second": 42
					}
				},
				"redis_ssl": true,
				"redis_ssl_verify": true,
				"redis_server_name": "test.com"
			}`,
			expectedConfig: `{
				"limits": {
					"sms": {
						"second": 42
					}
				},
				"redis_ssl": true,
				"redis_ssl_verify": true,
				"redis_server_name": "test.com"
			}`,
		},
	}
	expectedConfig := &v1.TestingConfig{
		Plugins: make([]*v1.Plugin, 0, len(tests)),
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

		var expected structpb.Struct
		require.NoError(t, json.ProtoJSONUnmarshal([]byte(test.expectedConfig), &expected))
		expectedConfig.Plugins = append(expectedConfig.Plugins, &v1.Plugin{
			Id:        plugin.Id,
			Name:      plugin.Name,
			Config:    &expected,
			Enabled:   plugin.Enabled,
			Protocols: plugin.Protocols,
		})
	}

	dpCleanup := run.KongDP(kong.GetKongConfForShared())
	defer dpCleanup()
	require.NoError(t, util.WaitForKong(t))
	require.NoError(t, util.WaitForKongAdminAPI(t))

	kongClient.RunWhenKong(t, ">=3.1.0")

	util.WaitFunc(t, func() error {
		err := util.EnsureConfig(expectedConfig)
		t.Log("plugin validation failed", err)
		return err
	})
}

func intMin(lhs int, rhs int) int {
	if lhs < rhs {
		return lhs
	}
	return rhs
}

func TestVersionCompatibility_KeysEntities(t *testing.T) {
	cleanup := run.Koko(t)
	defer cleanup()

	dpCleanup := run.KongDP(kong.GetKongConfForShared())
	defer dpCleanup()
	require.NoError(t, util.WaitForKong(t))
	require.NoError(t, util.WaitForKongAdminAPI(t))

	admin := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  "http://localhost:3000",
		Reporter: httpexpect.NewRequireReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewCompactPrinter(t),
		},
	})

	// Create a service
	service := &v1.Service{
		Id:   uuid.NewString(),
		Name: "foo",
		Host: "httpbin.org",
		Path: "/",
	}
	res := admin.POST("/v1/services").WithJSON(service).Expect()
	res.Status(http.StatusCreated)

	admin.POST("/v1/keys").WithJSON(&v1.Key{
		Name: "mellon",
		Jwk:  &v1.JwkKey{},
	}).Expect().Status(http.StatusCreated)

	t.Run("pre 3.0", func(t *testing.T) {
		kongClient.RunWhenKong(t, "< 3.0.0")

		util.WaitFunc(t, func() error {
			return util.EnsureConfig(&v1.TestingConfig{
				Services: []*v1.Service{{Name: "foo"}},
			})
		})
	})

	t.Run("pre 3.1", func(t *testing.T) {
		kongClient.RunWhenKong(t, ">= 3.0.0 < 3.1.0")

		util.WaitFunc(t, func() error {
			return util.EnsureConfig(&v1.TestingConfig{
				Services: []*v1.Service{{Name: "foo"}},
			})
		})
	})

	t.Run("3.1 and above", func(t *testing.T) {
		kongClient.RunWhenKong(t, ">= 3.1.0")

		util.WaitFunc(t, func() error {
			return util.EnsureConfig(&v1.TestingConfig{
				Services: []*v1.Service{{Name: "foo"}},
				Keys:     []*v1.Key{{Name: "mellon"}},
			})
		})
	})
}
