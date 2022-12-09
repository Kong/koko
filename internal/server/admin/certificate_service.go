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

type CertificateService struct {
	v1.UnimplementedCertificateServiceServer
	CommonOpts
}

func (s *CertificateService) GetCertificate(ctx context.Context,
	req *v1.GetCertificateRequest,
) (*v1.GetCertificateResponse, error) {
	if req.Id == "" {
		return nil, s.err(ctx, util.ErrClient{Message: "required ID is missing"})
	}
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	result := resource.NewCertificate()
	s.logger(ctx).With(zap.String("id", req.Id)).Debug("reading certificate by id")
	err = db.Read(ctx, result, store.GetByID(req.Id))
	if err != nil {
		return nil, s.err(ctx, err)
	}
	// retrieve the snis currently using this certificate
	if err := addSNIMetadata(ctx, result.Certificate, db); err != nil {
		return nil, err
	}
	return &v1.GetCertificateResponse{
		Item: result.Certificate,
	}, nil
}

func (s *CertificateService) CreateCertificate(ctx context.Context,
	req *v1.CreateCertificateRequest,
) (*v1.CreateCertificateResponse, error) {
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	res := resource.NewCertificate()
	res.Certificate = req.Item
	if err := db.Create(ctx, res); err != nil {
		return nil, s.err(ctx, err)
	}
	util.SetHeader(ctx, http.StatusCreated)
	return &v1.CreateCertificateResponse{
		Item: res.Certificate,
	}, nil
}

func (s *CertificateService) UpsertCertificate(ctx context.Context,
	req *v1.UpsertCertificateRequest,
) (*v1.UpsertCertificateResponse, error) {
	if err := validUUID(req.Item.Id); err != nil {
		return nil, s.err(ctx, err)
	}
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	res := resource.NewCertificate()
	res.Certificate = req.Item
	if err := db.Upsert(ctx, res); err != nil {
		return nil, s.err(ctx, err)
	}
	return &v1.UpsertCertificateResponse{
		Item: res.Certificate,
	}, nil
}

func (s *CertificateService) DeleteCertificate(ctx context.Context,
	req *v1.DeleteCertificateRequest,
) (*v1.DeleteCertificateResponse, error) {
	if err := validUUID(req.Id); err != nil {
		return nil, s.err(ctx, err)
	}
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	err = db.Delete(ctx, store.DeleteByID(req.Id),
		store.DeleteByType(resource.TypeCertificate))
	if err != nil {
		return nil, s.err(ctx, err)
	}
	util.SetHeader(ctx, http.StatusNoContent)
	return &v1.DeleteCertificateResponse{}, nil
}

func (s *CertificateService) ListCertificates(ctx context.Context,
	req *v1.ListCertificatesRequest,
) (*v1.ListCertificatesResponse, error) {
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	list := resource.NewList(resource.TypeCertificate)
	listOptFns, err := ListOptsFromReq(req.Page)
	if err != nil {
		return nil, s.err(ctx, err)
	}
	if err := db.List(ctx, list, listOptFns...); err != nil {
		return nil, s.err(ctx, err)
	}
	allCerts, err := certificatesFromObjects(ctx, db, list.GetAll())
	if err != nil {
		return nil, err
	}
	return &v1.ListCertificatesResponse{
		Items: allCerts,
		Page:  getPaginationResponse(list.GetTotalCount(), list.GetNextPage()),
	}, nil
}

func (s *CertificateService) err(ctx context.Context, err error) error {
	return util.HandleErr(ctx, s.logger(ctx), err)
}

func (s *CertificateService) logger(ctx context.Context) *zap.Logger {
	return util.LoggerFromContext(ctx).With(s.loggerFields...)
}

func addSNIMetadata(ctx context.Context, cert *pb.Certificate, db store.Store) error {
	snis, err := getAllSNIsForCertificate(ctx, cert.Id, db)
	if err != nil {
		return fmt.Errorf(
			"retrieving snis used by certificate '%s': %w",
			cert.Id, err,
		)
	}
	if len(snis) > 0 {
		if cert.Metadata == nil {
			cert.Metadata = &pb.CertificateMetadata{Snis: make([]string, len(snis))}
		} else if cert.Metadata.Snis == nil {
			cert.Metadata.Snis = make([]string, len(snis))
		}
	}
	for i, sni := range snis {
		cert.Metadata.Snis[i] = sni.Name
	}
	return nil
}

func certificatesFromObjects(
	ctx context.Context, db store.Store, objects []model.Object,
) ([]*pb.Certificate, error) {
	res := make([]*pb.Certificate, 0, len(objects))
	for _, object := range objects {
		cert, ok := object.Resource().(*pb.Certificate)
		if !ok {
			panic(fmt.Sprintf("expected type '%T' but got '%T'",
				&pb.Certificate{}, object.Resource()))
		}
		// retrieve the snis currently using this certificate
		if err := addSNIMetadata(ctx, cert, db); err != nil {
			return nil, err
		}
		res = append(res, cert)
	}
	return res, nil
}

func getAllSNIsForCertificate(
	ctx context.Context, certID string, db store.Store,
) ([]*pb.SNI, error) {
	var snis []*pb.SNI
	var page int32 = 1
	for {
		res, page, err := getSNIsPage(ctx, certID, page, db)
		if err != nil {
			return nil, err
		}
		snis = append(snis, res...)
		if page == 0 {
			break
		}
	}
	return snis, nil
}

func getSNIsPage(
	ctx context.Context, certID string, page int32, db store.Store,
) ([]*pb.SNI, int, error) {
	listFn := []store.ListOptsFunc{}
	listFn = append(listFn, store.ListFor(resource.TypeCertificate, certID))
	listOptFns, err := ListOptsFromReq(&pb.PaginationRequest{
		Number: page,
		Size:   store.MaxPageSize,
	})
	if err != nil {
		return nil, 0, err
	}
	listFn = append(listFn, listOptFns...)
	list := resource.NewList(resource.TypeSNI)
	if err := db.List(ctx, list, listFn...); err != nil {
		return nil, 0, err
	}
	return snisFromObjects(list.GetAll()), list.GetNextPage(), nil
}
