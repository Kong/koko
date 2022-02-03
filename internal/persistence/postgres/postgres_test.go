package postgres

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListQuery(t *testing.T) {
	query := listQueryPaging("foo/bar/42", 20, 0)
	require.Equal(t, "SELECT key, value, COUNT(*) OVER() AS full_count FROM "+
		"store WHERE key LIKE 'foo/bar/42%' ORDER BY key LIMIT 20 OFFSET 0;",
		query)
}
