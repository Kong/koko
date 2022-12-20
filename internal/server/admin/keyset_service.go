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

type KeySetService struct {
	v1.UnimplementedKeySetServiceServer
	CommonOpts
}

func (s *KeySetService) GetKeySet(
	ctx context.Context,
	req *v1.GetKeySetRequest,
) (*v1.GetKeySetResponse, error) {
	if req.Id == "" {
		return nil, s.err(ctx, util.ErrClient{Message: "required ID is missing"})
	}
	logger := s.logger(ctx)
	db, err := s.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	result := resource.NewKeySet()
	err = getEntityByIDOrName(ctx, req.Id, result, store.GetByName(req.Id), db, logger)
	if err != nil {
		return nil, util.HandleErr(ctx, logger, err)
	}

	return &v1.GetKeySetResponse{
		Item: result.KeySet,
	}, nil
}

func (s *KeySetService) CreateKeySet(
	ctx context.Context,
	req *v1.CreateKeySetRequest,
) (*v1.CreateKeySetResponse, error) {
	db, err := s.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}

	res := resource.NewKeySet()
	res.KeySet = req.Item
	if err := db.Create(ctx, res); err != nil {
		s.logger(ctx).Error("unable to create keyset entity", zap.Error(err))
		return nil, s.err(ctx, err)
	}

	util.SetHeader(ctx, http.StatusCreated)
	return &v1.CreateKeySetResponse{
		Item: res.KeySet,
	}, nil
}

func (s *KeySetService) UpsertKeySet(
	ctx context.Context,
	req *v1.UpsertKeySetRequest,
) (*v1.UpsertKeySetResponse, error) {
	if err := validUUID(req.Item.Id); err != nil {
		return nil, s.err(ctx, err)
	}
	db, err := s.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	res := resource.NewKeySet()
	res.KeySet = req.Item
	if err := db.Upsert(ctx, res); err != nil {
		return nil, s.err(ctx, err)
	}
	return &v1.UpsertKeySetResponse{
		Item: res.KeySet,
	}, nil
}

func (s *KeySetService) DeleteKeySet(
	ctx context.Context,
	req *v1.DeleteKeySetRequest,
) (*v1.DeleteKeySetResponse, error) {
	if err := validUUID(req.Id); err != nil {
		return nil, s.err(ctx, err)
	}
	db, err := s.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	err = db.Delete(ctx, store.DeleteByID(req.Id), store.DeleteByType(resource.TypeKeySet))
	if err != nil {
		return nil, s.err(ctx, err)
	}
	util.SetHeader(ctx, http.StatusNoContent)
	return &v1.DeleteKeySetResponse{}, nil
}

func (s *KeySetService) ListKeySets(
	ctx context.Context,
	req *v1.ListKeySetsRequest,
) (*v1.ListKeySetsResponse, error) {
	db, err := s.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}

	listFn := []store.ListOptsFunc{}

	list := resource.NewList(resource.TypeKeySet)
	listOptFns, err := ListOptsFromReq(req.Page)
	if err != nil {
		return nil, s.err(ctx, err)
	}

	listFn = append(listFn, listOptFns...)

	if err := db.List(ctx, list, listFn...); err != nil {
		return nil, s.err(ctx, err)
	}

	return &v1.ListKeySetsResponse{
		Items: keySetsFromObjects(list.GetAll()),
		Page:  getPaginationResponse(list.GetTotalCount(), list.GetNextPage()),
	}, nil
}

func keySetsFromObjects(objects []model.Object) []*pbModel.KeySet {
	res := make([]*pbModel.KeySet, 0, len(objects))
	for _, object := range objects {
		keySet, ok := object.Resource().(*pbModel.KeySet)
		if !ok {
			panic(fmt.Sprintf("expected type '%T' but got '%T'",
				&pbModel.KeySet{}, object.Resource()))
		}
		res = append(res, keySet)
	}
	return res
}
