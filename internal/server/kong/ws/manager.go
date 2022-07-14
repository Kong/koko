package ws

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/gorilla/websocket"
	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	admin "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
	relay "github.com/kong/koko/internal/gen/grpc/kong/relay/service/v1"
	grpcKongUtil "github.com/kong/koko/internal/gen/grpc/kong/util/v1"
	"github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/metrics"
	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/server/kong/ws/config"
	"github.com/kong/koko/internal/store"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ManagerOpts struct {
	Ctx     context.Context
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
	payload, err := NewPayload(PayloadOpts{
		VersionCompatibilityProcessor: opts.DPVersionCompatibility,
		Logger:                        opts.Logger,
	})
	if err != nil {
		panic(err)
	}

	m := &Manager{
		ctx:          opts.Ctx,
		Cluster:      opts.Cluster,
		configClient: opts.Client,
		logger:       opts.Logger,
		configLoader: opts.DPConfigLoader,
		payload:      payload,
		nodes:        &NodeList{},
		config:       opts.Config,
	}
	m.streamer = &streamer{
		Logger:      opts.Logger.With(zap.String("component", "manager-streamer")),
		EventClient: opts.Client.Event,
		OnRecvFunc: func() {
			m.updateEventCount.Add(1)
		},
		Cluster: opts.Cluster,
		Ctx:     opts.Ctx,
	}
	return m
}

type ManagerConfig struct {
	DataPlaneRequisites []*grpcKongUtil.DataPlanePrerequisite
}

type Manager struct {
	ctx          context.Context
	init         sync.Once
	configClient ConfigClient
	Cluster      Cluster
	logger       *zap.Logger

	payload *Payload
	nodes   *NodeList

	broadcastMutex sync.Mutex

	updateEventCount atomic.Uint32

	configLoader config.Loader

	config   ManagerConfig
	configMu sync.RWMutex

	latestExpectedHash string
	hashMu             sync.RWMutex

	// nodeTrackingMu protects the critical section of node admission and
	// removal.
	nodeTrackingMu sync.Mutex
	streamer       *streamer
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

func increaseMetricCounter(code int) {
	tags := metrics.Tag{
		Key:   "ws_close_code",
		Value: strconv.Itoa(code),
	}
	metrics.Count("websocket_connection_closed_count", 1, tags)
}

func (m *Manager) AddNode(node *Node) {
	m.init.Do(func() {
		go m.nodeCleanThread(m.ctx)
		go m.eventHandlerThread(m.ctx)
		// initial load of config data,
		// done synchronously to ensure it's ready
		// for the first push.
		_ = m.reconcileKongPayload(m.ctx)
	})
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
	m.addNode(node)
	// spawn a goroutine for each data-plane node that connects.
	go func() {
		err := node.readThread()
		if err != nil {
			wsErr, ok := err.(ErrConnClosed)
			if ok {
				increaseMetricCounter(wsErr.Code)
				if wsErr.Code == websocket.CloseAbnormalClosure {
					loggerWithNode.Info("node disconnected")
				} else {
					loggerWithNode.With(zap.Error(err)).Error("read thread: connection closed")
				}
			} else {
				loggerWithNode.With(zap.Error(err)).Error("read thread")
			}
		}
		// if there are any ws errors, remove the node
		m.removeNode(node)
	}()
	go m.broadcast()
}

func (m *Manager) addNode(node *Node) {
	m.nodeTrackingMu.Lock()
	defer m.nodeTrackingMu.Unlock()
	if err := m.nodes.Add(node); err != nil {
		m.logger.With(zap.Error(err)).Error("track node")
	}
	m.streamer.Enable()
}

func (m *Manager) removeNode(node *Node) {
	m.nodeTrackingMu.Lock()
	defer m.nodeTrackingMu.Unlock()
	// TODO(hbagdi): may need more graceful error handling
	if err := m.nodes.Remove(node); err != nil {
		nodeLogger(node, m.logger).With(zap.Error(err)).
			Error("remove node")
	}
	if len(m.nodes.All()) == 0 {
		m.logger.Info("no nodes connected, disabling stream")
		m.streamer.Disable()
	}
}

// broadcast sends the most recent configuration to all connected nodes.
func (m *Manager) broadcast() {
	m.broadcastMutex.Lock()
	defer m.broadcastMutex.Unlock()
	for _, node := range m.nodes.All() {
		payload, err := m.payload.Payload(context.Background(), node.Version)
		if err != nil {
			m.logger.With(zap.Error(err)).Error("unable to gather payload, payload not sent to node")
			// one node failure shouldn't result in no sync activity to all
			// subsequent/nodes even though it is likely that all nodes are
			// of same version
			continue
			// TODO(hbagdi) add a metric
		}
		loggerWithNode := nodeLogger(node, m.logger)
		loggerWithNode.Info("broadcasting to node",
			zap.String("hash", payload.Hash))
		// TODO(hbagdi): perf: use websocket.PreparedMessage
		hash, err := truncateHash(payload.Hash)
		if err != nil {
			m.logger.With(zap.Error(err)).Sugar().Errorf("invalid hash [%v]", hash[:])
			continue
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
	err = m.payload.UpdateBinary(context.Background(), config)
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
		// abort clean up thread if the cluster was deleted for any reason
		if s, ok := status.FromError(err); ok {
			if s.Code() == codes.InvalidArgument &&
				strings.Contains(s.Message(), "cluster not found") {
				m.logger.Info("cluster not found, aborting node clean up thread")
				return
			}
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

// eventHandlerThread 'watches' updateEventCount and reconciles as well as
// refreshes payload. The code ensures that events are coalesced to ensure
// that events are not processed wastefully and also ensures that the events
// are consumed once per second at the most.
func (m *Manager) eventHandlerThread(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			eventCount := m.updateEventCount.Swap(0)
			if eventCount == 0 {
				continue
			}
			if eventCount > 1 {
				m.logger.Info("coalesced update events",
					zap.Uint32("coalesce-count", eventCount))
			}
			m.updateEventHandler(ctx)
		}
	}
}

func (m *Manager) updateEventHandler(ctx context.Context) {
	m.logger.Info("reconciling payload")
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
		// TODO(hbagdi): retry again in some time
		// TODO(hbagdi): add a metric to track total reconciliation failures
		m.logger.With(
			zap.Error(err)).
			Error("failed to reconcile configuration")
		// skip broadcasting if configuration could not be updated
		return
	}
	m.logger.Info("broadcast configuration to all nodes")
	m.broadcast()
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
