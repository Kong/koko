package ws

import (
	"crypto/sha256"
	"fmt"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type sum [sha256.Size]byte

func (s sum) String() string {
	return string(s[:])
}

// If the string has more than 32 bytes, the trailing bytes get truncated.
func fromString(s32 string) sum {
	s := sum{}
	nodeHash := []byte(s32)
	for i := 0; i < sha256.Size; i++ {
		s[i] = nodeHash[i]
	}
	return s
}

type Node struct {
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
		n.logger.Sugar().Debugf("recv: %s", message)
	}
}

func (n *Node) write(payload []byte) error {
	sum := hashPayload(payload)
	if n.hash == sum {
		n.logger.With(zap.String("config_hash",
			sum.String())).Info("hash matched, skipping update")
		return nil
	}

	err := n.conn.WriteMessage(websocket.BinaryMessage, payload)
	if err != nil {
		if wsCloseErr, ok := err.(*websocket.CloseError); ok {
			return ErrConnClosed{Code: wsCloseErr.Code}
		}
		return err
	}
	n.hash = sum
	return nil
}

func hashPayload(payload []byte) sum {
	return sha256.Sum256(payload)
}
