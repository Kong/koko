package ws

import (
	"context"
	"fmt"

	"github.com/kong/go-wrpc/wrpc"
	config_service "github.com/kong/koko/internal/gen/wrpc/kong/services/config/v1"
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
// validate versions and, if positive, only then
// add the node to the manager.
func (c *Configer) ReportMetadata(
	ctx context.Context,
	peer *wrpc.Peer,
	req *config_service.ReportMetadataRequest,
) (resp *config_service.ReportMetadataResponse, err error) {
	// find out the Node
	// hope it's waiting for the metadata report
	// push it there
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
