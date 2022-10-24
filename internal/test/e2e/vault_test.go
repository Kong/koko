//go:build integration

package e2e

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"testing"

	"github.com/cenkalti/backoff/v4"
	"github.com/gavv/httpexpect/v2"
	"github.com/google/uuid"
	kongClient "github.com/kong/go-kong/kong"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/test/kong"
	"github.com/kong/koko/internal/test/run"
	"github.com/kong/koko/internal/test/util"
	"github.com/stretchr/testify/require"
)

var (
	goodCACertPEM = []byte(`-----BEGIN CERTIFICATE-----
MIIE6DCCAtACCQCjgi452nKnUDANBgkqhkiG9w0BAQsFADA2MQswCQYDVQQGEwJV
UzETMBEGA1UECAwKQ2FsaWZvcm5pYTESMBAGA1UEAwwJbG9jYWxob3N0MB4XDTIy
MTAwNDE4NTEyOFoXDTMyMTAwMTE4NTEyOFowNjELMAkGA1UEBhMCVVMxEzARBgNV
BAgMCkNhbGlmb3JuaWExEjAQBgNVBAMMCWxvY2FsaG9zdDCCAiIwDQYJKoZIhvcN
AQEBBQADggIPADCCAgoCggIBALUwleXMo+CxQFvgtmJbWHO4k3YBJwzWqcr2xWn+
vgeoLiKFDQC11F/nnWNKkPZyilLeJda5c9YEVaA9IW6/PZhxQ430RM53EJHoiIPB
B9j7BHGzsvWYHEkjXvGQWeD3mR4TAkoCVTfPAjBji/SL+WvLpgPW5hKRVuedD8ja
cTvkNfk6u2TwPYGgekh9+wS9zcEQs4OwsEiQxmi3Z8if1m1uD09tjqAHb0klPEzM
64tPvlzJrIcH3Z5iF+B9qr91PCQJVYOCjGWlUgPULaqIoTVtY+AnaNnNcol0LM/i
oq7uD0JbeyIFDFMDJVqZwDf/zowzLLlP8Hkok4M8JTefXvB0puQoxmGwOAhwlA0G
KF5etrmhg+dOb+f3nWdgbyjPEytyOeMOOA/4Lb8dHRlf9JnEc4DJqwRVPM9BMeUu
9ZlrSWvURRk8nUZfkjTstLqO2aeubfOvb+tDKUq5Ue2B+AFs0ETLy3bds8TU9syV
5Kl+tIwek2TXzc7afvmeCDoRunAx5nVhmW8dpGhknOmJM0GxOi5s2tiu8/3T9XdH
WcH/GMrocZrkhvzkZccSLYoo1jcDn9LwxHVr/BZ43NymjVa6T3QRTta4Kg5wWpfS
yXi4gIW7VJM12CmNfSDEXqhF03+fjFzoWH+YfBK/9GgUMNjnXWIL9PgFFOBomwEL
tv5zAgMBAAEwDQYJKoZIhvcNAQELBQADggIBAKH8eUGgH/OSS3mHB3Gqv1m2Ea04
Cs03KNEt1weelcHIBWVnPp+jGcSIIfMBnDFAwgxtBKhwptJ9ZKXIzjh7YFxbOT01
NU+KQ6tD+NFDf+SAUC4AWV9Cam63JIaCVNDoo5UjVMlssnng7NefM1q2+ucoP+gs
+bvUCTJcp3FZsq8aUI9Rka575HqRhl/8kyhcwICCgT5UHQJvCQYrInJ0Faem6dr0
tHw+PZ1bo6qB7uxBjK9kyu7dK/vEKliUGM4/MXMDKIc5qXUs47wPLbjxvKsuDglK
KftgUWNYRxx9Bf9ylbjd+ayo3+1Lb9cbvdZnh0UHN6677NvXlWNheCmeysLGQHtm
5H6iIhZ75r6QuC7m6hBSJYtLU3fsQECrmaS/+xBGoSSZjacciO7b7qjQdWOfQREn
7vc5eu0N+CJkp8t3SsyQP6v2Su3ILeTt2EWrmmE4K7SYlJe1HrUVj0AWUwzLa6+Z
+Dx16p3M0RBdFMGNNhLqvG3WRfE5c5md34Aq/C5ePjN7pQGmJhI6weowuX9wCrnh
nJJJRfqyJvqgnVBZ6IawNcOyIofITZHlYVKuaDB1odzWCDNEvFftgJvH0MnO7OY9
Pb9hILPoCy+91jQAVh6Z/ghIcZKHV+N6zV3uS3t5vCejhCNK8mUPSOwAeDf3Bq5r
wQPXd0DdsYGmXVIh
-----END CERTIFICATE-----`)

	badCACertPEM = []byte(`-----BEGIN CERTIFICATE-----
MIIDkzCCAnugAwIBAgIUYGc07pbHSjOBPreXh7OcNT2+sD4wDQYJKoZIhvcNAQEL
BQAwWTELMAkGA1UEBhMCVVMxCzAJBgNVBAgMAkNBMRUwEwYDVQQKDAxZb2xvNDIs
IEluYy4xJjAkBgNVBAMMHVlvbG80MiBzZWxmLXNpZ25lZCB0ZXN0aW5nIENBMB4X
DTIyMDMyOTE5NDczM1oXDTMyMDMyNjE5NDczM1owWTELMAkGA1UEBhMCVVMxCzAJ
BgNVBAgMAkNBMRUwEwYDVQQKDAxZb2xvNDIsIEluYy4xJjAkBgNVBAMMHVlvbG80
MiBzZWxmLXNpZ25lZCB0ZXN0aW5nIENBMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A
MIIBCgKCAQEAvnhTgdJALnuLKDA0ZUZRVMqcaaC+qvfJkiEFGYwX2ZJiFtzU65F/
sB2L0ToFqY4tmMVlOmiSZFnRLDZecmQDbbNwc3wtNikmxIOzx4qR4kbRP8DDdyIf
gaNmGCuaXTM5+FYy2iNBn6CeibIjqdErQlAbFLwQs5t3mLsjii2U4cyvfRtO+0RV
HdJ6Np5LsVziN0c5gVIesIrrbxLcOjtXDzwd/w/j5NXqL/OwD5EBH2vqd3QKKX4t
s83BLl2EsbUse47VAImavrwDhmV6S/p/NuJHqjJ6dIbXLYxNS7g26ijcrXxvNhiu
YoZTykSgdI3BXMNAm1ahP/BtJPZpU7CVdQIDAQABo1MwUTAdBgNVHQ4EFgQUe1WZ
fMfZQ9QIJIttwTmcrnl40ccwHwYDVR0jBBgwFoAUe1WZfMfZQ9QIJIttwTmcrnl4
0ccwDwYDVR0TAQH/BAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEAs4Z8VYbvEs93
haTHdbbaKk0V6xAL/Q8I8GitK9E8cgf8C5rwwn+wU/Gf39dtMUlnW8uxyzRPx53u
CAAcJAWkabT+xwrlrqjO68H3MgIAwgWA5yZC+qW7ECA8xYEK6DzEHIaOpagJdKcL
IaZr/qTJlEQClvwDs4x/BpHRB5XbmJs86GqEB7XWAm+T2L8DluHAXvek+welF4Xo
fQtLlNS/vqTDqPxkSbJhFv1L7/4gdwfAz51wH/iL7AG/ubFEtoGZPK9YCJ40yTWz
8XrUoqUC+2WIZdtmo6dFFJcLfQg4ARJZjaK6lmxJun3iRMZjKJdQKm/NEKz4y9kA
u8S6yNlu2Q==
-----END CERTIFICATE-----`)
)

// This test makes sure secrets management works as expected end-to-end
// by configuring:
// - a Service and a Route to verify the overall flow works end-to-end
// - a Certificate with secret references
// - an SNI for localhost
// - an {env} Vault using 'KONG_MY_SECRET_' as env variables prefix
//
// The needed secrets are injected in the testing DP using the
// KONG_MY_SECRET_CERT and KONG_MY_SECRET_KEY env variables storing
// cert/key signed with `goodCACertPEM`.
//
// After the configuration is synced in the DP, an HTTPS client is
// created using the `goodCACertPEM` used to sign the
// deployed certificate, and then a GET is performed to test the
// proxy functionality, which should return a 200.
func TestSecretsManagementEnvVault(t *testing.T) {
	// ensure that Vaults can be synced to Kong gateway
	cleanup := run.Koko(t)
	defer cleanup()

	c := httpexpect.New(t, "http://localhost:3000")

	// create vault entity
	vault := &v1.Vault{
		Id:     uuid.NewString(),
		Name:   "env",
		Prefix: "my-env-vault",
		Config: &v1.Vault_Config{
			Config: &v1.Vault_Config_Env{
				Env: &v1.Vault_EnvConfig{
					Prefix: "KONG_MY_SECRET_",
				},
			},
		},
	}
	kongVault := kongClient.Vault{
		ID:     &vault.Id,
		Name:   &vault.Name,
		Prefix: &vault.Prefix,
		Config: map[string]interface{}{
			"prefix": vault.Config.GetEnv().Prefix,
		},
	}

	vaultBytes, err := json.ProtoJSONMarshal(vault)
	require.NoError(t, err)
	c.POST("/v1/vaults").WithBytes(vaultBytes).Expect().Status(http.StatusCreated)

	// create service
	service := &v1.Service{
		Id:   uuid.NewString(),
		Name: "foo",
		Host: "mockbin.org",
		Path: "/status/200",
	}
	kongService := kongClient.Service{
		ID:   &service.Id,
		Name: &service.Name,
		Host: &service.Host,
		Path: &service.Path,
	}
	c.POST("/v1/services").WithJSON(service).Expect().Status(http.StatusCreated)

	// create route
	route := &v1.Route{
		Name:  "bar",
		Paths: []string{"/r1"},
		Service: &v1.Service{
			Id: service.Id,
		},
	}
	kongRoute := kongClient.Route{
		Name: &route.Name,
		Paths: func() []*string {
			ps := []*string{}
			for i := range route.Paths {
				ps = append(ps, &route.Paths[i])
			}
			return ps
		}(),
		Service: &kongClient.Service{
			ID: &route.Service.Id,
		},
	}
	c.POST("/v1/routes").WithJSON(route).Expect().Status(http.StatusCreated)

	// create certificate
	certificate := &v1.Certificate{
		Id:   uuid.NewString(),
		Cert: "{vault://my-env-vault/cert}",
		Key:  "{vault://my-env-vault/key}",
	}
	kongCertificate := kongClient.Certificate{
		ID:   &certificate.Id,
		Cert: &certificate.Cert,
		Key:  &certificate.Key,
	}
	c.POST("/v1/certificates").WithJSON(certificate).Expect().Status(http.StatusCreated)

	// create sni
	sni := &v1.SNI{
		Id:   uuid.NewString(),
		Name: "localhost",
		Certificate: &v1.Certificate{
			Id: certificate.Id,
		},
	}
	kongSni := kongClient.SNI{
		ID:   &sni.Id,
		Name: &sni.Name,
		Certificate: &kongClient.Certificate{
			ID: &sni.Certificate.Id,
		},
	}
	c.POST("/v1/snis").WithJSON(sni).Expect().Status(http.StatusCreated)

	dpCleanup := run.KongDP(kong.GetKongConfForSharedWithSecrets())
	defer dpCleanup()

	require.NoError(t, util.WaitForKongAdminAPI(t))
	kongClient.RunWhenKong(t, ">= 3.0.0")

	expectedConfig := util.KongConfig{
		Vaults:       []*kongClient.Vault{&kongVault},
		Services:     []*kongClient.Service{&kongService},
		Routes:       []*kongClient.Route{&kongRoute},
		Certificates: []*kongClient.Certificate{&kongCertificate},
		SNIs:         []*kongClient.SNI{&kongSni},
	}

	util.WaitFunc(t, func() error {
		err := util.EnsureKongConfig(expectedConfig)
		t.Log("configuration mismatch for vault", err)
		return err
	})

	require.NoError(t, backoff.RetryNotify(func() error {
		// build simple http client
		client := &http.Client{}
		// use simple http client with https should result
		// in a failure due to missing certificate.
		_, err := client.Get("https://localhost:8443/r1") //nolint:bodyclose
		require.Error(t, err)

		// use transport with wrong CA cert should result
		// in a failure due to unknown authority.
		badCACertPool := x509.NewCertPool()
		badCACertPool.AppendCertsFromPEM(badCACertPEM)

		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs:    badCACertPool,
					ClientAuth: tls.RequireAndVerifyClientCert,
				},
			},
		}

		_, err = client.Get("https://localhost:8443/r1") //nolint:bodyclose
		require.Error(t, err)

		// use transport with good CA cert should pass
		// if referenced secrets are resolved correctly
		// using the ENV vault.
		goodCACertPool := x509.NewCertPool()
		goodCACertPool.AppendCertsFromPEM(goodCACertPEM)

		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs:    goodCACertPool,
					ClientAuth: tls.RequireAndVerifyClientCert,
				},
			},
		}

		resp, err := client.Get("https://localhost:8443/r1")
		if err != nil {
			return fmt.Errorf("unexpected response: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return nil
		}
		return fmt.Errorf("unexpected code: %v", resp.StatusCode)
	}, util.TestBackoff, nil))
}
