package postgres

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListQuery(t *testing.T) {
	query := listQuery("foo/bar/42")
	require.Equal(t, "SELECT * FROM store WHERE key LIKE 'foo/bar/42%';",
		query)
}
