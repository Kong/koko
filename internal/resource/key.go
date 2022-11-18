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
			Name:      "id",
			Type:      model.IndexUnique,
			Value:     k.Key.Id,
			FieldName: "id",
		},
		{
			Name:      "kid",
			Type:      model.IndexUnique,
			Value:     k.Key.Jwk.Kid,
			FieldName: "kid",
		},
		{
			Name:      "name",
			Type:      model.IndexUnique,
			Value:     k.Key.Name,
			FieldName: "name",
		},
	}
	return res
}

const (
	TypeKeySet = model.Type("key-set")
)

type KeySet struct {
	KeySet *v1.KeySet
}

func NewKeySet() KeySet {
	return KeySet{
		KeySet: &v1.KeySet{},
	}
}

func (ks KeySet) ID() string {
	if ks.KeySet == nil {
		return ""
	}
	return ks.KeySet.Id
}

func (ks KeySet) Type() model.Type {
	return TypeKeySet
}

func (ks KeySet) Validate(ctx context.Context) error {
	return validation.Validate(string(TypeKeySet), ks.KeySet)
}

func (ks KeySet) ProcessDefaults(ctx context.Context) error {
	if ks.KeySet == nil {
		return fmt.Errorf("invalid nil resource")
	}
	defaultID(&ks.KeySet.Id)
	return nil
}

func (ks KeySet) Resource() model.Resource {
	return ks.KeySet
}

func (ks KeySet) SetResource(r model.Resource) error {
	return model.SetResource(ks, r)
}

func (ks KeySet) Indexes() []model.Index {
	res := []model.Index{
		{
			Name:      "id",
			Type:      model.IndexUnique,
			Value:     ks.KeySet.Id,
			FieldName: "id",
		},
		{
			Name:      "name",
			Type:      model.IndexUnique,
			Value:     ks.KeySet.Name,
			FieldName: "name",
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
		Type: "object",
		Properties: map[string]*generator.Schema{
			"id":         typedefs.ID,
			"created_at": typedefs.UnixEpoch,
			"updated_at": typedefs.UnixEpoch,
			"name":       typedefs.Name,
			"set":        typedefs.ReferenceObject,
			"jwk":        typedefs.JwkKey,
			"pem":        typedefs.PemKey,
			"tags":       typedefs.Tags,
		},
		AdditionalProperties: &falsy,
		Required:             []string{"id"},
	}
	err = generator.DefaultRegistry.Register(string(TypeKey), keySchema)
	if err != nil {
		panic(err)
	}

	err = model.RegisterType(TypeKeySet, &v1.KeySet{}, func() model.Object {
		return NewKeySet()
	})
	if err != nil {
		panic(err)
	}

	keysetSchema := &generator.Schema{
		Type: "object",
		Properties: map[string]*generator.Schema{
			"id":         typedefs.ID,
			"created_at": typedefs.UnixEpoch,
			"updated_at": typedefs.UnixEpoch,
			"name":       typedefs.Name,
		},
		AdditionalProperties: &falsy,
		Required:             []string{"id", "name"},
	}
	err = generator.DefaultRegistry.Register(string(TypeKeySet), keysetSchema)
	if err != nil {
		panic(err)
	}
}
