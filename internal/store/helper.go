package store

import (
	"context"

	"github.com/kong/koko/internal/persistence"
)

func getFullList(ctx context.Context, tx persistence.Tx, keyPrefix string) ([]*persistence.KVResult, error) {
	kvs, err := tx.List(ctx, keyPrefix, persistence.NewDefaultListOpts())
	if err != nil {
		return nil, err
	}
	var tCount int
	if len(kvs) > 0 {
		tCount = kvs[0].TotalCount
	}
	for kvl := len(kvs); (kvl > 0) && (tCount > kvl); {
		currKvs, err := tx.List(ctx, keyPrefix, persistence.NewDefaultListOpts())
		if err != nil {
			return nil, err
		}
		kvs = append(kvs, currKvs...)
	}
	return kvs, nil
}
