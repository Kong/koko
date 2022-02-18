package resource

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/google/uuid"
	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/stretchr/testify/require"
)

var certTemplate = x509.Certificate{
	SerialNumber: big.NewInt(1),
	Subject: pkix.Name{
		Organization: []string{"kong_clustering"},
	},
	NotBefore: time.Now(),
	NotAfter:  time.Now().Add(time.Hour * 24),

	KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
	ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	BasicConstraintsValid: true,
}

func TestNewCertificate(t *testing.T) {
	r := NewCertificate()
	require.NotNil(t, r)
	require.NotNil(t, r.Certificate)
}

func TestCertificate_Type(t *testing.T) {
	require.Equal(t, TypeCertificate, NewCertificate().Type())
}

func TestCertificate_ProcessDefaults(t *testing.T) {
	cert := NewCertificate()
	require.Nil(t, cert.ProcessDefaults())
	require.NotPanics(t, func() {
		uuid.MustParse(cert.ID())
	})
}

func TestCertificate_Validate(t *testing.T) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.Nil(t, err)
	require.NotNil(t, rsaKey)

	certPEM, err := createCert(rsaKey.Public(), rsaKey)
	require.Nil(t, err)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(rsaKey)})

	ecdsaKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.Nil(t, err)
	require.NotNil(t, ecdsaKey)

	certAltPEM, err := createCert(ecdsaKey.Public(), ecdsaKey)
	require.Nil(t, err)
	keyDER, err := x509.MarshalECPrivateKey(ecdsaKey)
	require.Nil(t, err)
	require.NotNil(t, keyDER)
	keyAltPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})

	tests := []struct {
		name        string
		Certificate func() Certificate
		wantErr     bool
		Errs        []*model.ErrorDetail
	}{
		{
			name: "empty certificate throws an error",
			Certificate: func() Certificate {
				return NewCertificate()
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"missing properties: 'id', 'cert', 'key'",
					},
				},
			},
		},
		{
			name: "certificate with a cert and no key throws an error",
			Certificate: func() Certificate {
				cert := NewCertificate()
				_ = cert.ProcessDefaults()
				cert.Certificate.Cert = string(certPEM)
				return cert
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"missing properties: 'key'",
					},
				},
			},
		},
		{
			name: "certificate with a key and no cert throws an error",
			Certificate: func() Certificate {
				cert := NewCertificate()
				_ = cert.ProcessDefaults()
				cert.Certificate.Key = string(keyPEM)
				return cert
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"missing properties: 'cert'",
					},
				},
			},
		},
		{
			name: "certificate with an alt cert and no alt key throws an error",
			Certificate: func() Certificate {
				cert := NewCertificate()
				_ = cert.ProcessDefaults()
				cert.Certificate.Cert = string(certPEM)
				cert.Certificate.Key = string(keyPEM)
				cert.Certificate.CertAlt = string(certPEM)
				return cert
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"missing properties: 'key_alt'",
					},
				},
			},
		},
		{
			name: "certificate with an alt key and no alt cert throws an error",
			Certificate: func() Certificate {
				cert := NewCertificate()
				_ = cert.ProcessDefaults()
				cert.Certificate.Cert = string(certPEM)
				cert.Certificate.Key = string(keyPEM)
				cert.Certificate.KeyAlt = string(keyPEM)
				return cert
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"missing properties: 'cert_alt'",
					},
				},
			},
		},
		{
			name: "certificate with invalid alt cert and alt key throws an error",
			Certificate: func() Certificate {
				cert := NewCertificate()
				_ = cert.ProcessDefaults()
				cert.Certificate.Cert = string(certPEM)
				cert.Certificate.Key = string(keyPEM)
				cert.Certificate.CertAlt = "a"
				cert.Certificate.KeyAlt = "b"
				return cert
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:     model.ErrorType_ERROR_TYPE_FIELD,
					Field:    "cert_alt",
					Messages: []string{"'a' is not valid 'pem-encoded-cert'"},
				},
				{
					Type:     model.ErrorType_ERROR_TYPE_FIELD,
					Field:    "key_alt",
					Messages: []string{"'b' is not valid 'pem-encoded-private-key'"},
				},
			},
		},
		{
			name: "certificate with invalid alt cert and alt key throws an error",
			Certificate: func() Certificate {
				cert := NewCertificate()
				_ = cert.ProcessDefaults()
				cert.Certificate.Cert = string(certPEM)
				cert.Certificate.Key = string(keyPEM)
				cert.Certificate.CertAlt = "a"
				cert.Certificate.KeyAlt = "b"
				return cert
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:     model.ErrorType_ERROR_TYPE_FIELD,
					Field:    "cert_alt",
					Messages: []string{"'a' is not valid 'pem-encoded-cert'"},
				},
				{
					Type:     model.ErrorType_ERROR_TYPE_FIELD,
					Field:    "key_alt",
					Messages: []string{"'b' is not valid 'pem-encoded-private-key'"},
				},
			},
		},
		{
			name: "certificate and key match alternate cert and key encryption algo throws an error",
			Certificate: func() Certificate {
				cert := NewCertificate()
				_ = cert.ProcessDefaults()
				cert.Certificate.Cert = string(certPEM)
				cert.Certificate.Key = string(keyPEM)
				cert.Certificate.CertAlt = string(certPEM)
				cert.Certificate.KeyAlt = string(keyPEM)
				return cert
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{"certificate and alternative certificate need to have different type" +
						" (e.g. RSA and ECDSA), the provided certificates were both of the same type 'RSA'"},
				},
			},
		},
		{
			name: "cert with non matching key throws an error",
			Certificate: func() Certificate {
				cert := NewCertificate()
				_ = cert.ProcessDefaults()
				cert.Certificate.Cert = string(certPEM)
				cert.Certificate.Key = string(keyAltPEM)
				return cert
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:     model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{"certificate does not match key"},
				},
			},
		},
		{
			name: "alternate cert with non matching key throws an error",
			Certificate: func() Certificate {
				cert := NewCertificate()
				_ = cert.ProcessDefaults()
				cert.Certificate.Cert = string(certPEM)
				cert.Certificate.Key = string(keyPEM)
				cert.Certificate.CertAlt = string(certAltPEM)
				cert.Certificate.KeyAlt = string(keyPEM)
				return cert
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:     model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{"alternate certificate does not match key"},
				},
			},
		},
		{
			name: "valid certificate and key passes",
			Certificate: func() Certificate {
				cert := NewCertificate()
				_ = cert.ProcessDefaults()
				cert.Certificate.Cert = string(certPEM)
				cert.Certificate.Key = string(keyPEM)
				return cert
			},
			wantErr: false,
		},
		{
			name: "valid certificate and key with valid alternate cert and key passes",
			Certificate: func() Certificate {
				cert := NewCertificate()
				_ = cert.ProcessDefaults()
				cert.Certificate.Cert = string(certPEM)
				cert.Certificate.Key = string(keyPEM)
				cert.Certificate.CertAlt = string(certAltPEM)
				cert.Certificate.KeyAlt = string(keyAltPEM)
				return cert
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.Certificate().Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.Errs != nil {
				verr, ok := err.(validation.Error)
				require.True(t, ok)
				require.ElementsMatch(t, tt.Errs, verr.Errs)
			}
		})
	}
}

func createCert(pubKey, privKey interface{}) (certPEM []byte, err error) {
	der, err := x509.CreateCertificate(rand.Reader, &certTemplate, &certTemplate, pubKey, privKey)
	if err != nil {
		return nil, err
	}

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), nil
}
