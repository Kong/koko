package ws

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

// streamer streams events and invokes a callback.
// All exported functions are thread-safe unless noted otherwise.
type streamer struct {
	// OnRecvFunc is the callback invoked for each event.
	OnRecvFunc func(ctx context.Context)
	Cluster    Cluster
	// Ctx is used to manage the lifetime of background goroutines that
	// EventStream may invoke.
	Ctx    context.Context
	Logger *zap.Logger

	// mutex protects the read/write access of the eventStreams
	mutex sync.Mutex
	// eventStreams is a list of registered EventStream instances.
	eventStreams []EventStream
}

// Enable enables the stream.
func (s *streamer) Enable() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, eventStream := range s.eventStreams {
		if err := eventStream.Register(s.Ctx, s.Cluster, s); err != nil {
			s.Logger.Error("failed to enable event stream",
				zap.String("event-stream", eventStream.Name()),
				zap.Error(err))
		}
	}
}

// Disable disables the stream.
func (s *streamer) Disable() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, eventStream := range s.eventStreams {
		if err := eventStream.Unregister(s.Ctx, s.Cluster); err != nil {
			s.Logger.Error("failed to disable event stream",
				zap.String("event-stream", eventStream.Name()),
				zap.Error(err))
		}
	}
}

// OnEvent handles the EventStream event.
func (s *streamer) OnEvent(ctx context.Context, e Event) error {
	if e.EventType == ReconfigureEvent {
		s.OnRecvFunc(ctx)
	}
	return nil
}

// addStream registers an EventStream for streaming events to the streamer.
func (s *streamer) addStream(eventStream EventStream) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, e := range s.eventStreams {
		if e.Name() == eventStream.Name() {
			return fmt.Errorf("event stream '%s' already registered", eventStream.Name())
		}
	}
	s.eventStreams = append(s.eventStreams, eventStream)
	return nil
}
