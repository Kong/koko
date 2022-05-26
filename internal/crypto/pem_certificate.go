package crypto

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

func ParsePEMCert(cert []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(cert)
	if block == nil {
		return nil, fmt.Errorf("invalid PEM-encoded value")
	}
	return x509.ParseCertificate(block.Bytes)
}

func ParsePEMCerts(certs ...[]byte) ([]*x509.Certificate, error) {
	res := make([]*x509.Certificate, 0, len(certs))
	for _, cert := range certs {
		cert, err := ParsePEMCert(cert)
		if err != nil {
			return nil, err
		}
		res = append(res, cert)
	}
	return res, nil
}
