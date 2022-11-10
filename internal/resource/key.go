package resource

import (
	"context"
	"fmt"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/model/json/generator"
	"github.com/kong/koko/internal/model/json/validation"
)

const (
	TypeKey = model.Type("key")
)

func NewKey() Key {
	return Key{
		Key: &v1.Key{},
	}
}

type Key struct {
	Key *v1.Key
}

func (k Key) ID() string {
	if k.Key == nil {
		return ""
	}
	return k.Key.Id
}

func (k Key) Type() model.Type {
	return TypeKey
}

func (k Key) Validate(ctx context.Context) error {
	return validation.Validate(string(TypeKey), k.Key)
}

func (k Key) ProcessDefaults(ctx context.Context) error {
	if k.Key == nil {
		return fmt.Errorf("invalid nil resource")
	}
	defaultID(&k.Key.Id)
	return nil
}

func (k Key) Resource() model.Resource {
	return k.Key
}

func (k Key) SetResource(r model.Resource) error {
	return model.SetResource(k, r)
}

func (k Key) Indexes() []model.Index {
	res := []model.Index{
		{
			Name:      "kid",
			Type:      model.IndexUnique,
			Value:     k.Key.Kid,
			FieldName: "kid",
		},
	}
	return res
}

func init() {
	err := model.RegisterType(TypeKey, &v1.Key{}, func() model.Object {
		return NewKey()
	})
	if err != nil {
		panic(err)
	}

	keySchema := &generator.Schema{
		Type:                 "object",
		Properties:           map[string]*generator.Schema{},
		AdditionalProperties: &falsy,
		Required:             []string{"id"},
	}
	err = generator.DefaultRegistry.Register(string(TypeKey), keySchema)
	if err != nil {
		panic(err)
	}
}
