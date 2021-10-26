//go:build integration
package e2e

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/gavv/httpexpect/v2"
	"github.com/google/uuid"
	"github.com/kong/koko/internal/cmd"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/log"
	"github.com/kong/koko/internal/test/kong"
	"github.com/stretchr/testify/require"
)

func TestGatewayConfig(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		require.Nil(t, cmd.Run(ctx, log.Logger))
	}()
	defer cancel()
	service := &v1.Service{
		Id:   uuid.NewString(),
		Name: "foo",
		Host: "httpbin.org",
		Path: "/",
	}
	require.Nil(t, waitForAdminAPI(t))
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
		_ = kong.RunDP(ctx, getKongConf())
	}()
	testing.Verbose()
	require.Nil(t, waitForKong(t))

	// test the route
	require.Nil(t, backoff.Retry(func() error {
		res, err := http.Get("http://localhost:8000/headers")
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status code: %v", res.StatusCode)
		}
		return nil
	}, backoffer))
}

var path404 = "61d624f6-59fb-45a0-9892-9f6e81264f3e"

var backoffer backoff.BackOff

func init() {
	backoffer = backoff.NewConstantBackOff(1 * time.Second)
	backoffer = backoff.WithMaxRetries(backoffer, 30)
}

func waitForKong(t *testing.T) error {
	return backoff.RetryNotify(func() error {
		res, err := http.Get("http://localhost:8000/" + path404)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if res.StatusCode == http.StatusNotFound {
			return nil
		}
		panic(fmt.Sprintf("unexpected status code: %v", res.StatusCode))
	}, backoffer, func(err error, duration time.Duration) {
		if err != nil {
			t.Log("waiting for kong DP")
		}
	})
}

func waitForAdminAPI(t *testing.T) error {
	return backoff.RetryNotify(func() error {
		res, err := http.Get("http://localhost:3000/v1/meta/version")
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if res.StatusCode == http.StatusOK {
			return nil
		}
		panic(fmt.Sprintf("unexpected status code: %v", res.StatusCode))
	}, backoffer, func(err error, duration time.Duration) {
		if err != nil {
			t.Log("waiting for admin API")
		}
	})
}

func getKongConf() kong.DockerInput {
	kongVersion := os.Getenv("KOKO_TEST_KONG_DP_VERSION")
	if kongVersion == "" {
		panic("no KOKO_TEST_KONG_DP_VERSION set")
	}
	res := kong.DockerInput{Version: kongVersion}
	if testing.Verbose() {
		k := "KONG_LOG_LEVEL"
		v := "debug"
		if _, ok := res.EnvVars[k]; !ok {
			res.EnvVars[k] = v
		}
	}
	return res
}
