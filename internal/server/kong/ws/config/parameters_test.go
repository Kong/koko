package config

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewParametersLoader(t *testing.T) {
	type args struct {
		clusterID string
	}
	tests := []struct {
		name    string
		args    args
		want    *ParametersLoader
		wantErr error
	}{
		{
			name: "creates a new parameters loader",
			args: args{
				clusterID: "b9d640b2-8551-498b-8da4-a55278beefb1",
			},
			want: &ParametersLoader{
				ClusterID: "b9d640b2-8551-498b-8da4-a55278beefb1",
			},
			wantErr: nil,
		},
		{
			name: "fails when clusterID is not a valid UUID",
			args: args{
				clusterID: "not-a-valid-uuid",
			},
			want:    nil,
			wantErr: fmt.Errorf("invalid UUID length: 16"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewParametersLoader(tt.args.clusterID)
			assert.Equal(t, tt.want, got)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestParametersLoader_Name(t *testing.T) {
	t.Run("should return correct loader name", func(t *testing.T) {
		l := ParametersLoader{}
		assert.Equal(t, "parameters", l.Name())
	})
}

func TestParametersLoader_Mutate(t *testing.T) {
	t.Run("sets cluster_id correctly", func(t *testing.T) {
		l := &ParametersLoader{
			ClusterID: "b9d640b2-8551-498b-8da4-a55278beefb1",
		}
		expected := []Map{{"key": "cluster_id", "value": "b9d640b2-8551-498b-8da4-a55278beefb1"}}
		config := DataPlaneConfig{}
		err := l.Mutate(context.Background(), MutatorOpts{}, config)
		require.NoError(t, err)
		require.Equal(t, expected, config["parameters"])
	})
}
