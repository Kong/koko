package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAllTypes(t *testing.T) {
	require.Nil(t, RegisterType(Type("foo"), func() Object {
		return nil
	}))
	require.Nil(t, RegisterType(Type("bar"), func() Object {
		return nil
	}))
	require.Nil(t, RegisterType(Type("baz"), func() Object {
		return nil
	}))
	types := AllTypes()
	require.ElementsMatch(t, []Type{Type("foo"), Type("bar"), Type("baz")}, types)
}
