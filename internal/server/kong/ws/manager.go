package ws

import (
	"context"
	"sync"
	"time"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
	"github.com/kong/koko/internal/server/kong/ws/config"
	"github.com/kong/koko/internal/server/kong/ws/mold"
	"go.uber.org/zap"
)

type ManagerOpts struct {
	Client ConfigClient
	Logger *zap.Logger
}

func NewManager(opts ManagerOpts) *Manager {
	return &Manager{
		configClient: opts.Client,
		logger:       opts.Logger,
		payload:      &config.Payload{},
		nodes:        &NodeList{},
	}
}

type Manager struct {
	configClient ConfigClient
	logger       *zap.Logger

	payload *config.Payload
	nodes   *NodeList

	broadcastMutex sync.Mutex
}

func (m *Manager) AddNode(node Node) {
	loggerWithNode := m.logger.With(zap.String("client-ip",
		node.conn.RemoteAddr().String()))
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
	Service v1.ServiceServiceClient
	Route   v1.RouteServiceClient
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

var defaultTimeout = 5 * time.Second

func (m *Manager) fetchContent(ctx context.Context) (mold.GrpcContent, error) {
	var err error

	reqCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()
	serviceList, err := m.configClient.Service.ListServices(reqCtx,
		&v1.ListServicesRequest{})
	if err != nil {
		return mold.GrpcContent{}, err
	}

	reqCtx, cancel = context.WithTimeout(ctx, defaultTimeout)
	defer cancel()
	routesList, err := m.configClient.Route.ListRoutes(reqCtx,
		&v1.ListRoutesRequest{})
	if err != nil {
		return mold.GrpcContent{}, err
	}

	return mold.GrpcContent{
		Services: serviceList.Items,
		Routes:   routesList.Items,
	}, nil
}
