package resource

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/stretchr/testify/require"
)

func TestNewNode(t *testing.T) {
	r := NewNode()
	require.NotNil(t, r)
	require.NotNil(t, r.Node)
}

func TestNode_ID(t *testing.T) {
	var r Node
	id := r.ID()
	require.Empty(t, id)
	r = NewNode()
	id = r.ID()
	require.Empty(t, id)
}

func TestNode_Type(t *testing.T) {
	require.Equal(t, TypeNode, NewNode().Type())
}

func TestNode_ProcessDefaults(t *testing.T) {
	t.Run("defaults are correctly injected", func(t *testing.T) {
		r := NewNode()
		err := r.ProcessDefaults()
		require.Nil(t, err)
		require.True(t, validUUID(r.ID()))
	})
	t.Run("defaults do not override explicit values", func(t *testing.T) {
		r := NewNode()
		id := uuid.NewString()
		r.Node.Id = id
		err := r.ProcessDefaults()
		require.Nil(t, err)
		require.Equal(t, id, r.Node.Id)
	})
}

func TestNode_Indexes(t *testing.T) {
	node := NewNode()
	require.Nil(t, node.Indexes())
}

func TestNode_Validate(t *testing.T) {
	tests := []struct {
		name    string
		Node    func() Node
		wantErr bool
		Errs    []*model.ErrorDetail
	}{
		{
			name: "empty node throws an error",
			Node: func() Node {
				return NewNode()
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"missing properties: 'id', 'hostname', 'type', " +
							"'last_ping', 'version'",
					},
				},
			},
		},
		{
			name: "node with details filled in is valid",
			Node: func() Node {
				node := NewNode()
				node.Node = &model.Node{
					Id:       uuid.NewString(),
					Version:  "1.1a",
					Hostname: "secure-server",
					LastPing: 42,
					Type:     NodeTypeKongProxy,
				}
				return node
			},
			wantErr: false,
		},
		{
			name: "node with invalid uuid errors",
			Node: func() Node {
				node := NewNode()
				node.Node = &model.Node{
					Id:       "borked",
					Version:  "1.1a",
					Hostname: "secure-server",
					LastPing: 42,
					Type:     NodeTypeKongProxy,
				}
				return node
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
			name: "node with invalid type errors",
			Node: func() Node {
				node := NewNode()
				node.Node = &model.Node{
					Id:       uuid.NewString(),
					Version:  "1.1a",
					Hostname: "secure-server",
					LastPing: 42,
					Type:     "borked",
				}
				return node
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "type",
					Messages: []string{
						`value must be "kong-proxy"`,
					},
				},
			},
		},
		{
			name: "node with long version",
			Node: func() Node {
				node := NewNode()
				version := strings.Repeat("1.1", 43)
				node.Node = &model.Node{
					Id:       uuid.NewString(),
					Version:  version,
					Hostname: "secure-server",
					LastPing: 42,
					Type:     NodeTypeKongProxy,
				}
				return node
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "version",
					Messages: []string{
						"length must be <= 128, but got 129",
					},
				},
			},
		},
		{
			name: "node with ultra-long hostname",
			Node: func() Node {
				node := NewNode()
				hostname := strings.Repeat("secure-server-under-attack", 420)
				node.Node = &model.Node{
					Id:       uuid.NewString(),
					Version:  "1.1a",
					Hostname: hostname,
					LastPing: 42,
					Type:     NodeTypeKongProxy,
				}
				return node
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "hostname",
					Messages: []string{
						"length must be <= 1024, but got 10920",
					},
				},
			},
		},
		{
			name: "node with config_hash does not error",
			Node: func() Node {
				node := NewNode()
				node.Node = &model.Node{
					Id:         uuid.NewString(),
					Version:    "1.1a",
					Hostname:   "secure-server",
					LastPing:   42,
					ConfigHash: strings.Repeat("0", 32),
					Type:       NodeTypeKongProxy,
				}
				return node
			},
			wantErr: false,
		},
		{
			name: "node with a long config_hash errors",
			Node: func() Node {
				node := NewNode()
				node.Node = &model.Node{
					Id:         uuid.NewString(),
					Version:    "1.1a",
					Hostname:   "secure-server",
					LastPing:   42,
					ConfigHash: strings.Repeat("0", 42),
					Type:       NodeTypeKongProxy,
				}
				return node
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "config_hash",
					Messages: []string{
						"length must be <= 32, but got 42",
					},
				},
			},
		},
		{
			name: "node with invalid hash errors",
			Node: func() Node {
				node := NewNode()
				node.Node = &model.Node{
					Id:         uuid.NewString(),
					Version:    "1.1a",
					Hostname:   "secure-server",
					LastPing:   42,
					ConfigHash: "${{ jndi.foo }}",
					Type:       NodeTypeKongProxy,
				}
				return node
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "config_hash",
					Messages: []string{
						"length must be >= 32, but got 15",
						"must match pattern '[a-z0-9]{32}'",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.Node().Validate()
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
