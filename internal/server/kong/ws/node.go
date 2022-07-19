package ws

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"net"
	"regexp"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/kong/go-wrpc/wrpc"
	config_service "github.com/kong/koko/internal/gen/wrpc/kong/services/config/v1"
	"github.com/kong/koko/internal/json"
	"go.uber.org/zap"
)

type sum [hashSize]byte

const (
	hashRegexPat = `^[a-zA-Z0-9]+$`
	hashSize     = sha256.Size
)

var hashRegex = regexp.MustCompile(hashRegexPat)

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

type Node struct {
	lock     sync.RWMutex
	ID       string
	Version  string
	Hostname string
	conn     *websocket.Conn
	peer     *wrpc.Peer
	logger   *zap.Logger
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
	return &Node{
		ID:       opts.id,
		Version:  opts.version,
		Hostname: opts.hostname,
		conn:     opts.connection,
		peer:     opts.peer,
		logger:   opts.logger,
	}, nil
}

// Close ends the Node's lifetime and of its connection.
func (n *Node) Close() error {
	if n.conn != nil {
		return n.conn.Close()
	}

	if n.peer != nil {
		return n.peer.Close()
	}

	return nil
}

// RemoteAddr returns the network address of the client.
func (n *Node) RemoteAddr() net.Addr {
	if n.conn != nil {
		return n.conn.RemoteAddr()
	}
	if n.peer != nil {
		return n.peer.RemoteAddr()
	}
	return &net.IPAddr{}
}

// GetPluginList receives the list of plugins the DP sends
// right after connection on the old WebSocket protocol.
func (n *Node) GetPluginList() ([]string, error) {
	if n.conn == nil {
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
	if n.conn == nil {
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
		n.logger.Info("received message from DP")
		n.logger.Sugar().Debugf("recv: %s", message)
	}
}

// write sends the provided config payload to the DP, if the hash
// is different from the last one reported.
// Used only on WebSocket protocol.
func (n *Node) write(payload []byte, hash sum) error {
	if n.conn == nil {
		return fmt.Errorf("node.write is only for plain WebSocket nodes")
	}
	n.lock.RLock()
	defer n.lock.RUnlock()

	if bytes.Equal(n.hash[:], hash[:]) {
		n.logger.With(zap.String("config_hash",
			hash.String())).Info("hash matched, skipping update")
		return nil
	}

	if n.conn != nil {
		err := n.conn.WriteMessage(websocket.BinaryMessage, payload)
		if err != nil {
			if wsCloseErr, ok := err.(*websocket.CloseError); ok {
				return ErrConnClosed{Code: wsCloseErr.Code}
			}
			return err
		}
	}

	return nil
}

func (n *Node) sendConfig(ctx context.Context, payload *Payload) error {
	if n.conn != nil {
		return n.sendJSONConfig(ctx, payload)
	}

	if n.peer != nil {
		return n.sendWRPCConfig(ctx, payload)
	}

	return fmt.Errorf("node disconnected")
}

func (n *Node) sendJSONConfig(ctx context.Context, payload *Payload) error {
	ctx, cancel := context.WithTimeout(ctx, defaultBroadcastTimeout)
	defer cancel()

	content, err := payload.Payload(ctx, n.Version)
	if err != nil {
		return fmt.Errorf("unable to gather payload: %w", err)
	}
	n.logger.Info("broadcasting to node",
		zap.String("hash", content.Hash))
	// TODO(hbagdi): perf: use websocket.PreparedMessage
	hash, err := truncateHash(content.Hash)
	if err != nil {
		n.logger.With(zap.Error(err)).Sugar().Errorf("invalid hash [%v]", hash)
		return err
	}
	err = n.write(content.CompressedPayload, hash)
	if err != nil {
		n.logger.Error("broadcast to node failed", zap.Error(err))
		// TODO(hbagdi: remove the node if connection has been closed?
		return err
	}
	n.logger.Info("successfully sent payload to node")
	return nil
}

func (n *Node) sendWRPCConfig(ctx context.Context, payload *Payload) error {
	content, err := payload.WrpcConfigPayload(ctx, n.Version)
	if err != nil {
		n.logger.With(zap.Error(err)).Error("preparing wrpc config payload")
		return err
	}

	hash, err := truncateHash(content.Hash)
	if err != nil {
		n.logger.Error("invalid hash", zap.Error(err), zap.String("hash", hash.String()))
		return err
	}

	if bytes.Equal(n.hash[:], hash[:]) {
		n.logger.With(zap.String("config_hash",
			hash.String())).Info("hash matched, skipping update")
		return nil
	}

	var out config_service.SyncConfigResponse
	go func() {
		ctx, cancel := context.WithTimeout(ctx, defaultBroadcastTimeout)
		defer cancel()

		err := n.peer.DoRequest(ctx, content.Req, &out)
		if err != nil {
			n.logger.With(zap.Error(err)).Error("sending config")
		}
	}()
	return nil
}
