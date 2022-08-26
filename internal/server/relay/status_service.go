package relay

import (
	"context"
	"fmt"

	adminModel "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	relay "github.com/kong/koko/internal/gen/grpc/kong/relay/service/v1"
	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/server/util"
	"github.com/kong/koko/internal/store"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StatusServiceOpts struct {
	StoreLoader util.StoreLoader
	Logger      *zap.Logger
}

func NewStatusService(opts StatusServiceOpts) *StatusService {
	res := &StatusService{
		storeLoader: opts.StoreLoader,
		logger:      opts.Logger,
	}
	return res
}

type StatusService struct {
	relay.UnimplementedStatusServiceServer
	storeLoader util.StoreLoader
	logger      *zap.Logger
}

func (s StatusService) UpdateExpectedHash(ctx context.Context,
	req *relay.UpdateExpectedHashRequest,
) (*relay.UpdateExpectedHashResponse, error) {
	db, err := s.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}

	res := resource.NewHash()
	if req.Hash == "" {
		return nil, fmt.Errorf("invalid hash: '%v'", req.Hash)
	}
	res.Hash.ExpectedHash = req.Hash

	err = db.Upsert(ctx, res)
	if err != nil {
		return nil, err
	}
	return &relay.UpdateExpectedHashResponse{}, nil
}

func (s StatusService) getDB(ctx context.Context,
	cluster *adminModel.RequestCluster,
) (store.Store, error) {
	store, err := s.storeLoader.Load(ctx, cluster)
	if err != nil {
		if storeLoadErr, ok := err.(util.StoreLoadErr); ok {
			return nil, status.Error(storeLoadErr.Code, storeLoadErr.Message)
		}
		return nil, err
	}
	return store, nil
}

func (s StatusService) UpdateNodeStatus(
	ctx context.Context,
	req *relay.UpdateNodeStatusRequest,
) (*relay.UpdateNodeStatusResponse, error) {
	nodeStatus := req.Item
	// node-status.id == node.id always
	// ensure the caller provides an ID for an update operation
	if nodeStatus.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "node-status ID is required")
	}
	res := resource.NewNodeStatus()
	res.NodeStatus = nodeStatus

	db, err := s.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, util.HandleErr(ctx, s.logger, err)
	}

	err = db.Upsert(ctx, res)
	if err != nil {
		return nil, util.HandleErr(ctx, s.logger, err)
	}
	return &relay.UpdateNodeStatusResponse{}, nil
}
