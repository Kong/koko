package ws

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNodeList(t *testing.T) {
	nl := NodeList{}

	{
		node, ok := nl.FindNode("127.0.0.1:8000")
		require.False(t, ok)
		require.Empty(t, node)

		require.Empty(t, nl.All())
	}

	{
		n1 := Node{}
		err := nl.Add(&n1)
		require.NoError(t, err)

		require.Equal(t, []*Node{&n1}, nl.All())

		n2, ok := nl.FindNode("")
		require.True(t, ok)
		require.Equal(t, n2, &n1)

		n3 := Node{}
		require.NotSame(t, &n3, &n1)
		err = nl.Add(&n3)
		require.Error(t, err)

		require.Equal(t, []*Node{&n1}, nl.All())
	}
}
