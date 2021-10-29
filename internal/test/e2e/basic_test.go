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
	"github.com/kong/koko/internal/cmd"
	"github.com/kong/koko/internal/crypto"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/log"
	"github.com/kong/koko/internal/test/certs"
	"github.com/kong/koko/internal/test/kong"
	"github.com/kong/koko/internal/test/util"
	"github.com/stretchr/testify/require"
)

func TestSharedMTLS(t *testing.T) {
	// ensure that Kong Gateway can connect using Shared MTLS mode
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cert, err := tls.X509KeyPair(certs.DefaultSharedCert, certs.DefaultSharedKey)
	require.Nil(t, err)
	go func() {
		require.Nil(t, cmd.Run(ctx, cmd.ServerConfig{
			DPAuthCert: cert,
			KongCPCert: cert,
			Logger:     log.Logger,
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
	}, util.TestBackoff))
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
		require.Nil(t, cmd.Run(ctx, cmd.ServerConfig{
			Logger: log.Logger,

			KongCPCert: cpCert,

			DPAuthMode:    cmd.DPAuthPKIMTLS,
			DPAuthCACerts: dpCACert,
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
	testing.Verbose()
	require.Nil(t, util.WaitForKong(t))

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
	}, util.TestBackoff))
}
