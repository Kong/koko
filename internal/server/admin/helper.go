package admin

import (
	"fmt"

	pbModel "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
)

func validateListOptions(listOpts *pbModel.ListOpts) error {
	if listOpts.Page < 1 {
		return fmt.Errorf("invalid page '%d', page must be >= 1", listOpts.Page)
	}
	if listOpts.PageSize < 1 || listOpts.PageSize > 1000 {
		return fmt.Errorf("invalid page_size '%d', must be within range [1 - 1000]", listOpts.PageSize)
	}
	return nil
}
