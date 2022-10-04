package admin

import (
	"context"
	"fmt"
	"net/http"

	pb "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/server/util"
	"github.com/kong/koko/internal/store"
	"go.uber.org/zap"
)

type VaultService struct {
	v1.UnimplementedVaultServiceServer
	CommonOpts
}

func (s *VaultService) GetVault(ctx context.Context,
	req *v1.GetVaultRequest,
) (*v1.GetVaultResponse, error) {
	if req.Id == "" {
		return nil, s.err(ctx, util.ErrClient{Message: "required ID is missing"})
	}
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	result := resource.NewVault()
	s.logger(ctx).With(zap.String("id", req.Id)).Debug("reading vault by id")
	err = getEntityByIDOrName(ctx, req.Id, result, store.GetByIndex("prefix", req.Id), db, s.logger(ctx))
	if err != nil {
		return nil, s.err(ctx, err)
	}
	return &v1.GetVaultResponse{
		Item: result.Vault,
	}, nil
}

func (s *VaultService) CreateVault(ctx context.Context,
	req *v1.CreateVaultRequest,
) (*v1.CreateVaultResponse, error) {
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	res := resource.NewVault()
	res.Vault = req.Item
	if err := db.Create(ctx, res); err != nil {
		return nil, s.err(ctx, err)
	}
	util.SetHeader(ctx, http.StatusCreated)
	return &v1.CreateVaultResponse{
		Item: res.Vault,
	}, nil
}

func (s *VaultService) UpsertVault(ctx context.Context,
	req *v1.UpsertVaultRequest,
) (*v1.UpsertVaultResponse, error) {
	if err := validUUID(req.Item.Id); err != nil {
		return nil, s.err(ctx, err)
	}
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	res := resource.NewVault()
	res.Vault = req.Item
	if err := db.Upsert(ctx, res); err != nil {
		return nil, s.err(ctx, err)
	}
	return &v1.UpsertVaultResponse{
		Item: res.Vault,
	}, nil
}

func (s *VaultService) DeleteVault(ctx context.Context,
	req *v1.DeleteVaultRequest,
) (*v1.DeleteVaultResponse, error) {
	if err := validUUID(req.Id); err != nil {
		return nil, s.err(ctx, err)
	}
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	err = db.Delete(ctx, store.DeleteByID(req.Id),
		store.DeleteByType(resource.TypeVault))
	if err != nil {
		return nil, s.err(ctx, err)
	}
	util.SetHeader(ctx, http.StatusNoContent)
	return &v1.DeleteVaultResponse{}, nil
}

func (s *VaultService) ListVaults(ctx context.Context,
	req *v1.ListVaultsRequest,
) (*v1.ListVaultsResponse, error) {
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	list := resource.NewList(resource.TypeVault)
	listOptFns, err := ListOptsFromReq(req.Page)
	if err != nil {
		return nil, s.err(ctx, err)
	}
	if err := db.List(ctx, list, listOptFns...); err != nil {
		return nil, s.err(ctx, err)
	}
	return &v1.ListVaultsResponse{
		Items: vaultsFromObjects(list.GetAll()),
		Page:  getPaginationResponse(list.GetTotalCount(), list.GetNextPage()),
	}, nil
}

func (s *VaultService) err(ctx context.Context, err error) error {
	return util.HandleErr(ctx, s.logger(ctx), err)
}

func (s *VaultService) logger(ctx context.Context) *zap.Logger {
	return util.LoggerFromContext(ctx).With(s.loggerFields...)
}

func vaultsFromObjects(objects []model.Object) []*pb.Vault {
	res := make([]*pb.Vault, len(objects))
	for i, object := range objects {
		var ok bool
		if res[i], ok = object.Resource().(*pb.Vault); !ok {
			panic(fmt.Sprintf("expected type '%T' but got '%T'",
				&pb.Vault{}, object.Resource()))
		}
	}
	return res
}
