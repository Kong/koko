package ws

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type ErrAuth struct {
	HTTPStatus int
	Message    string
}

func (e ErrAuth) Error() string {
	return fmt.Sprintf("%s (code %d)", e.Message, e.HTTPStatus)
}

type AuthFn func(http *http.Request) error

type Authenticator interface {
	// Authenticate takes a request, authenticates it and returns a manager that
	// will handle the request.
	// If err is of type ErrAuth, then the HTTP response is returned with
	// ErrAuth.HTTPStatus code and following JSON body:
	// { "message" : "ErrAuth.Message" }
	// If err is of any other type, the error is logged and a 500 code is
	// returned to the client.
	Authenticate(r *http.Request) (*Manager, error)
}

type DefaultAuthenticator struct {
	once    sync.Once
	Manager *Manager
	// Context is passed on the Manager.Run.
	// Use this context to shut down the manager.
	Context context.Context

	AuthFn AuthFn
}

func (d *DefaultAuthenticator) Authenticate(r *http.Request) (*Manager, error) {
	if err := d.AuthFn(r); err != nil {
		return nil, err
	}
	d.once.Do(func() {
		go d.Manager.Run(d.Context)
	})
	return d.Manager, nil
}

func AuthFnSharedTLS(cert tls.Certificate) (AuthFn, error) {
	var sharedCert []byte
	switch len(cert.Certificate) {
	case 0:
		return nil, fmt.Errorf("no certificate provided")
	case 1:
		sharedCert = cert.Certificate[0]
	default:
		// more than 1
		return nil, fmt.Errorf("only one shared certificate must be provided")
	}
	return func(r *http.Request) error {
		if r.TLS == nil {
			return ErrAuth{
				HTTPStatus: http.StatusBadRequest,
				Message:    "invalid non-TLS request",
			}
		}
		if len(r.TLS.PeerCertificates) == 0 {
			return ErrAuth{
				HTTPStatus: http.StatusUnauthorized,
				Message:    "no client certificate provided in request",
			}
		}
		peerCert := r.TLS.PeerCertificates[0]
		if !bytes.Equal(peerCert.Raw, sharedCert) {
			return ErrAuth{
				HTTPStatus: http.StatusUnauthorized,
				Message:    "client certificate authentication failed",
			}
		}
		// all good
		return nil
	}, nil
}

func AuthFnPKITLS(rootCAs []*x509.Certificate) (AuthFn, error) {
	if len(rootCAs) == 0 {
		return nil, fmt.Errorf("no root CAs provided")
	}
	caCertPool := x509.NewCertPool()
	for _, ca := range rootCAs {
		if !ca.IsCA {
			return nil, fmt.Errorf("certificate (serial number: %s) "+
				"is not a CA certificate", ca.SerialNumber)
		}
		caCertPool.AddCert(ca)
	}
	return func(r *http.Request) error {
		if r.TLS == nil {
			return ErrAuth{
				HTTPStatus: http.StatusBadRequest,
				Message:    "invalid non-TLS request",
			}
		}
		if len(r.TLS.PeerCertificates) == 0 {
			return ErrAuth{
				HTTPStatus: http.StatusUnauthorized,
				Message:    "no client certificate provided in request",
			}
		}
		peerCert := r.TLS.PeerCertificates[0]

		var intermediates *x509.CertPool
		if len(r.TLS.PeerCertificates) > 1 {
			intermediates = x509.NewCertPool()
			for _, cert := range r.TLS.PeerCertificates[1:] {
				intermediates.AddCert(cert)
			}
		}

		opts := x509.VerifyOptions{
			Intermediates: intermediates,
			Roots:         caCertPool,
			CurrentTime:   time.Now(),
			// KeyUsages:                 nil,
		}
		if _, err := peerCert.Verify(opts); err != nil {
			return ErrAuth{
				HTTPStatus: http.StatusUnauthorized,
				Message:    "client certificate authentication failed",
			}
		}
		return nil
	}, nil
}
