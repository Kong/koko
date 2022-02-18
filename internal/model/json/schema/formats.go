package schema

import (
	"crypto/x509"
	"encoding/pem"

	"github.com/kong/koko/internal/crypto"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

func isPEMEncodedCertificate(v interface{}) bool {
	switch v := v.(type) {
	case []byte:
		_, err := crypto.ParsePEMCert(v)
		return err == nil
	case string:
		_, err := crypto.ParsePEMCert([]byte(v))
		return err == nil
	default:
		return false
	}
}

func isPEMEncodedPrivateKey(v interface{}) bool {
	var block *pem.Block
	switch v := v.(type) {
	case []byte:
		block, _ = pem.Decode(v)
	case string:
		block, _ = pem.Decode([]byte(v))
	default:
		return false
	}
	if block == nil {
		return false
	}
	_, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err == nil {
		return true
	}
	_, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err == nil {
		return true
	}
	_, err = x509.ParseECPrivateKey(block.Bytes)
	return err == nil
}

func init() {
	jsonschema.Formats["pem-encoded-cert"] = isPEMEncodedCertificate
	jsonschema.Formats["pem-encoded-private-key"] = isPEMEncodedPrivateKey
}
