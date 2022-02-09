package admin

import (
	"fmt"
	"testing"

	pbModel "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/store"
	"github.com/stretchr/testify/require"
)

func Test_listOptsFromRequest(t *testing.T) {
	t.Run("Page 1, Size 1000 is successful", func(t *testing.T) {
		p := &pbModel.Pagination{Page: 1, Size: 1000}
		listOptFns, err := listOptsFromReq(p)
		require.NoError(t, err)
		require.Len(t, listOptFns, 2)
		listOpts := &store.ListOpts{}
		for _, fn := range listOptFns {
			fn(listOpts)
		}
		require.Equal(t, 1, listOpts.Page)
		require.Equal(t, 1000, listOpts.PageSize)
	})
	t.Run("Page 0, Size 10 fails with error", func(t *testing.T) {
		p := &pbModel.Pagination{Page: 0, Size: 10}
		_, err := listOptsFromReq(p)
		expectedErr := fmt.Errorf("invalid page '%d', page must be >= 1", 0)
		require.Error(t, expectedErr, err)
	})
	t.Run("Page 1, Size 0 fails with error", func(t *testing.T) {
		p := &pbModel.Pagination{Page: 1, Size: 0}
		_, err := listOptsFromReq(p)
		expectedErr := fmt.Errorf("invalid page_size '%d', must be within range [1 - 1000]", 0)
		require.Error(t, expectedErr, err)
	})
	t.Run("Page 1, Size 1001 fails with error", func(t *testing.T) {
		p := &pbModel.Pagination{Page: 1, Size: 1001}
		_, err := listOptsFromReq(p)
		expectedErr := fmt.Errorf("invalid page_size '%d', must be within range [1 - 1000]", 1001)
		require.Error(t, expectedErr, err)
	})
}
