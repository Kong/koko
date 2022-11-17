package admin

import (
	"context"
	"net/http"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/server/util"
	"github.com/kong/koko/internal/store"
	"go.uber.org/zap"
	// "github.com/kong/koko/internal/store"
)

type KeyService struct {
	v1.UnimplementedKeyServiceServer
	CommonOpts
}

func (s *KeyService) GetKey(
	ctx context.Context,
	req *v1.GetKeyRequest,
) (*v1.GetKeyResponse, error) {
	logger := s.logger(ctx)
	db, err := s.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	result := resource.NewKey()
	err = getEntityByIDOrName(ctx, req.Id, result, store.GetByName(req.Id), db, logger)
	if err != nil {
		return nil, util.HandleErr(ctx, logger, err)
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
		s.logger(ctx).Error("error creating", zap.Error(err))
		return nil, s.err(ctx, err)
	}

	util.SetHeader(ctx, http.StatusCreated)
	return &v1.CreateKeyResponse{
		Item: res.Key,
	}, nil
}

func (s *KeyService) UpsertKey(ctx context.Context,
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

func (s *KeyService) DeleteKey(ctx context.Context,
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
