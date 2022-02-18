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
