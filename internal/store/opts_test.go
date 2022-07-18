package store

import (
	"testing"

	"github.com/kong/koko/internal/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

func TestNewListOpts(t *testing.T) {
	filter := &expr.Expr{}

	t.Run("sets default page size", func(t *testing.T) {
		opts, err := NewListOpts()
		require.NoError(t, err)

		assert.Equal(t, &ListOpts{
			PageSize: DefaultPageSize,
			Page:     DefaultPage,
		}, opts)
	})

	t.Run("with ListWithFilter()", func(t *testing.T) {
		opts, err := NewListOpts(
			ListWithPageNum(10),
			ListWithPageSize(123),
			ListWithFilter(filter),
		)
		require.NoError(t, err)

		assert.Exactly(t, &ListOpts{
			PageSize: 123,
			Page:     10,
			Filter:   filter,
		}, opts)
	})

	t.Run("with ListFor()", func(t *testing.T) {
		opts, err := NewListOpts(
			ListWithPageNum(10),
			ListWithPageSize(123),
			ListFor(resource.TypeConsumer, "ref-id"),
		)
		require.NoError(t, err)

		assert.Exactly(t, &ListOpts{
			PageSize:      123,
			Page:          10,
			ReferenceType: resource.TypeConsumer,
			ReferenceID:   "ref-id",
		}, opts)
	})

	t.Run("with ListWithFilter() & ListFor()", func(t *testing.T) {
		_, err := NewListOpts(
			ListWithFilter(filter),
			ListFor(resource.TypeConsumer, "ref-id"),
		)
		assert.EqualError(
			t,
			err,
			`listing results with a pagination filter is currently unsupported `+
				`when results are scoped to the "consumer" (ID: "ref-id") resource`,
		)
	})
}
