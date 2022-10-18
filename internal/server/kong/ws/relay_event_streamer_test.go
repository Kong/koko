package ws

import (
	"context"
	"errors"
	"testing"
	"time"

	serviceRelay "github.com/kong/koko/internal/gen/grpc/kong/relay/service/v1"
	"github.com/kong/koko/internal/log"
	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/server/relay"
	serverUtil "github.com/kong/koko/internal/server/util"
	"github.com/kong/koko/internal/store"
	"github.com/kong/koko/internal/test/util"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"google.golang.org/grpc"
)

func TestRelayEventStreamer_NewRelayEventStreamer(t *testing.T) {
	t.Run("ensure client connection is required", func(t *testing.T) {
		streamer, err := NewRelayEventStreamer(RelayEventStreamerOpts{})
		require.Nil(t, streamer)
		require.EqualError(t, err, "event service client connection is required")
	})

	t.Run("ensure logger is required", func(t *testing.T) {
		streamer, err := NewRelayEventStreamer(RelayEventStreamerOpts{
			EventServiceClient: serviceRelay.NewEventServiceClient(&grpc.ClientConn{}),
		})
		require.Nil(t, streamer)
		require.EqualError(t, err, "logger is required")
	})

	t.Run("ensure relay event streamer instance is instantiated", func(t *testing.T) {
		eventClient := serviceRelay.NewEventServiceClient(&grpc.ClientConn{})
		streamer, err := NewRelayEventStreamer(RelayEventStreamerOpts{
			EventServiceClient: eventClient,
			Logger:             log.Logger,
		})
		require.NotNil(t, streamer)
		require.NoError(t, err)
		require.Equal(t, log.Logger, streamer.logger)
		require.NotNil(t, streamer.eventClient)
		require.Nil(t, streamer.streamCancel)
	})
}

type relayEventHandler struct {
	calledCount atomic.Int32
}

func (r *relayEventHandler) OnEvent(ctx context.Context, e Event) error {
	r.calledCount.Add(1)
	return nil
}

func TestRelayEventStreamer_RegisterUnregister(t *testing.T) {
	persister, err := util.GetPersister(t)
	require.Nil(t, err)
	db := store.New(persister, log.Logger).ForCluster(store.DefaultCluster)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	opts := relay.EventServiceOpts{
		Store:  db,
		Logger: log.Logger,
	}
	server := relay.NewEventService(ctx, opts)
	require.NotNil(t, server)
	l := setup()
	grpcServOpts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(serverUtil.LoggerInterceptor(opts.Logger),
			serverUtil.PanicInterceptor(opts.Logger)),
		grpc.ChainStreamInterceptor(serverUtil.PanicStreamInterceptor(opts.Logger)),
	}
	s := grpc.NewServer(grpcServOpts...)
	serviceRelay.RegisterEventServiceServer(s, server)
	cc := clientConn(t, l)
	client := serviceRelay.NewEventServiceClient(cc)
	go func() {
		_ = s.Serve(l)
	}()
	defer s.Stop()

	t.Run("ensure relay event stream can be registered", func(t *testing.T) {
		streamer, err := NewRelayEventStreamer(RelayEventStreamerOpts{
			EventServiceClient: client,
			Logger:             log.Logger,
		})
		require.NoError(t, err)

		handler := &relayEventHandler{}
		err = streamer.Register(context.Background(), DefaultCluster{}, handler)
		require.NoError(t, err)
		require.NotNil(t, streamer.streamCancel)
	})

	t.Run("ensure relay event stream can be re-registered", func(t *testing.T) {
		streamer, err := NewRelayEventStreamer(RelayEventStreamerOpts{
			EventServiceClient: client,
			Logger:             log.Logger,
		})
		require.NoError(t, err)

		handler := &relayEventHandler{}
		err = streamer.Register(context.Background(), DefaultCluster{}, handler)
		require.NoError(t, err)
		err = streamer.Register(context.Background(), DefaultCluster{}, handler)
		require.NoError(t, err)
	})

	t.Run("registering invalid handler fails", func(t *testing.T) {
		streamer, err := NewRelayEventStreamer(RelayEventStreamerOpts{
			EventServiceClient: client,
			Logger:             log.Logger,
		})
		require.NoError(t, err)

		err = streamer.Register(context.Background(), DefaultCluster{}, nil)
		require.EqualError(t, err, "handler is required")
	})

	t.Run("ensure relay event stream can be unregistered", func(t *testing.T) {
		streamer, err := NewRelayEventStreamer(RelayEventStreamerOpts{
			EventServiceClient: client,
			Logger:             log.Logger,
		})
		require.NoError(t, err)

		handler := &relayEventHandler{}
		err = streamer.Register(context.Background(), DefaultCluster{}, handler)
		require.NoError(t, err)
		require.NotNil(t, streamer.streamCancel)
		err = streamer.Unregister(context.Background(), DefaultCluster{})
		require.NoError(t, err)
		require.Nil(t, streamer.streamCancel)
	})

	t.Run("ensure relay event stream can be unregistered when not registered", func(t *testing.T) {
		streamer, err := NewRelayEventStreamer(RelayEventStreamerOpts{
			EventServiceClient: client,
			Logger:             log.Logger,
		})
		require.NoError(t, err)

		err = streamer.Unregister(context.Background(), DefaultCluster{})
		require.NoError(t, err)
		require.Nil(t, streamer.streamCancel)
	})

	t.Run("ensure registered handler callback is called", func(t *testing.T) {
		streamer, err := NewRelayEventStreamer(RelayEventStreamerOpts{
			EventServiceClient: client,
			Logger:             log.Logger,
		})
		require.NoError(t, err)

		handler := &relayEventHandler{}
		err = streamer.Register(context.Background(), DefaultCluster{}, handler)
		require.NoError(t, err)

		// Fire a relay event
		res := resource.NewService()
		res.Service.Host = "example.com"
		res.Service.Path = "/"
		err = db.Create(ctx, res)
		require.NoError(t, err)

		util.WaitFunc(t, func() error {
			if handler.calledCount.Load() != 1 {
				return errors.New("handler has not been called")
			}
			return nil
		})
		require.Equal(t, int32(1), handler.calledCount.Load())
	})

	t.Run("ensure unregistered handler callback is not called", func(t *testing.T) {
		streamer, err := NewRelayEventStreamer(RelayEventStreamerOpts{
			EventServiceClient: client,
			Logger:             log.Logger,
		})
		require.NoError(t, err)

		// Register and unregister handler
		handler := &relayEventHandler{}
		err = streamer.Register(context.Background(), DefaultCluster{}, handler)
		require.NoError(t, err)
		err = streamer.Unregister(context.Background(), DefaultCluster{})
		require.NoError(t, err)

		// Fire a relay event
		res := resource.NewService()
		res.Service.Host = "example.com"
		res.Service.Path = "/"
		err = db.Create(ctx, res)
		require.NoError(t, err)

		// Due to refresh interval for EventService defaulting to 5s and not being configurable
		// currently a sleep is required to ensure that the relay event occurs and does not
		// trigger the relay event handler.
		time.Sleep(6 * time.Second)
		require.Zero(t, handler.calledCount.Load())
	})
}
