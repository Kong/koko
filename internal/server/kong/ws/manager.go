package ws

import (
	"bytes"
	"context"
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
	"github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/server/kong/ws/config"
	"github.com/kong/koko/internal/store"
	"go.uber.org/zap"
)

type ManagerOpts struct {
	Client  ConfigClient
	Cluster Cluster
	Logger  *zap.Logger
	Config  ManagerConfig

	DPConfigLoader         config.Loader
	DPVersionCompatibility config.VersionCompatibility
}

type Cluster interface {
	Get() string
}

type DefaultCluster struct{}

func (d DefaultCluster) Get() string {
	return "default"
}

func NewManager(opts ManagerOpts) *Manager {
	payload, err := config.NewPayload(config.PayloadOpts{
		VersionCompatibilityProcessor: opts.DPVersionCompatibility,
	})
	if err != nil {
		panic(err)
	}

	return &Manager{
		Cluster:      opts.Cluster,
		configClient: opts.Client,
		logger:       opts.Logger,
		configLoader: opts.DPConfigLoader,
		payload:      payload,
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

	latestExpectedHash string
	hashMu             sync.RWMutex
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

func (m *Manager) updateNodeStatus(node *Node) {
	m.writeNode(node)
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

func (m *Manager) writeNode(node *Node) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultRequestTimeout)
	defer cancel()
	node.lock.RLock()
	defer node.lock.RUnlock()

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

func (m *Manager) setupPingHandler(node *Node) {
	c := node.conn
	c.SetPingHandler(func(appData string) error {
		// code inspired from the upstream library
		writeWait := time.Second
		err := c.WriteControl(websocket.PongMessage, nil,
			time.Now().Add(writeWait))
		if err == websocket.ErrCloseSent {
			return nil
		} else if _, ok := err.(net.Error); ok {
			return nil
		}
		loggerWithNode := nodeLogger(node, m.logger)
		loggerWithNode.Info("websocket ping handler received hash",
			zap.String("hash", appData))

		node.lock.Lock()
		node.hash, err = truncateHash(appData)
		node.lock.Unlock()
		if err != nil {
			// Logging for now
			loggerWithNode.With(zap.Error(err), zap.String("appData", appData)).
				Error("ping handler: received invalid hash from kong data-plane")
		}
		m.updateNodeStatus(node)

		return err
	})
}

func nodeLogger(node *Node, logger *zap.Logger) *zap.Logger {
	return logger.With(
		zap.String("node-id", node.ID),
		zap.String("node-hostname", node.Hostname),
		zap.String("node-version", node.Version),
		zap.String("client-ip", node.conn.RemoteAddr().String()))
}

func (m *Manager) AddNode(node *Node) {
	loggerWithNode := nodeLogger(node, m.logger)
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

	// refresh the payload on DP connect
	go func() {
		err = m.reconcileKongPayload(context.Background())
		if err != nil {
			m.logger.With(zap.Error(err)).
				Error("reconcile configuration")
		}
		m.broadcast()
	}()
}

// broadcast sends the most recent configuration to all connected nodes.
func (m *Manager) broadcast() {
	m.broadcastMutex.Lock()
	defer m.broadcastMutex.Unlock()
	for _, node := range m.nodes.All() {
		payload, err := m.payload.Payload(node.Version)
		if err != nil {
			m.logger.With(zap.Error(err)).Error("unable to gather payload")
			return
		}
		loggerWithNode := nodeLogger(node, m.logger)
		loggerWithNode.Info("broadcasting to node",
			zap.String("hash", payload.Hash))
		// TODO(hbagdi): perf: use websocket.PreparedMessage
		hash, err := truncateHash(payload.Hash)
		if err != nil {
			m.logger.With(zap.Error(err)).Sugar().Errorf("invalid hash [%v]", hash)
		}
		err = node.write(payload.CompressedPayload, hash)
		if err != nil {
			loggerWithNode.Error("broadcast to node failed", zap.Error(err))
			// TODO(hbagdi: remove the node if connection has been closed?
		}
		loggerWithNode.Info("successfully sent payload to node")
	}
}

type ConfigClient struct {
	Status relay.StatusServiceClient
	Node   admin.NodeServiceClient
	Event  relay.EventServiceClient
}

func (m *Manager) reconcileKongPayload(ctx context.Context) error {
	config, err := m.configLoader.Load(ctx, m.Cluster.Get())
	if err != nil {
		return err
	}

	m.updateExpectedHash(ctx, config.Hash)
	err = m.payload.UpdateBinary(config)
	if err != nil {
		return err
	}
	m.logger.Info("payload reconciled successfully",
		zap.String("hash", config.Hash))

	return nil
}

func (m *Manager) updateExpectedHash(ctx context.Context, hash string) {
	m.hashMu.Lock()
	defer m.hashMu.Unlock()
	if m.latestExpectedHash == hash {
		return
	}
	// TODO(hbagdi): add retry with backoff, take a new hash during retry into account
	ctx, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()
	m.logger.Info("update expected hash in database", zap.String("hash", hash))
	_, err := m.configClient.Status.UpdateExpectedHash(ctx, &relay.UpdateExpectedHashRequest{
		Cluster: m.reqCluster(),
		Hash:    hash,
	})
	if err != nil {
		m.logger.Error("failed to update expected hash", zap.Error(err))
	}
	m.latestExpectedHash = hash
}

var defaultRequestTimeout = 5 * time.Second

func (m *Manager) Run(ctx context.Context) {
	go m.nodeCleanThread(ctx)
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

func (m *Manager) nodeCleanThread(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.logger.Debug("cleaning up nodes")
			m.cleanupNodesWithRetry(ctx)
		}
	}
}

func (m *Manager) cleanupNodesWithRetry(ctx context.Context) {
	const backoffLimit = 5 * time.Minute
	backoffer := newBackOff(ctx, backoffLimit)
	for {
		err := m.cleanupNodes(ctx)
		if err == nil {
			break
		}
		if ctx.Err() != nil {
			return
		}
		waitDuration := backoffer.NextBackOff()
		if waitDuration == backoff.Stop {
			m.logger.Error("failed to clean up nodes after retries",
				zap.Error(err))
			break
		}
		m.logger.Error("failed to clean up nodes, "+
			"retrying with backoff",
			zap.Error(err), zap.Duration("retry-after", waitDuration))
		time.Sleep(waitDuration)
	}
}

func (m *Manager) cleanupNodes(ctx context.Context) error {
	const cleanupDelay = 24 * time.Hour
	nodes, err := m.listNodes(ctx)
	if err != nil {
		return fmt.Errorf("list nodes: %w", err)
	}
	cutoffTime := int32(time.Now().Add(-cleanupDelay).Unix())
	for _, node := range nodes {
		if node.LastPing < cutoffTime {
			err := m.deleteNode(ctx, node.Id)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Manager) deleteNode(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()
	_, err := m.configClient.Node.DeleteNode(ctx, &admin.DeleteNodeRequest{
		Cluster: m.reqCluster(),
		Id:      id,
	})
	if err != nil {
		return fmt.Errorf("delete node: %v", err)
	}
	return nil
}

func (m *Manager) listNodes(ctx context.Context) ([]*model.Node, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()
	var nodes []*model.Node
	var page int32 = 1
	for {
		resp, err := m.configClient.Node.ListNodes(ctx, &admin.ListNodesRequest{
			Cluster: m.reqCluster(),
			Page: &model.PaginationRequest{
				Number: page,
				Size:   store.MaxPageSize,
			},
		})
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, resp.Items...)
		if resp.Page == nil || resp.Page.NextPageNum == 0 {
			break
		}
		page = resp.Page.NextPageNum
	}
	return nodes, nil
}

func (m *Manager) setupStream(ctx context.Context) (relay.
	EventService_FetchReconfigureEventsClient, error,
) {
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
	EventService_FetchReconfigureEventsClient,
) {
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
			m.logger.Info("reconfigure event received")
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
			m.logger.Info("broadcast configuration to all nodes")
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

func (m *Manager) getPluginList(node *Node) ([]string, error) {
	messageType, message, err := node.conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("read websocket message: %v", err)
	}
	if messageType != websocket.BinaryMessage {
		return nil, fmt.Errorf("kong data-plane sent a message of type %v, "+
			"expected %v", messageType, websocket.BinaryMessage)
	}
	var info basicInfo
	err = json.Unmarshal(message, &info)
	if err != nil {
		return nil, fmt.Errorf("unmarshal basic-info json message: %v", err)
	}

	plugins := make([]string, 0, len(info.Plugins))
	for _, p := range info.Plugins {
		plugins = append(plugins, p.Name)
	}
	return plugins, nil
}

func (m *Manager) validateNode(node *Node) error {
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
