package util

import (
	"context"
	"fmt"
	"net/http"
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
	defaultTimeout    = 5 * time.Second
)

func init() {
	TestBackoff = backoff.NewConstantBackOff(1 * time.Second)
	TestBackoff = backoff.WithMaxRetries(TestBackoff, maxRetriesInTests)
}

func WaitFor(t *testing.T, port int, method, path, component string,
	wantHTTPCode int) error {
	return backoff.RetryNotify(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
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

var DefaultKongHTTPPort = 8000

func WaitForKong(t *testing.T) error {
	return WaitForKongPort(t, DefaultKongHTTPPort)
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

var DefaultAdminPort = 3000

func WaitForAdminAPI(t *testing.T) error {
	return WaitFor(t,
		DefaultAdminPort,
		http.MethodGet,
		"/v1/meta/version",
		fmt.Sprintf("admin-%d", DefaultAdminPort),
		http.StatusOK,
	)
}
