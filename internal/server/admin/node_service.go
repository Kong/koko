package admin

import (
	"context"
	"fmt"
	"net/http"

	pbModel "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/server/util"
	"github.com/kong/koko/internal/store"
	"go.uber.org/zap"
)

type NodeService struct {
	v1.UnimplementedNodeServiceServer
	CommonOpts
}

func (s *NodeService) GetNode(ctx context.Context,
	req *v1.GetNodeRequest) (*v1.GetNodeResponse, error) {
	if req.Id == "" {
		return nil, s.err(util.ErrClient{Message: "required ID is missing"})
	}
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	result := resource.NewNode()
	s.logger.With(zap.String("id", req.Id)).Debug("reading node by id")
	err = db.Read(ctx, result, store.GetByID(req.Id))
	if err != nil {
		return nil, s.err(err)
	}
	return &v1.GetNodeResponse{
		Item: result.Node,
	}, nil
}

func (s *NodeService) CreateNode(ctx context.Context,
	req *v1.CreateNodeRequest) (*v1.CreateNodeResponse, error) {
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	res := resource.NewNode()
	res.Node = req.Item
	if err := db.Create(ctx, res); err != nil {
		return nil, s.err(err)
	}
	util.SetHeader(ctx, http.StatusCreated)
	return &v1.CreateNodeResponse{
		Item: res.Node,
	}, nil
}

func (s *NodeService) UpsertNode(ctx context.Context,
	req *v1.UpsertNodeRequest) (*v1.UpsertNodeResponse, error) {
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	res := resource.NewNode()
	res.Node = req.Item
	if err := db.Upsert(ctx, res); err != nil {
		return nil, s.err(err)
	}
	return &v1.UpsertNodeResponse{
		Item: res.Node,
	}, nil
}

func (s *NodeService) DeleteNode(ctx context.Context,
	req *v1.DeleteNodeRequest) (*v1.DeleteNodeResponse, error) {
	if req.Id == "" {
		return nil, s.err(util.ErrClient{Message: "required ID is missing"})
	}
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	err = db.Delete(ctx, store.DeleteByID(req.Id),
		store.DeleteByType(resource.TypeNode))
	if err != nil {
		return nil, s.err(err)
	}
	util.SetHeader(ctx, http.StatusNoContent)
	return &v1.DeleteNodeResponse{}, nil
}

func (s *NodeService) ListNodes(ctx context.Context,
	req *v1.ListNodesRequest) (*v1.ListNodesResponse, error) {
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	list := resource.NewList(resource.TypeNode)
	listOptFns, err := listOptsFromReq(req.ListOptions)
	if err != nil {
		return nil, s.err(util.ErrClient{Message: err.Error()})
	}
	if err := db.List(ctx, list, listOptFns...); err != nil {
		return nil, s.err(err)
	}

	return &v1.ListNodesResponse{
		Items:  nodesFromObjects(list.GetAll()),
		Offset: getOffset(list.GetCount()),
	}, nil
}

func (s *NodeService) err(err error) error {
	return util.HandleErr(s.logger, err)
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
