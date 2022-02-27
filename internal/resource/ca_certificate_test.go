package resource

import (
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

var caCertTemplate = x509.Certificate{
	SerialNumber: big.NewInt(1),
	Subject: pkix.Name{
		Organization: []string{"kong_clustering"},
	},
	NotBefore:             time.Now(),
	NotAfter:              time.Now().Add(time.Hour * 24),
	BasicConstraintsValid: true,
	ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
	KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	IsCA:                  true,
}

func TestNewCACertificate(t *testing.T) {
	r := NewCACertificate()
	require.NotNil(t, r)
	require.NotNil(t, r.CACertificate)
}

func TestCACertificate_Type(t *testing.T) {
	require.Equal(t, TypeCACertificate, NewCACertificate().Type())
}

func TestCACertificate_ProcessDefaults(t *testing.T) {
	cert := NewCACertificate()
	require.Nil(t, cert.ProcessDefaults())
	require.NotPanics(t, func() {
		uuid.MustParse(cert.ID())
	})
}

func TestCACertificate_Validate(t *testing.T) {
	tests := []struct {
		name          string
		CACertificate func() CACertificate
		wantErr       bool
		Errs          []*model.ErrorDetail
	}{
		{
			name: "empty certificate throws an error",
			CACertificate: func() CACertificate {
				return NewCACertificate()
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"missing properties: 'id', 'cert'",
					},
				},
			},
		},
		{
			name: "invalid certificate fails",
			CACertificate: func() CACertificate {
				cert := NewCACertificate()
				_ = cert.ProcessDefaults()
				cert.CACertificate.Cert = "a"
				return cert
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:     model.ErrorType_ERROR_TYPE_FIELD,
					Field:    "cert",
					Messages: []string{"'a' is not valid 'pem-encoded-cert'"},
				},
			},
		},
		{
			name: "valid certificate, but invalid CA fails",
			CACertificate: func() CACertificate {
				cert := NewCACertificate()
				_ = cert.ProcessDefaults()
				template := caCertTemplate
				template.BasicConstraintsValid = false
				caCert, err := createCACert(&template)
				require.Nil(t, err)
				cert.CACertificate.Cert = string(caCert)
				return cert
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "cert",
					Messages: []string{"certificate does not appear to be a CA because" +
						"it is missing the \"CA\" basic constraint"},
				},
			},
		},
		{
			name: "valid but expired certificate fails",
			CACertificate: func() CACertificate {
				cert := NewCACertificate()
				_ = cert.ProcessDefaults()
				template := caCertTemplate
				template.NotAfter = time.Now().Add(-24 * time.Hour)
				caCert, err := createCACert(&template)
				require.Nil(t, err)
				cert.CACertificate.Cert = string(caCert)
				return cert
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:     model.ErrorType_ERROR_TYPE_FIELD,
					Field:    "cert",
					Messages: []string{"certificate expired, \"Not After\" time is in the past"},
				},
			},
		},
		{
			name: "valid but multiple certificates fails",
			CACertificate: func() CACertificate {
				cert := NewCACertificate()
				_ = cert.ProcessDefaults()
				caCertOne, err := createCACert(&caCertTemplate)
				require.Nil(t, err)
				caCertTwo, err := createCACert(&caCertTemplate)
				require.Nil(t, err)
				cert.CACertificate.Cert = string(caCertOne) + string(caCertTwo)
				return cert
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:     model.ErrorType_ERROR_TYPE_FIELD,
					Field:    "cert",
					Messages: []string{"please submit only one certificate at a time"},
				},
			},
		},
		{
			name: "valid certificate",
			CACertificate: func() CACertificate {
				cert := NewCACertificate()
				_ = cert.ProcessDefaults()
				caCert, err := createCACert(&caCertTemplate)
				require.Nil(t, err)
				cert.CACertificate.Cert = string(caCert)
				return cert
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.CACertificate().Validate()
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

func createCACert(template *x509.Certificate) (certPEM []byte, err error) {
	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), nil
}
