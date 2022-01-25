package ws

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSum_truncateHash(t *testing.T) {
	s := "99f6b391347bcb6d052d5ba59e9098ed"
	sumWithHash, err := truncateHash(s)
	require.NoError(t, err)
	require.Len(t, sumWithHash, 32)
	require.Equal(t, s, sumWithHash.String())
}

func TestSum_truncateHashTruncated(t *testing.T) {
	s := "99f6b391347bcb6d052d5ba59e9098ed12345"
	sumWithHash, err := truncateHash(s)
	require.Error(t, err)
	require.Len(t, sumWithHash, 32)
	require.Equal(t, sum{}, sumWithHash)
}

func TestSum_truncateHashNonHashChars(t *testing.T) {
	s := "99f6b391347bcb6d052d5ba59e9098ed?"
	sumWithHash, err := truncateHash(s)
	require.Error(t, err)
	require.Len(t, sumWithHash, 32)
	require.Equal(t, sum{}, sumWithHash)
}
