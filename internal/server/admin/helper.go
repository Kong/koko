package admin

import (
	"fmt"

	pbModel "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/store"
)

func validateListOptions(listOpts *pbModel.Pagination) error {
	if listOpts.Page < 0 {
		return fmt.Errorf("invalid page '%d', page must be > 0", listOpts.Page)
	}
	if listOpts.Size < 0 || listOpts.Size > store.MaxPageSize {
		return fmt.Errorf("invalid page_size '%d', must be within range [1 - 1000]", listOpts.Size)
	}
	return nil
}

func listOptsFromReq(listOpts *pbModel.Pagination) ([]store.ListOptsFunc, error) {
	if listOpts == nil {
		return []store.ListOptsFunc{}, nil
	}
	err := validateListOptions(listOpts)
	if err != nil {
		return nil, err
	}
	pageNumOption := store.ListWithPageNum(int(listOpts.Page))

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
		TotalCount: int32(totalCount),
		NextPage:   int32(nextPage),
	}
}
