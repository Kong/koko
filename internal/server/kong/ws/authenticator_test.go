package ws

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

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

func loadCert(filepath string) (*x509.Certificate, error) {
	certPEM, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(certPEM)
	return x509.ParseCertificate(block.Bytes)
}
