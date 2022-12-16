package ws

import (
	"context"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
	"github.com/kong/koko/internal/server/kong/ws/config"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type mockConfigLoader struct {
	mock.Mock
}

func newMockConfigLoader() *mockConfigLoader {
	return &mockConfigLoader{}
}

func (m *mockConfigLoader) Load(ctx context.Context, clusterID string) (config.Content, error) {
	args := m.Called(ctx, clusterID)
	content, ok := args.Get(0).(config.Content)
	if !ok {
		panic("return value must be of type config.Content")
	}
	return content, args.Error(1)
}

type mockNodeClient struct {
	mock.Mock
}

func newMockNodeClient() *mockNodeClient {
	return &mockNodeClient{}
}

func (m *mockNodeClient) GetNode(ctx context.Context, in *v1.GetNodeRequest, opts ...grpc.CallOption) (*v1.GetNodeResponse, error) { //nolint:lll
	args := m.Called(ctx, in, opts)
	resp, ok := args.Get(0).(*v1.GetNodeResponse)
	if !ok {
		panic("return value must be of type *v1.GetNodeResponse")
	}
	return resp, args.Error(1)
}

func (m *mockNodeClient) CreateNode(ctx context.Context, in *v1.CreateNodeRequest, opts ...grpc.CallOption) (*v1.CreateNodeResponse, error) { //nolint:lll
	args := m.Called(ctx, in, opts)
	resp, ok := args.Get(0).(*v1.CreateNodeResponse)
	if !ok {
		panic("return value must be of type *v1.CreateNodeResponse")
	}
	return resp, args.Error(1)
}

func (m *mockNodeClient) UpsertNode(ctx context.Context, in *v1.UpsertNodeRequest, opts ...grpc.CallOption) (*v1.UpsertNodeResponse, error) { //nolint:lll
	args := m.Called(ctx, in, opts)
	resp, ok := args.Get(0).(*v1.UpsertNodeResponse)
	if !ok {
		panic("return value must be of type *v1.UpsertNodeResponse")
	}
	return resp, args.Error(1)
}

func (m *mockNodeClient) DeleteNode(ctx context.Context, in *v1.DeleteNodeRequest, opts ...grpc.CallOption) (*v1.DeleteNodeResponse, error) { //nolint:lll
	args := m.Called(ctx, in, opts)
	resp, ok := args.Get(0).(*v1.DeleteNodeResponse)
	if !ok {
		panic("return value must be of type *v1.DeleteNodeResponse")
	}
	return resp, args.Error(1)
}

func (m *mockNodeClient) ListNodes(ctx context.Context, in *v1.ListNodesRequest, opts ...grpc.CallOption) (*v1.ListNodesResponse, error) { //nolint:lll
	args := m.Called(ctx, in, opts)
	resp, ok := args.Get(0).(*v1.ListNodesResponse)
	if !ok {
		panic("return value must be of type *v1.ListNodesResponse")
	}
	return resp, args.Error(1)
}
