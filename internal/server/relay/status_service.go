package relay

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	adminModel "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	relay "github.com/kong/koko/internal/gen/grpc/kong/relay/service/v1"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/server/util"
	"github.com/kong/koko/internal/store"
	"go.uber.org/zap"
	"google.golang.org/grpc/status"
)

type StatusServiceOpts struct {
	StoreLoader util.StoreLoader
	Logger      *zap.Logger
}

func NewStatusService(opts StatusServiceOpts) *StatusService {
	res := &StatusService{
		storeLoader: opts.StoreLoader,
		logger:      opts.Logger,
	}
	return res
}

type StatusService struct {
	relay.UnimplementedStatusServiceServer
	storeLoader util.StoreLoader
	logger      *zap.Logger
}

func (s StatusService) UpdateStatus(ctx context.Context,
	req *relay.UpdateStatusRequest,
) (*relay.UpdateStatusResponse, error) {
	err := validateRef(req.Item.ContextReference)
	if err != nil {
		return nil, util.ErrClient{Message: err.Error()}
	}
	db, err := s.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}

	res := resource.NewStatus()
	res.Status = req.Item

	refType := req.Item.ContextReference.Type
	refID := req.Item.ContextReference.Id
	id, err := s.getCurrentStatusID(ctx, db, refType, refID)
	if err != nil {
		return nil, err
	}
	if id != "" {
		res.Status.Id = id
	}

	if err := db.Upsert(ctx, res); err != nil {
		return nil, err
	}
	return &relay.UpdateStatusResponse{
		Item: res.Status,
	}, nil
}

func (s StatusService) getCurrentStatusID(ctx context.Context, db store.Store,
	refType, refID string) (string,
	error,
) {
	currentStatus := resource.NewStatus()
	err := db.Read(ctx, currentStatus,
		store.GetByIndex("ctx_ref", model.MultiValueIndex(refType, refID)),
	)
	switch err {
	case nil:
		return currentStatus.Status.Id, nil
	default:
		if errors.Is(err, store.ErrNotFound) {
			return "", nil
		}
		return "", err
	}
}

func (s StatusService) ClearStatus(ctx context.Context,
	req *relay.ClearStatusRequest,
) (*relay.ClearStatusResponse, error) {
	err := validateRef(req.ContextReference)
	if err != nil {
		return nil, util.ErrClient{Message: err.Error()}
	}
	typ := req.ContextReference.Type
	id := req.ContextReference.Id
	s.logger.
		With(zap.String("type", typ), zap.String("id", id)).
		Debug("clear status invoked")

	db, err := s.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	statusID, err := s.getCurrentStatusID(ctx, db, typ, id)
	if err != nil {
		return nil, err
	}
	if statusID == "" {
		return &relay.ClearStatusResponse{}, nil
	}
	err = db.Delete(ctx,
		store.DeleteByType(resource.TypeStatus),
		store.DeleteByID(statusID),
	)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return &relay.ClearStatusResponse{}, nil
		}
		return nil, err
	}
	return &relay.ClearStatusResponse{}, nil
}

func (s StatusService) UpdateExpectedHash(ctx context.Context,
	req *relay.UpdateExpectedHashRequest,
) (*relay.UpdateExpectedHashResponse, error) {
	db, err := s.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}

	res := resource.NewHash()
	if req.Hash == "" {
		return nil, fmt.Errorf("invalid hash: '%v'", req.Hash)
	}
	res.Hash.ExpectedHash = req.Hash

	err = db.Upsert(ctx, res)
	if err != nil {
		return nil, err
	}
	return &relay.UpdateExpectedHashResponse{}, nil
}

func (s StatusService) getDB(ctx context.Context,
	cluster *adminModel.RequestCluster,
) (store.Store, error) {
	store, err := s.storeLoader.Load(ctx, cluster)
	if err != nil {
		if storeLoadErr, ok := err.(util.StoreLoadErr); ok {
			return nil, status.Error(storeLoadErr.Code, storeLoadErr.Message)
		}
		return nil, err
	}
	return store, nil
}

func validateRef(ref *adminModel.EntityReference) error {
	if ref == nil {
		return fmt.Errorf("no context reference")
	}
	id := ref.Id
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid id")
	}
	typ := model.Type(ref.Type)
	if _, err := model.NewObject(typ); err != nil {
		return fmt.Errorf("invalid type")
	}
	return nil
}
