package admin

import (
	"fmt"

	pbModel "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/store"
)

func validateListOptions(listOpts *pbModel.Pagination) error {
	if listOpts.Page < 1 {
		return fmt.Errorf("invalid page '%d', page must be >= 1", listOpts.Page)
	}
	if listOpts.Size < 1 || listOpts.Size > store.MaxPageSize {
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

	listOptFns := []store.ListOptsFunc{
		store.ListWithPageNum(int(listOpts.Page)),
		store.ListWithPageSize(int(listOpts.Size)),
	}
	return listOptFns, nil
}

func getPagination(totalCount int, nextPage int) *pbModel.PaginationResponse {
	if totalCount == 0 {
		return nil
	}
	return &pbModel.PaginationResponse{
		TotalCount: int32(totalCount),
		NextPage:   int32(nextPage),
	}
}
