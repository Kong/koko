package resource

import (
	"context"
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

const (
	certChain = `
-----BEGIN CERTIFICATE-----
MIIC+DCCAeACAQIwDQYJKoZIhvcNAQEFBQAwQjELMAkGA1UEBhMCVVMxCzAJBgNV
BAgMAkNBMRUwEwYDVQQKDAxZb2xvNDIsIEluYy4xDzANBgNVBAMMBllvbG80MjAe
Fw0yMjA2MjMwODMzNDNaFw0zMjA2MjAwODMzNDNaMEIxCzAJBgNVBAYTAlVTMQsw
CQYDVQQIDAJDQTEVMBMGA1UECgwMWW9sbzQyLCBJbmMuMQ8wDQYDVQQDDAZZb2xv
NDIwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDRe6zuZYrcNisP429l
aTdUxSKXigd8Hdkje2jDCvlaVg72DeBj5VarYLRkuL/aeBNxfDbtBBAPh2oOw0dO
6cSLET9opHZxJefwzaSVa4vU6pSKQIT7MT5dTvH4FVDwVxUhD/LV6WB1LZMnNJcF
hokK4+lvyVFB+UIED2uRMB/H2Ilf4L+5hwM7PSxZebNye/34Qd4R3BhTrPnBosk9
WbXO+/jYoCaFOzzLCxPnGUlPxQlfyPU1lTQUQP9LL/t7hxLNn+SuKIarb5XAOb8a
i8Wgw7eORcTx2wlSr5ZsP/Q3ldlxgfSVl0F78Ra8b2Lne7jM9gRW5cG31xHRF7FR
3t+xAgMBAAEwDQYJKoZIhvcNAQEFBQADggEBAFWZkvzJq+Ha9fLntB/2hU/UxWBf
OofxKSrgVR7snAnLbVpxj3bJESec/+3ZEFvYpi8aJm0KSmxh9QP6sLZ7P+xNJs2q
sPeAihr3dGDAtcv7CmgGeaiSjxHILUX54VUKr4O/ff0Vi01m243rOngJIfRnI7Kw
vX6bdM/Kws5v8rVSA6uWzSDAXMmS9Klhd5fEtdENfk1w7maoy+z51PW3I10EkOOl
aebFAjlBquLeDkO3Ym6Y6GS8g5AG6CGXOwipncS2Q80Z2lFbdWl/3vbis08+YJtx
+nujRQ37xtrwgaXi+fgmSyIC+Rk+ItJUa8JObEOBfZoAkgoknyT+ftWicrA=
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIDDzCCAfegAwIBAgIBATANBgkqhkiG9w0BAQUFADBCMQswCQYDVQQGEwJVUzEL
MAkGA1UECAwCQ0ExFTATBgNVBAoMDFlvbG80MiwgSW5jLjEPMA0GA1UEAwwGWW9s
bzQyMB4XDTIyMDYyMzA4MzMyMVoXDTMyMDYyMDA4MzMyMVowQjELMAkGA1UEBhMC
VVMxCzAJBgNVBAgMAkNBMRUwEwYDVQQKDAxZb2xvNDIsIEluYy4xDzANBgNVBAMM
BllvbG80MjCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAOhpaXd1lFdf
ja3mqL9r8Ktz3hZYbniPKB65PMlEO3BLJj+HdaPNgZCErEdOrshSIEQBnkK3Im1t
SQmIlEJyIDpqS+k/ODkNGCNYmgCxNnK+jNt8LsLZsq0DkTMM1slqokrhdEwQ38Za
6JlwHmdJLPWerl2RtvNwXVRbCka38PmM7LbqmqR/238otcQSNuYEnBSAik+vy9XR
F5G+l5709tPexHSI09jUM54tVtsdFVHhlSwp/qlXhkOuY15NWTAx7POCXd/hiC9f
Ygg8bNuHzBXk1bvs6ANLpGY10qFk3JMz2N1kuxgRJ6kW6b54CTAvjY3LpWH2I+On
ms+m0X6zTEcCAwEAAaMQMA4wDAYDVR0TBAUwAwEB/zANBgkqhkiG9w0BAQUFAAOC
AQEAFMzEqZzqtZJKqpeCZvxXMRWbj7eV8UfdHSXUBOxvz7tONoSzXM1P/Hfkkciu
biYMrtLhxlivsZPY+M/6wAxD77QfWAWWG/ZdwQCtCRl3We6OSs6+b0M/7ity35A+
lPqHL18SEAK8yHXH6+iGfTOg4+W2hu28PywSmYJgWf8BIB2i1myORtnTFgH32R2P
WJ2EUfUqb9+ZKjtNrDrqPtX068AC6hTcFkc8t02EPvJ2TuXyXjSPKA+1DgnXPN9+
BqjK5V2VnCSO8KV9h0DvmWqijbqQyFLuJ5no342jaRyeZThpl0vtdH8Wa3qHbljb
2VpYsrrm/y9fk89fmVi0jd0oGQ==
-----END CERTIFICATE-----
`
	key = `
-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQDRe6zuZYrcNisP
429laTdUxSKXigd8Hdkje2jDCvlaVg72DeBj5VarYLRkuL/aeBNxfDbtBBAPh2oO
w0dO6cSLET9opHZxJefwzaSVa4vU6pSKQIT7MT5dTvH4FVDwVxUhD/LV6WB1LZMn
NJcFhokK4+lvyVFB+UIED2uRMB/H2Ilf4L+5hwM7PSxZebNye/34Qd4R3BhTrPnB
osk9WbXO+/jYoCaFOzzLCxPnGUlPxQlfyPU1lTQUQP9LL/t7hxLNn+SuKIarb5XA
Ob8ai8Wgw7eORcTx2wlSr5ZsP/Q3ldlxgfSVl0F78Ra8b2Lne7jM9gRW5cG31xHR
F7FR3t+xAgMBAAECggEBAKfp7rAZDLl/Yf0WXVB4ijWU3ymBJobCli7u2QaeYUmb
+doZPWhViKdOmMqznHVOEqfA3XYW75jC/qxes2X50+V1KdKDIb2ImOZYsDhlQGym
q/I1zWJcEpVQlnw4+evsoa8izY/RxdOneHDQos13DZqBHbjRMiUj21rN0XdLj+3r
nNF55or8VAFF4oeNy01dBuSJ2L+eO/kcEyge5ywhmn1Dwp6DYGORGdrnJzaWu2Pa
DzYwc6SJ7svGBsjL5t+4BhoeljpWm0STmZVkL5TUscGoHfNn2jzYyJYQrASbMLGz
QD72XQtrJI7G99B2lAXnf5fecB4wHcXTHvVE7YqUaQECgYEA5/zdoJ64GmkB58S9
fT84g9aTyVDk316LrchnC0wENkG7Fitp6RbhmAA6q01dPppilnknhinTMlBQHpZj
GlrWpkFXhfoaj5jKqbr/ZPT8HUHM6LzHhRWHvN2+9g+4t34iejMdxnLSmMjuvpoD
+MmaBD+wNhVn047Oz1H9SqfxPgkCgYEA5yp9NI/jUsvUIcx2uaUUtmBdkbcqbqnj
JvSkeww/9yBp6yUaVN3clwe+cgHRKZeXpkNwFp8TzWE0mmG1ZGAG4rQQdnDvvktG
zL+/JwMOsx6b5R7aA5DVkAZUykjOKluXvHyjjTGj7jiToa8RjMrOKoadnwWC3Llk
ZhauKbFEfmkCgYEAnzPSKHsj3sP3UcWbQIuVTiyAiSRhnMS2WJFx3bfSICXlrSYn
7ZUNRhHKMWrLNb4fMCJ+tDyZuiqRgRw1cI2sRrYKyV/EwIzbb7VrtS3GopFYfNOo
nLUUzNDkTtqlKg9+u5u+sER2L/GcneL2HNLFRmsqk0MHWJDlbjNW/tfX33kCgYAx
I2QIB0oQMInAQYE/RysW9XcOYXwgl/ZUMo7AJUN3malKNdHaFmsso5XFEEPQ7otq
6UzrUhdYggA3jOuNEaiFCjexpaIgtkmvflb4yPqX8rq6wosfVOtAuUfO1BkXAe9I
PspZWiL5oYcoSFmXrwiSG5ln0zkVCEeiN9H/xNHFeQKBgBfCRd5Hso0iAgiwUS/T
OSSGmeqEIC8Krk/G0V9iw4i9OQoBTFeFrja/3JGqUzvoXHJW012MiJ5ErY1bK8du
Z2uSW/FjsT7HM69XuC0ibPNJ+5Cw7iQJ6QIXTjID3dhv4NywmENhJW71nSyg/RPT
xV73bGApHGnNU8lCGx/9s1dL
-----END PRIVATE KEY-----
`
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
	require.Nil(t, cert.ProcessDefaults(context.Background()))
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
				_ = cert.ProcessDefaults(context.Background())
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
				_ = cert.ProcessDefaults(context.Background())
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
				_ = cert.ProcessDefaults(context.Background())
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
			name: "only alternate cert and key defined throws an error",
			Certificate: func() Certificate {
				cert := NewCertificate()
				_ = cert.ProcessDefaults(context.Background())
				cert.Certificate.CertAlt = string(certAltPEM)
				cert.Certificate.KeyAlt = string(keyAltPEM)
				return cert
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"missing properties: 'cert', 'key'",
					},
				},
			},
		},
		{
			name: "certificate with an alt key and no alt cert throws an error",
			Certificate: func() Certificate {
				cert := NewCertificate()
				_ = cert.ProcessDefaults(context.Background())
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
				_ = cert.ProcessDefaults(context.Background())
				cert.Certificate.Cert = string(certPEM)
				cert.Certificate.Key = string(keyPEM)
				cert.Certificate.CertAlt = "a"
				cert.Certificate.KeyAlt = "b"
				return cert
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "cert_alt",
					Messages: []string{
						"'a' is not valid 'pem-encoded-cert'",
						"referenceable field must contain a valid 'Reference'",
					},
				},
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "key_alt",
					Messages: []string{
						"'b' is not valid 'pem-encoded-private-key'",
						"referenceable field must contain a valid 'Reference'",
					},
				},
			},
		},
		{
			name: "certificate with invalid alt cert and alt key throws an error",
			Certificate: func() Certificate {
				cert := NewCertificate()
				_ = cert.ProcessDefaults(context.Background())
				cert.Certificate.Cert = string(certPEM)
				cert.Certificate.Key = string(keyPEM)
				cert.Certificate.CertAlt = "a"
				cert.Certificate.KeyAlt = "b"
				return cert
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "cert_alt",
					Messages: []string{
						"'a' is not valid 'pem-encoded-cert'",
						"referenceable field must contain a valid 'Reference'",
					},
				},
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "key_alt",
					Messages: []string{
						"'b' is not valid 'pem-encoded-private-key'",
						"referenceable field must contain a valid 'Reference'",
					},
				},
			},
		},
		{
			name: "certificate and key match alternate cert and key encryption algo throws an error",
			Certificate: func() Certificate {
				cert := NewCertificate()
				_ = cert.ProcessDefaults(context.Background())
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
				_ = cert.ProcessDefaults(context.Background())
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
				_ = cert.ProcessDefaults(context.Background())
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
				_ = cert.ProcessDefaults(context.Background())
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
				_ = cert.ProcessDefaults(context.Background())
				cert.Certificate.Cert = string(certPEM)
				cert.Certificate.Key = string(keyPEM)
				cert.Certificate.CertAlt = string(certAltPEM)
				cert.Certificate.KeyAlt = string(keyAltPEM)
				return cert
			},
			wantErr: false,
		},
		{
			name: "valid cert chain and valid key passes",
			Certificate: func() Certificate {
				cert := NewCertificate()
				_ = cert.ProcessDefaults(context.Background())
				cert.Certificate.Cert = certChain
				cert.Certificate.Key = key
				return cert
			},
			wantErr: false,
		},
		{
			name: "valid certificate and key reference passes",
			Certificate: func() Certificate {
				cert := NewCertificate()
				_ = cert.ProcessDefaults(context.Background())
				cert.Certificate.Cert = "{vault://env/test-cert}"
				cert.Certificate.Key = "{vault://env/test-key}"
				return cert
			},
			wantErr: false,
		},
		{
			name: "valid certificate and key reference with valid alternate cert and key reference passes",
			Certificate: func() Certificate {
				cert := NewCertificate()
				_ = cert.ProcessDefaults(context.Background())
				cert.Certificate.Cert = "{vault://env/test-cert}"
				cert.Certificate.Key = "{vault://env/test-key}"
				cert.Certificate.CertAlt = "{vault://env/test-cert-alt}"
				cert.Certificate.KeyAlt = "{vault://env/test-key-alt}"
				return cert
			},
			wantErr: false,
		},
		{
			name: "valid cert reference combinations 1 passes",
			Certificate: func() Certificate {
				cert := NewCertificate()
				_ = cert.ProcessDefaults(context.Background())
				cert.Certificate.Cert = string(certPEM)
				cert.Certificate.Key = "{vault://env/test-key}"
				cert.Certificate.CertAlt = string(certAltPEM)
				cert.Certificate.KeyAlt = "{vault://env/test-key-alt}"
				return cert
			},
			wantErr: false,
		},
		{
			name: "valid cert reference combinations 2 passes",
			Certificate: func() Certificate {
				cert := NewCertificate()
				_ = cert.ProcessDefaults(context.Background())
				cert.Certificate.Cert = "{vault://env/test-cert}"
				cert.Certificate.Key = string(keyPEM)
				cert.Certificate.CertAlt = "{vault://env/test-cert-alt}"
				cert.Certificate.KeyAlt = string(keyAltPEM)
				return cert
			},
			wantErr: false,
		},
		{
			name: "invalid cert reference throws an error",
			Certificate: func() Certificate {
				cert := NewCertificate()
				_ = cert.ProcessDefaults(context.Background())
				cert.Certificate.Cert = "vault://env/test-cert}"
				cert.Certificate.Key = "{vault://env/test-key"
				cert.Certificate.CertAlt = "{vaults://env/test-cert-alt}"
				cert.Certificate.KeyAlt = "{vault:/env/test-key-alt}"
				return cert
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "cert",
					Messages: []string{
						"'vault://env/test-cert}' is not valid 'pem-encoded-cert'",
						"referenceable field must contain a valid 'Reference'",
					},
				},
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "cert_alt",
					Messages: []string{
						"'{vaults://env/test-cert-alt}' is not valid 'pem-encoded-cert'",
						"referenceable field must contain a valid 'Reference'",
					},
				},
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "key",
					Messages: []string{
						"'{vault://env/test-key' is not valid 'pem-encoded-private-key'",
						"referenceable field must contain a valid 'Reference'",
					},
				},
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "key_alt",
					Messages: []string{
						"'{vault:/env/test-key-alt}' is not valid 'pem-encoded-private-key'",
						"referenceable field must contain a valid 'Reference'",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.Certificate().Validate(context.Background())
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
