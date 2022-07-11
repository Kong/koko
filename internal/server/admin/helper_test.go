package admin

import (
	"testing"

	pbModel "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/kong/koko/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ListOptsFromRequest(t *testing.T) {
	t.Run("Page 1, Size 1000 is successful", func(t *testing.T) {
		p := &pbModel.PaginationRequest{Number: 1, Size: 1000}
		listOptFns, err := ListOptsFromReq(p)
		require.NoError(t, err)
		require.Len(t, listOptFns, 2)
		listOpts := &store.ListOpts{}
		for _, fn := range listOptFns {
			fn(listOpts)
		}
		require.Equal(t, 1, listOpts.Page)
		require.Equal(t, 1000, listOpts.PageSize)
	})
	t.Run("Page 0, Size 10 succeeds with default Page", func(t *testing.T) {
		p := &pbModel.PaginationRequest{Number: 0, Size: 10}
		listOptFns, err := ListOptsFromReq(p)
		require.NoError(t, err)
		listOpts := &store.ListOpts{}
		for _, fn := range listOptFns {
			fn(listOpts)
		}
		require.Equal(t, store.DefaultPage, listOpts.Page)
		require.Equal(t, 10, listOpts.PageSize)
	})
	t.Run("Page 1, Size 0 succeeds with default Page Size", func(t *testing.T) {
		p := &pbModel.PaginationRequest{Number: 1, Size: 0}
		listOptFns, err := ListOptsFromReq(p)
		require.NoError(t, err)
		listOpts := &store.ListOpts{}
		for _, fn := range listOptFns {
			fn(listOpts)
		}
		require.Equal(t, 1, listOpts.Page)
		require.Equal(t, store.DefaultPageSize, listOpts.PageSize)
	})
	t.Run("Page 1, Size 1001 fails with error", func(t *testing.T) {
		p := &pbModel.PaginationRequest{Number: 1, Size: 1001}
		_, err := ListOptsFromReq(p)
		require.EqualError(t, err, "invalid page_size '1001', must be within range [1 - 1000]")
	})
	t.Run("setting filter expression", func(t *testing.T) {
		listOptFns, err := ListOptsFromReq(&pbModel.PaginationRequest{Filter: `"tag1" in tags`})
		require.NoError(t, err)
		listOpts := &store.ListOpts{}
		for _, fn := range listOptFns {
			fn(listOpts)
		}
		require.NotNil(t, listOpts.Filter)
	})
	t.Run("setting invalid expression", func(t *testing.T) {
		_, err := ListOptsFromReq(&pbModel.PaginationRequest{Filter: `"tag1" in undefined`})
		assert.Equal(t, validation.Error{
			Errs: []*pbModel.ErrorDetail{{
				Type:     pbModel.ErrorType_ERROR_TYPE_FIELD,
				Field:    "page.filter",
				Messages: []string{"invalid filter expression: undeclared reference to 'undefined'"},
			}},
		}, err)
	})
}
