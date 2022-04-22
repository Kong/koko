package admin

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
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
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	result := resource.NewSNI()
	err = getEntityByIDOrName(ctx, req.Id, result, store.GetByName(req.Id), db, s.logger(ctx))
	if err != nil {
		return nil, s.err(ctx, err)
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
		return nil, s.err(ctx, err)
	}
	util.SetHeader(ctx, http.StatusCreated)
	return &v1.CreateSNIResponse{
		Item: res.SNI,
	}, nil
}

func (s *SNIService) UpsertSNI(ctx context.Context, req *v1.UpsertSNIRequest) (*v1.UpsertSNIResponse, error) {
	if err := validUUID(req.Item.Id); err != nil {
		return nil, s.err(ctx, err)
	}
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	res := resource.NewSNI()
	res.SNI = req.Item
	if err := db.Upsert(ctx, res); err != nil {
		return nil, s.err(ctx, err)
	}
	return &v1.UpsertSNIResponse{
		Item: res.SNI,
	}, nil
}

func (s *SNIService) DeleteSNI(ctx context.Context, req *v1.DeleteSNIRequest) (*v1.DeleteSNIResponse, error) {
	if err := validUUID(req.Id); err != nil {
		return nil, s.err(ctx, err)
	}
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	if err = db.Delete(ctx, store.DeleteByID(req.Id),
		store.DeleteByType(resource.TypeSNI)); err != nil {
		return nil, s.err(ctx, err)
	}
	util.SetHeader(ctx, http.StatusNoContent)
	return &v1.DeleteSNIResponse{}, nil
}

func (s *SNIService) ListSNIs(ctx context.Context, req *v1.ListSNIsRequest) (*v1.ListSNIsResponse, error) {
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}

	certID := strings.TrimSpace(req.CertificateId)
	listFn := []store.ListOptsFunc{}
	if len(certID) > 0 {
		if _, err := uuid.Parse(certID); err != nil {
			return nil, s.err(ctx, util.ErrClient{
				Message: fmt.Sprintf("certificate_id '%s' is not a UUID", req.CertificateId),
			})
		}
		listFn = append(listFn, store.ListFor(resource.TypeCertificate, certID))
	}

	list := resource.NewList(resource.TypeSNI)
	listOptFns, err := listOptsFromReq(req.Page)
	if err != nil {
		return nil, s.err(ctx, util.ErrClient{Message: err.Error()})
	}

	listFn = append(listFn, listOptFns...)

	if err := db.List(ctx, list, listFn...); err != nil {
		return nil, s.err(ctx, err)
	}

	return &v1.ListSNIsResponse{
		Items: snisFromObjects(list.GetAll()),
		Page:  getPaginationResponse(list.GetTotalCount(), list.GetNextPage()),
	}, nil
}

func (s *SNIService) err(ctx context.Context, err error) error {
	return util.HandleErr(ctx, s.logger(ctx), err)
}

func (s *SNIService) logger(ctx context.Context) *zap.Logger {
	return util.LoggerFromContext(ctx).With(s.loggerFields...)
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
