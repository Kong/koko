package ws

import (
	"context"
	"fmt"

	"github.com/kong/go-wrpc/wrpc"
	config_service "github.com/kong/koko/internal/gen/wrpc/kong/services/config/v1"
	"go.uber.org/zap"
)

// Configurer is the Registerer to be added to a Negotiator object.
// It also implements the ConfigService interface so it adds itself
// as the wrpc.Service.  This allows the service to get a reference to
// the Manager object.
type Configurer struct {
	manager *Manager
	ping    chan *Node
}

// NewConfigurer creates a new configurer with the given manager.
func NewConfigurer(m *Manager) *Configurer {
	c := &Configurer{
		manager: m,
	}

	c.AnswerPingThread(m.ctx, m.logger)
	return c
}

// AnswerPingThread starts a background goroutine to send
// config messages whenever a node requests it.
// Currently only the PingCP rpc does this request.
func (c *Configurer) AnswerPingThread(ctx context.Context, logger *zap.Logger) {
	c.ping = make(chan *Node)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			case node := <-c.ping:
				_ = node.sendConfig(ctx, c.manager.payload)
			}
		}
	}()
}

// Register the "v1" config service.
func (c *Configurer) Register(peer *wrpc.Peer) error {
	return peer.Register(&config_service.ConfigServiceServer{ConfigService: c})
}

// GetCapabilities is a wRPC method, should only be CP to DP.
func (c *Configurer) GetCapabilities(
	ctx context.Context,
	peer *wrpc.Peer,
	req *config_service.GetCapabilitiesRequest,
) (*config_service.GetCapabilitiesResponse, error) {
	c.manager.logger.Warn("Received a GetCapabilities rpc call from DP",
		zap.String("wrpc-client-ip", peer.RemoteAddr().String()))

	return nil, fmt.Errorf("invalid RPC")
}

// PingCP handles the incoming ping method from the CP.
// (Different from a websocket Ping frame)
// Records the given hashes from CP and signals the answerPing thread
// to send a config if necessary.
func (c *Configurer) PingCP(
	_ context.Context,
	peer *wrpc.Peer,
	req *config_service.PingCPRequest,
) (*config_service.PingCPResponse, error) {
	// find out the Node
	// update the reported hash
	node := c.manager.FindNode(peer.RemoteAddr().String())
	if node == nil {
		return nil, fmt.Errorf("can't find node from %v", peer.RemoteAddr())
	}
	node.Logger.Debug("received PingCP method", zap.String("config_hash", req.Hash))

	node.lock.Lock()
	var err error
	node.hash, err = truncateHash(req.Hash)
	node.lock.Unlock()
	if err != nil {
		node.Logger.Error("Invalid hash in PingCP method", zap.Error(err))
		peer.ErrLogger(fmt.Errorf("PingCP: Received invalid hash from kong data-plane: %w", err))
		return nil, err
	}

	c.manager.updateNodeStatus(node)
	c.ping <- node

	return &config_service.PingCPResponse{}, nil
}

// ReportMetadata handles the initial information
// from the CP (currently the list of plugins it has available).
// Then the manager can validate and promote the
// node from "pending" to fully working.
func (c *Configurer) ReportMetadata(
	ctx context.Context,
	peer *wrpc.Peer,
	req *config_service.ReportMetadataRequest,
) (*config_service.ReportMetadataResponse, error) {
	c.manager.logger.Debug("received ReportMetadata method",
		zap.String("wrpc-client-ip", peer.RemoteAddr().String()))

	node := c.manager.pendingNodes.FindNode(peer.RemoteAddr().String())
	if node == nil {
		c.manager.logger.Error("can't find pending node",
			zap.String("wrpc-client-ip", peer.RemoteAddr().String()))
		return nil, fmt.Errorf("invalid RPC")
	}

	plugins := make([]string, len(req.Plugins))
	for i, p := range req.Plugins {
		plugins[i] = p.Name
	}
	node.Logger.Debug("plugin list reported by the DP", zap.Strings("plugins", plugins))

	err := c.manager.addWRPCNode(node, plugins)
	if err != nil {
		node.Logger.With(zap.Error(err)).Error("error when adding validated node")
		return nil, err
	}
	node.Logger.Debug("validated node added")

	return &config_service.ReportMetadataResponse{
		Response: &config_service.ReportMetadataResponse_Ok{Ok: "valid"},
	}, nil
}

// SyncConfig is a wRPC method, should only be CP to DP.
func (c *Configurer) SyncConfig(
	ctx context.Context,
	peer *wrpc.Peer,
	req *config_service.SyncConfigRequest,
) (*config_service.SyncConfigResponse, error) {
	// this is a CP->DP method
	c.manager.logger.Warn("Received a SyncConfig rpc call from DP",
		zap.String("wrpc-client-ip", peer.RemoteAddr().String()))

	return nil, fmt.Errorf("invalid RPC")
}
