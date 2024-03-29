package resource

import (
	"context"
	"fmt"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/model/json/generator"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/kong/koko/internal/model/json/validation/typedefs"
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
	defaultID(&k.Key.Kid)
	return nil
}

func (k Key) Resource() model.Resource {
	return k.Key
}

func (k Key) SetResource(r model.Resource) error {
	return model.SetResource(k, r)
}

func (k Key) Indexes() []model.Index {
	if k.Key == nil {
		return nil
	}

	res := []model.Index{
		{
			Name:      "kid",
			Type:      model.IndexUnique,
			Value:     k.Key.Kid,
			FieldName: "kid",
		},
	}

	if k.Key.Name != "" {
		res = append(res, model.Index{
			Name:      "name",
			Type:      model.IndexUnique,
			Value:     k.Key.Name,
			FieldName: "name",
		})
	}

	if k.Key.Set != nil {
		res = append(res, model.Index{
			Name:        "set_id",
			Type:        model.IndexForeign,
			ForeignType: TypeKeySet,
			FieldName:   "set.id",
			Value:       k.Key.Set.Id,
		})
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
		Type: "object",
		Properties: map[string]*generator.Schema{
			"id":         typedefs.ID,
			"created_at": typedefs.UnixEpoch,
			"updated_at": typedefs.UnixEpoch,
			"name":       typedefs.Name,
			"set":        typedefs.ReferenceObject,
			"kid":        {Type: "string"},
			"jwk":        typedefs.JWKKey,
			"pem":        typedefs.PEMKey,
			"tags":       typedefs.Tags,
		},
		AdditionalProperties: &falsy,
		Required:             []string{"id", "kid"},
		AllOf: []*generator.Schema{
			{
				Title:       "one key format",
				Description: "Keys must be defined either in JWK or PEM format",
				OneOf: []*generator.Schema{
					{
						Required: []string{"jwk"},
						Properties: map[string]*generator.Schema{
							"pem": {Not: &generator.Schema{Description: "there's a JWK, don't set PEM"}},
						},
					},
					{
						Required: []string{"pem"},
						Properties: map[string]*generator.Schema{
							"jwk": {Not: &generator.Schema{Description: "there's a PEM, don't set JWK"}},
						},
					},
				},
			},
		},
	}
	err = generator.DefaultRegistry.Register(string(TypeKey), keySchema)
	if err != nil {
		panic(err)
	}
}
