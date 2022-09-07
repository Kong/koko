package admin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	pbModel "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
	nonPublic "github.com/kong/koko/internal/gen/grpc/kong/nonpublic/v1"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/server/kong/ws/config"
	"github.com/kong/koko/internal/server/util"
	"github.com/kong/koko/internal/store"
	"go.uber.org/zap"
)

type NodeService struct {
	v1.UnimplementedNodeServiceServer
	CommonOpts
}

func (s *NodeService) GetNode(ctx context.Context,
	req *v1.GetNodeRequest,
) (*v1.GetNodeResponse, error) {
	if req.Id == "" {
		return nil, s.err(ctx, util.ErrClient{Message: "required ID is missing"})
	}
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	result := resource.NewNode()
	s.logger(ctx).With(zap.String("id", req.Id)).Debug("reading node by id")
	err = db.Read(ctx, result, store.GetByID(req.Id))
	if err != nil {
		return nil, s.err(ctx, err)
	}

	status, err := s.statusForNode(ctx, db, req.Id)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return nil, s.err(ctx, err)
	}

	var nodeStatus *nonPublic.NodeStatus
	if status != nil {
		nodeStatus = status.NodeStatus
	}
	result.Node.CompatibilityStatus = s.nodeStatusToCompatibilityStatus(ctx, result.Node.ConfigHash, nodeStatus)
	return &v1.GetNodeResponse{
		Item: result.Node,
	}, nil
}

func (s *NodeService) statusForNode(ctx context.Context, db store.Store, nodeID string) (*resource.NodeStatus, error) {
	result := resource.NewNodeStatus()
	err := db.Read(ctx, result, store.GetByID(nodeID))
	if err != nil {
		return nil, err
	}
	return &result, nil
}

var emptyConfigHash = strings.Repeat("0", 32) //nolint:gomnd

func (s *NodeService) nodeStatusToCompatibilityStatus(ctx context.Context,
	configHash string, nodeStatus *nonPublic.NodeStatus,
) *pbModel.CompatibilityStatus {
	// no config hash means the configuration state cannot be reliably tracked
	// no nodeStatus implies the compat status is not tracked
	if emptyConfigHash == configHash || configHash == "" || nodeStatus == nil {
		return &pbModel.CompatibilityStatus{
			State: pbModel.CompatibilityState_COMPATIBILITY_STATE_UNKNOWN,
		}
	}
	compatIssues := make([]*pbModel.CompatibilityIssue, len(nodeStatus.Issues))
	for i, issue := range nodeStatus.Issues {
		metadata, err := config.ChangeRegistry.GetMetadata(config.ChangeID(issue.GetCode()))
		if err != nil {
			s.logger(ctx).Error("failed to get change metadata",
				zap.String("change-id", issue.GetCode()),
			)
			return &pbModel.CompatibilityStatus{
				State: pbModel.CompatibilityState_COMPATIBILITY_STATE_UNKNOWN,
			}
		}
		issue.Severity = string(metadata.Severity)
		issue.Resolution = metadata.Resolution
		issue.Description = metadata.Description
		compatIssues[i] = &pbModel.CompatibilityIssue{
			Code:              issue.GetCode(),
			Severity:          string(metadata.Severity),
			Description:       metadata.Description,
			Resolution:        metadata.Resolution,
			AffectedResources: issue.AffectedResources,
		}
	}
	state := pbModel.CompatibilityState_COMPATIBILITY_STATE_FULLY_COMPATIBLE
	if len(compatIssues) > 0 {
		state = pbModel.CompatibilityState_COMPATIBILITY_STATE_INCOMPATIBLE
	}
	return &pbModel.CompatibilityStatus{
		State:  state,
		Issues: compatIssues,
	}
}

func (s *NodeService) CreateNode(ctx context.Context,
	req *v1.CreateNodeRequest,
) (*v1.CreateNodeResponse, error) {
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	res := resource.NewNode()
	res.Node = req.Item
	if err := db.Create(ctx, res); err != nil {
		return nil, s.err(ctx, err)
	}
	util.SetHeader(ctx, http.StatusCreated)
	return &v1.CreateNodeResponse{
		Item: res.Node,
	}, nil
}

func (s *NodeService) UpsertNode(ctx context.Context,
	req *v1.UpsertNodeRequest,
) (*v1.UpsertNodeResponse, error) {
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	res := resource.NewNode()
	res.Node = req.Item
	if err := db.Upsert(ctx, res); err != nil {
		return nil, s.err(ctx, err)
	}
	return &v1.UpsertNodeResponse{
		Item: res.Node,
	}, nil
}

func (s *NodeService) DeleteNode(ctx context.Context,
	req *v1.DeleteNodeRequest,
) (*v1.DeleteNodeResponse, error) {
	if req.Id == "" {
		return nil, s.err(ctx, util.ErrClient{Message: "required ID is missing"})
	}
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	err = db.Delete(ctx, store.DeleteByID(req.Id),
		store.DeleteByType(resource.TypeNode))
	if err != nil {
		return nil, s.err(ctx, err)
	}

	err = db.Delete(ctx, store.DeleteByID(req.Id),
		store.DeleteByType(resource.TypeNodeStatus))
	// node-status may not be present and that is okay
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return nil, s.err(ctx, err)
	}

	util.SetHeader(ctx, http.StatusNoContent)
	return &v1.DeleteNodeResponse{}, nil
}

func (s *NodeService) ListNodes(ctx context.Context,
	req *v1.ListNodesRequest,
) (*v1.ListNodesResponse, error) {
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	list := resource.NewList(resource.TypeNode)
	listOptFns, err := ListOptsFromReq(req.Page)
	if err != nil {
		return nil, s.err(ctx, err)
	}
	if err := db.List(ctx, list, listOptFns...); err != nil {
		return nil, s.err(ctx, err)
	}
	nodeStatuses, err := s.listAllNodeStatus(ctx, db)
	if err != nil {
		return nil, s.err(ctx, err)
	}
	nodes := nodesFromObjects(list.GetAll())
	s.addStatusToNodes(ctx, nodes, nodeStatuses)
	return &v1.ListNodesResponse{
		Items: nodes,
		Page:  getPaginationResponse(list.GetTotalCount(), list.GetNextPage()),
	}, nil
}

func (s *NodeService) addStatusToNodes(ctx context.Context, nodes []*pbModel.Node, statuses []*nonPublic.NodeStatus) {
	for _, node := range nodes {
		var nodeStatus *nonPublic.NodeStatus
		for _, status := range statuses {
			if status.Id == node.Id {
				nodeStatus = status
				break
			}
		}
		node.CompatibilityStatus = s.nodeStatusToCompatibilityStatus(ctx, node.ConfigHash, nodeStatus)
	}
}

// listAllNodeStatus fetches all node-statuses.
// For most clusters, this should result in a single page of instances since
// the number of Kong gateway nodes are fewer than 100 for most clusters.
// This is an optimization to avoid N queries for node status.
func (s *NodeService) listAllNodeStatus(ctx context.Context, db store.Store) ([]*nonPublic.NodeStatus, error) {
	var allNodeStatuses []*nonPublic.NodeStatus

	page := 1
	for page != 0 {
		list := resource.NewList(resource.TypeNodeStatus)
		if err := db.List(ctx, list,
			store.ListWithPageSize(store.MaxPageSize),
			store.ListWithPageNum(page)); err != nil {
			return nil, s.err(ctx, err)
		}
		page = list.GetNextPage()
		allNodeStatuses = append(allNodeStatuses,
			nodeStatusesFromObjects(list.GetAll())...)
	}
	return allNodeStatuses, nil
}

func (s *NodeService) err(ctx context.Context, err error) error {
	return util.HandleErr(ctx, s.logger(ctx), err)
}

func (s *NodeService) logger(ctx context.Context) *zap.Logger {
	return util.LoggerFromContext(ctx).With(s.loggerFields...)
}

func nodesFromObjects(objects []model.Object) []*pbModel.Node {
	res := make([]*pbModel.Node, 0, len(objects))
	for _, object := range objects {
		node, ok := object.Resource().(*pbModel.Node)
		if !ok {
			panic(fmt.Sprintf("expected type '%T' but got '%T'",
				&pbModel.Node{}, object.Resource()))
		}
		res = append(res, node)
	}
	return res
}

func nodeStatusesFromObjects(objects []model.Object) []*nonPublic.NodeStatus {
	res := make([]*nonPublic.NodeStatus, len(objects))
	for i, object := range objects {
		nodeStatus, ok := object.Resource().(*nonPublic.NodeStatus)
		if !ok {
			panic(fmt.Sprintf("expected type '%T' but got '%T'",
				&nonPublic.NodeStatus{}, object.Resource()))
		}
		res[i] = nodeStatus
	}
	return res
}
