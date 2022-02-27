package ws

import (
	"bytes"
	"context"
	encodingJSON "encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/gorilla/websocket"
	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	admin "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
	relay "github.com/kong/koko/internal/gen/grpc/kong/relay/service/v1"
	grpcKongUtil "github.com/kong/koko/internal/gen/grpc/kong/util/v1"
	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/server/kong/ws/config"
	"go.uber.org/zap"
)

type ManagerOpts struct {
	Client  ConfigClient
	Cluster Cluster
	Logger  *zap.Logger
	Config  ManagerConfig

	DPConfigLoader config.Loader
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
		Cluster:      opts.Cluster,
		configClient: opts.Client,
		logger:       opts.Logger,
		configLoader: opts.DPConfigLoader,
		payload:      &config.Payload{},
		nodes:        &NodeList{},
		config:       opts.Config,
	}
}

type ManagerConfig struct {
	DataPlaneRequisites []*grpcKongUtil.DataPlanePrerequisite
}

type Manager struct {
	configClient ConfigClient
	Cluster      Cluster
	logger       *zap.Logger

	payload *config.Payload
	nodes   *NodeList

	broadcastMutex sync.Mutex

	configLoader config.Loader

	config   ManagerConfig
	configMu sync.RWMutex
}

func (m *Manager) reqCluster() *model.RequestCluster {
	return &model.RequestCluster{Id: m.Cluster.Get()}
}

func (m *Manager) UpdateConfig(c ManagerConfig) {
	m.configMu.Lock()
	defer m.configMu.Unlock()
	m.config = c
}

func (m *Manager) ReadConfig() ManagerConfig {
	m.configMu.RLock()
	defer m.configMu.RUnlock()
	return m.config
}

func (m *Manager) updateNodeStatus(node Node) {
	m.writeNode(node)
	// TODO(hbagdi): Make this robust
	// assuming happy state once a valid ping is received
	// instead compare received hash with expected hash and update status
	// accordingly
	ctx, cancel := context.WithTimeout(context.Background(), defaultRequestTimeout)
	defer cancel()
	_, err := m.configClient.Status.ClearStatus(ctx, &relay.ClearStatusRequest{
		ContextReference: &model.EntityReference{
			Id:   node.ID,
			Type: string(resource.TypeNode),
		},
		Cluster: m.reqCluster(),
	})
	if err != nil {
		m.logger.Error("failed to clear status", zap.Error(err),
			zap.String("node-id", node.ID))
	}
}

var emptySum sum

func (m *Manager) writeNode(node Node) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultRequestTimeout)
	defer cancel()

	nodeToUpsert := &model.Node{
		Id:       node.ID,
		Version:  node.Version,
		Hostname: node.Hostname,
		Type:     resource.NodeTypeKongProxy,
		LastPing: int32(time.Now().Unix()),
	}
	if !bytes.Equal(node.hash[:], emptySum[:]) {
		nodeToUpsert.ConfigHash = node.hash.String()
	}

	_, err := m.configClient.Node.UpsertNode(ctx,
		&admin.UpsertNodeRequest{
			Item:    nodeToUpsert,
			Cluster: m.reqCluster(),
		})
	if err != nil {
		m.logger.Error("update kong node resource", zap.Error(err),
			zap.String("node-id", node.ID))
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
		m.logger.Debug("pingHandler received hash", zap.String("hash", appData))
		node.hash, err = truncateHash(appData)
		if err != nil {
			// Logging for now
			m.logger.With(zap.Error(err), zap.String("appData", appData)).
				Error("ping handler: received invalid hash from kong data-plane")
		}
		m.updateNodeStatus(node)

		return err
	})
}

func (m *Manager) AddNode(node Node) {
	loggerWithNode := m.logger.With(zap.String("client-ip",
		node.conn.RemoteAddr().String()))
	// track each authenticated node
	m.writeNode(node)
	// check if node is compatible
	err := m.validateNode(node)
	if err != nil {
		// node has failed to meet the pre-requisites, close the connection
		// TODO(hbagdi): send an error to Kong DP once wRPC is supported
		m.logger.With(
			zap.Error(err),
			zap.String("node-id", node.ID),
		).Info("kong DP node rejected")
		err := node.conn.Close()
		if err != nil {
			m.logger.With(zap.Error(err)).Error(
				"failed to close websocket connection")
		}
		return
	}
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
	err = m.reconcileKongPayload(context.Background())
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
	Service       admin.ServiceServiceClient
	Route         admin.RouteServiceClient
	Plugin        admin.PluginServiceClient
	Upstream      admin.UpstreamServiceClient
	Target        admin.TargetServiceClient
	Status        relay.StatusServiceClient
	Node          admin.NodeServiceClient
	Consumer      admin.ConsumerServiceClient
	Certificate   admin.CertificateServiceClient
	CACertificate admin.CACertificateServiceClient

	Event relay.EventServiceClient
}

func (m *Manager) reconcileKongPayload(ctx context.Context) error {
	config, err := m.configLoader.Load(ctx, m.Cluster.Get())
	if err != nil {
		return err
	}
	return m.payload.UpdateBinary(config)
}

var defaultRequestTimeout = 5 * time.Second

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
				Cluster: m.reqCluster(),
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
				return m.reconcileKongPayload(ctx)
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

type basicInfoPlugin struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

type basicInfo struct {
	Type    string            `json:"type,omitempty"`
	Plugins []basicInfoPlugin `json:"plugins,omitempty"`
}

type nodeAttributes struct {
	Plugins []string
	Version string
}

func (m *Manager) getPluginList(node Node) ([]string, error) {
	messageType, message, err := node.conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("read websocket message: %v", err)
	}
	if messageType != websocket.BinaryMessage {
		return nil, fmt.Errorf("kong data-plane sent a message of type %v, "+
			"expected %v", messageType, websocket.BinaryMessage)
	}
	var info basicInfo
	err = encodingJSON.Unmarshal(message, &info)
	if err != nil {
		return nil, fmt.Errorf("unmarshal basic-info json message: %v", err)
	}
	var plugins []string
	for _, p := range info.Plugins {
		plugins = append(plugins, p.Name)
	}
	return plugins, nil
}

func (m *Manager) validateNode(node Node) error {
	pluginList, err := m.getPluginList(node)
	if err != nil {
		return err
	}
	mConfig := m.ReadConfig()
	conditions := checkPreReqs(nodeAttributes{
		Plugins: pluginList,
		Version: node.Version,
	}, mConfig.DataPlaneRequisites)
	if len(conditions) == 0 {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultRequestTimeout)
	defer cancel()
	// update status before returning
	_, err = m.configClient.Status.UpdateStatus(ctx,
		&relay.UpdateStatusRequest{
			Item: &model.Status{
				ContextReference: &model.EntityReference{
					Id:   node.ID,
					Type: string(resource.TypeNode),
				},
				Conditions: conditions,
			},
			Cluster: m.reqCluster(),
		})
	if err != nil {
		m.logger.Error("failed to update status of a node", zap.Error(err))
	}
	return fmt.Errorf("node failed to meet pre-requisites")
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
