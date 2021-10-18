package resource

import (
	"testing"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model/json/validation/typedefs"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestNewRoute(t *testing.T) {
	r := NewRoute()
	assert.NotNil(t, r)
	assert.NotNil(t, r.Route)
}

func TestRoute_ID(t *testing.T) {
	var r Route
	id := r.ID()
	assert.Empty(t, id)
	r = NewRoute()
	id = r.ID()
	assert.Empty(t, id)
}

func TestRoute_Type(t *testing.T) {
	assert.Equal(t, TypeRoute, NewRoute().Type())
}

func TestRoute_ProcessDefaults(t *testing.T) {
	t.Run("defaults are correctly injected", func(t *testing.T) {
		r := NewRoute()
		err := r.ProcessDefaults()
		assert.Nil(t, err)
		assert.True(t, validUUID(r.ID()))
		// empty out the id for equality comparison
		r.Route.Id = ""
		r.Route.CreatedAt = 0
		r.Route.UpdatedAt = 0
		assert.Equal(t, r.Resource(), defaultRoute)
	})
	t.Run("defaults do not override explicit values", func(t *testing.T) {
		r := NewRoute()
		r.Route.HttpsRedirectStatusCode = 302
		r.Route.Protocols = []string{"grpc"}
		r.Route.PathHandling = "v1"
		r.Route.RegexPriority = wrapperspb.Int32(1)
		r.Route.PreserveHost = wrapperspb.Bool(true)
		r.Route.StripPath = wrapperspb.Bool(false)
		r.Route.RequestBuffering = wrapperspb.Bool(false)
		r.Route.ResponseBuffering = wrapperspb.Bool(false)
		err := r.ProcessDefaults()
		assert.Nil(t, err)
		assert.True(t, validUUID(r.ID()))
		assert.NotEmpty(t, r.Route.CreatedAt)
		assert.NotEmpty(t, r.Route.UpdatedAt)
		// empty out the id and ts for equality comparison
		r.Route.Id = ""
		r.Route.CreatedAt = 0
		r.Route.UpdatedAt = 0
		assert.Equal(t, &v1.Route{
			Protocols:               []string{typedefs.ProtocolGRPC},
			RegexPriority:           wrapperspb.Int32(1),
			PreserveHost:            wrapperspb.Bool(true),
			StripPath:               wrapperspb.Bool(false),
			RequestBuffering:        wrapperspb.Bool(false),
			ResponseBuffering:       wrapperspb.Bool(false),
			PathHandling:            "v1",
			HttpsRedirectStatusCode: 302,
		}, r.Resource())
	})
}
