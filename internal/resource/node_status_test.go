package resource

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	nonPublic "github.com/kong/koko/internal/gen/grpc/kong/nonpublic/v1"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/stretchr/testify/require"
)

func TestNewNodeStatus(t *testing.T) {
	r := NewNodeStatus()
	require.NotNil(t, r)
	require.NotNil(t, r.NodeStatus)
}

func TestNodeStatus_ID(t *testing.T) {
	var r NodeStatus
	id := r.ID()
	require.Empty(t, id)
	r = NewNodeStatus()
	id = r.ID()
	require.Empty(t, id)
}

func TestNodeStatus_Type(t *testing.T) {
	require.Equal(t, TypeNodeStatus, NewNodeStatus().Type())
}

func TestNodeStatus_ProcessDefaults(t *testing.T) {
	t.Run("defaults are correctly injected", func(t *testing.T) {
		r := NewNodeStatus()
		err := r.ProcessDefaults(context.Background())
		require.NoError(t, err)
		require.True(t, validUUID(r.ID()))
	})
	t.Run("defaults do not override explicit values", func(t *testing.T) {
		r := NewNodeStatus()
		id := uuid.NewString()
		r.NodeStatus.Id = id
		err := r.ProcessDefaults(context.Background())
		require.NoError(t, err)
		require.Equal(t, id, r.NodeStatus.Id)
	})
	t.Run("empty resource return an error", func(t *testing.T) {
		var r NodeStatus
		require.Error(t, r.ProcessDefaults(context.Background()))
	})
}

func TestNodeStatus_Indexes(t *testing.T) {
	nodeStatus := NewNodeStatus()
	require.Nil(t, nodeStatus.Indexes())
}

func TestNodeStatus_Validate(t *testing.T) {
	tests := []struct {
		name       string
		NodeStatus func() NodeStatus
		wantErr    bool
		Errs       []*model.ErrorDetail
	}{
		{
			name: "empty node-status throws an error",
			NodeStatus: func() NodeStatus {
				return NewNodeStatus()
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"missing properties: 'id'",
					},
				},
			},
		},
		{
			name: "node-status with details filled in is valid",
			NodeStatus: func() NodeStatus {
				nodeStatus := NewNodeStatus()
				nodeStatus.NodeStatus = &nonPublic.NodeStatus{
					Id: uuid.NewString(),
					Issues: []*model.CompatibilityIssue{
						{
							Code: "F220",
						},
						{
							Code: "P420",
							AffectedResources: []*model.Resource{
								{
									Type: "foo",
									Id:   uuid.NewString(),
								},
							},
						},
					},
				}
				return nodeStatus
			},
			wantErr: false,
		},
		{
			name: "node-status with invalid uuid errors",
			NodeStatus: func() NodeStatus {
				nodeStatus := NewNodeStatus()
				nodeStatus.NodeStatus = &nonPublic.NodeStatus{
					Id: "borked",
				}
				return nodeStatus
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "id",
					Messages: []string{
						"must be a valid UUID",
					},
				},
			},
		},
		{
			name: "node-status with invalid issue code errors",
			NodeStatus: func() NodeStatus {
				nodeStatus := NewNodeStatus()
				nodeStatus.NodeStatus = &nonPublic.NodeStatus{
					Id: uuid.NewString(),
					Issues: []*model.CompatibilityIssue{
						{
							Code: "F4224",
						},
					},
				}
				return nodeStatus
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "issues[0].code",
					Messages: []string{
						`length must be <= 4, but got 5`,
						`must match pattern '^[A-Z][A-Z\d]{3}$'`,
					},
				},
			},
		},
		{
			name: "node-status with invalid affected resources errors",
			NodeStatus: func() NodeStatus {
				nodeStatus := NewNodeStatus()
				nodeStatus.NodeStatus = &nonPublic.NodeStatus{
					Id: uuid.NewString(),
					Issues: []*model.CompatibilityIssue{
						{
							Code: "F424",
							AffectedResources: []*model.Resource{
								{
									Type: "Foo",
								},
							},
						},
					},
				}
				return nodeStatus
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "issues[0].affected_resources[0]",
					Messages: []string{
						"missing properties: 'id'",
					},
				},
			},
		},
		{
			name: "node-status with invalid affected resources errors",
			NodeStatus: func() NodeStatus {
				nodeStatus := NewNodeStatus()
				nodeStatus.NodeStatus = &nonPublic.NodeStatus{
					Id: uuid.NewString(),
					Issues: []*model.CompatibilityIssue{
						{
							Code: "F424",
							AffectedResources: []*model.Resource{
								{
									Id: uuid.NewString(),
								},
							},
						},
					},
				}
				return nodeStatus
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "issues[0].affected_resources[0]",
					Messages: []string{
						"missing properties: 'type'",
					},
				},
			},
		},
		{
			name: "node-status with invalid affected resources type errors",
			NodeStatus: func() NodeStatus {
				nodeStatus := NewNodeStatus()
				nodeStatus.NodeStatus = &nonPublic.NodeStatus{
					Id: uuid.NewString(),
					Issues: []*model.CompatibilityIssue{
						{
							Code: "F424",
							AffectedResources: []*model.Resource{
								{
									Type: strings.Repeat("node-", 7),
									Id:   uuid.NewString(),
								},
							},
						},
					},
				}
				return nodeStatus
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "issues[0].affected_resources[0].type",
					Messages: []string{
						"length must be <= 32, but got 35",
					},
				},
			},
		},
		{
			name: "node-status with too many issues errors",
			NodeStatus: func() NodeStatus {
				nodeStatus := NewNodeStatus()
				nodeStatus.NodeStatus = &nonPublic.NodeStatus{
					Id: uuid.NewString(),
				}
				for i := 0; i <= 128; i++ {
					nodeStatus.NodeStatus.Issues = append(nodeStatus.NodeStatus.Issues,
						&model.CompatibilityIssue{
							Code: fmt.Sprintf("F%03d", i),
						})
				}
				return nodeStatus
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "issues",
					Messages: []string{
						"maximum 128 items required, but found 129 items",
					},
				},
			},
		},
		{
			name: "node-status issue with too many resources errors",
			NodeStatus: func() NodeStatus {
				nodeStatus := NewNodeStatus()
				var resources []*model.Resource
				for i := 0; i <= 128; i++ {
					resources = append(resources, &model.Resource{
						Type: "plugin",
						Id:   uuid.NewString(),
					})
				}
				nodeStatus.NodeStatus = &nonPublic.NodeStatus{
					Id: uuid.NewString(),
					Issues: []*model.CompatibilityIssue{
						{
							Code:              "F424",
							AffectedResources: resources,
						},
					},
				}
				return nodeStatus
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "issues[0].affected_resources",
					Messages: []string{
						"maximum 128 items required, but found 129 items",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.NodeStatus().Validate(context.Background())
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
