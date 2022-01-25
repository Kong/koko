package ws

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSumHash(t *testing.T) {
	t.Run("hash conversion succeeds", func(t *testing.T) {
		s := "99f6b391347bcb6d052d5ba59e9098ed"
		sumWithHash, err := truncateHash(s)
		require.NoError(t, err)
		require.Len(t, sumWithHash, 32)
		require.Equal(t, s, sumWithHash.String())
	})
	t.Run("hash conversion with string > 32 bytes errors", func(t *testing.T) {
		s := "99f6b391347bcb6d052d5ba59e9098ed12345"
		sumWithHash, err := truncateHash(s)
		require.Error(t, err)
		require.Len(t, sumWithHash, 32)
		require.Equal(t, sum{}, sumWithHash)
	})

	t.Run("hash conversion with chars not matching regex errors", func(t *testing.T) {
		s := "99f6b39137bcb6d052d5ba59e_098e?"
		sumWithHash, err := truncateHash(s)
		require.Error(t, err)
		require.Len(t, sumWithHash, 32)
		require.Equal(t, sum{}, sumWithHash)
	})
}
