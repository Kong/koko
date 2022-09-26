package resource

import (
	"context"
	"fmt"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/model/json/extension"
	"github.com/kong/koko/internal/model/json/generator"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/kong/koko/internal/model/json/validation/typedefs"
)

var TypeHash = model.Type("hash")

func NewHash() Hash {
	return Hash{
		Hash: &v1.ConfigHash{},
	}
}

type Hash struct {
	Hash *v1.ConfigHash
}

func (r Hash) ID() string {
	if r.Hash == nil {
		return ""
	}
	return "config-hash-id"
}

func (r Hash) Type() model.Type {
	return TypeHash
}

func (r Hash) Resource() model.Resource {
	return r.Hash
}

// SetResource implements the Object.SetResource interface.
func (r Hash) SetResource(pr model.Resource) error { return model.SetResource(r, pr) }

func (r Hash) Validate(ctx context.Context) error {
	return validation.Validate(string(TypeHash), r.Hash)
}

func (r Hash) ProcessDefaults(ctx context.Context) error {
	if r.Hash == nil {
		return fmt.Errorf("invalid nil resource")
	}
	return nil
}

func (r Hash) Indexes() []model.Index {
	return nil
}

func init() {
	err := model.RegisterType(TypeHash, &v1.ConfigHash{}, func() model.Object {
		return NewHash()
	})
	if err != nil {
		panic(err)
	}

	nodeSchema := &generator.Schema{
		Type: "object",
		Properties: map[string]*generator.Schema{
			"expected_hash": {
				Type:    "string",
				Pattern: "^[0-9a-f]{32}$",
			},
			"created_at": typedefs.UnixEpoch,
			"updated_at": typedefs.UnixEpoch,
		},
		AdditionalProperties: &falsy,
		Required: []string{
			"expected_hash",
		},
		XKokoConfig: &extension.Config{DisableValidateEndpoint: true},
	}
	err = generator.DefaultRegistry.Register(string(TypeHash), nodeSchema)
	if err != nil {
		panic(err)
	}
}
