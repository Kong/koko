package ws

import "context"

// EventType represents the type of event.
type EventType int

const (
	// ReconfigureEvent is a event for indicating a payload reconfiguration should occur.
	ReconfigureEvent EventType = iota
)

// Event represents the event taking place.
type Event struct {
	// EventType associated with the event.
	EventType
}

// EventStream interface represents the mechanism responsible for detecting and triggering
// events for a cluster.
type EventStream interface {
	// Name of the event stream.
	Name() string
	// Register will register and enable the event stream.
	// Registering a registered stream is a no-op.
	Register(ctx context.Context, cluster Cluster, handler EventHandler) error
	// Unregister will unregister and disable the event stream.
	// Un-registering an unregistered stream is a no-op.
	Unregister(ctx context.Context, cluster Cluster) error
}

// EventHandler interface represents an event.
type EventHandler interface {
	// OnEvent executes when an event occurs from an EventStream.
	OnEvent(ctx context.Context, e Event) error
}

// String will return the display name of the event enumeration.
func (e Event) String() string {
	switch e.EventType {
	case ReconfigureEvent:
		return "reconfigure event"
	default:
		return "unknown"
	}
}
