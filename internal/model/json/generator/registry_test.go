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
	err := Default.Register(string(TypeTestObj), testSchema)
	require.NoError(t, err)

	// make sure a type is registered
	_, ok := Default.Schema.Definitions[string(TypeTestObj)]
	require.True(t, ok)

	// unregister a type
	_, err = Default.Unregister(string(TypeTestObj))
	require.NoError(t, err)

	// make sure a type is unregistered
	_, ok = Default.Schema.Definitions[string(TypeTestObj)]
	require.False(t, ok)

	// unregister a not-registered type must fail
	_, err = Default.Unregister(string("unregistered-type"))
	require.EqualError(t, err, "type not registered yet: 'unregistered-type'")
}

func TestSchemaRegistry(t *testing.T) {
	// create a new newRegistry
	newRegistry := NewSchemaRegistry()

	testSchema := &Schema{}
	err := newRegistry.Register(string(TypeTestObj), testSchema)
	require.NoError(t, err)

	// make sure a type is registered
	_, ok := newRegistry.Schema.Definitions[string(TypeTestObj)]
	require.True(t, ok)

	// make sure the new registry is independent of the global one
	_, ok = Default.Schema.Definitions[string(TypeTestObj)]
	require.False(t, ok)

	// unregister a type
	_, err = newRegistry.Unregister(string(TypeTestObj))
	require.NoError(t, err)

	// make sure a type is unregistered
	_, ok = newRegistry.Schema.Definitions[string(TypeTestObj)]
	require.False(t, ok)
}
