package util

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
)

var (
	path404 = "61d624f6-59fb-45a0-9892-9f6e81264f3e"
	hc      = http.DefaultClient

	// TestBackoff retries every second for 30 seconds and then gives up.
	TestBackoff backoff.BackOff
)

const (
	maxRetriesInTests = 30
	retryTimeout      = 5 * time.Second
	requestTimeout    = 30 * time.Second
)

func init() {
	TestBackoff = backoff.NewConstantBackOff(1 * time.Second)
	TestBackoff = backoff.WithMaxRetries(TestBackoff, maxRetriesInTests)
}

func WaitFunc(t *testing.T, fn func() error) {
	err := backoff.RetryNotify(fn, TestBackoff, func(err error,
		duration time.Duration,
	) {
		if err != nil {
			t.Log("waiting for func to complete")
		}
	})
	if err != nil {
		t.Errorf("failed to complete operation: %v", err)
	}
}

func WaitFor(t *testing.T, port int, method, path, component string,
	wantHTTPCode int,
) error {
	return backoff.RetryNotify(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
		defer cancel()
		req, _ := http.NewRequestWithContext(
			ctx,
			method,
			fmt.Sprintf("http://localhost:%d/%s",
				port,
				strings.TrimPrefix(path, "/"),
			),
			nil,
		)
		res, err := hc.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if res.StatusCode == wantHTTPCode {
			return nil
		}
		return fmt.Errorf("unexpected status code while waiting for '%v': %v",
			component, res.StatusCode)
	}, TestBackoff, func(err error, duration time.Duration) {
		if err != nil {
			t.Log("waiting for " + component)
		}
	})
}

const (
	defaultKongHTTPProxyPort = 8000
	defaultKongHTTPAdminPort = 8001
)

func WaitForKong(t *testing.T) error {
	return WaitForKongPort(t, defaultKongHTTPProxyPort)
}

func WaitForKongAdminAPI(t *testing.T) error {
	err := WaitFor(t,
		defaultKongHTTPAdminPort,
		http.MethodGet,
		"/",
		fmt.Sprintf("kong-dp-admin-%d", defaultKongHTTPAdminPort),
		http.StatusOK,
	)
	if err == nil {
		time.Sleep(retryTimeout)
	}
	return err
}

func WaitForKongPort(t *testing.T, port int) error {
	return WaitFor(t,
		port,
		http.MethodGet,
		path404,
		fmt.Sprintf("kong-dp-%d", port),
		http.StatusNotFound,
	)
}

const defaultAdminPort = 3000

func WaitForAdminAPI(t *testing.T) error {
	return WaitFor(t,
		defaultAdminPort,
		http.MethodGet,
		"/v1/meta/version",
		fmt.Sprintf("admin-%d", defaultAdminPort),
		http.StatusOK,
	)
}

// GenerateCertificate creates a random certificate with the given
// amount of bits, and returns the PEM encoded public & private keys.
func GenerateCertificate(bits int) (publicKey string, privateKey string, err error) {
	caCert := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{Organization: []string{"Test"}},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour), //nolint:gomnd
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	pk, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return "", "", err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, caCert, caCert, &pk.PublicKey, pk)
	if err != nil {
		return "", "", err
	}

	certPEM := &bytes.Buffer{}
	if err := pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	}); err != nil {
		return "", "", err
	}

	pkPEM := &bytes.Buffer{}
	if err := pem.Encode(pkPEM, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(pk),
	}); err != nil {
		return "", "", err
	}

	return certPEM.String(), pkPEM.String(), nil
}

// SkipTestIfEnterpriseTesting skips OSS test when skip is true and KOKO_TEST_ENTERPRISE_TESTING
// environment variable is set to true.
func SkipTestIfEnterpriseTesting(t *testing.T, skip bool) {
	if skip && strings.EqualFold(os.Getenv("KOKO_TEST_ENTERPRISE_TESTING"), "true") {
		t.Skip("Skipping test for enterprise level testing")
	}
}
