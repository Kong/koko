package generator

import (
	"testing"

	"github.com/kong/koko/internal/model"
	"github.com/stretchr/testify/require"
)

const TypeTestObj = model.Type("testObj")

func TestRegisterUnregister(t *testing.T) {
	// register a type
	testSchema := &Schema{}
	err := Register(string(TypeTestObj), testSchema)
	require.NoError(t, err)

	// make sure a type is registered
	_, ok := globalSchema.Definitions[string(TypeTestObj)]
	require.True(t, ok)

	// unregister a type
	_, err = Unregister(string(TypeTestObj))
	require.NoError(t, err)

	// make sure a type is unregistered
	_, ok = globalSchema.Definitions[string(TypeTestObj)]
	require.False(t, ok)

	// unregister a not-registered type must fail
	_, err = Unregister(string("unregistered-type"))
	require.EqualError(t, err, "type not registered yet: 'unregistered-type'")
}
