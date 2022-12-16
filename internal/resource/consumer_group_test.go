package resource

import (
	"context"
	"testing"

	"github.com/google/uuid"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/kong/koko/internal/test/util"
	"github.com/stretchr/testify/require"
)

func TestNewConsumerGroup(t *testing.T) {
	s := NewConsumerGroup()
	require.NotNil(t, s)
	require.NotNil(t, s.ConsumerGroup)
}

func TestConsumerGroup_ID(t *testing.T) {
	var s ConsumerGroup
	id := s.ID()
	require.Empty(t, id)
	s = NewConsumerGroup()
	id = s.ID()
	require.Empty(t, id)
}

func TestConsumerGroup_Type(t *testing.T) {
	require.Equal(t, TypeConsumerGroup, NewConsumerGroup().Type())
}

func TestConsumerGroup_ProcessDefaults(t *testing.T) {
	t.Run("defaults are correctly injected", func(t *testing.T) {
		r := NewConsumerGroup()
		err := r.ProcessDefaults(context.Background())
		require.Nil(t, err)
		require.NotPanics(t, func() {
			uuid.MustParse(r.ID())
		})
	})
}

func TestConsumerGroup_Validate(t *testing.T) {
	tests := []struct {
		name                    string
		ConsumerGroup           func() ConsumerGroup
		wantErr                 bool
		skipIfEnterpriseTesting bool
		Errs                    []*v1.ErrorDetail
	}{
		{
			name: "empty consumer-group throws an error",
			ConsumerGroup: func() ConsumerGroup {
				return NewConsumerGroup()
			},
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
			name: "missing name throws an error",
			ConsumerGroup: func() ConsumerGroup {
				r := NewConsumerGroup()
				r.ProcessDefaults(context.Background())
				return r
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
			name: "valid consumer-group fails because it's not implemented",
			ConsumerGroup: func() ConsumerGroup {
				r := NewConsumerGroup()
				r.ProcessDefaults(context.Background())
				r.ConsumerGroup.Name = "test"
				return r
			},
			skipIfEnterpriseTesting: true,
			wantErr:                 true,
			Errs: []*v1.ErrorDetail{
				{
					Type: v1.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						`Consumer Groups are a Kong Enterprise-only feature. ` +
							`Please upgrade to Kong Enterprise to use this feature.`,
					},
				},
			},
		},
		{
			name: "consumer-group with invalid member fails",
			ConsumerGroup: func() ConsumerGroup {
				r := NewConsumerGroup()
				r.ProcessDefaults(context.Background())
				r.ConsumerGroup.Name = "test"
				r.MemberIDsToAdd = []string{"foo"}
				return r
			},
			wantErr: true,
			Errs: []*v1.ErrorDetail{
				{
					Type:  v1.ErrorType_ERROR_TYPE_FIELD,
					Field: "consumer_id",
					Messages: []string{
						"must be a valid UUID",
					},
				},
			},
		},
		{
			name: "consumer-group with valid member fails because it's not implemented",
			ConsumerGroup: func() ConsumerGroup {
				r := NewConsumerGroup()
				r.ProcessDefaults(context.Background())
				r.ConsumerGroup.Name = "test"
				r.MemberIDsToAdd = []string{uuid.NewString()}
				return r
			},
			skipIfEnterpriseTesting: true,
			wantErr:                 true,
			Errs: []*v1.ErrorDetail{
				{
					Type: v1.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						`Consumer Groups are a Kong Enterprise-only feature. ` +
							`Please upgrade to Kong Enterprise to use this feature.`,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			util.SkipTestIfEnterpriseTesting(t, tt.skipIfEnterpriseTesting)
			err := tt.ConsumerGroup().Validate(context.Background())
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
