package ws

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"sync"

	"github.com/gorilla/websocket"
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
	logger   *zap.Logger
	hash     sum
}

type ErrConnClosed struct {
	Code int
}

func (e ErrConnClosed) Error() string {
	return fmt.Sprintf("websocket connection closed (code: %v)", e.Code)
}

// readThread continuously reads messages from connected DP node.
func (n *Node) readThread() error {
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

func (n *Node) write(payload []byte, hash sum) error {
	n.lock.Lock()
	defer n.lock.Unlock()

	if n.hash == hash {
		n.logger.With(zap.String("config_hash",
			hash.String())).Info("hash matched, skipping update")
		return nil
	}

	err := n.conn.WriteMessage(websocket.BinaryMessage, payload)
	if err != nil {
		if wsCloseErr, ok := err.(*websocket.CloseError); ok {
			return ErrConnClosed{Code: wsCloseErr.Code}
		}
		return err
	}
	n.hash = hash
	return nil
}
