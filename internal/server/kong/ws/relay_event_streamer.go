package ws

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	relay "github.com/kong/koko/internal/gen/grpc/kong/relay/service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	grpcStatus "google.golang.org/grpc/status"
)

// RelayEventStreamerOpts represents the options to create a new instance of a RelayStreamEvent
// object.
type RelayEventStreamerOpts struct {
	// EventServiceClient from which to fetch the events.
	EventServiceClient relay.EventServiceClient
	// Logger is the defined logger for the relay event streamer.
	Logger *zap.Logger
}

// RelayEventStreamer represents the handling and notification for streamed events with the
// relay service.
type RelayEventStreamer struct {
	// eventClient from which to fetch the service events.
	eventClient relay.EventServiceClient
	// logger is used for appending log message for relayEventStreamer.
	logger *zap.Logger

	// mutex protects the critical sections of register and unregister.
	mutex sync.Mutex
	// streamCancel allows for the relay EventServiceClient stream to be canceled (e.g. disabled).
	streamCancel context.CancelFunc
}

// NewRelayEventStreamer will create a new relay event streamer for communication of relay events.
func NewRelayEventStreamer(opts RelayEventStreamerOpts) (*RelayEventStreamer, error) {
	if opts.EventServiceClient == nil {
		return nil, errors.New("event service client connection is required")
	}
	if opts.Logger == nil {
		return nil, errors.New("logger is required")
	}

	return &RelayEventStreamer{
		eventClient: opts.EventServiceClient,
		logger:      opts.Logger,
	}, nil
}

// Name is the name of the relay event streamer.
func (r *RelayEventStreamer) Name() string {
	return "relay-event-streamer"
}

// Register enables the relay event streamer and registers the handler.
// Calling enable when the stream is already registered is a no-op.
func (r *RelayEventStreamer) Register(ctx context.Context, cluster Cluster, handler EventHandler) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.streamCancel != nil {
		// no-op for already registered streamer
		return nil
	}

	// Handler must be supplied in order to register
	if handler == nil {
		return errors.New("handler is required")
	}

	r.logger.Info("relay event stream enabled")
	streamCtx, streamCancel := context.WithCancel(ctx)
	r.streamCancel = streamCancel
	go func() {
		for {
			if err := streamCtx.Err(); err != nil {
				if errors.Is(err, context.Canceled) {
					r.logger.Info("shutting down relay event stream due to context cancellation")
					return
				}
				r.logger.Sugar().Errorf("shutting down relay event stream: %v", err)
				return
			}
			stream, err := r.setupStream(streamCtx, cluster.Get())
			if err != nil {
				r.logger.With(zap.Error(err)).Error("relay event stream setup failure")
				continue
			}
			r.streamUpdateEvents(streamCtx, stream, handler)
		}
	}()
	return nil
}

// Unregister disables the relay event streamer and un-registers the event handler.
// Calling unregister when stream is already unregistered is a no-op.
func (r *RelayEventStreamer) Unregister(ctx context.Context, _ Cluster) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.streamCancel == nil {
		// no-op for already unregistered streamer
		return nil
	}

	r.logger.Info("relay event stream disabled")
	r.streamCancel()
	r.streamCancel = nil
	return nil
}

// setupStream creates the stream for fetching reconfigure events from the relay server.
// The stream creation attempt use an exponential backoff that is executed indefinitely
// or until the context is canceled.
func (r *RelayEventStreamer) setupStream(ctx context.Context,
	clusterID string,
) (relay.EventService_FetchReconfigureEventsClient, error) {
	var (
		stream    relay.EventService_FetchReconfigureEventsClient
		backoffer = newBackOff(ctx, 0) // retry forever
		ctxErr    error
	)
	err := backoff.RetryNotify(func() error {
		var err error
		stream, err = r.eventClient.FetchReconfigureEvents(ctx,
			&relay.FetchReconfigureEventsRequest{
				Cluster: &v1.RequestCluster{Id: clusterID},
			})
		if err != nil {
			if ctx.Err() != nil {
				ctxErr = err
				// stop retrying if ctx is cancelled
				return nil
			}
		}
		return err
	}, backoffer, func(err error, duration time.Duration) {
		if err != nil {
			r.logger.With(
				zap.Error(err),
				zap.Duration("retry-in", duration)).
				Error("failed to setup a stream with relay server, retrying")
		}
	})
	if err != nil {
		r.logger.Error("failed to setup stream with relay server", zap.Error(err))
	}
	return stream, ctxErr
}

// streamUpdateEvents listens for reconfigure events from the relay server and notifies
// the handler of events.
func (r *RelayEventStreamer) streamUpdateEvents(ctx context.Context, stream relay.
	EventService_FetchReconfigureEventsClient, handler EventHandler,
) {
	r.logger.Debug("start read from relay event stream")
	for {
		updateEvent, err := stream.Recv()
		if err != nil {
			if err == io.EOF || grpcStatus.Code(err) == codes.Canceled {
				r.logger.Info("relay event stream closed")
				return
			}
			r.logger.With(zap.Error(err)).Error("receive event")
			// return on any error, caller will re-establish a stream if needed
			return
		}
		if updateEvent != nil {
			r.logger.Info("reconfigure event received")
			eventCtx, eventCancel := context.WithTimeout(ctx, defaultRequestTimeout)
			defer eventCancel()
			event := Event{
				EventType: ReconfigureEvent,
			}
			if err := handler.OnEvent(eventCtx, event); err != nil {
				r.logger.Error("error handling reconfigure event",
					zap.Error(err))
			}
		}
	}
}
