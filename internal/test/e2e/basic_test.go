//go:build integration

package e2e

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"testing"

	"github.com/cenkalti/backoff/v4"
	"github.com/gavv/httpexpect/v2"
	"github.com/google/uuid"
	kongClient "github.com/kong/go-kong/kong"
	"github.com/kong/koko/internal/cmd"
	"github.com/kong/koko/internal/config"
	"github.com/kong/koko/internal/crypto"
	"github.com/kong/koko/internal/db"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/log"
	"github.com/kong/koko/internal/test/certs"
	"github.com/kong/koko/internal/test/kong"
	"github.com/kong/koko/internal/test/util"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestSharedMTLS(t *testing.T) {
	// ensure that Kong Gateway can connect using Shared MTLS mode
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cert, err := tls.X509KeyPair(certs.DefaultSharedCert, certs.DefaultSharedKey)
	require.Nil(t, err)
	go func() {
		require.Nil(t, util.CleanDB())
		require.Nil(t, cmd.Run(ctx, cmd.ServerConfig{
			DPAuthCert: cert,
			KongCPCert: cert,
			Logger:     log.Logger,
			Database: config.Database{
				Dialect: db.DialectSQLite3,
				SQLite: config.SQLite{
					InMemory: true,
				},
			},
		}))
	}()
	require.Nil(t, util.WaitForAdminAPI(t))

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

	go func() {
		_ = kong.RunDP(ctx, kong.GetKongConfForShared())
	}()

	require.Nil(t, util.WaitForKongPort(t, 8001))

	expectedConfig := &v1.TestingConfig{
		Services: []*v1.Service{service},
		Routes:   []*v1.Route{route},
	}
	util.WaitFunc(t, func() error {
		return util.EnsureConfig(expectedConfig)
	})
}

func TestPKIMTLS(t *testing.T) {
	// ensure that Kong Gateway can connect using PKI MTLS mode
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cpCert, err := tls.X509KeyPair(certs.CPCert, certs.CPKey)
	require.Nil(t, err)

	dpCACert, err := crypto.ParsePEMCerts(certs.DPTree1CACert)
	require.Nil(t, err)
	go func() {
		require.Nil(t, util.CleanDB())
		require.Nil(t, cmd.Run(ctx, cmd.ServerConfig{
			Logger: log.Logger,

			KongCPCert: cpCert,

			DPAuthMode:    cmd.DPAuthPKIMTLS,
			DPAuthCACerts: dpCACert,
			Database: config.Database{
				Dialect: db.DialectSQLite3,
				SQLite: config.SQLite{
					InMemory: true,
				},
			},
		}))
	}()
	require.Nil(t, util.WaitForAdminAPI(t))

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

	go func() {
		_ = kong.RunDP(ctx, kong.GetKongConf())
	}()

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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cert, err := tls.X509KeyPair(certs.DefaultSharedCert, certs.DefaultSharedKey)
	require.Nil(t, err)
	go func() {
		require.Nil(t, util.CleanDB())
		require.Nil(t, cmd.Run(ctx, cmd.ServerConfig{
			DPAuthCert: cert,
			KongCPCert: cert,
			Logger:     log.Logger,
			Database: config.Database{
				Dialect: db.DialectSQLite3,
				SQLite: config.SQLite{
					InMemory: true,
				},
			},
		}))
	}()
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cert, err := tls.X509KeyPair(certs.DefaultSharedCert, certs.DefaultSharedKey)
	require.Nil(t, err)
	go func() {
		require.Nil(t, util.CleanDB())
		require.Nil(t, cmd.Run(ctx, cmd.ServerConfig{
			DPAuthCert: cert,
			KongCPCert: cert,
			Logger:     log.Logger,
			Database: config.Database{
				Dialect: db.DialectSQLite3,
				SQLite: config.SQLite{
					InMemory: true,
				},
			},
		}))
	}()
	require.Nil(t, util.WaitForAdminAPI(t))

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

	go func() {
		_ = kong.RunDP(ctx, kong.GetKongConfForShared())
	}()
	testing.Verbose()
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

	// once node is up, check the status endpoint
	res = c.GET("/v1/nodes").Expect()
	res.Status(http.StatusOK)
	nodes := res.JSON().Object().Value("items").Array()
	nodes.Length().Equal(1)
}

func TestPluginSync(t *testing.T) {
	// ensure that plugins can be synced to Kong gateway
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cert, err := tls.X509KeyPair(certs.DefaultSharedCert, certs.DefaultSharedKey)
	require.Nil(t, err)
	go func() {
		require.Nil(t, util.CleanDB())
		require.Nil(t, cmd.Run(ctx, cmd.ServerConfig{
			DPAuthCert: cert,
			KongCPCert: cert,
			Logger:     log.Logger,
			Database: config.Database{
				Dialect: db.DialectSQLite3,
				SQLite: config.SQLite{
					InMemory: true,
				},
			},
		}))
	}()
	require.Nil(t, util.WaitForAdminAPI(t))

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

	go func() {
		_ = kong.RunDP(ctx, kong.GetKongConfForShared())
	}()

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

func TestRouteHeader(t *testing.T) {
	// ensure that routes with headers can be synced to Kong gateway
	// this is done because the data-structures for headers in Koko and Kong
	// are different
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cert, err := tls.X509KeyPair(certs.DefaultSharedCert, certs.DefaultSharedKey)
	require.Nil(t, err)
	go func() {
		require.Nil(t, util.CleanDB())
		require.Nil(t, cmd.Run(ctx, cmd.ServerConfig{
			DPAuthCert: cert,
			KongCPCert: cert,
			Logger:     log.Logger,
			Database: config.Database{
				Dialect: db.DialectSQLite3,
				SQLite: config.SQLite{
					InMemory: true,
				},
			},
		}))
	}()
	require.Nil(t, util.WaitForAdminAPI(t))

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

	go func() {
		_ = kong.RunDP(ctx, kong.GetKongConfForShared())
	}()

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
