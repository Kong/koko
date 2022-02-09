package admin

import (
	pbModel "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/store"
)

func listOptsFromReq(listOpts *pbModel.Pagination) ([]store.ListOptsFunc, error) {
	if listOpts == nil {
		return []store.ListOptsFunc{}, nil
	}
	pageNumOption, err := store.ListWithPageNum(int(listOpts.Page))
	if err != nil {
		return nil, err
	}
	pageSizeOption, err := store.ListWithPageSize(int(listOpts.Size))
	if err != nil {
		return nil, err
	}

	listOptFns := []store.ListOptsFunc{
		pageNumOption,
		pageSizeOption,
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
