package resource

import (
	"fmt"

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

func (r Certificate) Validate() error {
	return validation.Validate(string(TypeCertificate), r.Certificate)
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
	err := model.RegisterType(TypeCertificate, func() model.Object {
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
