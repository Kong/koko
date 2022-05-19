package resource

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"time"

	"github.com/kong/koko/internal/crypto"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/model/json/generator"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/kong/koko/internal/model/json/validation/typedefs"
)

const (
	TypeCACertificate model.Type = "ca_certificate"

	maxDigestLength = 64
)

var certRegex = regexp.MustCompile("BEGIN CERTIFICATE")

func NewCACertificate() CACertificate {
	return CACertificate{
		CACertificate: &v1.CACertificate{},
	}
}

type CACertificate struct {
	CACertificate *v1.CACertificate
}

func (r CACertificate) ID() string {
	if r.CACertificate == nil {
		return ""
	}
	return r.CACertificate.Id
}

func (r CACertificate) Type() model.Type {
	return TypeCACertificate
}

func (r CACertificate) Resource() model.Resource {
	return r.CACertificate
}

// SetResource implements the Object.SetResource interface.
func (r CACertificate) SetResource(pr model.Resource) error { return SetResource(r, pr) }

func (r CACertificate) Validate() error {
	err := validation.Validate(string(TypeCACertificate), r.CACertificate)
	if err != nil {
		return err
	}
	cert, err := crypto.ParsePEMCert([]byte(r.CACertificate.Cert))
	if err != nil {
		return validation.Error{
			Errs: []*v1.ErrorDetail{
				{
					Type:     v1.ErrorType_ERROR_TYPE_FIELD,
					Field:    "cert",
					Messages: []string{fmt.Sprintf("invalid certificate: %v", err)},
				},
			},
		}
	}

	matches := certRegex.FindAllString(r.CACertificate.Cert, -1)
	if len(matches) > 1 {
		errStr := "only one certificate must be present"
		return validation.Error{
			Errs: []*v1.ErrorDetail{
				{
					Type:     v1.ErrorType_ERROR_TYPE_FIELD,
					Field:    "cert",
					Messages: []string{errStr},
				},
			},
		}
	}

	if time.Now().After(cert.NotAfter) {
		errStr := `certificate expired, "Not After" time is in the past`
		return validation.Error{
			Errs: []*v1.ErrorDetail{
				{
					Type:     v1.ErrorType_ERROR_TYPE_FIELD,
					Field:    "cert",
					Messages: []string{errStr},
				},
			},
		}
	}
	if !cert.BasicConstraintsValid {
		errStr := `certificate does not appear to be a CA because` +
			`it is missing the "CA" basic constraint`
		return validation.Error{
			Errs: []*v1.ErrorDetail{
				{
					Type:     v1.ErrorType_ERROR_TYPE_FIELD,
					Field:    "cert",
					Messages: []string{errStr},
				},
			},
		}
	}

	return nil
}

func (r CACertificate) ProcessDefaults() error {
	if r.CACertificate == nil {
		return fmt.Errorf("invalid nil resource")
	}
	defaultID(&r.CACertificate.Id)

	if r.CACertificate.Cert == "" {
		return nil
	}
	// add cert_digest field
	cert, err := crypto.ParsePEMCert([]byte(r.CACertificate.Cert))
	if err != nil {
		return validation.Error{
			Errs: []*v1.ErrorDetail{
				{
					Type:     v1.ErrorType_ERROR_TYPE_FIELD,
					Field:    "cert",
					Messages: []string{fmt.Sprintf("invalid certificate: %v", err)},
				},
			},
		}
	}
	digest := sha256.Sum256(cert.Raw)
	r.CACertificate.CertDigest = hex.EncodeToString(digest[:])
	return nil
}

func (r CACertificate) Indexes() []model.Index {
	return []model.Index{
		{
			Name:      "cert_digest",
			Type:      model.IndexUnique,
			Value:     r.CACertificate.CertDigest,
			FieldName: "cert_digest",
		},
	}
}

func init() {
	err := model.RegisterType(TypeCACertificate, &v1.CACertificate{}, func() model.Object {
		return NewCACertificate()
	})
	if err != nil {
		panic(err)
	}

	caCertificateSchema := &generator.Schema{
		Properties: map[string]*generator.Schema{
			"id": typedefs.ID,
			"cert": {
				Type:   "string",
				Format: "pem-encoded-cert",
			},
			"cert_digest": {
				Type:      "string",
				MaxLength: maxDigestLength,
			},
			"tags":       typedefs.Tags,
			"created_at": typedefs.UnixEpoch,
			"updated_at": typedefs.UnixEpoch,
		},
		Required: []string{
			"id",
			"cert",
		},
	}
	err = generator.Register(string(TypeCACertificate), caCertificateSchema)
	if err != nil {
		panic(err)
	}
}
