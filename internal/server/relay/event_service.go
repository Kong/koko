package relay

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	relay "github.com/kong/koko/internal/gen/grpc/kong/relay/service/v1"
	"github.com/kong/koko/internal/store"
	storeEvent "github.com/kong/koko/internal/store/event"
	"go.uber.org/zap"
	"google.golang.org/grpc/peer"
)

type EventService struct {
	relay.UnimplementedEventServiceServer

	store  store.Store
	logger *zap.Logger

	clients sync.Map
}

type EventServiceOpts struct {
	Store  store.Store
	Logger *zap.Logger
}

func NewEventService(ctx context.Context, opts EventServiceOpts) relay.EventServiceServer {
	res := &EventService{
		store:  opts.Store,
		logger: opts.Logger,
	}
	go res.run(ctx)
	return res
}

type client struct {
	done       chan struct{}
	stream     relay.EventService_FetchReconfigureEventsServer
	seenID     string
	remoteAddr string
}

func (e *EventService) FetchReconfigureEvents(req *relay.FetchReconfigureEventsRequest,
	stream relay.EventService_FetchReconfigureEventsServer) error {
	if req.Cluster == nil || req.Cluster.Id == "" {
		return fmt.Errorf("no cluster")
	}

	peer, ok := peer.FromContext(stream.Context())
	if !ok {
		panic("failed to find peer in context")
	}

	e.logger.With(
		zap.String("cluster", req.Cluster.Id),
		zap.String("peer", peer.Addr.String()),
	).Debug("received request for fetching events")

	done := make(chan struct{})
	streamID := uuid.NewString()
	e.clients.Store(streamID, &client{
		done:       done,
		stream:     stream,
		remoteAddr: peer.Addr.String(),
	})
	select {
	case <-done:
	// server encountered an error and shut down this stream
	case <-stream.Context().Done():
		// stream was likely closed by the client so
		// remove from tracking list
	}
	e.clients.Delete(streamID)
	return nil
}

func (e *EventService) run(ctx context.Context) {
	var latestID string
	for {
		// return only when ctx.Done(), otherwise log errors and keep running
		eventID, err := e.lastEvent(ctx)
		if err != nil {
			e.logger.With(zap.Error(err)).Error("fetch event")
		} else {
			latestID = eventID
		}
		// update clients unconditionally since there could be new clients
		// that need to be sent old updates
		e.updateClients(latestID)
		select {
		case <-ctx.Done():
			e.logger.Info("shutting down due to context cancellation")
			return
		case <-time.After(1 * time.Second):
		}
	}
}

func (e *EventService) lastEvent(ctx context.Context) (string, error) {
	event := storeEvent.New()
	err := e.store.Read(ctx, event, store.GetByID(storeEvent.ID))
	if err != nil {
		if err == store.ErrNotFound {
			return "", nil
		}
		return "", err
	}
	return event.StoreEvent.Value, nil
}

func (e *EventService) updateClients(eventID string) {
	e.clients.Range(func(_, value interface{}) bool {
		node, ok := value.(*client)
		if !ok {
			panic(fmt.Sprintf("unexpected type: %T, expected %T",
				value, client{}))
		}
		clientLogger := e.logger.With(zap.String("client", node.remoteAddr))
		if node.seenID == eventID {
			clientLogger.Debug("skipping re-configure as seenID is up-to-date")
			return true
		}
		clientLogger.Debug("reconfigure event sent")
		// TODO(hbagdi): can this block indefinitely?
		err := node.stream.Send(&relay.FetchReconfigureEventsResponse{})
		if err != nil {
			clientLogger.With(zap.Error(err)).Error("send re-configure event")
			close(node.done)
		} else {
			node.seenID = eventID
		}
		return true
	})
}
