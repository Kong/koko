package ws

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadPassthroughCertificate(t *testing.T) {
	t.Run("empty certificate returns nil for fallthrough on auth methods",
		func(t *testing.T) {
			cert, err := readPassthroughCertificate(&http.Request{
				Header: http.Header{},
			})
			require.Nil(t, cert)
			require.Equal(t, "passthrough certificate not found in the http header 'x-client-cert'", err.Error())
		})
	t.Run("invalid certificate returns an error",
		func(t *testing.T) {
			header := http.Header{}
			header.Set(clientCertHeaderKey, "invalid")
			cert, err := readPassthroughCertificate(&http.Request{
				Header: header,
			})
			require.Nil(t, cert)
			require.NotNil(t, err)
			require.EqualError(t, err, "failed to parse PEM certificate from 'x-client-cert' header")
		})
	t.Run("invalid urlencoding returns an error",
		func(t *testing.T) {
			header := http.Header{}
			header.Set(clientCertHeaderKey, "invalid%%cert")
			cert, err := readPassthroughCertificate(&http.Request{
				Header: header,
			})
			require.Nil(t, cert)
			require.NotNil(t, err)
			require.EqualError(t, err,
				"failed to url decode client certificate from 'x-client-cert' header. invalid URL escape \"%%c\"")
		})
}

func TestAuthFnSharedTLSWithPassThruCert(t *testing.T) {
	cert, err := tls.LoadX509KeyPair("testdata/shared/shared.crt",
		"testdata/shared/shared.key")
	require.Nil(t, err)
	require.NotNil(t, cert)

	t.Run("runtime verification", func(t *testing.T) {
		fn, err := AuthFnSharedTLS(cert)
		require.Nil(t, err)
		require.NotNil(t, fn)

		t.Run("request with non-matching certificate fails",
			func(t *testing.T) {
				encodedCert, err := loadURLEncodedCert("testdata/shared/mismatched_cert.crt")
				require.Nil(t, err)
				require.NotNil(t, encodedCert)

				header := http.Header{}
				header.Set(clientCertHeaderKey, encodedCert)
				err = fn(&http.Request{
					Header: header,
				})
				require.NotNil(t, err)
				errAuth, ok := err.(ErrAuth)
				require.True(t, ok)
				require.Equal(t, http.StatusUnauthorized, errAuth.HTTPStatus)
				require.Equal(t, "client certificate authentication failed", errAuth.Message)
			})
		t.Run("request with valid/matching certificate succeeds",
			func(t *testing.T) {
				encodedCert, err := loadURLEncodedCert("testdata/shared/shared.crt")
				require.Nil(t, err)
				require.NotNil(t, encodedCert)

				header := http.Header{}
				header.Set(clientCertHeaderKey, encodedCert)
				err = fn(&http.Request{
					Header: header,
				})
				require.Nil(t, err)
			})
	})
}

func TestAuthFnPKITLSWithPassThruCert(t *testing.T) {
	rootCACert, err := loadCert("testdata/pki/root_ca.crt")
	require.Nil(t, err)
	require.NotNil(t, rootCACert)

	t.Run("runtime verification", func(t *testing.T) {
		fn, err := AuthFnPKITLS([]*x509.Certificate{rootCACert})
		require.Nil(t, err)
		require.NotNil(t, fn)

		t.Run("request with the wrong passthrough certificate errors",
			func(t *testing.T) {
				encodedCert, err := loadURLEncodedCert("testdata/shared/mismatched_cert.crt")
				require.Nil(t, err)
				require.NotNil(t, encodedCert)

				header := http.Header{}
				header.Set(clientCertHeaderKey, encodedCert)
				err = fn(&http.Request{
					Header: header,
				})
				require.NotNil(t, err)
				errAuth, ok := err.(ErrAuth)
				require.True(t, ok)
				require.Equal(t, http.StatusUnauthorized, errAuth.HTTPStatus)
				require.Equal(t, "client certificate authentication failed", errAuth.Message)
			})
		t.Run("request with the right passthrough certificate succeeds",
			func(t *testing.T) {
				encodedCert, err := loadURLEncodedCert("testdata/pki/valid_client.crt")
				require.Nil(t, err)
				require.NotNil(t, encodedCert)

				header := http.Header{}
				header.Set(clientCertHeaderKey, encodedCert)
				err = fn(&http.Request{
					Header: header,
				})
				require.Nil(t, err)
			})
	})
}

func TestAuthFnSharedTLS(t *testing.T) {
	cert, err := tls.LoadX509KeyPair("testdata/shared/shared.crt",
		"testdata/shared/shared.key")
	require.Nil(t, err)
	require.NotNil(t, cert)

	t.Run("valid certificate returns an auth func", func(t *testing.T) {
		fn, err := AuthFnSharedTLS(cert)
		require.Nil(t, err)
		require.NotNil(t, fn)
	})
	t.Run("empty certificate returns an error", func(t *testing.T) {
		fn, err := AuthFnSharedTLS(tls.Certificate{})
		require.NotNil(t, err)
		require.Nil(t, fn)
	})
	t.Run("runtime verification", func(t *testing.T) {
		fn, err := AuthFnSharedTLS(cert)
		require.Nil(t, err)
		require.NotNil(t, fn)
		t.Run("non-TLS request errors", func(t *testing.T) {
			err := fn(&http.Request{})
			require.NotNil(t, err)
			errAuth, ok := err.(ErrAuth)
			require.True(t, ok)
			require.Equal(t, http.StatusBadRequest, errAuth.HTTPStatus)
			require.Equal(t, "invalid non-TLS request", errAuth.Message)
		})
		t.Run("request without a certificate errors", func(t *testing.T) {
			err = fn(&http.Request{
				TLS: &tls.ConnectionState{
					PeerCertificates: nil,
				},
			})
			require.NotNil(t, err)
			errAuth, ok := err.(ErrAuth)
			require.True(t, ok)
			require.Equal(t, http.StatusUnauthorized, errAuth.HTTPStatus)
			require.Equal(t, "no client certificate provided in request",
				errAuth.Message)
		})
		t.Run("request with the wrong certificate errors",
			func(t *testing.T) {
				cert, err := loadCert("testdata/shared/mismatched_cert.crt")
				require.Nil(t, err)
				require.NotNil(t, cert)

				err = fn(&http.Request{
					TLS: &tls.ConnectionState{
						PeerCertificates: []*x509.Certificate{cert},
					},
				})
				require.NotNil(t, err)
				errAuth, ok := err.(ErrAuth)
				require.True(t, ok)
				require.Equal(t, http.StatusUnauthorized, errAuth.HTTPStatus)
				require.Equal(t, "client certificate authentication failed",
					errAuth.Message)
			})
		t.Run("request with the right certificate succeeds",
			func(t *testing.T) {
				cert, err := loadCert("testdata/shared/shared.crt")
				require.Nil(t, err)
				require.NotNil(t, cert)

				err = fn(&http.Request{
					TLS: &tls.ConnectionState{
						PeerCertificates: []*x509.Certificate{cert},
					},
				})
				require.Nil(t, err)
			})
	})
}

func TestAuthFnPKITLS(t *testing.T) {
	rootCACert, err := loadCert("testdata/pki/root_ca.crt")
	require.Nil(t, err)
	require.NotNil(t, rootCACert)

	t.Run("valid certificate returns an auth func", func(t *testing.T) {
		fn, err := AuthFnPKITLS([]*x509.Certificate{rootCACert})
		require.Nil(t, err)
		require.NotNil(t, fn)
	})
	t.Run("empty certificate returns an error", func(t *testing.T) {
		fn, err := AuthFnPKITLS([]*x509.Certificate{})
		require.NotNil(t, err)
		require.Nil(t, fn)
	})
	t.Run("loading non-CA certificate returns an error", func(t *testing.T) {
		invalidCACert, err := loadCert("testdata/pki/valid_client.crt")
		require.Nil(t, err)
		require.NotNil(t, invalidCACert)
		fn, err := AuthFnPKITLS([]*x509.Certificate{invalidCACert})
		require.NotNil(t, err)
		require.Nil(t, fn)
	})
	t.Run("runtime verification", func(t *testing.T) {
		fn, err := AuthFnPKITLS([]*x509.Certificate{rootCACert})
		require.Nil(t, err)
		require.NotNil(t, fn)
		t.Run("non-TLS request errors", func(t *testing.T) {
			err := fn(&http.Request{})
			require.NotNil(t, err)
			errAuth, ok := err.(ErrAuth)
			require.True(t, ok)
			require.Equal(t, http.StatusBadRequest, errAuth.HTTPStatus)
			require.Equal(t, "invalid non-TLS request", errAuth.Message)
		})
		t.Run("request without a certificate errors", func(t *testing.T) {
			err = fn(&http.Request{
				TLS: &tls.ConnectionState{
					PeerCertificates: nil,
				},
			})
			require.NotNil(t, err)
			errAuth, ok := err.(ErrAuth)
			require.True(t, ok)
			require.Equal(t, http.StatusUnauthorized, errAuth.HTTPStatus)
			require.Equal(t, "no client certificate provided in request",
				errAuth.Message)
		})
		t.Run("request with the wrong certificate errors",
			func(t *testing.T) {
				cert, err := loadCert("testdata/shared/mismatched_cert.crt")
				require.Nil(t, err)
				require.NotNil(t, cert)

				err = fn(&http.Request{
					TLS: &tls.ConnectionState{
						PeerCertificates: []*x509.Certificate{cert},
					},
				})
				require.NotNil(t, err)
				errAuth, ok := err.(ErrAuth)
				require.True(t, ok)
				require.Equal(t, http.StatusUnauthorized, errAuth.HTTPStatus)
				require.Equal(t, "client certificate authentication failed",
					errAuth.Message)
			})
		t.Run("request with the right certificate succeeds",
			func(t *testing.T) {
				cert, err := loadCert("testdata/pki/valid_client.crt")
				require.Nil(t, err)
				require.NotNil(t, cert)

				err = fn(&http.Request{
					TLS: &tls.ConnectionState{
						PeerCertificates: []*x509.Certificate{cert},
					},
				})
				require.Nil(t, err)
			})
	})
}

func loadCertFile(filepath string) ([]byte, error) {
	certPEM, err := ioutil.ReadFile(filepath)
	return certPEM, err
}

func loadURLEncodedCert(filepath string) (string, error) {
	certPEM, err := loadCertFile(filepath)
	if err != nil {
		return "", err
	}
	return url.QueryEscape(string(certPEM)), nil
}

func loadCert(filepath string) (*x509.Certificate, error) {
	certPEM, err := loadCertFile(filepath)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(certPEM)
	return x509.ParseCertificate(block.Bytes)
}
