//go:build integration

package e2e

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/cenkalti/backoff/v4"
	"github.com/gavv/httpexpect/v2"
	"github.com/google/uuid"
	kongClient "github.com/kong/go-kong/kong"
	"github.com/kong/koko/internal/cmd"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/test/kong"
	"github.com/kong/koko/internal/test/run"
	"github.com/kong/koko/internal/test/util"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func goodService() *v1.Service {
	return &v1.Service{
		Name: "foo",
		Host: "example.com",
		Path: "/",
	}
}

func disabledService() *v1.Service {
	return &v1.Service{
		Name:    "bar",
		Host:    "example-bar.com",
		Path:    "/",
		Enabled: wrapperspb.Bool(false),
	}
}

func TestSharedMTLS(t *testing.T) {
	// ensure that Kong Gateway can connect using Shared MTLS mode
	cleanup := run.Koko(t)
	defer cleanup()

	service := &v1.Service{
		Id:   uuid.NewString(),
		Name: "foo",
		Host: "httpbin.org",
		Path: "/",
	}
	c := httpexpect.New(t, "http://localhost:3000")
	res := c.POST("/v1/services").WithJSON(service).Expect()
	res.Status(201)
	route := &v1.Route{
		Name:  "bar",
		Paths: []string{"/"},
		Service: &v1.Service{
			Id: service.Id,
		},
	}
	res = c.POST("/v1/routes").WithJSON(route).Expect()
	res.Status(201)

	dpCleanup := run.KongDP(kong.GetKongConfForShared())
	defer dpCleanup()

	require.Nil(t, util.WaitForKongPort(t, 8001))

	expectedConfig := &v1.TestingConfig{
		Services: []*v1.Service{service},
		Routes:   []*v1.Route{route},
	}
	util.WaitFunc(t, func() error {
		err := util.EnsureConfig(expectedConfig)
		return err
	})
}

func TestPKIMTLS(t *testing.T) {
	// ensure that Kong Gateway can connect using PKI MTLS mode
	cleanup := run.Koko(t, run.WithDPAuthMode(cmd.DPAuthPKIMTLS))
	defer cleanup()

	service := &v1.Service{
		Id:   uuid.NewString(),
		Name: "foo",
		Host: "httpbin.org",
		Path: "/",
	}
	c := httpexpect.New(t, "http://localhost:3000")
	res := c.POST("/v1/services").WithJSON(service).Expect()
	res.Status(201)
	route := &v1.Route{
		Name:  "bar",
		Paths: []string{"/"},
		Service: &v1.Service{
			Id: service.Id,
		},
	}
	res = c.POST("/v1/routes").WithJSON(route).Expect()
	res.Status(201)

	dpCleanup := run.KongDP(kong.GetKongConf())
	defer dpCleanup()

	require.Nil(t, util.WaitForKongPort(t, 8001))

	expectedConfig := &v1.TestingConfig{
		Services: []*v1.Service{service},
		Routes:   []*v1.Route{route},
	}
	util.WaitFunc(t, func() error {
		return util.EnsureConfig(expectedConfig)
	})
}

func TestHealthEndpointOnCPPort(t *testing.T) {
	// ensure that health-check is enabled on the CP port
	cleanup := run.Koko(t)
	defer cleanup()

	// test the endpoint
	require.Nil(t, backoff.Retry(func() error {
		client := insecureHTTPClient()
		res, err := client.Get("https://localhost:3100/health")
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status code: %v", res.StatusCode)
		}
		return nil
	}, util.TestBackoff))
}

func insecureHTTPClient() *http.Client {
	transport := http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	return &http.Client{Transport: &transport}
}

func TestNodesEndpoint(t *testing.T) {
	// ensure that gateway nodes are tracked in database
	cleanup := run.Koko(t)
	defer cleanup()

	service := &v1.Service{
		Id:   uuid.NewString(),
		Name: "foo",
		Host: "httpbin.org",
		Path: "/",
	}
	c := httpexpect.New(t, "http://localhost:3000")
	res := c.POST("/v1/services").WithJSON(service).Expect()
	res.Status(201)
	route := &v1.Route{
		Name:  "bar",
		Paths: []string{"/"},
		Service: &v1.Service{
			Id: service.Id,
		},
	}
	res = c.POST("/v1/routes").WithJSON(route).Expect()
	res.Status(201)

	dpCleanup := run.KongDP(kong.GetKongConfForShared())
	defer dpCleanup()

	require.Nil(t, util.WaitForKong(t))

	// ensure kong node is up
	require.Nil(t, util.WaitForKongPort(t, 8001))

	expectedConfig := &v1.TestingConfig{
		Services: []*v1.Service{service},
		Routes:   []*v1.Route{route},
	}
	util.WaitFunc(t, func() error {
		return util.EnsureConfig(expectedConfig)
	})

	util.WaitFunc(t, func() error {
		// once node is up, check the status endpoint
		res = c.GET("/v1/nodes").Expect()
		res.Status(http.StatusOK)
		body := gjson.Parse(res.Body().Raw())
		hash := body.Get("items.0.config_hash").String()
		if len(hash) != 32 {
			return fmt.Errorf(
				"expected config hash to be 32 character long")
		}
		if hash == strings.Repeat("0", 32) {
			return fmt.Errorf("expected hash to not be a string of 0s")
		}
		return nil
	})
}

func TestPluginSync(t *testing.T) {
	// ensure that plugins can be synced to Kong gateway
	cleanup := run.Koko(t)
	defer cleanup()

	service := &v1.Service{
		Id:   uuid.NewString(),
		Name: "foo",
		Host: "example.com",
		Path: "/",
	}
	c := httpexpect.New(t, "http://localhost:3000")
	res := c.POST("/v1/services").WithJSON(service).Expect()
	res.Status(201)

	route := &v1.Route{
		Id:    uuid.NewString(),
		Name:  "bar",
		Paths: []string{"/bar"},
		Service: &v1.Service{
			Id: service.Id,
		},
	}
	res = c.POST("/v1/routes").WithJSON(route).Expect()
	res.Status(201)

	var expectedPlugins []*v1.Plugin
	plugin := &v1.Plugin{
		Name:      "key-auth",
		Enabled:   wrapperspb.Bool(true),
		Service:   &v1.Service{Id: service.Id},
		Protocols: []string{"http", "https"},
	}
	pluginBytes, err := json.Marshal(plugin)
	require.Nil(t, err)
	res = c.POST("/v1/plugins").WithBytes(pluginBytes).Expect()
	res.Status(201)
	expectedPlugins = append(expectedPlugins, plugin)

	plugin = &v1.Plugin{
		Name:      "basic-auth",
		Enabled:   wrapperspb.Bool(true),
		Route:     &v1.Route{Id: route.Id},
		Protocols: []string{"http", "https"},
	}
	pluginBytes, err = json.Marshal(plugin)
	require.Nil(t, err)
	res = c.POST("/v1/plugins").WithBytes(pluginBytes).Expect()
	res.Status(201)
	expectedPlugins = append(expectedPlugins, plugin)

	plugin = &v1.Plugin{
		Name:      "request-transformer",
		Enabled:   wrapperspb.Bool(true),
		Protocols: []string{"http", "https"},
	}
	pluginBytes, err = json.Marshal(plugin)
	require.Nil(t, err)
	res = c.POST("/v1/plugins").WithBytes(pluginBytes).Expect()
	res.Status(201)
	expectedPlugins = append(expectedPlugins, plugin)

	dpCleanup := run.KongDP(kong.GetKongConfForShared())
	defer dpCleanup()

	require.Nil(t, util.WaitForKongPort(t, 8001))

	expectedConfig := &v1.TestingConfig{
		Services: []*v1.Service{service},
		Routes:   []*v1.Route{route},
		Plugins:  expectedPlugins,
	}
	util.WaitFunc(t, func() error {
		err := util.EnsureConfig(expectedConfig)
		t.Log("configuration mismatch", err)
		return err
	})
}

func TestUpstreamSync(t *testing.T) {
	// ensure that upstreams can be synced to Kong gateway
	cleanup := run.Koko(t)
	defer cleanup()

	upstream := &v1.Upstream{
		Id:   uuid.NewString(),
		Name: "foo",
	}
	c := httpexpect.New(t, "http://localhost:3000")
	res := c.POST("/v1/upstreams").WithJSON(upstream).Expect()
	res.Status(201)

	dpCleanup := run.KongDP(kong.GetKongConfForShared())
	defer dpCleanup()

	require.Nil(t, util.WaitForKongPort(t, 8001))

	expectedConfig := &v1.TestingConfig{
		Upstreams: []*v1.Upstream{upstream},
	}
	util.WaitFunc(t, func() error {
		err := util.EnsureConfig(expectedConfig)
		t.Log("configuration mismatch", err)
		return err
	})
}

func TestConsumerSync(t *testing.T) {
	// ensure that consumers can be synced to Kong gateway
	cleanup := run.Koko(t)
	defer cleanup()

	consumer := &v1.Consumer{
		Id:       uuid.NewString(),
		Username: "testConsumer",
	}
	// create the consumer in CP
	c := httpexpect.New(t, "http://localhost:3000")
	res := c.POST("/v1/consumers").WithJSON(consumer).Expect()
	res.Status(201)

	// launch the DP
	dpCleanup := run.KongDP(kong.GetKongConfForShared())
	defer dpCleanup()

	// wait for DP to come-up
	require.Nil(t, util.WaitForKongPort(t, 8001))

	expectedConfig := &v1.TestingConfig{
		Consumers: []*v1.Consumer{consumer},
	}
	util.WaitFunc(t, func() error {
		err := util.EnsureConfig(expectedConfig)
		t.Log("configuration mismatch for consumer", err)
		return err
	})
}

func TestTargetSync(t *testing.T) {
	// ensure that target can be synced to Kong gateway
	cleanup := run.Koko(t)
	defer cleanup()

	uid := uuid.NewString()
	upstream := &v1.Upstream{
		Id:   uid,
		Name: "foo",
	}
	c := httpexpect.New(t, "http://localhost:3000")
	res := c.POST("/v1/upstreams").WithJSON(upstream).Expect()
	res.Status(201)

	target := &v1.Target{
		Target:   "10.0.42.42:8000",
		Upstream: &v1.Upstream{Id: uid},
	}
	res = c.POST("/v1/targets").WithJSON(target).Expect()
	res.Status(201)

	dpCleanup := run.KongDP(kong.GetKongConfForShared())
	defer dpCleanup()

	require.Nil(t, util.WaitForKongPort(t, 8001))

	expectedConfig := &v1.TestingConfig{
		Upstreams: []*v1.Upstream{upstream},
		Targets:   []*v1.Target{target},
	}
	util.WaitFunc(t, func() error {
		err := util.EnsureConfig(expectedConfig)
		t.Log("configuration mismatch", err)
		return err
	})
}

func TestServiceSync(t *testing.T) {
	// ensure that services can be synced to Kong gateway
	// only enabled services should be synced though
	cleanup := run.Koko(t)
	defer cleanup()

	enabledService := goodService()
	c := httpexpect.New(t, "http://localhost:3000")
	res := c.POST("/v1/services").WithJSON(enabledService).Expect()
	res.Status(201)

	disabled, err := json.Marshal(disabledService())
	require.Nil(t, err)
	res = c.POST("/v1/services").WithBytes(disabled).Expect()
	res.Status(201)

	dpCleanup := run.KongDP(kong.GetKongConfForShared())
	defer dpCleanup()

	require.Nil(t, util.WaitForKongPort(t, 8001))

	expectedConfig := &v1.TestingConfig{
		Services: []*v1.Service{enabledService},
	}
	util.WaitFunc(t, func() error {
		err := util.EnsureConfig(expectedConfig)
		t.Log("configuration mismatch", err)
		return err
	})
}

func TestRouteHeader(t *testing.T) {
	// ensure that routes with headers can be synced to Kong gateway
	// this is done because the data-structures for headers in Koko and Kong
	// are different
	cleanup := run.Koko(t)
	defer cleanup()

	service := &v1.Service{
		Id:   uuid.NewString(),
		Name: "foo",
		Host: "httpbin.org",
		Path: "/",
	}
	c := httpexpect.New(t, "http://localhost:3000")
	res := c.POST("/v1/services").WithJSON(service).Expect()
	res.Status(201)
	route := &v1.Route{
		Name:  "bar",
		Paths: []string{"/"},
		Headers: map[string]*v1.HeaderValues{
			"foo": {
				Values: []string{"bar", "baz"},
			},
		},
		Service: &v1.Service{
			Id: service.Id,
		},
	}
	res = c.POST("/v1/routes").WithJSON(route).Expect()
	res.Status(201)

	dpCleanup := run.KongDP(kong.GetKongConfForShared())
	defer dpCleanup()

	require.Nil(t, util.WaitForKongPort(t, 8001))
	util.WaitFunc(t, func() error {
		ctx := context.Background()
		client, err := kongClient.NewClient(util.BasedKongAdminAPIAddr, nil)
		if err != nil {
			return fmt.Errorf("create go client for kong: %v", err)
		}
		routes, err := client.Routes.ListAll(ctx)
		if err != nil {
			return fmt.Errorf("fetch routes: %v", err)
		}
		if len(routes) != 1 {
			return fmt.Errorf("expected %v routes but got %v routes", 1,
				len(routes))
		}
		route := routes[0]
		if len(route.Headers["foo"]) != 2 {
			return fmt.Errorf("expected route.Headers."+
				"foo to have 2 values but got %v", len(route.Headers["foo"]))
		}
		return nil
	})
}

func TestDataPlanePluginCheck(t *testing.T) {
	// ensure that a data-plane that doesn't meet the pre-requisites is
	// tracked as a node and has a corresponding status entry
	cleanup := run.Koko(t)
	defer cleanup()

	conf := kong.GetKongConfForShared()
	conf.EnvVars["KONG_PLUGINS"] = "datadog,acl"
	dpCleanup := run.KongDP(conf)
	defer dpCleanup()

	require.Nil(t, util.WaitForKongPort(t, 8001))
	c := httpexpect.New(t, "http://localhost:3000")

	util.WaitFunc(t, func() error {
		res := c.GET("/v1/nodes").Expect()
		res.Status(http.StatusOK)
		body := gjson.Parse(res.Body().Raw())
		nodeID := body.Get("items.0.id").String()
		if nodeID == "" {
			return fmt.Errorf("expected a node entry")
		}

		res = c.GET("/v1/statuses").Expect()
		res.Status(http.StatusOK)
		body = gjson.Parse(res.Body().Raw())
		refType := body.Get("items.0.context_reference.type").String()
		if refType != "node" {
			return fmt.Errorf("expected a status entry for node")
		}

		refID := body.Get("items.0.context_reference.id").String()
		if refType != "node" {
			return fmt.Errorf("expected a status entry for node")
		}
		if refID != nodeID {
			return fmt.Errorf("expected node ID and status' reference id to match" +
				" up")
		}

		conditionCode := body.Get("items.0.conditions.0.code").String()
		if conditionCode != "DP001" {
			return fmt.Errorf("expected condition code to be DP001")
		}
		conditionMessage := body.Get("items.0.conditions.0.message").String()
		if conditionMessage != "kong data-plane node missing plugin[DP001]: rate-limiting" {
			return fmt.Errorf("unexpected condition code")
		}
		return nil
	})
}
