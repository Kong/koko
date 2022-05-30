//go:build integration

package e2e

import (
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/google/uuid"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/persistence"
	"github.com/kong/koko/internal/test/certs"
	"github.com/kong/koko/internal/test/kong"
	"github.com/kong/koko/internal/test/run"
	"github.com/kong/koko/internal/test/util"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// setting a number higher than the max page size
// to make sure pagination works.
const numberOfEntities = persistence.MaxLimit + 1

func goodServiceWithIntNumber(i int) *v1.Service {
	return &v1.Service{
		Id:   uuid.NewString(),
		Name: fmt.Sprintf("foo%d", i),
		Host: fmt.Sprintf("example%d.com", i),
		Path: "/",
	}
}

func goodRoute(i int) *v1.Route {
	return &v1.Route{
		Name:  fmt.Sprintf("foo%d", i),
		Paths: []string{"/"},
	}
}

func goodCert() *v1.Certificate {
	return &v1.Certificate{
		Id:   uuid.NewString(),
		Cert: string(certs.DefaultSharedCert),
		Key:  string(certs.DefaultSharedKey),
	}
}

func goodConsumer(i int) *v1.Consumer {
	return &v1.Consumer{
		Id:       uuid.NewString(),
		Username: fmt.Sprintf("consumer%d", i),
	}
}

func goodPlugin(serviceID string) *v1.Plugin {
	return &v1.Plugin{
		Name:      "key-auth",
		Enabled:   wrapperspb.Bool(true),
		Service:   &v1.Service{Id: serviceID},
		Protocols: []string{"http", "https"},
	}
}

func goodSNI(certID string, i int) *v1.SNI {
	return &v1.SNI{
		Id:   uuid.NewString(),
		Name: fmt.Sprintf("test-%d.example.com", i),
		Certificate: &v1.Certificate{
			Id: certID,
		},
	}
}

func goodUpstream(i int) *v1.Upstream {
	return &v1.Upstream{
		Id:   uuid.NewString(),
		Name: fmt.Sprintf("upstream%d", i),
	}
}

func goodTarget(upstreamID string) *v1.Target {
	return &v1.Target{
		Target:   "10.0.42.42:8000",
		Upstream: &v1.Upstream{Id: upstreamID},
	}
}

func TestCPPagination(t *testing.T) {
	// ensure pagination works for CP
	cleanup := run.Koko(t)
	defer cleanup()

	dpCleanup := run.KongDP(kong.GetKongConfForShared())
	defer dpCleanup()

	c := httpexpect.New(t, "http://localhost:3000")

	services := make([]*v1.Service, numberOfEntities)
	routes := make([]*v1.Route, numberOfEntities)
	plugins := make([]*v1.Plugin, numberOfEntities)
	consumers := make([]*v1.Consumer, numberOfEntities)
	certs := make([]*v1.Certificate, numberOfEntities)
	snis := make([]*v1.SNI, numberOfEntities)
	targets := make([]*v1.Target, numberOfEntities)
	upstreams := make([]*v1.Upstream, numberOfEntities)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < numberOfEntities; i++ {
			service := goodServiceWithIntNumber(i)
			c.POST("/v1/services").WithJSON(service).Expect().Status(http.StatusCreated)
			services[i] = service
		}
	}()
	wg.Wait()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < numberOfEntities; i++ {
			plugin := goodPlugin(services[i].Id)
			pluginBytes, err := json.ProtoJSONMarshal(plugin)
			require.NoError(t, err)
			c.POST("/v1/plugins").WithBytes(pluginBytes).Expect().Status(http.StatusCreated)
			plugins[i] = plugin
		}
	}()
	wg.Wait()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < numberOfEntities; i++ {
			route := goodRoute(i)
			c.POST("/v1/routes").WithJSON(route).Expect().Status(http.StatusCreated)
			routes[i] = route
		}
	}()
	wg.Wait()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < numberOfEntities; i++ {
			consumer := goodConsumer(i)
			c.POST("/v1/consumers").WithJSON(consumer).Expect().Status(http.StatusCreated)
			consumers[i] = consumer
		}
	}()
	wg.Wait()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < numberOfEntities; i++ {
			cert := goodCert()
			c.POST("/v1/certificates").WithJSON(cert).Expect().Status(http.StatusCreated)
			certs[i] = cert
		}
	}()
	wg.Wait()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < numberOfEntities; i++ {
			sni := goodSNI(certs[i].Id, i)
			c.POST("/v1/snis").WithJSON(sni).Expect().Status(http.StatusCreated)
			snis[i] = sni
		}
	}()
	wg.Wait()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < numberOfEntities; i++ {
			upstream := goodUpstream(i)
			c.POST("/v1/upstreams").WithJSON(upstream).Expect().Status(http.StatusCreated)
			upstreams[i] = upstream
		}
	}()
	wg.Wait()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < numberOfEntities; i++ {
			target := goodTarget(upstreams[i].Id)
			c.POST("/v1/targets").WithJSON(target).Expect().Status(http.StatusCreated)
			targets[i] = target
		}
	}()
	wg.Wait()

	require.NoError(t, util.WaitForKongPort(t, 8001))

	expectedConfig := &v1.TestingConfig{
		Services:     services,
		Routes:       routes,
		Plugins:      plugins,
		Consumers:    consumers,
		Certificates: certs,
		Snis:         snis,
		Upstreams:    upstreams,
		Targets:      targets,
	}
	util.WaitFunc(t, func() error {
		err := util.EnsureConfig(expectedConfig)
		t.Log("configuration mismatch", err)
		return err
	})
}
