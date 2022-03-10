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

type SNIService struct {
	v1.UnimplementedSNIServiceServer
	CommonOpts
}

func (s *SNIService) GetSNI(ctx context.Context, req *v1.GetSNIRequest) (*v1.GetSNIResponse, error) {
	if req.Id == "" {
		return nil, s.err(util.ErrClient{Message: "required ID is missing"})
	}
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	result := resource.NewSNI()
	s.logger.With(zap.String("id", req.Id)).Debug("reading sni by id")
	err = db.Read(ctx, result, store.GetByID(req.Id))
	if err != nil {
		return nil, s.err(err)
	}
	return &v1.GetSNIResponse{
		Item: result.SNI,
	}, nil
}

func (s *SNIService) CreateSNI(ctx context.Context, req *v1.CreateSNIRequest) (*v1.CreateSNIResponse, error) {
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	res := resource.NewSNI()
	res.SNI = req.Item
	if err := db.Create(ctx, res); err != nil {
		return nil, s.err(err)
	}
	util.SetHeader(ctx, http.StatusCreated)
	return &v1.CreateSNIResponse{
		Item: res.SNI,
	}, nil
}

func (s *SNIService) UpsertSNI(ctx context.Context, req *v1.UpsertSNIRequest) (*v1.UpsertSNIResponse, error) {
	if err := validUUID(req.Item.Id); err != nil {
		return nil, s.err(err)
	}
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	res := resource.NewSNI()
	res.SNI = req.Item
	if err := db.Upsert(ctx, res); err != nil {
		return nil, s.err(err)
	}
	return &v1.UpsertSNIResponse{
		Item: res.SNI,
	}, nil
}

func (s *SNIService) DeleteSNI(ctx context.Context, req *v1.DeleteSNIRequest) (*v1.DeleteSNIResponse, error) {
	if err := validUUID(req.Id); err != nil {
		return nil, s.err(err)
	}
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	if err = db.Delete(ctx, store.DeleteByID(req.Id),
		store.DeleteByType(resource.TypeSNI)); err != nil {
		return nil, s.err(err)
	}
	util.SetHeader(ctx, http.StatusNoContent)
	return &v1.DeleteSNIResponse{}, nil
}

func (s *SNIService) ListSNIs(ctx context.Context, req *v1.ListSNIsRequest) (*v1.ListSNIsResponse, error) {
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	list := resource.NewList(resource.TypeSNI)
	listOptFns, err := listOptsFromReq(req.Page)
	if err != nil {
		return nil, s.err(util.ErrClient{Message: err.Error()})
	}
	if err := db.List(ctx, list, listOptFns...); err != nil {
		return nil, s.err(err)
	}
	return &v1.ListSNIsResponse{
		Items: snisFromObjects(list.GetAll()),
		Page:  getPaginationResponse(list.GetTotalCount(), list.GetNextPage()),
	}, nil
}

func (s *SNIService) err(err error) error {
	return util.HandleErr(s.logger, err)
}

func snisFromObjects(objects []model.Object) []*pb.SNI {
	res := make([]*pb.SNI, 0, len(objects))
	for _, object := range objects {
		sni, ok := object.Resource().(*pb.SNI)
		if !ok {
			panic(fmt.Sprintf("expected type '%T' but got '%T'",
				&pb.SNI{}, object.Resource()))
		}
		res = append(res, sni)
	}
	return res
}
