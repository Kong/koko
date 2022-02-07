package admin

import (
	"encoding/base64"
	"fmt"
	"strconv"

	pbModel "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/persistence"
)

func validateListOptions(listOpts *pbModel.ListOpts) error {
	if listOpts.Page < 1 {
		return fmt.Errorf("invalid page '%d', page must be >= 1", listOpts.Page)
	}
	if listOpts.PageSize < 1 || listOpts.PageSize > persistence.MaxPageSize {
		return fmt.Errorf("invalid page_size '%d', must be within range [1 - 1000]", listOpts.PageSize)
	}
	return nil
}

func getOffset(totalCount int) string {
	if totalCount == 0 {
		return ""
	}
	// Converting to string first offset may not be int
	return base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(totalCount)))
}
