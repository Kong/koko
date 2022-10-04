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

const (
	TypeVault model.Type = "vault"
)

var (
	VaultTypes = []interface{}{
		"env",
	}
	EnterpriseVaultTypes = []interface{}{
		"aws",
		"gcp",
		"hcv",
	}
)

func NewVault() Vault {
	return Vault{
		Vault: &v1.Vault{},
	}
}

type Vault struct {
	Vault *v1.Vault
}

func (r Vault) ID() string {
	if r.Vault == nil {
		return ""
	}
	return r.Vault.Id
}

func (r Vault) Type() model.Type {
	return TypeVault
}

func (r Vault) Resource() model.Resource {
	return r.Vault
}

// SetResource implements the Object.SetResource interface.
func (r Vault) SetResource(pr model.Resource) error { return model.SetResource(r, pr) }

func (r Vault) Validate(ctx context.Context) error {
	return validation.Validate(string(TypeVault), r.Vault)
}

func (r Vault) ProcessDefaults(ctx context.Context) error {
	if r.Vault == nil {
		return fmt.Errorf("invalid nil resource")
	}
	defaultID(&r.Vault.Id)
	return nil
}

func (r Vault) Indexes() []model.Index {
	return []model.Index{
		{
			Name:      "prefix",
			Type:      model.IndexUnique,
			Value:     r.Vault.Prefix,
			FieldName: "prefix",
		},
	}
}

func init() {
	err := model.RegisterType(TypeVault, &v1.Vault{}, func() model.Object {
		return NewVault()
	})
	if err != nil {
		panic(err)
	}

	envConfigSchema := &generator.Schema{
		Type: "object",
		Properties: map[string]*generator.Schema{
			"prefix": {
				Type:    "string",
				Pattern: "^[a-zA-Z_][a-zA-Z0-9_]*$",
			},
		},
		AdditionalProperties: &falsy,
	}
	vaultSchema := &generator.Schema{
		Properties: map[string]*generator.Schema{
			"id": typedefs.ID,
			"prefix": {
				Type:    "string",
				Pattern: "^[a-z][a-z0-9-]*[a-z0-9]+$",
				Not: &generator.Schema{
					AnyOf: []*generator.Schema{
						{
							Enum: VaultTypes,
						},
						{
							Enum: EnterpriseVaultTypes,
						},
					},
					Description: fmt.Sprintf("must not be any of %v or %v", VaultTypes, EnterpriseVaultTypes),
				},
			},
			"name": {
				Type: "string",
				Enum: VaultTypes,
			},
			"config": {
				OneOf: []*generator.Schema{
					{
						Properties: map[string]*generator.Schema{
							"env": envConfigSchema,
						},
					},
				},
			},
			"description": {Type: "string"},
			"tags":        typedefs.Tags,
			"created_at":  typedefs.UnixEpoch,
			"updated_at":  typedefs.UnixEpoch,
		},
		Required: []string{
			"prefix",
			"name",
		},
		XKokoConfig: &extension.Config{
			ResourceAPIPath: "vaults",
		},
	}
	err = generator.DefaultRegistry.Register(string(TypeVault), vaultSchema)
	if err != nil {
		panic(err)
	}
}
