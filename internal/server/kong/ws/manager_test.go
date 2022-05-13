package ws

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/google/uuid"
	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
	"github.com/kong/koko/internal/log"
	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/server/admin"
	serverUtil "github.com/kong/koko/internal/server/util"
	"github.com/kong/koko/internal/store"
	"github.com/kong/koko/internal/test/util"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func TestCleanupNodes(t *testing.T) {
	persister, err := util.GetPersister(t)
	require.Nil(t, err)
	store := store.New(persister, log.Logger).ForCluster("default")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := grpc.NewServer()
	admin.RegisterAdminService(s, admin.HandlerOpts{
		Logger:      log.Logger,
		StoreLoader: serverUtil.DefaultStoreLoader{Store: store},
	})

	l := setup()
	go func() {
		_ = s.Serve(l)
	}()
	defer s.Stop()

	cc := clientConn(t, l)
	nodeClient := v1.NewNodeServiceClient(cc)

	m := Manager{
		configClient: ConfigClient{
			Node: nodeClient,
		},
		Cluster: DefaultCluster{},
		logger:  log.Logger,
	}

	oldNodePing := int32(time.Now().Add(-time.Hour * 25).Unix())
	for i := 0; i < 1500; i++ {
		id := uuid.NewString()
		res, err := nodeClient.CreateNode(ctx, &v1.CreateNodeRequest{
			Item: &model.Node{
				Id:         id,
				Version:    "2.8.0",
				Hostname:   fmt.Sprintf("foobar-%d", i),
				LastPing:   oldNodePing,
				Type:       resource.NodeTypeKongProxy,
				CreatedAt:  oldNodePing,
				UpdatedAt:  oldNodePing,
				ConfigHash: "bcd086b1ba3914e70a859db671f75eb9",
			},
		})
		require.Nil(t, err)
		require.NotNil(t, res)
	}

	var expectedLiveNodeIDs []string
	liveNodePing := int32(time.Now().Unix())
	for i := 0; i < 2; i++ {
		id := uuid.NewString()
		expectedLiveNodeIDs = append(expectedLiveNodeIDs, id)
		res, err := nodeClient.CreateNode(ctx, &v1.CreateNodeRequest{
			Item: &model.Node{
				Id:         id,
				Version:    "2.8.0",
				Hostname:   fmt.Sprintf("foobar-%d", i),
				LastPing:   liveNodePing,
				Type:       resource.NodeTypeKongProxy,
				CreatedAt:  oldNodePing,
				UpdatedAt:  liveNodePing,
				ConfigHash: "bcd086b1ba3914e70a859db671f75eb9",
			},
		})
		require.Nil(t, err)
		require.NotNil(t, res)
	}

	err = m.cleanupNodes(ctx)
	require.Nil(t, err)

	res, err := nodeClient.ListNodes(ctx, &v1.ListNodesRequest{
		Cluster: m.reqCluster(),
	})
	require.Nil(t, err)
	require.NotNil(t, res)
	require.Len(t, res.Items, 2)
	var gotLiveNodeIDs []string
	for _, node := range res.Items {
		gotLiveNodeIDs = append(gotLiveNodeIDs, node.Id)
	}
	require.ElementsMatch(t, expectedLiveNodeIDs, gotLiveNodeIDs)
}

func setup() *bufconn.Listener {
	const bufSize = 1024 * 1024
	return bufconn.Listen(bufSize)
}

func clientConn(t *testing.T, l *bufconn.Listener) grpc.ClientConnInterface {
	conn, err := grpc.DialContext(context.Background(),
		"bufnet",
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return l.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.Nil(t, err)
	return conn
}
