package store

import (
	"context"

	"github.com/kong/koko/internal/persistence"
)

// getFullList returns the full list despite pagination.
func getFullList(ctx context.Context, tx persistence.Tx, keyPrefix string) (persistence.ListResult, error) {
	listOptions := persistence.NewDefaultListOpts()
	listOptions.Limit = persistence.MaxLimit
	listResult, err := tx.List(ctx, keyPrefix, listOptions)
	if err != nil {
		return persistence.ListResult{}, err
	}
	tCount := listResult.TotalCount
	for kvl := len(listResult.KVList); (kvl > 0) && (tCount > kvl); kvl = len(listResult.KVList) {
		listOptions.Offset += listOptions.Limit
		currListRes, err := tx.List(ctx, keyPrefix, listOptions)
		if err != nil {
			return persistence.ListResult{}, err
		}
		listResult.KVList = append(listResult.KVList, currListRes.KVList...)
	}
	return listResult, nil
}

// getPersistenceListOptions Converts store Page and Page Size to Limit and Offset.
func getPersistenceListOptions(opts *ListOpts) *persistence.ListOpts {
	return &persistence.ListOpts{
		Limit:  opts.PageSize,
		Offset: toOffset(opts),
	}
}

func toOffset(opts *ListOpts) int {
	if opts.Page == 1 || opts.Page == 0 {
		return 0
	}
	return opts.PageSize * (opts.Page - 1)
}

// toLastPage calculates the last page based on totalItems and pageSize.
func toLastPage(pageSize int, totalItems int) int {
	if pageSize >= totalItems {
		return 1
	}
	if totalItems%pageSize == 0 {
		return totalItems / pageSize
	}
	return (totalItems / pageSize) + 1
}
