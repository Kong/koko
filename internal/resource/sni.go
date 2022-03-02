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
	TypeSNI model.Type = "sni"
)

type SNI struct {
	SNI *v1.SNI
}

func NewSNI() SNI {
	return SNI{
		SNI: &v1.SNI{},
	}
}

func (r SNI) ID() string {
	if r.SNI == nil {
		return ""
	}
	return r.SNI.Id
}

func (r SNI) Type() model.Type {
	return TypeSNI
}

func (r SNI) Resource() model.Resource {
	return r.SNI
}

func (r SNI) Validate() error {
	err := validation.Validate(string(TypeSNI), r.SNI)
	if err != nil {
		return err
	}
	return nil
}

func (r SNI) ProcessDefaults() error {
	if r.SNI == nil {
		return fmt.Errorf("invalid nil resource")
	}
	defaultID(&r.SNI.Id)
	return nil
}

func (r SNI) Indexes() []model.Index {
	res := []model.Index{
		{
			Name:      "unique-name",
			Type:      model.IndexUnique,
			Value:     r.SNI.Name,
			FieldName: "name",
		},
	}
	if r.SNI.Certificate != nil {
		res = append(res, model.Index{
			Name:        "certificate_id",
			Type:        model.IndexForeign,
			ForeignType: TypeCertificate,
			FieldName:   "certificate.id",
			Value:       r.SNI.Certificate.Id,
		})
	}
	return res
}

func init() {
	err := model.RegisterType(TypeSNI, func() model.Object {
		return NewSNI()
	})
	if err != nil {
		panic(err)
	}

	sniSchema := &generator.Schema{
		Type: "object",
		Properties: map[string]*generator.Schema{
			"id":          typedefs.ID,
			"name":        typedefs.WilcardHost,
			"certificate": typedefs.ReferenceObject,
			"tags":        typedefs.Tags,
			"created_at":  typedefs.UnixEpoch,
			"updated_at":  typedefs.UnixEpoch,
		},
		Required: []string{
			"id",
			"name",
			"certificate",
		},
	}
	err = generator.Register(string(TypeSNI), sniSchema)
	if err != nil {
		panic(err)
	}
}
