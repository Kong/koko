package validation

import (
	"fmt"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
)

type Error struct {
	Errs []*v1.ErrorDetail
}

func (v Error) Error() string {
	// Not used so not helpful error
	// this exists primarily to satisfy the error interface only
	// callers are expected to type cast and inspect Errs
	return fmt.Sprintf("%d errors", len(v.Errs))
}
