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
	"github.com/kong/koko/internal/gen/wrpc/kong/model"
	config_service "github.com/kong/koko/internal/gen/wrpc/kong/services/config/v1"
	"github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/server/kong/ws/config"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
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

	if n.peer != nil {
		uncomp, err := config.UncompressPayload(payload)
		if err != nil {
			n.logger.With(zap.Error(err)).Error("decompressing config payload")
			return err
		}

		configTable := model.Config{}
		err = protojson.Unmarshal(uncomp, &configTable)
		if err != nil {
			n.logger.With((zap.Error(err))).Error("unmarshaling config")
			return err
		}

		c := config_service.ConfigServiceClient{Peer: n.peer}
		_, err = c.SyncConfig(context.Background(), &config_service.SyncConfigRequest{
			Config: &configTable,
		})
		if err != nil {
			n.logger.With((zap.Error(err))).Error("calling SyncConfig")
			return err
		}
	}
	return nil
}
