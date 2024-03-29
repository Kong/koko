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

type KeyService struct {
	v1.UnimplementedKeyServiceServer
	CommonOpts
}

func (s *KeyService) GetKey(
	ctx context.Context,
	req *v1.GetKeyRequest,
) (*v1.GetKeyResponse, error) {
	if req.Id == "" {
		return nil, s.err(ctx, util.ErrClient{Message: "required ID is missing"})
	}
	logger := s.logger(ctx)
	db, err := s.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	result := resource.NewKey()
	err = getEntityByIDOrName(ctx, req.Id, result, store.GetByName(req.Id), db, logger)
	if err != nil {
		return nil, s.err(ctx, err)
	}

	return &v1.GetKeyResponse{
		Item: result.Key,
	}, nil
}

func (s *KeyService) CreateKey(
	ctx context.Context,
	req *v1.CreateKeyRequest,
) (*v1.CreateKeyResponse, error) {
	db, err := s.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}

	res := resource.NewKey()
	res.Key = req.Item
	if err := db.Create(ctx, res); err != nil {
		s.logger(ctx).Error("unable to create new key entity", zap.Error(err))
		return nil, s.err(ctx, err)
	}

	util.SetHeader(ctx, http.StatusCreated)
	return &v1.CreateKeyResponse{
		Item: res.Key,
	}, nil
}

func (s *KeyService) UpsertKey(
	ctx context.Context,
	req *v1.UpsertKeyRequest,
) (*v1.UpsertKeyResponse, error) {
	if err := validUUID(req.Item.Id); err != nil {
		return nil, s.err(ctx, err)
	}
	db, err := s.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	res := resource.NewKey()
	res.Key = req.Item
	if err := db.Upsert(ctx, res); err != nil {
		return nil, s.err(ctx, err)
	}
	return &v1.UpsertKeyResponse{
		Item: res.Key,
	}, nil
}

func (s *KeyService) DeleteKey(
	ctx context.Context,
	req *v1.DeleteKeyRequest,
) (*v1.DeleteKeyResponse, error) {
	if err := validUUID(req.Id); err != nil {
		return nil, s.err(ctx, err)
	}
	db, err := s.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	err = db.Delete(ctx, store.DeleteByID(req.Id), store.DeleteByType(resource.TypeKey))
	if err != nil {
		return nil, s.err(ctx, err)
	}
	util.SetHeader(ctx, http.StatusNoContent)
	return &v1.DeleteKeyResponse{}, nil
}

func (s *KeyService) ListKeys(
	ctx context.Context,
	req *v1.ListKeysRequest,
) (*v1.ListKeysResponse, error) {
	db, err := s.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}

	listFn := []store.ListOptsFunc{}

	list := resource.NewList(resource.TypeKey)
	listOptFns, err := ListOptsFromReq(req.Page)
	if err != nil {
		return nil, s.err(ctx, err)
	}

	listFn = append(listFn, listOptFns...)

	if err := db.List(ctx, list, listFn...); err != nil {
		return nil, s.err(ctx, err)
	}

	return &v1.ListKeysResponse{
		Items: keysFromObjects(list.GetAll()),
		Page:  getPaginationResponse(list.GetTotalCount(), list.GetNextPage()),
	}, nil
}

func keysFromObjects(objects []model.Object) []*pbModel.Key {
	res := make([]*pbModel.Key, 0, len(objects))
	for _, object := range objects {
		key, ok := object.Resource().(*pbModel.Key)
		if !ok {
			panic(fmt.Sprintf("expected type '%T' but got '%T'",
				&pbModel.Key{}, object.Resource()))
		}
		res = append(res, key)
	}
	return res
}
