package schema

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var publicKey = `
-----BEGIN PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA2gujMwJavJnU9VA3U+RM
fKJAUvcptlncXSA0jJqTU1PNrK6vJzDbmmaNGC7L4hmue2im4fzujc0lYQM4AYBO
N/OVpbQs7zRBijMARiUZUAhUmQDPBNiazjIxh3ETIXOOYuNInGfeWiu1TaPraOss
Vx0gm9BD5O9af/meuBAq1QhwrV3gNxvfNvFAKMxHLiTFImPIXzct/7FrLyxjb1Uw
g14INW+ioNz7Qh8aVdO9XxfLo0mVD3sAsonf7+q0bxfvwvbAy7IWZCVijZdkiFB1
ycYDsNtZ6xWk00dXARM+q3EnWXNKcCIbMSb4OZjIyAudQ9pp/V2hJF9dWZZmZDOo
K6h3K2tYGQfrzD0ANlbRM+G6uS9yPaM5+aL9m8mH2w4ShwsJksp0QF0GMKNYhOl2
0Fcbp7IlegexF/4ZANWehs3/2TQP72P+fDGvheqZf+2fQ3tBGdoBIeHIW2jxIeh4
eaoMLG5WcAmPGVFK0bMC7eljXHSAmVb8kTO9/+hH5jz4GGgr885BB8suOdlM/g69
ZCjH7Wj6eKnaS6oaN/xnxXhL/LwijWA35vGDzF2lBfWTV/VHI98pBtOA+7r5qixd
L5prbjUUvusTCnyT24G2UlVlyNOI1qX5WstAmlqDuQTJI6xQFkhe+fXln205rDO3
B60cbIexAfjSnMk+rEwlWQ0CAwEAAQ==
-----END PUBLIC KEY-----
`

func TestRSA_PKCS8PrivateKeyFormat(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.Nil(t, err)
	require.NotNil(t, key)

	der, err := x509.MarshalPKCS8PrivateKey(key)
	require.Nil(t, err)
	require.NotNil(t, der)

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
	require.NotNil(t, keyPEM)
	require.True(t, isPEMEncodedPrivateKey(string(keyPEM)))
}

func TestRSA_PKCS1PrivateKeyFormat(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.Nil(t, err)
	require.NotNil(t, key)

	der := x509.MarshalPKCS1PrivateKey(key)
	require.NotNil(t, der)

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
	require.NotNil(t, keyPEM)
	require.True(t, isPEMEncodedPrivateKey(string(keyPEM)))
}

func TestECDSA_PKCS8PrivateKeyFormat(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	require.Nil(t, err)
	require.NotNil(t, key)

	der, err := x509.MarshalPKCS8PrivateKey(key)
	require.Nil(t, err)
	require.NotNil(t, der)

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
	require.NotNil(t, keyPEM)
	require.True(t, isPEMEncodedPrivateKey(string(keyPEM)))
}

func TestED25519_PKCS8PrivateKeyFormat(t *testing.T) {
	pubkey, pvtkey, err := ed25519.GenerateKey(rand.Reader)
	require.Nil(t, err)
	require.NotNil(t, pubkey)
	require.NotNil(t, pvtkey)

	der, err := x509.MarshalPKCS8PrivateKey(pvtkey)
	require.Nil(t, err)
	require.NotNil(t, der)

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
	require.NotNil(t, keyPEM)
	require.True(t, isPEMEncodedPrivateKey(string(keyPEM)))
}

func TestECDSAPrivateKeyFormat(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	require.Nil(t, err)
	require.NotNil(t, key)

	der, err := x509.MarshalECPrivateKey(key)
	require.Nil(t, err)
	require.NotNil(t, der)

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
	require.NotNil(t, keyPEM)
	require.True(t, isPEMEncodedPrivateKey(keyPEM))
}

func TestInvalidPrivateKeyFormat(t *testing.T) {
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte{}})
	require.NotNil(t, keyPEM)
	require.False(t, isPEMEncodedPrivateKey(keyPEM))
	require.False(t, isPEMEncodedPrivateKey(1))
}

func TestPEMEncodedCertificateFormat(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.Nil(t, err)
	require.NotNil(t, key)

	cert := x509.Certificate{
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

	der, err := x509.CreateCertificate(rand.Reader, &cert, &cert, &key.PublicKey, key)
	require.Nil(t, err)
	require.NotNil(t, der)

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	require.NotNil(t, certPEM)
	require.True(t, isPEMEncodedCertificate(certPEM))
	require.True(t, isPEMEncodedCertificate(string(certPEM)))
}

func TestInvalidPEMEncodedCertificateFormat(t *testing.T) {
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte{}})
	require.NotNil(t, certPEM)
	require.False(t, isPEMEncodedCertificate(string(certPEM)))
	require.False(t, isPEMEncodedCertificate(1))
}

func TestRSA_PKIXPublicKeyFormat(t *testing.T) {
	pubPEM, _ := pem.Decode([]byte(publicKey))
	parsedKey, err := x509.ParsePKIXPublicKey(pubPEM.Bytes)
	require.Nil(t, err)
	require.NotNil(t, parsedKey)

	der, err := x509.MarshalPKIXPublicKey(parsedKey)
	require.Nil(t, err)
	require.NotNil(t, der)

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})
	require.NotNil(t, keyPEM)
	require.True(t, isPEMEncodedPublicKey(string(keyPEM)))
}

func TestInvalidPublicKeyFormat(t *testing.T) {
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: []byte{}})
	require.NotNil(t, keyPEM)
	require.False(t, isPEMEncodedPublicKey(keyPEM))
	require.False(t, isPEMEncodedPublicKey(1))
}
