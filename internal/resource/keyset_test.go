package resource

import (
	"context"
	"testing"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/stretchr/testify/require"
)

func TestNewKeySet(t *testing.T) {
	ks := NewKeySet()
	require.NotNil(t, ks)
	require.NotNil(t, ks.KeySet)
}

func TestKeySet_ID(t *testing.T) {
	var ks KeySet
	id := ks.ID()
	require.Empty(t, id)

	ks = NewKeySet()
	id = ks.ID()
	require.Empty(t, id)
}

func TestKeySet_Type(t *testing.T) {
	require.Equal(t, TypeKeySet, NewKeySet().Type())
}

func TestKeySet_ProcessDefaults(t *testing.T) {
	ks := NewKeySet()
	err := ks.ProcessDefaults(context.Background())
	require.NoError(t, err)
	require.True(t, validUUID(ks.ID()))

	ks.KeySet.Id = ""
	ks.KeySet.CreatedAt = 0
	ks.KeySet.UpdatedAt = 0
	require.Equal(t, ks.Resource(), &v1.KeySet{})
}

func goodKeySet() KeySet {
	ks := NewKeySet()
	ks.KeySet.Name = "good_set-of-keys"
	ks.ProcessDefaults(context.Background())
	return ks
}

func TestKeySet_Validate(t *testing.T) {
	tests := []struct {
		name    string
		KeySet  func() KeySet
		wantErr bool
		Errs    []*v1.ErrorDetail
	}{
		{
			name:    "empty key isn't valid",
			KeySet:  NewKeySet,
			wantErr: true,
			Errs: []*v1.ErrorDetail{
				{
					Type: v1.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"missing properties: 'id', 'name'",
					},
				},
			},
		},
		{
			name: "needs a name",
			KeySet: func() KeySet {
				ks := NewKeySet()
				ks.ProcessDefaults(context.Background())
				return ks
			},
			wantErr: true,
			Errs: []*v1.ErrorDetail{
				{
					Type: v1.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"missing properties: 'name'",
					},
				},
			},
		},
		{
			name:   "good key set",
			KeySet: goodKeySet,
		},
		{
			name: "repeated tags isn't valid",
			KeySet: func() KeySet {
				ks := goodKeySet()
				ks.KeySet.Tags = []string{"A1", "X3", "A1"}
				return ks
			},
			wantErr: true,
			Errs: []*v1.ErrorDetail{
				{
					Type:     v1.ErrorType_ERROR_TYPE_FIELD,
					Field:    "tags",
					Messages: []string{"items at index 0 and 2 are equal"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ks := tt.KeySet()
			err := ks.Validate(context.Background())
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
