package admin

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	pbModel "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/server/util"
	"github.com/kong/koko/internal/store"
)

func (s *ConsumerGroupService) GetConsumerGroupRateLimitingAdvancedConfig(
	ctx context.Context,
	req *v1.GetConsumerGroupRateLimitingAdvancedConfigRequest,
) (*v1.GetConsumerGroupRateLimitingAdvancedConfigResponse, error) {
	if err := s.validateRequest(req); err != nil {
		return nil, s.err(ctx, err)
	}

	db, err := s.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}

	result := resource.NewConsumerGroupRateLimitingAdvancedConfig()
	if err := db.Read(ctx, result, store.GetByIndex("consumer_group_id", req.ConsumerGroupId)); err != nil {
		return nil, s.err(ctx, err)
	}

	return &v1.GetConsumerGroupRateLimitingAdvancedConfigResponse{
		Item: result.Config,
	}, nil
}

func (s *ConsumerGroupService) ListConsumerGroupRateLimitingAdvancedConfig(
	ctx context.Context,
	req *v1.ListConsumerGroupRateLimitingAdvancedConfigRequest,
) (*v1.ListConsumerGroupRateLimitingAdvancedConfigResponse, error) {
	db, err := s.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	list := resource.NewList(resource.TypeConsumerGroupRateLimitingAdvancedConfig)
	listOptFns, err := ListOptsFromReq(req.Page)
	if err != nil {
		return nil, s.err(ctx, err)
	}
	if err := db.List(ctx, list, listOptFns...); err != nil {
		return nil, s.err(ctx, err)
	}
	return &v1.ListConsumerGroupRateLimitingAdvancedConfigResponse{
		Items: consumerGroupRateLimitingConfigFromObjects(list.GetAll()),
		Page:  getPaginationResponse(list.GetTotalCount(), list.GetNextPage()),
	}, nil
}

func (s *ConsumerGroupService) CreateConsumerGroupRateLimitingAdvancedConfig(
	ctx context.Context,
	req *v1.CreateConsumerGroupRateLimitingAdvancedConfigRequest,
) (*v1.CreateConsumerGroupRateLimitingAdvancedConfigResponse, error) {
	if err := s.validateRequest(req); err != nil {
		return nil, s.err(ctx, err)
	}

	db, err := s.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}

	result := resource.NewConsumerGroupRateLimitingAdvancedConfig()
	result.Config = req.Item
	result.Config.ConsumerGroupId = req.ConsumerGroupId
	if err := db.Create(ctx, result); err != nil {
		return nil, s.err(ctx, err)
	}

	util.SetHeader(ctx, http.StatusCreated)

	return &v1.CreateConsumerGroupRateLimitingAdvancedConfigResponse{
		Item: result.Config,
	}, nil
}

func (s *ConsumerGroupService) UpsertConsumerGroupRateLimitingAdvancedConfig(
	ctx context.Context,
	req *v1.UpsertConsumerGroupRateLimitingAdvancedConfigRequest,
) (*v1.UpsertConsumerGroupRateLimitingAdvancedConfigResponse, error) {
	if err := s.validateRequest(req); err != nil {
		return nil, s.err(ctx, err)
	}

	db, err := s.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}

	// FIXME(tjasko): The persistence store does not allow updating based on an unique,
	//  foreign relation. It wouldn't be wise to force callers to update by ID here,
	//  as it's a 1:1 relationship anyway. This is a hack and should be fixed.
	currentConfig := resource.NewConsumerGroupRateLimitingAdvancedConfig()
	if err := db.Read(
		ctx,
		currentConfig,
		store.GetByIndex("consumer_group_id", req.ConsumerGroupId),
	); err != nil && !errors.Is(err, store.ErrNotFound) {
		return nil, s.err(ctx, err)
	}

	newConfig := resource.NewConsumerGroupRateLimitingAdvancedConfig()
	newConfig.Config = req.Item
	newConfig.Config.Id = currentConfig.ID()
	newConfig.Config.ConsumerGroupId = req.ConsumerGroupId
	if err := db.Upsert(ctx, newConfig); err != nil {
		return nil, s.err(ctx, err)
	}

	return &v1.UpsertConsumerGroupRateLimitingAdvancedConfigResponse{
		Item: newConfig.Config,
	}, nil
}

func (s *ConsumerGroupService) DeleteConsumerGroupRateLimitingAdvancedConfig(
	ctx context.Context,
	req *v1.DeleteConsumerGroupRateLimitingAdvancedConfigRequest,
) (*v1.DeleteConsumerGroupRateLimitingAdvancedConfigResponse, error) {
	if err := s.validateRequest(req); err != nil {
		return nil, s.err(ctx, err)
	}

	db, err := s.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}

	// FIXME(tjasko): The persistence store does not allow deletions based on an unique,
	//  foreign relation. It wouldn't be wise to force callers to update by ID here,
	//  as it's a 1:1 relationship anyway. This is a hack and should be fixed.
	result := resource.NewConsumerGroupRateLimitingAdvancedConfig()
	if err := db.Read(ctx, result, store.GetByIndex("consumer_group_id", req.ConsumerGroupId)); err != nil {
		return nil, s.err(ctx, err)
	}

	if err := db.Delete(
		ctx,
		store.DeleteByID(result.ID()),
		store.DeleteByType(resource.TypeConsumerGroupRateLimitingAdvancedConfig),
	); err != nil {
		return nil, s.err(ctx, err)
	}

	util.SetHeader(ctx, http.StatusNoContent)

	return &v1.DeleteConsumerGroupRateLimitingAdvancedConfigResponse{}, nil
}

func consumerGroupRateLimitingConfigFromObjects(
	objects []model.Object,
) []*pbModel.ConsumerGroupRateLimitingAdvancedConfig {
	res := make([]*pbModel.ConsumerGroupRateLimitingAdvancedConfig, len(objects))
	for i, obj := range objects {
		var ok bool
		if res[i], ok = obj.Resource().(*pbModel.ConsumerGroupRateLimitingAdvancedConfig); !ok {
			panic(fmt.Sprintf(
				"expected type '%T' but got '%T'",
				&pbModel.ConsumerGroupRateLimitingAdvancedConfig{},
				obj.Resource(),
			))
		}
	}
	return res
}
