package ws

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"fmt"
	"net"
	"regexp"
	"sync"
	"unsafe"

	"github.com/gorilla/websocket"
	"github.com/kong/go-wrpc/wrpc"
	config_service "github.com/kong/koko/internal/gen/wrpc/kong/services/config/v1"
	"github.com/kong/koko/internal/json"
	metricsv2 "github.com/kong/koko/internal/metrics/v2"
	"github.com/kong/koko/internal/server/kong/ws/config"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type sum [hashSize]byte

const (
	hashRegexPat = `^[a-zA-Z0-9]+$`
	hashSize     = sha256.Size
)

var hashRegex = regexp.MustCompile(hashRegexPat)

var configSize = metricsv2.NewHistogram(
	prometheus.NewRegistry(),
	metricsv2.HistogramOpts{
		Subsystem:  "cp",
		Name:       "data_plane_config_size",
		Help:       "Size (in bytes) of a dataplane config",
		Buckets:    []float64{10, 100, 200, 300, 400, 500, 1000, 2500, 5000, 10000, 15000, 25000, 50000, 100000, 1000000},
		LabelNames: []string{"dp_version"},
	})

func (s sum) String() string {
	return string(s[:])
}

// If the string has more than 32 bytes, an error is reported.
func truncateHash(s32 string) (sum, error) {
	s := sum{}
	nodeHash := []byte(s32)
	matched := hashRegex.Match(nodeHash)
	if !matched || len(nodeHash) > hashSize {
		return s, fmt.Errorf("hash input is invalid")
	}
	for i := 0; i < hashSize; i++ {
		s[i] = nodeHash[i]
	}
	return s, nil
}

type nodeType int

const (
	nodeTypeWebSocket nodeType = iota + 1
	nodeTypeWRPC
)

type Node struct {
	nodetype nodeType
	lock     sync.RWMutex
	ID       string
	Version  string
	Hostname string
	conn     *websocket.Conn
	peer     *wrpc.Peer
	Logger   *zap.Logger
	hash     sum
}

type nodeOpts struct {
	id         string
	version    string
	hostname   string
	connection *websocket.Conn
	peer       *wrpc.Peer
	logger     *zap.Logger
}

type ErrConnClosed struct {
	Code int
}

func (e ErrConnClosed) Error() string {
	return fmt.Sprintf("websocket connection closed (code: %v)", e.Code)
}

// NewNode creates a new Node object with the given options
// after verifying them for consistency.
func NewNode(opts nodeOpts) (*Node, error) {
	if opts.connection == nil && opts.peer == nil {
		return nil, fmt.Errorf("a Node requires either a WebSocket connection or a wRPC peer")
	}
	if opts.connection != nil && opts.peer != nil {
		return nil, fmt.Errorf("a Node can't have both a WebSocket connection and a wRPC peer")
	}

	nodetype := nodeTypeWebSocket
	protocolLabel := "ws"
	if opts.peer != nil {
		nodetype = nodeTypeWRPC
		protocolLabel = "wrpc"
	}

	node := &Node{
		nodetype: nodetype,
		ID:       opts.id,
		Version:  opts.version,
		Hostname: opts.hostname,
		conn:     opts.connection,
		peer:     opts.peer,
		Logger: opts.logger.With(
			zap.String("node-id", opts.id),
			zap.String("node-protocol", protocolLabel),
			zap.String("node-hostname", opts.hostname),
			zap.String("node-version", opts.version)),
	}

	return node, nil
}

// Close ends the Node's lifetime and of its connection.
func (n *Node) Close() error {
	switch n.nodetype {
	case nodeTypeWebSocket:
		return n.conn.Close()
	case nodeTypeWRPC:
		return n.peer.Close()
	}
	return nil
}

// RemoteAddr returns the network address of the client.
func (n *Node) RemoteAddr() net.Addr {
	switch n.nodetype {
	case nodeTypeWebSocket:
		return n.conn.RemoteAddr()
	case nodeTypeWRPC:
		return n.peer.RemoteAddr()
	}
	return &net.IPAddr{}
}

// GetPluginList receives the list of plugins the DP sends
// right after connection on the old WebSocket protocol.
func (n *Node) GetPluginList() ([]string, error) {
	if n.nodetype != nodeTypeWebSocket {
		return nil, fmt.Errorf("not implemented")
	}

	messageType, message, err := n.conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("read websocket message: %w", err)
	}
	if messageType != websocket.BinaryMessage {
		return nil, fmt.Errorf("kong data-plane sent a message of type %v, "+
			"expected %v", messageType, websocket.BinaryMessage)
	}
	var info basicInfo
	err = json.Unmarshal(message, &info)
	if err != nil {
		return nil, fmt.Errorf("unmarshal basic-info json message: %w", err)
	}
	plugins := make([]string, len(info.Plugins))
	for i, p := range info.Plugins {
		plugins[i] = p.Name
	}
	return plugins, nil
}

// readThread continuously reads messages from connected DP node.
func (n *Node) readThread() error {
	if n.nodetype != nodeTypeWebSocket {
		return fmt.Errorf("readThread is only for plain WebSocket nodes")
	}

	for {
		_, message, err := n.conn.ReadMessage()
		if err != nil {
			if wsCloseErr, ok := err.(*websocket.CloseError); ok {
				return ErrConnClosed{Code: wsCloseErr.Code}
			}
			return err
		}
		n.Logger.Info("received message from DP")
		n.Logger.Sugar().Debugf("recv: %s", message)
	}
}

func (n *Node) sendConfig(ctx context.Context, payload *Payload) error {
	content, err := n.getPayload(ctx, payload)
	configSize.Observe(float64(unsafe.Sizeof(content)), metricsv2.Label{
		Key:   "dp_version",
		Value: n.Version,
	})
	if err != nil {
		return fmt.Errorf("unable to gather payload: %w", err)
	}
	n.Logger.Info("broadcasting to node",
		zap.String("config_hash", content.Hash))
	// TODO(hbagdi): perf: use websocket.PreparedMessage
	hash, err := truncateHash(content.Hash)
	if err != nil {
		n.Logger.Error("invalid hash", zap.Error(err), zap.String("config_hash", hash.String()))
		return err
	}

	n.lock.RLock()
	defer n.lock.RUnlock()

	if bytes.Equal(n.hash[:], hash[:]) {
		n.Logger.With(zap.String("config_hash",
			hash.String())).Info("hash matched, skipping update")
		return nil
	}

	switch n.nodetype {
	case nodeTypeWebSocket:
		return n.writeConfig(ctx, content)

	case nodeTypeWRPC:
		return n.writeWRPCConfig(ctx, content, payload.configVersion)
	}

	return nil
}

func (n *Node) getPayload(ctx context.Context, payload *Payload) (config.Content, error) {
	if n.nodetype == nodeTypeWebSocket {
		return payload.Payload(ctx, n.Version)
	}

	key := n.Version + "-wrpc"
	entry, err := payload.configCache.load(key)
	if err == errNotFound {
		content, err := payload.Payload(ctx, n.Version)
		if err != nil {
			return content, err
		}

		content.CompressedPayload, err = peelPayload(content.CompressedPayload)
		if err != nil {
			return config.Content{}, err
		}

		entry = cacheEntry{
			Content: content,
			Error:   nil,
		}
		err = payload.configCache.store(key, entry)
		if err != nil {
			return config.Content{}, err
		}
	}

	return entry.Content, nil
}

func (n *Node) writeConfig(_ context.Context, c config.Content) error {
	err := n.conn.WriteMessage(websocket.BinaryMessage, c.CompressedPayload)
	if err != nil {
		if wsCloseErr, ok := err.(*websocket.CloseError); ok {
			return ErrConnClosed{Code: wsCloseErr.Code}
		}
		return err
	}

	n.Logger.Info("successfully sent payload to node",
		zap.Int("payload-size-byte", len(c.CompressedPayload)))
	return nil
}

func (n *Node) writeWRPCConfig(ctx context.Context, c config.Content, configVersion uint64) error {
	req := &config_service.SyncConfigRequest{
		Config:     c.CompressedPayload,
		Version:    configVersion,
		ConfigHash: c.Hash,
		Hashes: &config_service.GranularHashes{
			Config:    c.Hash,
			Routes:    c.GranularHashes["routes"],
			Services:  c.GranularHashes["services"],
			Plugins:   c.GranularHashes["plugins"],
			Upstreams: c.GranularHashes["upstreams"],
			Targets:   c.GranularHashes["targets"],
		},
	}

	go func() {
		ctx, cancel := context.WithTimeout(ctx, defaultBroadcastTimeout)
		defer cancel()

		cli := config_service.ConfigServiceClient{Peer: n.peer}
		defer func() {
			if err := recover(); err != nil {
				n.Logger.Error("panic sending wrpc config", zap.Any("panic-message", err))
			}
		}()
		out, err := cli.SyncConfig(ctx, req)
		if err != nil {
			n.Logger.Error("SyncConfig method failed", zap.Error(err))
			return
		}
		if out != nil && !out.Accepted {
			n.Logger.Info("configuration not accepted")
			for _, configerr := range out.Errors {
				n.Logger.Info("rejection description",
					zap.String("err-type", configerr.ErrType.String()),
					zap.String("err-id", configerr.Id),
					zap.String("err-entity", configerr.Entity))
			}
			return
		}

		n.Logger.Info("successfully sent payload to node",
			zap.Int("payload-size-byte", len(c.CompressedPayload)))
	}()

	return nil
}

func peelPayload(compressed []byte) ([]byte, error) {
	compressedReader := bytes.NewReader(compressed)
	decompressReader, err := gzip.NewReader(compressedReader)
	if err != nil {
		return nil, err
	}

	jsonDecoder := json.Marshaller.NewDecoder(decompressReader)
	var m map[string]any
	err = jsonDecoder.Decode(&m)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	jsonEncoder := json.Marshaller.NewEncoder(w)
	if err = jsonEncoder.Encode(m["config_table"]); err != nil {
		return nil, err
	}
	if err = w.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
