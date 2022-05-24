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
func (c *Configer) Register(peer registerer) error {
	return peer.Register(&config_service.ConfigServiceServer{ConfigService: c})
}

func (c *Configer) GetCapabilities(
	ctx context.Context,
	peer *wrpc.Peer,
	req *config_service.GetCapabilitiesRequest,
) (resp *config_service.GetCapabilitiesResponse, err error) {
	return nil, fmt.Errorf("not implemented")
}

// Got a ping RPC => record given hashes.
func (c *Configer) PingCP(
	ctx context.Context,
	peer *wrpc.Peer,
	req *config_service.PingCPRequest,
) (resp *config_service.PingCPResponse, err error) {
	// find out the Node
	// update the reported hash
	c.Manager.logger.Debug("received PingCP method",
		zap.String("nodeAddr", peer.RemoteAddr().String()),
		zap.String("hash", req.Hash))
	node, ok := c.Manager.FindNode(peer.RemoteAddr().String())
	if !ok {
		return nil, fmt.Errorf("can't find node from %v", peer.RemoteAddr())
	}

	node.lock.Lock()
	node.hash, err = truncateHash(req.Hash)
	node.lock.Unlock()
	if err != nil {
		peer.ErrLogger(fmt.Errorf("PingCP: Received invalid hash from kong data-plane: %w", err))
		return nil, err
	}

	c.Manager.updateNodeStatus(node) // nolint: contextcheck
	return &config_service.PingCPResponse{}, nil
}

// Got the initial metadata (list of plugins)
// then the manager can validate and promote the
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

	var plugins []string
	for _, p := range req.Plugins {
		plugins = append(plugins, p.Name)
	}
	node.logger.Debug("plugin list", zap.Strings("plugins", plugins))

	err = c.Manager.addWrpcNode(node, plugins) // nolint: contextcheck
	if err != nil {
		node.logger.Error("adding validated node")
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
	return nil, fmt.Errorf("wrong direction")
}
