package store

import (
	"context"

	"github.com/kong/koko/internal/persistence"
)

// returns the full list despite pagination.
func getFullList(ctx context.Context, tx persistence.Tx, keyPrefix string) (persistence.ListResult, error) {
	listResult, err := tx.List(ctx, keyPrefix, persistence.NewDefaultListOpts())
	if err != nil {
		return persistence.ListResult{}, err
	}
	if listResult.KVList == nil {
		listResult.KVList = []*persistence.KVResult{}
	}
	tCount := listResult.TotalCount
	for kvl := len(listResult.KVList); (kvl > 0) && (tCount > kvl); {
		currListRes, err := tx.List(ctx, keyPrefix, persistence.NewDefaultListOpts())
		if err != nil {
			return persistence.ListResult{}, err
		}
		listResult.KVList = append(listResult.KVList, currListRes.KVList...)
	}
	return listResult, nil
}
