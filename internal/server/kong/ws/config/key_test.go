package config

import (
	"context"
	"testing"

	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type mockKeyClient struct{}

func (mkc mockKeyClient) CreateKey(
	ctx context.Context,
	req *v1.CreateKeyRequest,
	opts ...grpc.CallOption,
) (*v1.CreateKeyResponse, error) {
	return &v1.CreateKeyResponse{Item: req.Item}, nil
}

func (mkc mockKeyClient) DeleteKey(
	ctx context.Context,
	req *v1.DeleteKeyRequest,
	opts ...grpc.CallOption,
) (*v1.DeleteKeyResponse, error) {
	return &v1.DeleteKeyResponse{}, nil
}

func (mkc mockKeyClient) GetKey(
	ctx context.Context,
	req *v1.GetKeyRequest,
	opts ...grpc.CallOption,
) (*v1.GetKeyResponse, error) {
	return &v1.GetKeyResponse{}, nil
}

func (mkc mockKeyClient) ListKeys(
	ctx context.Context,
	req *v1.ListKeysRequest,
	opts ...grpc.CallOption,
) (*v1.ListKeysResponse, error) {
	return &v1.ListKeysResponse{
		Items: []*model.Key{
			{
				Id:   "2bbf81d2-a42f-45c1-b41e-c406445ceda4",
				Name: "first",
			},
			{
				Id:   "8da62582-ce84-47ec-932f-7babda59c4e3",
				Name: "second",
			},
		},
	}, nil
}

func (mkc mockKeyClient) UpsertKey(
	ctx context.Context,
	req *v1.UpsertKeyRequest,
	opts ...grpc.CallOption,
) (*v1.UpsertKeyResponse, error) {
	return &v1.UpsertKeyResponse{
		Item: req.Item,
	}, nil
}

func TestKongKeyLoader_Mutate(t *testing.T) {
	config := DataPlaneConfig{}
	l := KongKeyLoader{Client: &mockKeyClient{}}

	opts := MutatorOpts{ClusterID: "xx"}

	err := l.Mutate(context.Background(), opts, config)
	require.NoError(t, err)
	require.Equal(t, []Map{
		{"id": "2bbf81d2-a42f-45c1-b41e-c406445ceda4", "name": "first"},
		{"id": "8da62582-ce84-47ec-932f-7babda59c4e3", "name": "second"},
	}, config["keys"])
}
