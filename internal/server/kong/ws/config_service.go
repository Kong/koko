package ws

import (
	"context"
	"fmt"

	"github.com/kong/go-wrpc/wrpc"
	config_service "github.com/kong/koko/internal/gen/wrpc/kong/services/config/v1"
	"go.uber.org/zap"
)

// Configer is the HandleRegisterer to be added to a Negotiator object.
// It also implements the ConfigService interface so it adds itself
// as the wrpc.Service.  This allows the service to get a reference to
// the Manager object.
type Configer struct {
	Manager *Manager
}

// Register the "v1" config service.
func (c *Configer) Register(peer *wrpc.Peer) error {
	return peer.Register(&config_service.ConfigServiceServer{ConfigService: c})
}

func (c *Configer) GetCapabilities(
	ctx context.Context,
	peer *wrpc.Peer,
	req *config_service.GetCapabilitiesRequest,
) (resp *config_service.GetCapabilitiesResponse, err error) {
	c.Manager.logger.Warn("Received a GetCapabilities rpc call from DP",
		zap.String("nodeAddr", peer.RemoteAddr().String()))

	return nil, fmt.Errorf("Not implemented")
}

// PingCP handles the incoming ping method from the CP.
// (Different from a websocket Ping frame)
// Records the given hashes from CP.
func (c *Configer) PingCP(
	ctx context.Context,
	peer *wrpc.Peer,
	req *config_service.PingCPRequest,
) (resp *config_service.PingCPResponse, err error) {
	// find out the Node
	// update the reported hash
	node, ok := c.Manager.FindNode(peer.RemoteAddr().String())
	if !ok {
		return nil, fmt.Errorf("can't find node from %v", peer.RemoteAddr())
	}
	node.logger.Debug("received PingCP method", zap.String("hash", req.Hash))

	node.lock.Lock()
	node.hash, err = truncateHash(req.Hash)
	node.lock.Unlock()
	if err != nil {
		node.logger.Error("Invalid hash in PingCP method", zap.Error(err))
		peer.ErrLogger(fmt.Errorf("PingCP: Received invalid hash from kong data-plane: %w", err))
		return nil, err
	}

	c.Manager.updateNodeStatus(node)
	return &config_service.PingCPResponse{}, nil
}

// ReportMetadata handles the initial information
// from the CP (currently the list of plugins it has available).
// Then the manager can validate and promote the
// node from "pending" to fully working.
func (c *Configer) ReportMetadata(
	ctx context.Context,
	peer *wrpc.Peer,
	req *config_service.ReportMetadataRequest,
) (resp *config_service.ReportMetadataResponse, err error) {
	c.Manager.logger.Debug("received ReportMetadata method",
		zap.String("nodeAddr", peer.RemoteAddr().String()))

	node, ok := c.Manager.pendingNodes.FindNode(peer.RemoteAddr().String())
	if !ok {
		return nil, fmt.Errorf("can't find node from %v", peer.RemoteAddr())
	}

	plugins := make([]string, 0, len(req.Plugins))
	for _, p := range req.Plugins {
		plugins = append(plugins, p.Name)
	}
	node.logger.Debug("plugin list", zap.Strings("plugins", plugins))

	err = c.Manager.addWrpcNode(node, plugins)
	if err != nil {
		node.logger.With(zap.Error(err)).Error("error when adding validated node")
		node.Close()
		return nil, err
	}
	node.logger.Debug("validated node added")

	return &config_service.ReportMetadataResponse{
		Response: &config_service.ReportMetadataResponse_Ok{
			Ok: "valid",
		},
	}, nil
}

func (c *Configer) SyncConfig(
	ctx context.Context,
	peer *wrpc.Peer,
	req *config_service.SyncConfigRequest,
) (resp *config_service.SyncConfigResponse, err error) {
	// this is a CP->DP method
	c.Manager.logger.Warn("Received a SyncConfig rpc call from DP",
		zap.String("nodeAddr", peer.RemoteAddr().String()))

	return nil, fmt.Errorf("Control plane nodes don't implement this method.")
}
