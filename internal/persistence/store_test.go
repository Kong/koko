package persistence

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrNotFound(t *testing.T) {
	err := ErrNotFound{Key: "foo"}
	require.Equal(t, "foo", err.Key)
	require.Equal(t, "foo not found", err.Error())
}
