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

	defaultHcvPort       = 8200
	defaultHcvHost       = "127.0.0.1"
	defaultHcvKV         = "v1"
	defaultHcvMount      = "secret"
	defaultHcvProtocol   = typedefs.ProtocolHTTP
	defaultHcvAuthMethod = "token"
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

	defaultHcv = &v1.Vault_Config_Hcv{
		Hcv: &v1.Vault_HcvConfig{
			Host:       defaultHcvHost,
			Port:       defaultHcvPort,
			Kv:         defaultHcvKV,
			Mount:      defaultHcvMount,
			Protocol:   defaultHcvProtocol,
			AuthMethod: defaultHcvAuthMethod,
		},
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

	if r.Vault.Name == "hcv" {
		if r.Vault.Config.GetHcv() == nil {
			r.Vault.Config.Config = defaultHcv
		} else {
			if r.Vault.Config.GetHcv().Host == "" {
				r.Vault.Config.GetHcv().Host = defaultHcvHost
			}
			if r.Vault.Config.GetHcv().Kv == "" {
				r.Vault.Config.GetHcv().Kv = defaultHcvKV
			}
			if r.Vault.Config.GetHcv().Mount == "" {
				r.Vault.Config.GetHcv().Mount = defaultHcvMount
			}
			if r.Vault.Config.GetHcv().Port == 0 {
				r.Vault.Config.GetHcv().Port = defaultHcvPort
			}
			if r.Vault.Config.GetHcv().Protocol == "" {
				r.Vault.Config.GetHcv().Protocol = defaultHcvProtocol
			}
			if r.Vault.Config.GetHcv().AuthMethod == "" {
				r.Vault.Config.GetHcv().AuthMethod = defaultHcvAuthMethod
			}
		}
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
		AdditionalProperties: &falsy,
		Properties: map[string]*generator.Schema{
			"env": {
				Type:                 "object",
				AdditionalProperties: &falsy,
				Properties: map[string]*generator.Schema{
					"prefix": {
						Type:    "string",
						Pattern: "^[a-zA-Z_][a-zA-Z0-9_]*$",
					},
				},
			},
		},
	}
	vaultSchema := &generator.Schema{
		AdditionalProperties: &falsy,
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
				Type:                 "object",
				AdditionalProperties: &falsy,
				Properties: func() map[string]*generator.Schema {
					s := map[string]*generator.Schema{}
					for _, v := range VaultTypes {
						vs, ok := v.(string)
						if !ok {
							panic(fmt.Sprintf("expected string, got unexpected type: %T", v))
						}
						s[vs] = &generator.Schema{}
					}
					return s
				}(),
			},
			"description": {Type: "string"},
			"tags":        typedefs.Tags,
			"created_at":  typedefs.UnixEpoch,
			"updated_at":  typedefs.UnixEpoch,
		},
		AllOf: []*generator.Schema{
			{
				If: &generator.Schema{
					Properties: map[string]*generator.Schema{
						"name": {
							Const: "env",
						},
					},
				},
				Then: &generator.Schema{
					Type: "object",
					Properties: map[string]*generator.Schema{
						"config": envConfigSchema,
					},
				},
			},
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
