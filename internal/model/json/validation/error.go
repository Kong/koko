package validation

import (
	"fmt"
	"sort"
	"strings"

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

// SortErrorDetails sorts a slice of error details for predictability,
// by modifying the inputted slice.
//
// Orders by errors without field names in ascending order first (using
// their messages), and then orders the field names in ascending order.
func SortErrorDetails(errs []*v1.ErrorDetail) {
	sort.Slice(errs, func(i, j int) bool {
		if errs[i].Field == errs[j].Field {
			return strings.Join(errs[i].Messages, "-") < strings.Join(errs[j].Messages, "-")
		}
		return errs[i].Field < errs[j].Field
	})
}
