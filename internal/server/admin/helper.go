package admin

import (
	"context"
	"fmt"
	"regexp"

	"github.com/google/uuid"
	pbModel "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/server/util"
	"github.com/kong/koko/internal/store"
	"go.uber.org/zap"
)

const namePattern = `^[0-9a-zA-Z.\-_~]{1,128}$`

var nameRegex = regexp.MustCompile(namePattern)

func validateListOptions(listOpts *pbModel.PaginationRequest) error {
	if listOpts.Number < 0 {
		return fmt.Errorf("invalid page number '%d', page must be > 0", listOpts.Number)
	}
	if listOpts.Size < 0 || listOpts.Size > store.MaxPageSize {
		return fmt.Errorf("invalid page_size '%d', must be within range [1 - %d]", listOpts.Size, store.MaxPageSize)
	}
	return nil
}

func validUUID(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return util.ErrClient{Message: fmt.Sprintf(" '%v' is not a valid uuid", id)}
	}
	return nil
}

func listOptsFromReq(listOpts *pbModel.PaginationRequest) ([]store.ListOptsFunc, error) {
	if listOpts == nil {
		return []store.ListOptsFunc{}, nil
	}
	err := validateListOptions(listOpts)
	if err != nil {
		return nil, err
	}
	pageNumOption := store.ListWithPageNum(int(listOpts.Number))

	pageSizeOption := store.ListWithPageSize(int(listOpts.Size))

	listOptFns := []store.ListOptsFunc{
		pageNumOption,
		pageSizeOption,
	}
	return listOptFns, nil
}

func getPaginationResponse(totalCount int, nextPage int) *pbModel.PaginationResponse {
	if totalCount == 0 {
		return nil
	}
	return &pbModel.PaginationResponse{
		TotalCount:  int32(totalCount),
		NextPageNum: int32(nextPage),
	}
}

func getEntityByIDOrName(ctx context.Context, idOrName string, entity model.Object, nameOpt store.ReadOptsFunc,
	s store.Store, logger *zap.Logger,
) error {
	logger = util.LoggerWithSpan(ctx, logger)
	if idOrName == "" {
		return util.ErrClient{Message: "required ID is missing"}
	}
	if err := validUUID(idOrName); err == nil {
		logger.With(zap.String("id", idOrName)).Debug(fmt.Sprintf("reading %v by id", entity.Type()))
		err = s.Read(ctx, entity, store.GetByID(idOrName))
		if err != nil {
			return err
		}
		return nil
	}
	if nameRegex.MatchString(idOrName) {
		logger.With(zap.String("name", idOrName)).Debug(fmt.Sprintf("attempting reading %v by name",
			entity.Type()))
		err := s.Read(ctx, entity, nameOpt)
		if err != nil {
			return err
		}
		return nil
	}
	return util.ErrClient{Message: fmt.Sprintf("invalid ID:'%s'", idOrName)}
}
