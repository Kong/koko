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

type CACertificateService struct {
	v1.UnimplementedCACertificateServiceServer
	CommonOpts
}

func (s *CACertificateService) GetCACertificate(ctx context.Context,
	req *v1.GetCACertificateRequest) (*v1.GetCACertificateResponse, error) {
	if req.Id == "" {
		return nil, s.err(util.ErrClient{Message: "required ID is missing"})
	}
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	result := resource.NewCACertificate()
	s.logger.With(zap.String("id", req.Id)).Debug("reading CA certificate by id")
	err = db.Read(ctx, result, store.GetByID(req.Id))
	if err != nil {
		return nil, s.err(err)
	}
	return &v1.GetCACertificateResponse{
		Item: result.CACertificate,
	}, nil
}

func (s *CACertificateService) CreateCACertificate(ctx context.Context,
	req *v1.CreateCACertificateRequest) (*v1.CreateCACertificateResponse, error) {
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	res := resource.NewCACertificate()
	res.CACertificate = req.Item
	if err := db.Create(ctx, res); err != nil {
		return nil, s.err(err)
	}
	util.SetHeader(ctx, http.StatusCreated)
	return &v1.CreateCACertificateResponse{
		Item: res.CACertificate,
	}, nil
}

func (s *CACertificateService) UpsertCACertificate(ctx context.Context,
	req *v1.UpsertCACertificateRequest) (*v1.UpsertCACertificateResponse, error) {
	if err := validUUID(req.Item.Id); err != nil {
		return nil, s.err(err)
	}
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	res := resource.NewCACertificate()
	res.CACertificate = req.Item
	if err := db.Upsert(ctx, res); err != nil {
		return nil, s.err(err)
	}
	return &v1.UpsertCACertificateResponse{
		Item: res.CACertificate,
	}, nil
}

func (s *CACertificateService) DeleteCACertificate(ctx context.Context,
	req *v1.DeleteCACertificateRequest) (*v1.DeleteCACertificateResponse, error) {
	if err := validUUID(req.Id); err != nil {
		return nil, s.err(err)
	}
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	err = db.Delete(ctx, store.DeleteByID(req.Id),
		store.DeleteByType(resource.TypeCACertificate))
	if err != nil {
		return nil, s.err(err)
	}
	util.SetHeader(ctx, http.StatusNoContent)
	return &v1.DeleteCACertificateResponse{}, nil
}

func (s *CACertificateService) ListCACertificates(ctx context.Context,
	req *v1.ListCACertificatesRequest) (*v1.ListCACertificatesResponse, error) {
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	list := resource.NewList(resource.TypeCACertificate)
	listOptFns, err := listOptsFromReq(req.Page)
	if err != nil {
		return nil, s.err(util.ErrClient{Message: err.Error()})
	}
	if err := db.List(ctx, list, listOptFns...); err != nil {
		return nil, s.err(err)
	}
	return &v1.ListCACertificatesResponse{
		Items: caCertificatesFromObjects(list.GetAll()),
		Page:  getPaginationResponse(list.GetTotalCount(), list.GetNextPage()),
	}, nil
}

func (s *CACertificateService) err(err error) error {
	return util.HandleErr(s.logger, err)
}

func caCertificatesFromObjects(objects []model.Object) []*pb.CACertificate {
	res := make([]*pb.CACertificate, 0, len(objects))
	for _, object := range objects {
		caCert, ok := object.Resource().(*pb.CACertificate)
		if !ok {
			panic(fmt.Sprintf("expected type '%T' but got '%T'",
				&pb.CACertificate{}, object.Resource()))
		}
		res = append(res, caCert)
	}
	return res
}
