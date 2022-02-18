package crypto

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

func RetrievePublicFromPrivateKey(keyPEM []byte) ([]byte, error) {
	block, _ := pem.Decode(keyPEM)
	if block == nil {
		return nil, errors.New("failed to PEM decode key")
	}
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		return MarshalPublicKey(key)
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return MarshalPublicKey(key)
	}
	if key, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
		return MarshalPublicKey(key)
	}
	return nil, errors.New("unsupported private key type")
}

func MarshalPublicKey(privateKey interface{}) ([]byte, error) {
	switch k := privateKey.(type) {
	case *rsa.PrivateKey:
		return x509.MarshalPKIXPublicKey(&k.PublicKey)
	case *ecdsa.PrivateKey:
		return x509.MarshalPKIXPublicKey(&k.PublicKey)
	case ed25519.PrivateKey:
		pubKey, ok := k.Public().(ed25519.PublicKey)
		if ok {
			return []byte(pubKey), nil
		}
		return nil, errors.New("invalid ed25519 public key")
	default:
		return nil, errors.New("unsupported private key type")
	}
}
