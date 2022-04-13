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

type ConsumerService struct {
	v1.UnimplementedConsumerServiceServer
	CommonOpts
}

func (s *ConsumerService) GetConsumer(ctx context.Context,
	req *v1.GetConsumerRequest,
) (*v1.GetConsumerResponse, error) {
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	result := resource.NewConsumer()
	err = getEntityByIDOrName(ctx, req.Id, result, store.GetByIndex("username", req.Id), db, s.logger)
	if err != nil {
		return nil, s.err(ctx, err)
	}
	return &v1.GetConsumerResponse{
		Item: result.Consumer,
	}, nil
}

func (s *ConsumerService) CreateConsumer(ctx context.Context,
	req *v1.CreateConsumerRequest,
) (*v1.CreateConsumerResponse, error) {
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	result := resource.NewConsumer()
	result.Consumer = req.Item
	if err := db.Create(ctx, result); err != nil {
		return nil, s.err(ctx, err)
	}
	util.SetHeader(ctx, http.StatusCreated)
	return &v1.CreateConsumerResponse{
		Item: result.Consumer,
	}, nil
}

func (s *ConsumerService) UpsertConsumer(ctx context.Context,
	req *v1.UpsertConsumerRequest,
) (*v1.UpsertConsumerResponse, error) {
	if err := validUUID(req.Item.Id); err != nil {
		return nil, s.err(ctx, err)
	}
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	result := resource.NewConsumer()
	result.Consumer = req.Item
	if err := db.Upsert(ctx, result); err != nil {
		return nil, s.err(ctx, err)
	}
	return &v1.UpsertConsumerResponse{
		Item: result.Consumer,
	}, nil
}

func (s *ConsumerService) DeleteConsumer(ctx context.Context,
	req *v1.DeleteConsumerRequest,
) (*v1.DeleteConsumerResponse, error) {
	if req.Id == "" {
		return nil, s.err(ctx, util.ErrClient{Message: "required ID is missing"})
	}
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	s.logger.With(zap.String("id", req.Id)).Debug("deleting consumer by id")
	err = db.Delete(ctx, store.DeleteByID(req.Id), store.DeleteByType(resource.TypeConsumer))
	if err != nil {
		return nil, s.err(ctx, err)
	}
	util.SetHeader(ctx, http.StatusNoContent)
	return &v1.DeleteConsumerResponse{}, nil
}

func (s *ConsumerService) ListConsumers(ctx context.Context,
	req *v1.ListConsumersRequest,
) (*v1.ListConsumersResponse, error) {
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	list := resource.NewList(resource.TypeConsumer)
	listOptFns, err := listOptsFromReq(req.Page)
	if err != nil {
		return nil, s.err(ctx, util.ErrClient{Message: err.Error()})
	}
	if err := db.List(ctx, list, listOptFns...); err != nil {
		return nil, s.err(ctx, err)
	}
	return &v1.ListConsumersResponse{
		Items: consumersFromObjects(list.GetAll()),
		Page:  getPaginationResponse(list.GetTotalCount(), list.GetNextPage()),
	}, nil
}

func consumersFromObjects(objects []model.Object) []*pbModel.Consumer {
	res := make([]*pbModel.Consumer, 0, len(objects))
	for _, obj := range objects {
		consumer, ok := obj.Resource().(*pbModel.Consumer)
		if !ok {
			panic(fmt.Sprintf("expected type '%T' but got '%T'",
				&pbModel.Consumer{}, obj.Resource()))
		}
		res = append(res, consumer)
	}
	return res
}

func (s *ConsumerService) err(ctx context.Context, err error) error {
	return util.HandleErr(ctx, s.logger, err)
}
