package resource

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestNewTarget(t *testing.T) {
	r := NewTarget()
	require.NotNil(t, r)
	require.NotNil(t, r.Target)
}

func TestTarget_ID(t *testing.T) {
	var r Target
	id := r.ID()
	require.Empty(t, id)
	r = NewTarget()
	id = r.ID()
	require.Empty(t, id)
}

func TestTarget_Type(t *testing.T) {
	require.Equal(t, TypeTarget, NewTarget().Type())
}

func TestTarget_ProcessDefaults(t *testing.T) {
	t.Run("defaults are correctly injected", func(t *testing.T) {
		r := NewTarget()
		err := r.ProcessDefaults()
		require.Nil(t, err)
		require.True(t, validUUID(r.ID()))
		// empty out the id for equality comparison
		r.Target.Id = ""
		r.Target.CreatedAt = 0
		r.Target.UpdatedAt = 0
		require.Equal(t, r.Resource(), &model.Target{Weight: wrapperspb.Int32(100)})
	})
	t.Run("defaults do not override explicit values", func(t *testing.T) {
		r := NewTarget()
		r.Target.Target = "10.42.42.42:42"
		r.Target.Weight = wrapperspb.Int32(420)
		err := r.ProcessDefaults()
		require.Nil(t, err)
		require.True(t, validUUID(r.ID()))
		// empty out the id and ts for equality comparison
		r.Target.Id = ""
		require.Equal(t, &model.Target{
			Weight: wrapperspb.Int32(420),
			Target: "10.42.42.42:42",
		}, r.Resource())
	})
}

func goodTarget() Target {
	r := NewTarget()
	r.Target.Target = "10.42.42.42:42"
	r.Target.Upstream = &model.Upstream{
		Id: uuid.NewString(),
	}
	_ = r.ProcessDefaults()
	return r
}

func TestTarget_Validate(t *testing.T) {
	tests := []struct {
		name    string
		Target  func() Target
		wantErr bool
		Errs    []*model.ErrorDetail
		Skip    bool
	}{
		{
			name: "empty target throws an error",
			Target: func() Target {
				return NewTarget()
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"missing properties: 'id', 'target', 'upstream'",
					},
				},
			},
		},
		{
			name: "target.upstream without id throws an error",
			Target: func() Target {
				t := goodTarget()
				t.Target.Upstream = &model.Upstream{}
				return t
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "upstream",
					Messages: []string{
						"missing properties: 'id'",
					},
				},
			},
		},
		{
			name: "good target doesn't throw any error",
			Target: func() Target {
				return goodTarget()
			},
			wantErr: false,
			Errs:    []*model.ErrorDetail{},
		},
		{
			name: "target with target longer than 1024 errors",
			Target: func() Target {
				t := goodTarget()
				t.Target.Target = strings.Repeat("foo.bar", 147)
				return t
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "target",
					Messages: []string{
						"length must be <= 1024, but got 1029",
					},
				},
			},
		},
		{
			name: "target with negative weight errors",
			Target: func() Target {
				t := goodTarget()
				t.Target.Weight = wrapperspb.Int32(-1)
				return t
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "weight",
					Messages: []string{
						"must be >= 0 but found -1",
					},
				},
			},
		},
		{
			name: "weight higher than 65535 errors",
			Target: func() Target {
				t := goodTarget()
				t.Target.Weight = wrapperspb.Int32(65536)
				return t
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "weight",
					Messages: []string{
						"must be <= 65535 but found 65536",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.Skip {
				t.Skip()
			}
			target := tt.Target()
			err := target.Validate()
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
