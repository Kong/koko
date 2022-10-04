package resource

import (
	"context"
	"testing"

	"github.com/google/uuid"
	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/kong/koko/internal/test/util"
	"github.com/stretchr/testify/require"
)

func TestNewVault(t *testing.T) {
	r := NewVault()
	require.NotNil(t, r)
	require.NotNil(t, r.Vault)
}

func TestVault_Type(t *testing.T) {
	require.Equal(t, TypeVault, NewVault().Type())
}

func TestVault_ProcessDefaults(t *testing.T) {
	vault := NewVault()
	require.NoError(t, vault.ProcessDefaults(context.Background()))
	require.NotPanics(t, func() {
		uuid.MustParse(vault.ID())
	})
}

func TestVault_Validate(t *testing.T) {
	tests := []struct {
		name                    string
		Vault                   func() Vault
		wantErr                 bool
		skipIfEnterpriseTesting bool
		Errs                    []*model.ErrorDetail
	}{
		{
			name: "empty vault throws an error",
			Vault: func() Vault {
				return NewVault()
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"missing properties: 'prefix', 'name'",
					},
				},
			},
		},
		{
			name: "basic env vault passes",
			Vault: func() Vault {
				v := NewVault()
				v.Vault.Name = "env"
				v.Vault.Prefix = "test-vault-1"
				return v
			},
			wantErr: false,
		},
		{
			name: "unknown vault throws an error",
			Vault: func() Vault {
				v := NewVault()
				v.Vault.Name = "unknown"
				v.Vault.Prefix = "test-vault-1"
				return v
			},
			wantErr:                 true,
			skipIfEnterpriseTesting: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "name",
					Messages: []string{
						"value must be \"env\"",
					},
				},
			},
		},
		{
			name: "no name vault throws an error",
			Vault: func() Vault {
				v := NewVault()
				v.Vault.Name = ""
				v.Vault.Prefix = "test-vault-1"
				return v
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"missing properties: 'name'",
					},
				},
			},
		},
		{
			name: "no prefix vault throws an error",
			Vault: func() Vault {
				v := NewVault()
				v.Vault.Name = "env"
				v.Vault.Prefix = ""
				return v
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"missing properties: 'prefix'",
					},
				},
			},
		},
		{
			name: "advanced env vault passes",
			Vault: func() Vault {
				v := NewVault()
				v.Vault.Name = "env"
				v.Vault.Prefix = "test-vault-1"
				v.Vault.Config = &model.Vault_Config{
					Config: &model.Vault_Config_Env{
						Env: &model.Vault_EnvConfig{
							Prefix: "MY_",
						},
					},
				}
				return v
			},
			wantErr: false,
		},
		{
			name: "bad advanced env vault throws an error",
			Vault: func() Vault {
				v := NewVault()
				v.Vault.Name = "env"
				v.Vault.Config = &model.Vault_Config{
					Config: &model.Vault_Config_Env{
						Env: &model.Vault_EnvConfig{
							Prefix: "MY_",
						},
					},
				}
				return v
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_ENTITY,
					Field: "",
					Messages: []string{
						"missing properties: 'prefix'",
					},
				},
			},
		},
		{
			name: "bad prefix 1 throws an error",
			Vault: func() Vault {
				v := NewVault()
				v.Vault.Name = "env"
				v.Vault.Prefix = "env"
				return v
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "prefix",
					Messages: []string{
						"must not be any of [env] or [aws gcp hcv]",
					},
				},
			},
		},
		{
			name: "bad prefix 2 throws an error",
			Vault: func() Vault {
				v := NewVault()
				v.Vault.Name = "env"
				v.Vault.Prefix = "-bad"
				return v
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "prefix",
					Messages: []string{
						"must match pattern '^[a-z][a-z0-9-]*[a-z0-9]+$'",
					},
				},
			},
		},
		{
			name: "bad config prefix throws an error",
			Vault: func() Vault {
				v := NewVault()
				v.Vault.Name = "env"
				v.Vault.Prefix = "test-prefix-1"
				v.Vault.Config = &model.Vault_Config{
					Config: &model.Vault_Config_Env{
						Env: &model.Vault_EnvConfig{
							Prefix: "MY-",
						},
					},
				}
				return v
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "config.env.prefix",
					Messages: []string{
						"must match pattern '^[a-zA-Z_][a-zA-Z0-9_]*$'",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		util.SkipTestIfEnterpriseTesting(t, tt.skipIfEnterpriseTesting)
		t.Run(tt.name, func(t *testing.T) {
			v := tt.Vault()
			err := v.Validate(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.Errs != nil {
				verr, ok := err.(validation.Error)
				require.True(t, ok)
				require.ElementsMatch(t, tt.Errs, verr.Errs)
			}
		})
	}
}
