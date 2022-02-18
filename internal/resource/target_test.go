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

func TestTarget_fortmatTarget(t *testing.T) {
	tests := []struct {
		name     string
		target   string
		expected string
		wantErr  bool
	}{
		{
			name:     "valid ipv4 ok",
			target:   "192.0.2.1",
			expected: "192.0.2.1:8000",
		},
		{
			name:     "valid ipv4 and port ok",
			target:   "192.0.2.1:80",
			expected: "192.0.2.1:80",
		},
		{
			name:    "valid ipv4 and invalid port fails",
			target:  "192.0.2.1:-80",
			wantErr: true,
		},
		{
			name:    "valid ipv4 and invalid port fails",
			target:  "192.0.2.1:99999999999999",
			wantErr: true,
		},
		{
			name:     "valid ipv6 ok",
			target:   "2001:DB8::1",
			expected: "[2001:0db8:0000:0000:0000:0000:0000:0001]:8000",
		},
		{
			name:    "invalid ipv6 fails",
			target:  "2001:DBk::1",
			wantErr: true,
		},
		{
			name:    "invalid ipv6 fails",
			target:  "2001:DBDBDB::1",
			wantErr: true,
		},
		{
			name:    "invalid ipv6 fails",
			target:  "2001:DB8:85a3:0000:0000:8a2e:370:7334:1234",
			wantErr: true,
		},
		{
			name:     "valid ipv6 ok",
			target:   "2001:DB8::",
			expected: "[2001:0db8:0000:0000:0000:0000:0000:0000]:8000",
		},
		{
			name:     "valid ipv6 ok",
			target:   "2001:DB8:85a3:0000:0000:8a2e:370:7334",
			expected: "[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:8000",
		},
		{
			name:     "valid ipv6 ok",
			target:   "2001:DB8:85a3::8a2e:370:7334",
			expected: "[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:8000",
		},
		{
			name:     "valid ipv6 ok",
			target:   "2001:DB8:2::3:4:5",
			expected: "[2001:0db8:0002:0000:0000:0003:0004:0005]:8000",
		},
		{
			name:     "valid ipv6 brackets ok",
			target:   "[2001:DB8::1]",
			expected: "[2001:0db8:0000:0000:0000:0000:0000:0001]:8000",
		},
		{
			name:    "invalid ipv6 with brackets fails",
			target:  "[2001:DBk::1]",
			wantErr: true,
		},
		{
			name:     "valid ipv6 brackets and port ok",
			target:   "[2001:DB8::1]:80",
			expected: "[2001:0db8:0000:0000:0000:0000:0000:0001]:80",
		},
		{
			name:    "valid ipv6 brackets and invalid port fails",
			target:  "[2001:DB8::1]:-80",
			wantErr: true,
		},
		{
			name:    "valid ipv6 brackets and invalid port fails",
			target:  "[2001:DB8::1]:99999999999999",
			wantErr: true,
		},
		{
			name:     "valid domain ok",
			target:   "valid",
			expected: "valid:8000",
		},
		{
			name:     "valid dotted domain ok",
			target:   "valid.domain",
			expected: "valid.domain:8000",
		},
		{
			name:     "valid domain and port ok",
			target:   "valid:80",
			expected: "valid:80",
		},
		{
			name:    "valid domain and invalid port fails",
			target:  "valid:-80",
			wantErr: true,
		},
		{
			name:    "valid domain and invalid port fails",
			target:  "valid:99999999999999",
			wantErr: true,
		},
		{
			name:     "valid dotted domain and port ok",
			target:   "valid.name:80",
			expected: "valid.name:80",
		},
		{
			name:    "valid dotted domain and invalid port fails",
			target:  "valid.name:-80",
			wantErr: true,
		},
		{
			name:    "valid dotted domain and invalid port fails",
			target:  "valid.name:99999999999999",
			wantErr: true,
		},
		{
			name:    "invalid domain throws an error",
			target:  "abc.cde:80:80",
			wantErr: true,
		},
		{
			name:    "invalid ipv4 throws an error",
			target:  "192.0.2.1.1",
			wantErr: true,
		},
		{
			name:    "invalid ipv4 throws an error",
			target:  "300.0.2.1",
			wantErr: true,
		},
		{
			name:    "invalid name fails",
			target:  "\\\\bad\\\\////name////",
			wantErr: true,
		},
		{
			name:    "invalid domain and port fails",
			target:  "1.2.3.4.5:80",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateAndFormatTarget(tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTarget() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.expected {
				t.Errorf("validateTarget() got = %s, expected %s", got, tt.expected)
			}
		})
	}
}
