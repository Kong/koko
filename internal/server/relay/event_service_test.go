package relay

import (
	"context"
	"net"
	"testing"

	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	relay "github.com/kong/koko/internal/gen/grpc/kong/relay/service/v1"
	"github.com/kong/koko/internal/log"
	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/server/admin"
	"github.com/kong/koko/internal/store"
	"github.com/kong/koko/internal/test/util"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func TestEventService(t *testing.T) {
	persister, err := util.GetPersister(t)
	require.Nil(t, err)
	store := store.New(persister, log.Logger).ForCluster("default")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	opts := EventServiceOpts{
		Store:  store,
		Logger: log.Logger,
	}
	server := NewEventService(ctx, opts)
	require.NotNil(t, server)
	l := setup()
	s := grpc.NewServer(grpc.ChainUnaryInterceptor(admin.LoggerInterceptor(log.Logger)))
	relay.RegisterEventServiceServer(s, server)
	cc := clientConn(t, l)
	client := relay.NewEventServiceClient(cc)
	go func() {
		_ = s.Serve(l)
	}()
	defer s.Stop()

	t.Run("errors without a cluster", func(t *testing.T) {
		stream, err := client.FetchReconfigureEvents(ctx,
			&relay.FetchReconfigureEventsRequest{})
		require.Nil(t, err)
		event, err := stream.Recv()
		require.NotNil(t, err)
		s, _ := status.FromError(err)
		require.Equal(t, "no cluster", s.Message())
		require.Nil(t, event)
	})
	t.Run("receives an event when there is activity in the cluster",
		func(t *testing.T) {
			ctx, cancel = context.WithCancel(ctx)
			defer cancel()
			stream, err := client.FetchReconfigureEvents(ctx,
				&relay.FetchReconfigureEventsRequest{
					Cluster: &model.RequestCluster{Id: "default"},
				})
			require.Nil(t, err)
			res := resource.NewService()
			res.Service.Host = "example.com"
			res.Service.Path = "/"
			err = store.Create(ctx, res)
			require.Nil(t, err)

			event, err := stream.Recv()
			require.Nil(t, err)
			require.NotNil(t, event)
		})
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
