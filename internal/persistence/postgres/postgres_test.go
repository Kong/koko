package postgres

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListQuery(t *testing.T) {
	query := listQueryPaging("foo/bar/42", 20, 0)
	require.Equal(t, "SELECT key, value, count(*) OVER() as full_count FROM "+
		"store WHERE key LIKE 'foo/bar/42%' order by key limit 20 offset 0;",
		query)
}
