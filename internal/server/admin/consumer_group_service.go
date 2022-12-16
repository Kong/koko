package admin

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	pb "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/server/util"
	"github.com/kong/koko/internal/store"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type ConsumerGroupService struct {
	v1.UnimplementedConsumerGroupServiceServer
	CommonOpts
}

func (s *ConsumerGroupService) GetConsumerGroup(ctx context.Context,
	req *v1.GetConsumerGroupRequest,
) (*v1.GetConsumerGroupResponse, error) {
	if err := s.validateRequest(req); err != nil {
		return nil, s.err(ctx, err)
	}
	db, err := s.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	result := resource.NewConsumerGroup()
	if err := getEntityByIDOrName(
		ctx,
		req.Id,
		result,
		store.GetByName(req.Id),
		db,
		s.logger(ctx),
	); err != nil {
		return nil, s.err(ctx, err)
	}
	return &v1.GetConsumerGroupResponse{
		Item: result.ConsumerGroup,
	}, nil
}

func (s *ConsumerGroupService) CreateConsumerGroup(ctx context.Context,
	req *v1.CreateConsumerGroupRequest,
) (*v1.CreateConsumerGroupResponse, error) {
	db, err := s.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	res := resource.NewConsumerGroup()
	res.ConsumerGroup = req.Item
	if err := db.Create(ctx, res); err != nil {
		return nil, s.err(ctx, err)
	}
	util.SetHeader(ctx, http.StatusCreated)
	return &v1.CreateConsumerGroupResponse{
		Item: res.ConsumerGroup,
	}, nil
}

func (s *ConsumerGroupService) UpsertConsumerGroup(ctx context.Context,
	req *v1.UpsertConsumerGroupRequest,
) (*v1.UpsertConsumerGroupResponse, error) {
	if err := s.validateRequest(req); err != nil {
		return nil, s.err(ctx, err)
	}
	db, err := s.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	res := resource.NewConsumerGroup()
	res.ConsumerGroup = req.Item
	if err := db.Upsert(ctx, res); err != nil {
		return nil, s.err(ctx, err)
	}
	return &v1.UpsertConsumerGroupResponse{
		Item: res.ConsumerGroup,
	}, nil
}

func (s *ConsumerGroupService) DeleteConsumerGroup(ctx context.Context,
	req *v1.DeleteConsumerGroupRequest,
) (*v1.DeleteConsumerGroupResponse, error) {
	if _, err := uuid.Parse(req.Id); err != nil {
		return nil, s.err(ctx, validation.Error{Errs: []*pb.ErrorDetail{{
			Field:    "id",
			Type:     pb.ErrorType_ERROR_TYPE_FIELD,
			Messages: []string{"'id' is not a valid UUID"},
		}}})
	}

	if err := s.validateRequest(req); err != nil {
		return nil, s.err(ctx, err)
	}
	db, err := s.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	err = db.Delete(ctx, store.DeleteByID(req.Id),
		store.DeleteByType(resource.TypeConsumerGroup))
	if err != nil {
		return nil, s.err(ctx, err)
	}
	util.SetHeader(ctx, http.StatusNoContent)
	return &v1.DeleteConsumerGroupResponse{}, nil
}

func (s *ConsumerGroupService) ListConsumerGroups(ctx context.Context,
	req *v1.ListConsumerGroupsRequest,
) (*v1.ListConsumerGroupsResponse, error) {
	db, err := s.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}
	list := resource.NewList(resource.TypeConsumerGroup)
	listOptFns, err := ListOptsFromReq(req.Page)
	if err != nil {
		return nil, s.err(ctx, err)
	}
	if err := db.List(ctx, list, listOptFns...); err != nil {
		return nil, s.err(ctx, err)
	}
	return &v1.ListConsumerGroupsResponse{
		Items: consumerGroupsFromObjects(list.GetAll()),
		Page:  getPaginationResponse(list.GetTotalCount(), list.GetNextPage()),
	}, nil
}

func (s *ConsumerGroupService) ListConsumerGroupMembers(
	ctx context.Context,
	req *v1.ListConsumerGroupMembersRequest,
) (*v1.ListConsumerGroupMembersResponse, error) {
	if err := s.validateRequest(req); err != nil {
		return nil, s.err(ctx, err)
	}

	db, err := s.getDB(ctx, req.Cluster)
	if err != nil {
		return nil, err
	}

	result := resource.NewConsumerGroup()
	if err := getEntityByIDOrName(
		ctx,
		req.Id,
		result,
		store.GetByName(req.Id),
		db,
		s.logger(ctx),
	); err != nil {
		return nil, s.err(ctx, err)
	}

	listFn := []store.ListOptsFunc{store.ListReverseFor(resource.TypeConsumerGroup, result.ID())}
	list := resource.NewList(resource.TypeConsumer)
	listOptFns, err := ListOptsFromReq(req.Page)
	if err != nil {
		return nil, s.err(ctx, err)
	}
	listFn = append(listFn, listOptFns...)

	if err := db.List(ctx, list, listFn...); err != nil {
		return nil, s.err(ctx, err)
	}

	return &v1.ListConsumerGroupMembersResponse{
		Items: consumersFromObjects(list.GetAll()),
		Page:  getPaginationResponse(list.GetTotalCount(), list.GetNextPage()),
	}, nil
}

func (s *ConsumerGroupService) CreateConsumerGroupMember(
	ctx context.Context,
	req *v1.CreateConsumerGroupMemberRequest,
) (*v1.CreateConsumerGroupMemberResponse, error) {
	if err := s.validateRequest(req); err != nil {
		return nil, s.err(ctx, err)
	}
	res, err := s.manageMembers(ctx, req)
	if err != nil {
		return nil, err
	}
	util.SetHeader(ctx, http.StatusCreated)
	return &v1.CreateConsumerGroupMemberResponse{
		Item: res.ConsumerGroup,
	}, nil
}

func (s *ConsumerGroupService) DeleteConsumerGroupMember(
	ctx context.Context,
	req *v1.DeleteConsumerGroupMemberRequest,
) (*v1.DeleteConsumerGroupMemberResponse, error) {
	if err := s.validateRequest(req); err != nil {
		return nil, s.err(ctx, err)
	}
	if _, err := s.manageMembers(ctx, req); err != nil {
		return nil, err
	}
	util.SetHeader(ctx, http.StatusNoContent)
	return &v1.DeleteConsumerGroupMemberResponse{}, nil
}

func (s *ConsumerGroupService) manageMembers(
	ctx context.Context,
	req interface {
		GetCluster() *pb.RequestCluster
		GetConsumerId() string
		GetConsumerGroupId() string
	},
) (*resource.ConsumerGroup, error) {
	consumerID, consumerGroupID := req.GetConsumerId(), req.GetConsumerGroupId()

	for key, value := range map[string]string{
		"consumer_group_id": consumerGroupID,
		"consumer_id":       consumerID,
	} {
		if _, err := uuid.Parse(value); err != nil {
			return nil, s.err(ctx, validation.Error{Errs: []*pb.ErrorDetail{{
				Field:    key,
				Type:     pb.ErrorType_ERROR_TYPE_FIELD,
				Messages: []string{fmt.Sprintf("'%s' is not a valid UUID", key)},
			}}})
		}
	}

	db, err := s.getDB(ctx, req.GetCluster())
	if err != nil {
		return nil, err
	}

	consumerGroup := resource.ConsumerGroup{
		ConsumerGroup: &pb.ConsumerGroup{Id: consumerGroupID},
	}
	if _, ok := req.(*v1.CreateConsumerGroupMemberRequest); ok {
		consumerGroup.MemberIDsToAdd = []string{consumerID}
	} else {
		consumerGroup.MemberIDsToRemove = []string{consumerID}
	}

	if err := db.UpdateForeignKeys(ctx, &consumerGroup); err != nil {
		return nil, s.err(ctx, err)
	}

	return &consumerGroup, nil
}

func (s *ConsumerGroupService) err(ctx context.Context, err error) error {
	return util.HandleErr(ctx, s.logger(ctx), err)
}

func (s *ConsumerGroupService) logger(ctx context.Context) *zap.Logger {
	return util.LoggerFromContext(ctx).With(s.loggerFields...)
}

func consumerGroupsFromObjects(objects []model.Object) []*pb.ConsumerGroup {
	res := make([]*pb.ConsumerGroup, len(objects))
	for i, object := range objects {
		var ok bool
		if res[i], ok = object.Resource().(*pb.ConsumerGroup); !ok {
			panic(fmt.Sprintf("expected type '%T' but got '%T'",
				&pb.ConsumerGroup{}, object.Resource()))
		}
	}
	return res
}

func (s *ConsumerGroupService) validateRequest(req proto.Message) error {
	errs := make([]*pb.ErrorDetail, 0)

	if req, ok := req.(interface{ GetId() string }); ok {
		if id := req.GetId(); id == "" {
			errs = append(errs, &pb.ErrorDetail{
				Field:    "id",
				Type:     pb.ErrorType_ERROR_TYPE_FIELD,
				Messages: []string{"missing properties: 'id'"},
			})
		}
	}

	if req, ok := req.(interface{ GetConsumerId() string }); ok {
		if consumerID := req.GetConsumerId(); consumerID == "" {
			errs = append(errs, &pb.ErrorDetail{
				Field:    "consumer_id",
				Type:     pb.ErrorType_ERROR_TYPE_FIELD,
				Messages: []string{"missing properties: 'consumer_id'"},
			})
		}
	}

	if req, ok := req.(interface{ GetConsumerGroupId() string }); ok {
		if consumerGroupID := req.GetConsumerGroupId(); consumerGroupID == "" {
			errs = append(errs, &pb.ErrorDetail{
				Field:    "consumer_group_id",
				Type:     pb.ErrorType_ERROR_TYPE_FIELD,
				Messages: []string{"missing properties: 'consumer_group_id'"},
			})
		}
	}

	if len(errs) > 0 {
		return validation.Error{Errs: errs}
	}

	return nil
}
