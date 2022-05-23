package resource

import (
	"bytes"
	"crypto/x509"
	"fmt"

	"github.com/kong/koko/internal/crypto"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/model/json/generator"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/kong/koko/internal/model/json/validation/typedefs"
)

const (
	TypeCertificate model.Type = "certificate"
)

func NewCertificate() Certificate {
	return Certificate{
		Certificate: &v1.Certificate{},
	}
}

type Certificate struct {
	Certificate *v1.Certificate
}

func (r Certificate) ID() string {
	if r.Certificate == nil {
		return ""
	}
	return r.Certificate.Id
}

func (r Certificate) Type() model.Type {
	return TypeCertificate
}

func (r Certificate) Resource() model.Resource {
	return r.Certificate
}

// SetResource implements the Object.SetResource interface.
func (r Certificate) SetResource(pr model.Resource) error { return model.SetResource(r, pr) }

func (r Certificate) Validate() error {
	err := validation.Validate(string(TypeCertificate), r.Certificate)
	if err != nil {
		return err
	}

	cert, _ := crypto.ParsePEMCert([]byte(r.Certificate.Cert))
	pubKey, err := crypto.RetrievePublicFromPrivateKey([]byte(r.Certificate.Key))
	if err != nil {
		return validation.Error{
			Errs: []*v1.ErrorDetail{
				{
					Type:     v1.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{fmt.Sprintf("failed to get public key from certificate: %v", err)},
				},
			},
		}
	}
	certPubKey, _ := x509.MarshalPKIXPublicKey(cert.PublicKey)
	if !bytes.Equal(certPubKey, pubKey) {
		return validation.Error{
			Errs: []*v1.ErrorDetail{
				{
					Type:     v1.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{"certificate does not match key"},
				},
			},
		}
	}

	if r.Certificate.CertAlt == "" {
		return nil
	}

	altCert, _ := crypto.ParsePEMCert([]byte(r.Certificate.CertAlt))
	if cert.PublicKeyAlgorithm == altCert.PublicKeyAlgorithm {
		return validation.Error{
			Errs: []*v1.ErrorDetail{
				{
					Type: v1.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{fmt.Sprintf("certificate and alternative certificate need to have "+
						"different type (e.g. RSA and ECDSA), the provided "+
						"certificates were both of the same type '%s'", cert.PublicKeyAlgorithm.String())},
				},
			},
		}
	}

	altPubKey, err := crypto.RetrievePublicFromPrivateKey([]byte(r.Certificate.KeyAlt))
	if err != nil {
		return validation.Error{
			Errs: []*v1.ErrorDetail{
				{
					Type:     v1.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{fmt.Sprintf("failed to get public key from alternate certificate: %v", err)},
				},
			},
		}
	}
	altCertPubKey, _ := x509.MarshalPKIXPublicKey(altCert.PublicKey)
	if !bytes.Equal(altCertPubKey, altPubKey) {
		return validation.Error{
			Errs: []*v1.ErrorDetail{
				{
					Type:     v1.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{"alternate certificate does not match key"},
				},
			},
		}
	}
	return nil
}

func (r Certificate) ProcessDefaults() error {
	if r.Certificate == nil {
		return fmt.Errorf("invalid nil resource")
	}
	defaultID(&r.Certificate.Id)
	return nil
}

func (r Certificate) Indexes() []model.Index {
	return nil
}

func init() {
	err := model.RegisterType(TypeCertificate, &v1.Certificate{}, func() model.Object {
		return NewCertificate()
	})
	if err != nil {
		panic(err)
	}

	certificateSchema := &generator.Schema{
		Properties: map[string]*generator.Schema{
			"id": typedefs.ID,
			"cert": {
				Type:   "string",
				Format: "pem-encoded-cert",
			},
			"key": {
				Type:   "string",
				Format: "pem-encoded-private-key",
			},
			"cert_alt": {
				Type:   "string",
				Format: "pem-encoded-cert",
			},
			"key_alt": {
				Type:   "string",
				Format: "pem-encoded-private-key",
			},
			"tags":       typedefs.Tags,
			"created_at": typedefs.UnixEpoch,
			"updated_at": typedefs.UnixEpoch,
		},
		Required: []string{
			"id",
			"cert",
			"key",
		},
		Dependencies: map[string]*generator.Schema{
			"cert_alt": {
				Required: []string{
					"key_alt",
				},
			},
			"key_alt": {
				Required: []string{
					"cert_alt",
				},
			},
		},
	}
	err = generator.Register(string(TypeCertificate), certificateSchema)
	if err != nil {
		panic(err)
	}
}
