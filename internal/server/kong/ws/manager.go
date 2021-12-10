package ws

import (
	"context"
	"io"
	"net"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/gorilla/websocket"
	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	admin "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
	relay "github.com/kong/koko/internal/gen/grpc/kong/relay/service/v1"
	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/server/kong/ws/config"
	"github.com/kong/koko/internal/server/kong/ws/mold"
	"go.uber.org/zap"
)

type ManagerOpts struct {
	Client  ConfigClient
	Cluster Cluster
	Logger  *zap.Logger
}

type Cluster interface {
	Get() string
}

type DefaultCluster struct{}

func (d DefaultCluster) Get() string {
	return "default"
}

func NewManager(opts ManagerOpts) *Manager {
	return &Manager{
		cluster:      opts.Cluster,
		configClient: opts.Client,
		logger:       opts.Logger,
		payload:      &config.Payload{},
		nodes:        &NodeList{},
	}
}

type Manager struct {
	configClient ConfigClient
	cluster      Cluster
	logger       *zap.Logger

	payload *config.Payload
	nodes   *NodeList

	broadcastMutex sync.Mutex
}

func (m *Manager) updateNodeStatus(node Node) {
	const timeout = 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	_, err := m.configClient.Node.UpsertNode(ctx, &admin.UpsertNodeRequest{
		Item: &model.Node{
			Id:       node.ID,
			Version:  node.Version,
			Hostname: node.Hostname,
			Type:     resource.NodeTypeKongProxy,
			LastPing: int32(time.Now().Unix()),
		},
		Cluster: &model.RequestCluster{
			Id: m.cluster.Get(),
		},
	})
	if err != nil {
		m.logger.With(zap.Error(err)).Error("update kong node status")
	}
}

func (m *Manager) setupPingHandler(node Node) {
	c := node.conn
	c.SetPingHandler(func(appData string) error {
		// code inspired from the upstream library
		writeWait := time.Second
		err := c.WriteControl(websocket.PongMessage, nil,
			time.Now().Add(writeWait))
		if err == websocket.ErrCloseSent {
			return nil
		} else if e, ok := err.(net.Error); ok && e.Temporary() {
			return nil
		}
		m.updateNodeStatus(node)

		return err
	})
}

func (m *Manager) AddNode(node Node) {
	loggerWithNode := m.logger.With(zap.String("client-ip",
		node.conn.RemoteAddr().String()))
	m.setupPingHandler(node)
	if err := m.nodes.Add(node); err != nil {
		m.logger.With(zap.Error(err)).Error("track node")
	}
	// spawn a goroutine for each data-plane node that connects.
	go func() {
		err := node.readThread()
		if err != nil {
			loggerWithNode.With(zap.Error(err)).
				Error("read thread")
		}
		// if there are any ws errors, remove the node
		// TODO(hbagdi): may need more graceful error handling
		err = m.nodes.Remove(node)
		if err != nil {
			loggerWithNode.With(zap.Error(err)).
				Error("remove node")
		}
	}()
	err := m.reconcilePayload(context.Background())
	if err != nil {
		m.logger.With(zap.Error(err)).
			Error("reconcile configuration")
	}
	m.broadcast()
}

// broadcast sends the most recent configuration to all connected nodes.
func (m *Manager) broadcast() {
	m.broadcastMutex.Lock()
	defer m.broadcastMutex.Unlock()
	payload := m.payload.Payload()
	for _, node := range m.nodes.All() {
		loggerWithNode := m.logger.With(zap.String("client-ip",
			node.conn.RemoteAddr().String()))
		loggerWithNode.Debug("broadcasting to node")
		// TODO(hbagdi): perf: use websocket.PreparedMessage
		err := node.write(payload)
		if err != nil {
			m.logger.With(zap.Error(err)).Error("broadcast failed")
			// TODO(hbagdi: remove the node if connection has been closed?
		}
	}
}

type ConfigClient struct {
	Service admin.ServiceServiceClient
	Route   admin.RouteServiceClient
	Plugin  admin.PluginServiceClient

	Node admin.NodeServiceClient

	Event relay.EventServiceClient
}

func (m *Manager) reconcilePayload(ctx context.Context) error {
	grpcContent, err := m.fetchContent(ctx)
	if err != nil {
		return err
	}
	wrpcContent, err := mold.GrpcToWrpc(grpcContent)
	if err != nil {
		return err
	}
	return m.payload.Update(wrpcContent)
}

var defaultRequestTimeout = 5 * time.Second

func (m *Manager) fetchContent(ctx context.Context) (mold.GrpcContent, error) {
	var err error

	serviceList, err := m.fetchServices(ctx)
	if err != nil {
		return mold.GrpcContent{}, err
	}

	routesList, err := m.fetchRoutes(ctx)
	if err != nil {
		return mold.GrpcContent{}, err
	}

	pluginsList, err := m.fetchPlugins(ctx)
	if err != nil {
		return mold.GrpcContent{}, err
	}
	return mold.GrpcContent{
		Services: serviceList.Items,
		Routes:   routesList.Items,
		Plugins:  pluginsList.Items,
	}, nil
}

func (m *Manager) fetchServices(ctx context.Context) (*admin.
	ListServicesResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()
	return m.configClient.Service.ListServices(ctx,
		&admin.ListServicesRequest{
			Cluster: &model.RequestCluster{
				Id: m.cluster.Get(),
			},
		})
}

func (m *Manager) fetchRoutes(ctx context.Context) (*admin.
	ListRoutesResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()
	return m.configClient.Route.ListRoutes(ctx,
		&admin.ListRoutesRequest{
			Cluster: &model.RequestCluster{
				Id: m.cluster.Get(),
			},
		})
}

func (m *Manager) fetchPlugins(ctx context.Context) (*admin.
	ListPluginsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()
	return m.configClient.Plugin.ListPlugins(ctx,
		&admin.ListPluginsRequest{
			Cluster: &model.RequestCluster{
				Id: m.cluster.Get(),
			},
		})
}

func (m *Manager) Run(ctx context.Context) {
	for {
		stream, err := m.setupStream(ctx)
		if err != nil {
			m.logger.With(zap.Error(err)).Error("event stream setup failure")
			return
		}
		m.streamUpdateEvents(ctx, stream)
		if err := ctx.Err(); err != nil {
			m.logger.Sugar().Errorf("shutting down manager: %v", err)
			return
		}
	}
}

func (m *Manager) setupStream(ctx context.Context) (relay.
	EventService_FetchReconfigureEventsClient, error) {
	var stream relay.EventService_FetchReconfigureEventsClient

	backoffer := newBackOff(ctx, 0) // retry forever
	err := backoff.RetryNotify(func() error {
		var err error
		stream, err = m.configClient.Event.FetchReconfigureEvents(ctx,
			&relay.FetchReconfigureEventsRequest{
				Cluster: &model.RequestCluster{
					Id: m.cluster.Get(),
				},
			})
		// triggers backoff if err != nil
		return err
	}, backoffer, func(err error, duration time.Duration) {
		if err != nil {
			m.logger.With(
				zap.Error(err),
				zap.Duration("retry-in", duration)).
				Error("failed to setup a stream with relay server, retrying")
		}
	})
	return stream, err
}

func (m *Manager) streamUpdateEvents(ctx context.Context, stream relay.
	EventService_FetchReconfigureEventsClient) {
	m.logger.Debug("start read from event stream")
	for {
		up, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				m.logger.Info("event stream closed")
			} else {
				m.logger.With(zap.Error(err)).Error("receive event")
			}
			// return on any error, caller will re-establish a stream if needed
			return
		}
		// TODO(hbagdi): make this concurrent, events can pile up and thrash
		// caches unnecessarily
		if up != nil {
			m.logger.Debug("reconfigure event received")
			// TODO(hbagdi): add a rate-limiter to de-duplicate events in case
			// of a short write burst
			m.logger.Debug("reconcile payload")
			backoffer := newBackOff(ctx, 1*time.Minute) // retry for a minute
			err := backoff.RetryNotify(func() error {
				return m.reconcilePayload(ctx)
			}, backoffer, func(err error, duration time.Duration) {
				m.logger.With(
					zap.Error(err),
					zap.Duration("retry-in", duration)).
					Error("configuration reconciliation failed, retrying")
			})
			if err != nil {
				m.logger.With(
					zap.Error(err)).
					Error("failed to reconcile configuration")
				// skip broadcasting if configuration could not be updated
				continue
			}
			m.logger.Debug("broadcast configuration to all nodes")
			go m.broadcast()
		}
	}
}

const (
	defaultInitialInterval     = 500 * time.Millisecond
	defaultRandomizationFactor = 0.3
	defaultMultiplier          = 1.5
	defaultMaxInterval         = 60 * time.Second
)

func newBackOff(ctx context.Context, limit time.Duration) backoff.BackOff {
	var backoffer backoff.BackOff
	backoffer = &backoff.ExponentialBackOff{
		InitialInterval:     defaultInitialInterval,
		RandomizationFactor: defaultRandomizationFactor,
		Multiplier:          defaultMultiplier,
		MaxInterval:         defaultMaxInterval,
		MaxElapsedTime:      limit,
		Stop:                backoff.Stop,
		Clock:               backoff.SystemClock,
	}
	backoffer = backoff.WithContext(backoffer, ctx)
	backoffer.Reset()
	return backoffer
}
