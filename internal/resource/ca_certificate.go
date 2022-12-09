package resource

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/kong/koko/internal/crypto"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/model/json/extension"
	"github.com/kong/koko/internal/model/json/generator"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/kong/koko/internal/model/json/validation/typedefs"
	"google.golang.org/protobuf/proto"
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
func (r CACertificate) SetResource(pr model.Resource) error { return model.SetResource(r, pr) }

func (r CACertificate) Validate(ctx context.Context) error {
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

	errStr := ""
	if !cert.BasicConstraintsValid {
		errStr = `certificate does not appear to be a CA because ` +
			`it is missing the "CA" basic constraint`
	} else if !cert.IsCA {
		errStr = `certificate does not appear to be a CA because ` +
			`the "CA" basic constraint is set to False`
	}
	if errStr != "" {
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

func (r CACertificate) ProcessDefaults(ctx context.Context) error {
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
	if r.CACertificate.CertDigest == "" {
		return nil
	}

	return []model.Index{
		{
			Name:      "cert_digest",
			Type:      model.IndexUnique,
			Value:     r.CACertificate.CertDigest,
			FieldName: "cert_digest",
		},
	}
}

// MarshalResourceJSON is used here to make sure that CA certificate
// metadata, which cannot be passed via API, are still included
// in the marshalled object.
//
// At the same time, we don't want to store the metadata object in the
// database, so here we are creating a clone of the original object and
// setting the metadata to nil. This way the original object, which
// includes the metadata, will be returned to the caller, while the
// clone object, which doesn't include the metadata, will be stored in
// the database.
func (r CACertificate) MarshalResourceJSON() ([]byte, error) {
	if !isReference(r.CACertificate.Cert) {
		if err := extractCACertInfo(r.CACertificate); err != nil {
			return nil, err
		}
	}
	newCert := proto.Clone(r.CACertificate)
	newCertObj, ok := newCert.(*v1.CACertificate)
	if !ok {
		return nil, errors.New("unexpected type while converting CACertificate clone")
	}
	newCertObj.Metadata = nil
	return json.Marshal(newCertObj)
}

// UnmarshalResourceJSON is used here to make sure that CA certificate
// metadata, which cannot be passed via API, are still included
// in the unmarshalled object.
func (r CACertificate) UnmarshalResourceJSON(cert []byte) error {
	if err := json.Unmarshal(cert, &r.CACertificate); err != nil {
		return err
	}
	if !isReference(r.CACertificate.Cert) {
		if err := extractCACertInfo(r.CACertificate); err != nil {
			return err
		}
	}
	return nil
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
		AdditionalProperties: &falsy,
		XKokoConfig: &extension.Config{
			ResourceAPIPath: "ca-certificates",
		},
	}
	err = generator.DefaultRegistry.Register(string(TypeCACertificate), caCertificateSchema)
	if err != nil {
		panic(err)
	}
}

func extractCACertInfo(cert *v1.CACertificate) error {
	decoded, err := crypto.ParsePEMCert([]byte(cert.Cert))
	if err != nil {
		return fmt.Errorf("decoding certificate: %w", err)
	}
	if cert.Metadata != nil {
		return nil
	}
	cert.Metadata = &v1.CertificateMetadata{
		Issuer:    decoded.Issuer.String(),
		Subject:   decoded.Subject.String(),
		KeyUsages: extractKeyUsages(decoded),
		Expiry:    int32(decoded.NotAfter.Unix()),
		SanNames:  decoded.DNSNames,
	}
	return nil
}
