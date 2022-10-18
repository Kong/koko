package relay

import (
	"context"
	"testing"

	"github.com/google/uuid"
	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	nonPublic "github.com/kong/koko/internal/gen/grpc/kong/nonpublic/v1"
	relay "github.com/kong/koko/internal/gen/grpc/kong/relay/service/v1"
	"github.com/kong/koko/internal/log"
	"github.com/kong/koko/internal/resource"
	serverUtil "github.com/kong/koko/internal/server/util"
	"github.com/kong/koko/internal/store"
	"github.com/kong/koko/internal/test/util"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	statusPb "google.golang.org/grpc/status"
)

func TestRelayStatusServiceUpdateNodeStatus(t *testing.T) {
	ctx := context.Background()
	persister, err := util.GetPersister(t)
	require.Nil(t, err)
	db := store.New(persister, log.Logger).ForCluster(store.DefaultCluster)
	opts := StatusServiceOpts{
		StoreLoader: serverUtil.DefaultStoreLoader{Store: db},
		Logger:      log.Logger,
	}
	server := NewStatusService(opts)
	require.NotNil(t, server)
	l := setup()
	s := grpc.NewServer(grpc.ChainUnaryInterceptor(
		serverUtil.LoggerInterceptor(opts.Logger),
		serverUtil.PanicInterceptor(opts.Logger)))
	relay.RegisterStatusServiceServer(s, server)
	cc := clientConn(t, l)
	client := relay.NewStatusServiceClient(cc)
	go func() {
		_ = s.Serve(l)
	}()
	defer s.Stop()

	t.Run("updates a given node-status", func(t *testing.T) {
		defer func() {
			util.CleanDB(t)
		}()
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		_, err := client.UpdateNodeStatus(ctx, &relay.UpdateNodeStatusRequest{
			Item: &nonPublic.NodeStatus{
				Id: uuid.NewString(),
				Issues: []*model.CompatibilityIssue{
					{
						Code: "T420",
					},
				},
			},
		})
		require.Nil(t, err)

		// verify the node-status in database
		list := resource.NewList(resource.TypeNodeStatus)
		err = db.List(ctx, list)
		require.NoError(t, err)
		require.Len(t, list.GetAll(), 1)
		item := list.GetAll()[0]
		nodeStatus, ok := item.Resource().(*nonPublic.NodeStatus)
		require.True(t, ok)
		require.Equal(t, "T420", nodeStatus.Issues[0].Code)
	})
	t.Run("updates an existing node-status", func(t *testing.T) {
		defer func() {
			util.CleanDB(t)
		}()
		nodeStatusID := uuid.NewString()
		nodeStatus := resource.NewNodeStatus()
		nodeStatus.NodeStatus = &nonPublic.NodeStatus{
			Id: nodeStatusID,
			Issues: []*model.CompatibilityIssue{
				{
					Code: "T420",
				},
			},
		}
		err := db.Upsert(ctx, nodeStatus)
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		_, err = client.UpdateNodeStatus(ctx, &relay.UpdateNodeStatusRequest{
			Item: &nonPublic.NodeStatus{
				Id: nodeStatusID,
				Issues: []*model.CompatibilityIssue{
					{
						Code: "T137",
					},
				},
			},
		})
		require.Nil(t, err)

		// verify the node-status in database
		list := resource.NewList(resource.TypeNodeStatus)
		err = db.List(ctx, list)
		require.NoError(t, err)
		require.Len(t, list.GetAll(), 1)
		item := list.GetAll()[0]
		res, ok := item.Resource().(*nonPublic.NodeStatus)
		require.True(t, ok)
		require.Equal(t, nodeStatusID, res.Id)
		require.Equal(t, "T137", res.Issues[0].Code)
	})
	t.Run("request without an id errors", func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		_, err := client.UpdateNodeStatus(ctx, &relay.UpdateNodeStatusRequest{
			Item: &nonPublic.NodeStatus{
				Issues: []*model.CompatibilityIssue{
					{
						Code: "T420",
					},
				},
			},
		})
		require.Error(t, err)
		statusErr, ok := statusPb.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.InvalidArgument, statusErr.Code())
		require.Equal(t, "node-status ID is required", statusErr.Message())
	})
	t.Run("update with an invalid node-status errors", func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		_, err := client.UpdateNodeStatus(ctx, &relay.UpdateNodeStatusRequest{
			Item: &nonPublic.NodeStatus{
				Id: uuid.NewString(),
				Issues: []*model.CompatibilityIssue{
					{
						Code: "T#420",
					},
				},
			},
		})
		require.Error(t, err)
		statusErr, ok := statusPb.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.InvalidArgument, statusErr.Code())
		require.Equal(t, "validation error", statusErr.Message())
	})
}
