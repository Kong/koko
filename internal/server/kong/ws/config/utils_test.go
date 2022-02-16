package config

import (
	"testing"

	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestFlattenForeign(t *testing.T) {
	t.Run("flattens when key found", func(t *testing.T) {
		m := Map{
			"foo": "bar",
			"baz": map[string]interface{}{
				"id": "long-id",
			},
		}
		flattenForeign(m, "baz")
		require.Equal(t, Map{"foo": "bar", "baz": "long-id"}, m)
	})
	t.Run("does nothing when key not found", func(t *testing.T) {
		m := Map{
			"foo": "bar",
			"baz": "qux",
		}
		flattenForeign(m, "fubaz")
		require.Equal(t, Map{"foo": "bar", "baz": "qux"}, m)
	})
	t.Run("panics when key is a JSON array", func(t *testing.T) {
		m := Map{
			"foo": "bar",
			"baz": []string{"qux"},
		}
		require.Panics(t, func() {
			flattenForeign(m, "baz")
		})
	})
	t.Run("panics when key is a JSON string", func(t *testing.T) {
		m := Map{
			"foo": "bar",
			"baz": "qux",
		}
		require.Panics(t, func() {
			flattenForeign(m, "baz")
		})
	})
}

func TestConvert(t *testing.T) {
	t.Run("converts a proto.Message to Map", func(t *testing.T) {
		svc := &model.Service{
			Name: "foo",
			Host: "example.com",
			Path: "/",
		}
		m, err := convert(svc)
		require.Nil(t, err)
		expected := Map{
			"name": "foo",
			"host": "example.com",
			"path": "/",
		}
		require.Equal(t, expected, m)
	})
	t.Run("converts wrapper proto types correctly", func(t *testing.T) {
		svc := &model.Route{
			Name:          "foo",
			Paths:         []string{"/foO"},
			RegexPriority: wrapperspb.Int32(42),
		}
		m, err := convert(svc)
		require.Nil(t, err)
		expected := Map{
			"name":           "foo",
			"paths":          []interface{}{"/foO"},
			"regex_priority": float64(42),
		}
		require.Equal(t, expected, m)
	})
}
