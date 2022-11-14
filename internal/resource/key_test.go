package resource

import (
	"context"
	"testing"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/stretchr/testify/require"
)

func TestNewKey(t *testing.T) {
	k := NewKey()
	require.NotNil(t, k)
	require.NotNil(t, k.Key)
}

func TestKey_ID(t *testing.T) {
	var k Key
	id := k.ID()
	require.Empty(t, id)

	k = NewKey()
	id = k.ID()
	require.Empty(t, id)
}

func TestKey_Type(t *testing.T) {
	require.Equal(t, TypeKey, NewKey().Type())
}

func TestKey_ProcessDefaults(t *testing.T) {
	k := NewKey()
	err := k.ProcessDefaults(context.Background())
	require.NoError(t, err)
	require.True(t, validUUID(k.ID()))

	k.Key.Id = ""
	k.Key.CreatedAt = 0
	k.Key.UpdatedAt = 0
	require.Equal(t, k.Resource(), &v1.Key{})
}

func goodKey() Key {
	k := NewKey()
	k.ProcessDefaults(context.Background())
	return k
}

func TestKey_Validate(t *testing.T) {
	tests := []struct {
		name    string
		Key     func() Key
		wantErr bool
		Errs    []*v1.ErrorDetail
	}{
		{
			name:    "empty key isn't valid",
			Key:     NewKey,
			wantErr: true,
			Errs: []*v1.ErrorDetail{
				{
					Type: v1.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"missing properties: 'id', 'kid'",
					},
				},
			},
		},
		{
			name: "key without kid isn't valid",
			Key: func() Key {
				k := goodKey()
				k.Key.Kid = ""
				return k
			},
			wantErr: true,
			Errs: []*v1.ErrorDetail{
				{
					Type: v1.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"missing properties: 'kid'",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.Key()
			err := k.Validate(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.Errs != nil {
				verr, _ := err.(validation.Error)
				require.ElementsMatch(t, tt.Errs, verr.Errs)
			}
		})
	}
}
