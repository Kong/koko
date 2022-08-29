package ws

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/cespare/xxhash/v2"
	"github.com/gorilla/websocket"
	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	admin "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
	nonPublic "github.com/kong/koko/internal/gen/grpc/kong/nonpublic/v1"
	relay "github.com/kong/koko/internal/gen/grpc/kong/relay/service/v1"
	grpcKongUtil "github.com/kong/koko/internal/gen/grpc/kong/util/v1"
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

func NewManager(opts ManagerOpts) (*Manager, error) {
	payload, err := NewPayload(PayloadOpts{
		VersionCompatibilityProcessor: opts.DPVersionCompatibility,
		Logger:                        opts.Logger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create new payload: %v", err)
	}

	m := &Manager{
		ctx:          opts.Ctx,
		Cluster:      opts.Cluster,
		configClient: opts.Client,
		logger:       opts.Logger,
		configLoader: opts.DPConfigLoader,
		payload:      payload,
		nodes:        &NodeList{},
		pendingNodes: &NodeList{},
		config:       opts.Config,
	}
	m.streamer = &streamer{
		Logger: opts.Logger.With(zap.String("component", "manager-streamer")),
		OnRecvFunc: func(_ context.Context) {
			m.updateEventCount.Add(1)
		},
		Cluster: opts.Cluster,
		Ctx:     opts.Ctx,
	}
	return m, nil
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

	pendingNodes *NodeList

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

	nodeStatusCache sync.Map
}

// AddEventStream registers an EventStream for streaming events to the streamer.
func (m *Manager) AddEventStream(eventStream EventStream) error {
	return m.streamer.addStream(eventStream)
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

	trackedChanges, err := m.payload.ChangesFor(node.hash.String(), node.Version)
	if err != nil {
		m.logger.Error("not found changes for key ", zap.String("hash",
			node.hash.String()))
	}

	if !m.nodeStatusTracked(node.ID, trackedChanges) {
		issues := trackedChangesToCompatIssues(trackedChanges)

		_, err = m.configClient.Status.UpdateNodeStatus(ctx, &relay.UpdateNodeStatusRequest{
			Item: &nonPublic.NodeStatus{
				Id:     node.ID,
				Issues: issues,
			},
			Cluster: m.reqCluster(),
		})
		if err != nil {
			m.logger.Error("update kong node status resource", zap.Error(err),
				zap.String("node-id", node.ID))
		}
	}
}

func hashTrackedChanges(changes config.TrackedChanges) (string, error) {
	h := xxhash.New()
	e := gob.NewEncoder(h)
	if err := e.Encode(changes); err != nil {
		return "", err
	}
	return string(h.Sum(nil)), nil
}

func (m *Manager) nodeStatusTracked(nodeID string, changes config.TrackedChanges) bool {
	newHash, err := hashTrackedChanges(changes)
	if err != nil {
		m.logger.Error("failed to hash tracked changes", zap.Error(err))
		// always assume the node status is not tracked
		return false
	}

	previousHash := ""
	value, loaded := m.nodeStatusCache.Load(nodeID)
	if loaded {
		var ok bool
		previousHash, ok = value.(string)
		if !ok {
			panic(fmt.Sprintf("expected %T but got %T", "", value))
		}
	}
	if previousHash == newHash {
		return true
	}
	m.nodeStatusCache.Store(nodeID, newHash)
	return false
}

func trackedChangesToCompatIssues(trackedChanges config.TrackedChanges) []*model.CompatibilityIssue {
	issues := make([]*model.CompatibilityIssue, len(trackedChanges.ChangeDetails))
	for i := 0; i < len(trackedChanges.ChangeDetails); i++ {
		change := trackedChanges.ChangeDetails[i]

		var affectedResources []*model.Resource
		if len(change.Resources) > 0 {
			affectedResources = make([]*model.Resource, len(change.Resources))
			for i, affectedResource := range change.Resources {
				affectedResources[i] = &model.Resource{
					Id:   affectedResource.ID,
					Type: affectedResource.Type,
				}
			}
		}

		issues[i] = &model.CompatibilityIssue{
			Code:              string(change.ID),
			AffectedResources: affectedResources,
		}
	}
	return issues
}

func (m *Manager) setupPingHandler(node *Node) {
	c := node.conn
	c.SetPingHandler(func(appData string) error {
		// code inspired from the upstream library
		metrics.Count("data_plane_ping_total", 1, metrics.Tag{
			Key:   "dp_version",
			Value: node.Version,
		},
			metrics.Tag{
				Key:   "protocol",
				Value: "ws",
			},
		)
		writeWait := time.Second
		err := c.WriteControl(websocket.PongMessage, nil,
			time.Now().Add(writeWait))
		if err == websocket.ErrCloseSent {
			return nil
		} else if _, ok := err.(net.Error); ok {
			return nil
		}
		node.Logger.Info("websocket ping handler received hash",
			zap.String("config_hash", appData))

		node.lock.Lock()
		node.hash, err = truncateHash(appData)
		node.lock.Unlock()
		if err != nil {
			// Logging for now
			node.Logger.With(zap.Error(err), zap.String("appData", appData)).
				Error("ping handler: received invalid hash from kong data-plane")
		}
		m.updateNodeStatus(node)

		return err
	})
}

func increaseMetricCounter(code int) {
	tags := metrics.Tag{
		Key:   "ws_close_code",
		Value: strconv.Itoa(code),
	}
	metrics.Count("websocket_connection_closed_count", 1, tags)
}

func (m *Manager) startThreads() {
	// setup relay event streamer
	if relayEventStreamer, err := NewRelayEventStreamer(
		RelayEventStreamerOpts{
			EventServiceClient: m.configClient.Event,
			Logger:             m.logger.With(zap.String("component", "relay-event-streamer")),
		},
	); err != nil {
		m.logger.Error("failed to create new relay event streamer", zap.Error(err))
	} else {
		if err := m.streamer.addStream(relayEventStreamer); err != nil {
			m.logger.Error("failed to add relay event streamer", zap.Error(err))
		}
	}

	go m.nodeCleanThread(m.ctx)
	go m.eventHandlerThread(m.ctx)
	// initial load of config data,
	// done synchronously to ensure it's ready
	// for the first push.
	_ = m.reconcileKongPayload(m.ctx)
}

func (m *Manager) AddWebsocketNode(node *Node) {
	m.init.Do(m.startThreads)
	// track each authenticated node
	m.writeNode(node)

	m.setupPingHandler(node)
	m.addToNodeList(node)
	// spawn a goroutine for each data-plane node that connects.
	go func() {
		err := node.readThread()
		if err != nil {
			wsErr, ok := err.(ErrConnClosed)
			if ok {
				increaseMetricCounter(wsErr.Code)
				if wsErr.Code == websocket.CloseAbnormalClosure {
					node.Logger.Info("node disconnected")
				} else {
					node.Logger.With(zap.Error(err)).Error("read thread: connection closed")
				}
			} else {
				node.Logger.With(zap.Error(err)).Error("read thread")
			}
		}
		// if there are any ws errors, remove the node
		m.removeNode(node)
	}()

	go m.broadcast()
}

func (m *Manager) addToNodeList(node *Node) {
	m.nodeTrackingMu.Lock()
	defer m.nodeTrackingMu.Unlock()
	if err := m.nodes.Add(node); err != nil {
		m.logger.Error("failed adding node to manager", zap.Error(err))
	}
	m.streamer.Enable()
}

func (m *Manager) removeNode(node *Node) {
	m.nodeTrackingMu.Lock()
	defer m.nodeTrackingMu.Unlock()
	// TODO(hbagdi): may need more graceful error handling

	nodeAddr := node.RemoteAddr().String()
	if m.nodes.FindNode(nodeAddr) == node {
		if err := m.nodes.Remove(node); err != nil {
			node.Logger.Error("failed to remove node", zap.Error(err))
		}
	}
	if m.pendingNodes.FindNode(nodeAddr) == node {
		if err := m.pendingNodes.Remove(node); err != nil {
			node.Logger.Error("failed to remove pending node", zap.Error(err))
		}
	}

	if err := node.Close(); err != nil {
		node.Logger.Info("error closing node", zap.Error(err))
	}
	if len(m.nodes.All()) == 0 {
		m.logger.Info("no nodes connected, disabling stream")
		m.streamer.Disable()
	}
}

// FindNode returns a pointer to the node given a remote address
// or nil if not found.
func (m *Manager) FindNode(remoteAddress string) *Node {
	return m.nodes.FindNode(remoteAddress)
}

// AddPendingNode registers a Node that hasn't been validated yet.
// In particular, wrpc nodes have to be actively listening in order
// to negotiate the services it will handle.  In the meantime,
// the manager keeps them as "pending"
// A wrpc node isn't directly added to the `m.nodes` list.
// Instead, as soon as their connection is live, they're added to
// the `m.pendingNodes` list until it finishes some initialization.
func (m *Manager) AddPendingNode(node *Node) error {
	if err := m.pendingNodes.Add(node); err != nil {
		return err
	}
	m.writeNode(node)
	return nil
}

// addWRPCNode is called on behalf of a node when it is ready
// to receive config updates. The ReportMetadata RPC method handler
// does this after receiving the list of plugins.
// Here, the node is moved from the `m.pendingNodes` to the
// final `m.nodes` list.
func (m *Manager) addWRPCNode(node *Node) error {
	m.init.Do(m.startThreads)

	err := m.pendingNodes.Remove(node)
	if err != nil {
		return fmt.Errorf("failed to remove node from pending list: %w", err)
	}

	m.addToNodeList(node)
	go m.broadcast()
	return nil
}

const defaultBroadcastTimeout = 1 * time.Minute // should this be configurable?

// broadcast sends the most recent configuration to all connected nodes.
func (m *Manager) broadcast() {
	m.broadcastMutex.Lock()
	defer m.broadcastMutex.Unlock()
	// NOTE: since wRPC syncs are done in goroutines, they continue after
	// this loop has finished and the lock has been released.
	// Avoid overlaps by either cancelling the passed context, or
	// by only releasing the lock after all configs have been acked.
	for _, node := range m.nodes.All() {
		if err := node.sendConfig(m.ctx, m.payload); err != nil {
			node.Logger.Error("failed to send config to node", zap.Error(err))
			// one node failure shouldn't result in no sync activity to all
			// subsequent/nodes even though it is likely that all nodes are
			// of same version
			continue
			// TODO(hbagdi) add a metric
		}
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
		zap.String("config_hash", config.Hash))

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
	m.logger.Info("update expected hash in database", zap.String("config_hash", hash))
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
