package ws

import (
	"context"
	"fmt"

	"github.com/kong/go-wrpc/wrpc"
	config_service "github.com/kong/koko/internal/gen/wrpc/kong/services/config/v1"
	"github.com/kong/koko/internal/metrics"
	"go.uber.org/zap"
)

// ConfigRegisterer is the Registerer to be added to a Negotiator object.
// Its only responsibility is to add a ConfigService to a peer.
type ConfigRegisterer struct{}

// Register the "v1" config service.
func (c *ConfigRegisterer) Register(peer *wrpc.Peer, m *Manager) error {
	return peer.Register(&config_service.ConfigServiceServer{
		ConfigService: &configService{manager: m},
	})
}

// configService implements the wrpc config service.
type configService struct {
	manager *Manager
}

// GetCapabilities is a wRPC method, should only be CP to DP.
func (c *configService) GetCapabilities(
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
// Records the given hashes from CP.
func (c *configService) PingCP(
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

	metrics.Count("data_plane_ping_total", 1, metrics.Tag{
		Key:   "dp_version",
		Value: node.Version,
	},
		metrics.Tag{
			Key:   "protocol",
			Value: "wrpc",
		},
	)
	c.manager.updateNodeStatus(node)
	return &config_service.PingCPResponse{}, nil
}

// ReportMetadata handles the initial information
// from the CP (currently the list of plugins it has available).
// Then the manager can validate and promote the
// node from "pending" to fully working.
func (c *configService) ReportMetadata(
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

	err := c.manager.addWRPCNode(node)
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
func (c *configService) SyncConfig(
	ctx context.Context,
	peer *wrpc.Peer,
	req *config_service.SyncConfigRequest,
) (*config_service.SyncConfigResponse, error) {
	// this is a CP->DP method
	c.manager.logger.Warn("Received a SyncConfig rpc call from DP",
		zap.String("wrpc-client-ip", peer.RemoteAddr().String()))

	return nil, fmt.Errorf("invalid RPC")
}
