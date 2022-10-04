package config

import (
	"testing"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/stretchr/testify/require"
)

func Test_flattenVaultConfig(t *testing.T) {
	tests := []struct {
		name     string
		vault    *v1.Vault
		expected map[string]interface{}
	}{
		{
			name: "do not modify koko vault without config",
			vault: &v1.Vault{
				Name:   "env",
				Prefix: "test-vault-1",
			},
			expected: map[string]interface{}{
				"name":   "env",
				"prefix": "test-vault-1",
			},
		},
		{
			name: "do not modify koko vault with nil config",
			vault: &v1.Vault{
				Name:   "env",
				Prefix: "test-vault-1",
				Config: nil,
			},
			expected: map[string]interface{}{
				"name":   "env",
				"prefix": "test-vault-1",
			},
		},
		{
			name: "modify koko vault with nil env config",
			vault: &v1.Vault{
				Name:   "env",
				Prefix: "test-vault-1",
				Config: &v1.Vault_Config{
					Config: &v1.Vault_Config_Env{
						Env: nil,
					},
				},
			},
			expected: map[string]interface{}{
				"name":   "env",
				"prefix": "test-vault-1",
				"config": map[string]interface{}{},
			},
		},
		{
			name: "modify koko vault with empty string env config",
			vault: &v1.Vault{
				Name:   "env",
				Prefix: "test-vault-1",
				Config: &v1.Vault_Config{
					Config: &v1.Vault_Config_Env{
						Env: &v1.Vault_EnvConfig{
							Prefix: "",
						},
					},
				},
			},
			expected: map[string]interface{}{
				"name":   "env",
				"prefix": "test-vault-1",
				"config": map[string]interface{}{},
			},
		},
		{
			name: "flatten koko vault format to kong vault format",
			vault: &v1.Vault{
				Name:   "env",
				Prefix: "test-vault-1",
				Config: &v1.Vault_Config{
					Config: &v1.Vault_Config_Env{
						Env: &v1.Vault_EnvConfig{
							Prefix: "TEST_",
						},
					},
				},
			},
			expected: map[string]interface{}{
				"name":   "env",
				"prefix": "test-vault-1",
				"config": map[string]interface{}{
					"prefix": "TEST_",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := convert(tt.vault)
			require.NoError(t, err)
			fv := flattenVaultConfig(m)

			expected, err := convert(tt.expected)
			require.NoError(t, err)

			require.Equal(t, expected, fv)
		})
	}
}
