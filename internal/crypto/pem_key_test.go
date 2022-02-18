package crypto

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRSAPrivateToPublicKey(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.Nil(t, err)
	require.NotNil(t, key)

	der := x509.MarshalPKCS1PrivateKey(key)
	require.NotNil(t, der)

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
	pubKey, err := RetrievePublicFromPrivateKey(keyPEM)
	require.Nil(t, err)
	require.NotNil(t, pubKey)
}

func TestECDSAPrivateToPublicKey(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.Nil(t, err)
	require.NotNil(t, key)

	der, err := x509.MarshalECPrivateKey(key)
	require.Nil(t, err)
	require.NotNil(t, der)

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
	pubKey, err := RetrievePublicFromPrivateKey(keyPEM)
	require.Nil(t, err)
	require.NotNil(t, pubKey)
}

func TestED25519PrivateToPublicKey(t *testing.T) {
	pubKey, pvtKey, err := ed25519.GenerateKey(rand.Reader)
	require.Nil(t, err)
	require.NotNil(t, pubKey)
	require.NotNil(t, pvtKey)

	der, err := x509.MarshalPKCS8PrivateKey(pvtKey)
	require.Nil(t, err)
	require.NotNil(t, der)

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
	pub, err := RetrievePublicFromPrivateKey(keyPEM)
	require.Nil(t, err)
	require.NotNil(t, pubKey)
	require.Equal(t, []byte(pubKey), pub)
}
