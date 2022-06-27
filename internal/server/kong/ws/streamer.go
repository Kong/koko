package ws

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	relay "github.com/kong/koko/internal/gen/grpc/kong/relay/service/v1"
	"go.uber.org/zap"
)

// streamer streams events and invokes a callback.
// All exported functions are thread-safe unless noted otherwise.
type streamer struct {
	// EventClient from which to fetch the events
	EventClient relay.EventServiceClient
	// OnRecvFunc is the callback invoked for each event.
	OnRecvFunc func()
	Cluster    Cluster
	// Ctx is used to manage the lifetime of background goroutines that
	// streamer  may invoke.
	Ctx    context.Context
	Logger *zap.Logger

	// mu protects the critical section of enablement/disablement.
	mu            sync.Mutex
	currentCancel context.CancelFunc
}

func (s *streamer) setupStream(ctx context.Context) (relay.EventService_FetchReconfigureEventsClient, error) {
	var (
		stream    relay.EventService_FetchReconfigureEventsClient
		backoffer = newBackOff(ctx, 0) // retry forever
		ctxErr    error
	)
	err := backoff.RetryNotify(func() error {
		var err error
		stream, err = s.EventClient.FetchReconfigureEvents(ctx,
			&relay.FetchReconfigureEventsRequest{
				Cluster: &v1.RequestCluster{Id: s.Cluster.Get()},
			})
		if err != nil {
			if ctx.Err() != nil {
				ctxErr = err
				// stop retrying if Ctx is cancelled
				return nil
			}
		}
		return err
	}, backoffer, func(err error, duration time.Duration) {
		if err != nil {
			s.Logger.With(
				zap.Error(err),
				zap.Duration("retry-in", duration)).
				Error("failed to setup a stream with relay server, retrying")
		}
	})
	if err != nil {
		s.Logger.Error("failed to setup stream with relay server", zap.Error(err))
	}
	return stream, ctxErr
}

func (s *streamer) streamUpdateEvents(_ context.Context, stream relay.
	EventService_FetchReconfigureEventsClient,
) {
	s.Logger.Debug("start read from event stream")
	for {
		updateEvent, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				s.Logger.Info("event stream closed")
				return
			}
			s.Logger.With(zap.Error(err)).Error("receive event")
			// return on any error, caller will re-establish a stream if needed
			return
		}
		if updateEvent != nil {
			s.Logger.Info("reconfigure event received")
			s.OnRecvFunc()
		}
	}
}

// Enable enables the stream.
// Calling enable when the stream is already enabled is a no-op.
func (s *streamer) Enable() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.currentCancel != nil {
		s.Logger.Info("stream enablement requested, already enabled")
		return
	}
	s.Logger.Info("stream enabled")

	ctx, cancel := context.WithCancel(s.Ctx)
	s.currentCancel = cancel
	go func() {
		for {
			if err := ctx.Err(); err != nil {
				s.Logger.Sugar().Errorf("shutting down streamer: %v", err)
				return
			}
			stream, err := s.setupStream(ctx)
			if err != nil {
				s.Logger.With(zap.Error(err)).Error("event stream setup failure")
				continue
			}
			s.streamUpdateEvents(ctx, stream)
		}
	}()
}

// Disable disables the stream.
// Calling disable when stream is already disabled is a no-op.
func (s *streamer) Disable() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.currentCancel == nil {
		s.Logger.Info("stream disablement requested, already disabled")
		return
	}

	s.Logger.Info("stream disabled")
	s.currentCancel()
	s.currentCancel = nil
}
