package config

import (
	"testing"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/stretchr/testify/require"
)

func Test_translateRouteHeaders(t *testing.T) {
	t.Run("translates a header to kong format", func(t *testing.T) {
		route := &v1.Route{
			Headers: map[string]*v1.HeaderValues{
				"foo": {
					Values: []string{"bar"},
				},
				"foo-multi": {
					Values: []string{"bar", "baz"},
				},
			},
		}
		m, err := convert(route)
		require.Nil(t, err)
		translateRouteHeaders(route, m)
		expected := map[string][]string{
			"foo":       {"bar"},
			"foo-multi": {"bar", "baz"},
		}
		require.Equal(t, expected, m["headers"])
	})
	t.Run("no headers translates nothing", func(t *testing.T) {
		route := &v1.Route{
			Headers: nil,
		}
		m, err := convert(route)
		require.Nil(t, err)
		translateRouteHeaders(route, m)
		require.Nil(t, err)
		translateRouteHeaders(route, m)
		require.Nil(t, m["headers"])
	})
}
