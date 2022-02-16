package admin

import (
	"fmt"

	"github.com/google/uuid"
	pbModel "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/store"
)

func validateListOptions(listOpts *pbModel.PaginationRequest) error {
	if listOpts.Number < 0 {
		return fmt.Errorf("invalid page number '%d', page must be > 0", listOpts.Number)
	}
	if listOpts.Size < 0 || listOpts.Size > store.MaxPageSize {
		return fmt.Errorf("invalid page_size '%d', must be within range [1 - %d]", listOpts.Size, store.MaxPageSize)
	}
	return nil
}

func validUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
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
