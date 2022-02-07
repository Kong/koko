package admin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/google/uuid"
	pbModel "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/persistence"
	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/server/util"
	"github.com/kong/koko/internal/store"
	"go.uber.org/zap"
)

type StatusService struct {
	v1.UnimplementedStatusServiceServer
	CommonOpts
}

func (s *StatusService) GetStatus(ctx context.Context,
	req *v1.GetStatusRequest) (*v1.GetStatusResponse, error) {
	if req.Id == "" {
		return nil, s.err(util.ErrClient{Message: "required ID is missing"})
	}
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	result := resource.NewStatus()
	s.logger.With(zap.String("id", req.Id)).Debug("reading route by id")
	err = db.Read(ctx, result, store.GetByID(req.Id))
	if err != nil {
		return nil, s.err(err)
	}
	return &v1.GetStatusResponse{
		Item: result.Status,
	}, nil
}

func (s *StatusService) DeleteStatus(ctx context.Context,
	req *v1.DeleteStatusRequest) (*v1.DeleteStatusResponse, error) {
	if req.Id == "" {
		return nil, s.err(util.ErrClient{Message: "required ID is missing"})
	}
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	err = db.Delete(ctx, store.DeleteByID(req.Id),
		store.DeleteByType(resource.TypeStatus))
	if err != nil {
		return nil, s.err(err)
	}
	util.SetHeader(ctx, http.StatusNoContent)
	return &v1.DeleteStatusResponse{}, nil
}

func (s *StatusService) ListStatuses(ctx context.Context,
	req *v1.ListStatusesRequest) (*v1.ListStatusesResponse, error) {
	db, err := s.CommonOpts.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}

	if req.RefType != "" && req.RefId != "" {
		return s.listStatusForEntity(ctx, db, req.RefType, req.RefId)
	}

	listOpts := req.ListOptions
	if listOpts == nil {
		listOpts = &pbModel.ListOpts{Page: persistence.DefaultPage, PageSize: persistence.DefaultPageSize}
	}
	// Validate what we got
	if err = validateListOptions(listOpts); err != nil {
		return nil, s.err(util.ErrClient{Message: err.Error()})
	}

	list := resource.NewList(resource.TypeStatus)
	if err := db.List(ctx, list, store.ListWithPageNum(int(listOpts.Page)),
		store.ListWithPageSize(int(listOpts.PageSize))); err != nil {
		return nil, s.err(err)
	}

	return &v1.ListStatusesResponse{
		Items:  statusesFromObjects(list.GetAll()),
		Offset: strconv.Itoa(persistence.ToLastPage(int(listOpts.PageSize), list.GetCount())),
	}, nil
}

func (s *StatusService) err(err error) error {
	return util.HandleErr(s.logger, err)
}

// TODO(hbagdi): change this regex to either include '-' or '_'.
var typeRegex = regexp.MustCompile("^[a-z]{1,16}$")

func validateRefs(refType, refID string) error {
	if _, err := uuid.Parse(refID); err != nil {
		return util.ErrClient{Message: "invalid id"}
	}
	if !typeRegex.MatchString(refType) {
		return util.ErrClient{Message: "invalid type"}
	}
	return nil
}

func (s *StatusService) listStatusForEntity(ctx context.Context,
	db store.Store, refType, refID string) (*v1.ListStatusesResponse, error) {
	if err := validateRefs(refType, refID); err != nil {
		return nil, err
	}

	result := resource.NewStatus()
	err := db.Read(ctx, result, store.GetByIndex("ctx_ref",
		model.MultiValueIndex(refType, refID)))
	if err == nil {
		return &v1.ListStatusesResponse{
			Items: []*pbModel.Status{result.Status},
		}, nil
	} else if errors.Is(err, store.ErrNotFound) {
		return &v1.ListStatusesResponse{
			Items: []*pbModel.Status{},
		}, nil
	}
	return nil, s.err(err)
}

func statusesFromObjects(objects []model.Object) []*pbModel.Status {
	res := make([]*pbModel.Status, 0, len(objects))
	for _, object := range objects {
		route, ok := object.Resource().(*pbModel.Status)
		if !ok {
			panic(fmt.Sprintf("expected type '%T' but got '%T'",
				&pbModel.Status{}, object.Resource()))
		}
		res = append(res, route)
	}
	return res
}
