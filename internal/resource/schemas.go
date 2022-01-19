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
	TypeSchemas = model.Type("schemas")
)

func NewSchemas() Schemas {
	return Schemas{
		Schemas: &v1.Schemas{},
	}
}

type Schemas struct {
	Schemas *v1.Schemas
}

func (r Schemas) ID() string {
	if r.Schemas == nil {
		return ""
	}
	return r.Schemas.Name
}

func (r Schemas) Type() model.Type {
	return TypeSchemas
}

func (r Schemas) Resource() model.Resource {
	return r.Schemas
}

func (r Schemas) Validate() error {
	return validation.Validate(string(TypeSchemas), r.Schemas)
}

func (r Schemas) ProcessDefaults() error {
	if r.Schemas == nil {
		return fmt.Errorf("invalid nil resource")
	}
	return nil
}

func (r Schemas) Indexes() []model.Index {
	return nil
}

func init() {
	err := model.RegisterType(TypeSchemas, func() model.Object {
		return NewSchemas()
	})
	if err != nil {
		panic(err)
	}

	schemasSchema := &generator.Schema{
		Type: "object",
		Properties: map[string]*generator.Schema{
			"name": typedefs.Name,
		},
		AdditionalProperties: false,
		Required: []string{
			"name",
		},
	}
	err = generator.Register(string(TypeSchemas), schemasSchema)
	if err != nil {
		panic(err)
	}
}
