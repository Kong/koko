package resource

import (
	"testing"

	"github.com/google/uuid"
	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/stretchr/testify/require"
)

const (
	certPEM = `-----BEGIN CERTIFICATE-----
MIIBxTCCAS4CCQD0SjBT9iaA9zANBgkqhkiG9w0BAQsFADAnMQswCQYDVQQGEwJV
UzEYMBYGA1UEAwwPa29uZ19jbHVzdGVyaW5nMB4XDTIyMDIwNDIwMDg1NVoXDTMy
MDIwMjIwMDg1NVowJzELMAkGA1UEBhMCVVMxGDAWBgNVBAMMD2tvbmdfY2x1c3Rl
cmluZzCBnzANBgkqhkiG9w0BAQEFAAOBjQAwgYkCgYEAyHXqTwrb6MrDFaZBoLnY
7cTXbiZu6Tjp2eqHAp1NfCxu3RUd3kbDmX27yWPbOCMo/gv2Nl5xQIp/ciOQPaPx
gqU8oYjcXomK3zc57nb7meyPn4H6fGIcVxknD+42LAG2DKEdDjLRJIveTeZqvDDt
OXj1IVQoHme7jLAPF+Wta2ECAwEAATANBgkqhkiG9w0BAQsFAAOBgQBtRKEEVYLs
/X0ZggtDy/WgRIlgXFzt8q4ECqxdL+h3o9/Cl051xdcAGbnz6Ji+0ZK1+iCLTNl8
n9pVqfh1Bate01962jGELKyXkGePn/6HzTYJbk1SqCXemYst7VZmiRjx4biz4kPl
odxViRGIvVwiG+wBvdgZc699Pfh4Hqa/DA==
-----END CERTIFICATE-----`

	keyPEM = `-----BEGIN PRIVATE KEY-----
MIG2AgEAMBAGByqGSM49AgEGBSuBBAAiBIGeMIGbAgEBBDBNvmhmUfKWbjUAvc3/
XoEiDjZuGfCKOXlS1pSUsUgV8DerlatiPc7aD8vOej+jsnmhZANiAASDiQiDvMcr
UFsXqHQ3lyuIUqWkPmrww0rQiFTBpDvyW8Ci723O+piU5n4GjR4JamxTKzlWMQxV
LJDtSk6WZf3IEzQ9REDXszOJWafqjVih+ge/MjzzQ1S/0BorZHZEZ6w=
-----END PRIVATE KEY-----`
)

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
				cert.Certificate.Cert = certPEM
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
			name: "certificate with a key and cert throws an error",
			Certificate: func() Certificate {
				cert := NewCertificate()
				_ = cert.ProcessDefaults()
				cert.Certificate.Key = keyPEM
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
				cert.Certificate.Cert = certPEM
				cert.Certificate.Key = keyPEM
				cert.Certificate.CertAlt = certPEM
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
				cert.Certificate.Cert = certPEM
				cert.Certificate.Key = keyPEM
				cert.Certificate.KeyAlt = keyPEM
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
				cert.Certificate.Cert = certPEM
				cert.Certificate.Key = keyPEM
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
				cert.Certificate.Cert = certPEM
				cert.Certificate.Key = keyPEM
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
			name: "valid certificate passes",
			Certificate: func() Certificate {
				cert := NewCertificate()
				_ = cert.ProcessDefaults()
				cert.Certificate.Cert = certPEM
				cert.Certificate.Key = keyPEM
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
