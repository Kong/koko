package admin

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/google/uuid"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/json"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func goodUpstream() *v1.Upstream {
	return &v1.Upstream{
		Name: "foo",
	}
}

func validateGoodUpstream(body *httpexpect.Object) {
}

func TestUpstreamCreate(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	t.Run("creates a valid upstream", func(t *testing.T) {
		res := c.POST("/v1/upstreams").WithJSON(goodUpstream()).Expect()
		res.Status(201)
		res.Header("grpc-metadata-koko-status-code").Empty()
		body := res.JSON().Path("$.item").Object()
		body.Value("name").String().Equal("foo")
		validateGoodUpstream(body)
	})
	t.Run("creating an invalid upstream fails with 400", func(t *testing.T) {
		upstream := &v1.Upstream{}
		res := c.POST("/v1/upstreams").WithJSON(upstream).Expect()
		res.Status(400)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		err := body.Value("details").Array().Element(0).Object()
		err.ValueEqual("type", v1.ErrorType_ERROR_TYPE_ENTITY.String())
		err.Value("messages").Array().Element(0).String().
			Equal("missing properties: 'name'")
	})
	t.Run("recreating an upstream with the same name fails", func(t *testing.T) {
		upstream := goodUpstream()
		res := c.POST("/v1/upstreams").WithJSON(upstream).Expect()
		res.Status(400)
		body := res.JSON().Object()
		body.ValueEqual("message", "data constraint error")
		body.Value("details").Array().Length().Equal(1)
		err := body.Value("details").Array().Element(0)
		err.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_REFERENCE.
			String())
		err.Object().ValueEqual("field", "name")
	})
	t.Run("upstream with a '-' in name can be created", func(t *testing.T) {
		upstream := goodUpstream()
		upstream.Name = "foo-with-dash"
		res := c.POST("/v1/upstreams").WithJSON(upstream).Expect()
		res.Status(201)
	})
}

func TestUpstreamUpsert(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	t.Run("upserts a valid upstream", func(t *testing.T) {
		res := c.PUT("/v1/upstreams/" + uuid.NewString()).
			WithJSON(goodUpstream()).
			Expect()
		res.Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		validateGoodUpstream(body)
	})
	t.Run("upserting an invalid upstream fails with 400", func(t *testing.T) {
		upstream := &v1.Upstream{
			Name: "$foo",
		}
		res := c.PUT("/v1/upstreams/" + uuid.NewString()).
			WithJSON(upstream).
			Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		errs := body.Value("details").Array()
		var fields []string
		for _, err := range errs.Iter() {
			err.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_FIELD.
				String())
			fields = append(fields, err.Object().Value("field").String().Raw())
		}
		require.ElementsMatch(t, []string{"name"}, fields)
	})
	t.Run("recreating the upstream with the same name but different id fails",
		func(t *testing.T) {
			upstream := goodUpstream()
			res := c.PUT("/v1/upstreams/" + uuid.NewString()).
				WithJSON(upstream).
				Expect()
			res.Status(http.StatusBadRequest)
			body := res.JSON().Object()
			body.ValueEqual("message", "data constraint error")
			body.Value("details").Array().Length().Equal(1)
			err := body.Value("details").Array().Element(0)
			err.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_REFERENCE.
				String())
			err.Object().ValueEqual("field", "name")
		})
	t.Run("upsert correctly updates an upstream", func(t *testing.T) {
		uid := uuid.NewString()
		upstream := &v1.Upstream{
			Id:   uid,
			Name: "foo.com",
		}
		res := c.POST("/v1/upstreams").
			WithJSON(upstream).
			Expect()
		res.Status(http.StatusCreated)

		upstream = &v1.Upstream{
			Id:           uid,
			Name:         "foo.com",
			HashOn:       "header",
			HashOnHeader: "apikey",
			HashFallback: "ip",
			Healthchecks: &v1.Healthchecks{
				Active: &v1.ActiveHealthcheck{
					Concurrency: wrapperspb.Int32(32),
					Healthy: &v1.ActiveHealthyCondition{
						Interval:  wrapperspb.Int32(1),
						Successes: wrapperspb.Int32(5),
					},
				},
			},
		}
		upstreamJSON, err := json.Marshal(upstream)
		require.Nil(t, err)
		res = c.PUT("/v1/upstreams/" + uid).
			WithBytes(upstreamJSON).
			Expect()
		res.Status(http.StatusOK)

		res = c.GET("/v1/upstreams/" + uid).Expect()
		res.Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		body.Value("hash_on").Equal("header")
		body.Value("hash_on_header").Equal("apikey")
		body.Value("hash_fallback").Equal("ip")
	})
}

func TestUpstreamDelete(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	upstream := goodUpstream()
	res := c.POST("/v1/upstreams").WithJSON(upstream).Expect()
	id := res.JSON().Path("$.item.id").String().Raw()
	res.Status(201)
	t.Run("deleting a non-existent upstream returns 404", func(t *testing.T) {
		randomID := "071f5040-3e4a-46df-9d98-451e79e318fd"
		c.DELETE("/v1/upstreams/" + randomID).Expect().Status(404)
	})
	t.Run("deleting a upstream return 204", func(t *testing.T) {
		c.DELETE("/v1/upstreams/" + id).Expect().Status(204)
	})
	t.Run("delete request without an ID returns 400", func(t *testing.T) {
		c.DELETE("/v1/upstreams/").Expect().Status(400)
	})
}

func TestUpstreamRead(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	upstream := goodUpstream()
	res := c.POST("/v1/upstreams").WithJSON(upstream).Expect()
	id := res.JSON().Path("$.item.id").String().Raw()
	res.Status(201)
	t.Run("reading a non-existent upstream returns 404", func(t *testing.T) {
		randomID := "071f5040-3e4a-46df-9d98-451e79e318fd"
		c.GET("/v1/upstreams/" + randomID).Expect().Status(404)
	})
	t.Run("reading a upstream return 200", func(t *testing.T) {
		c.GET("/v1/upstreams/" + id).Expect().Status(http.StatusOK)
		// body := res.JSON().Path("$.item").Object()
		// validateGoodUpstream(body)
	})
	t.Run("read request without an ID returns 400", func(t *testing.T) {
		c.GET("/v1/upstreams/").Expect().Status(400)
	})
}

func TestUpstreamList(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	upstream := goodUpstream()
	res := c.POST("/v1/upstreams").WithJSON(upstream).Expect()
	id1 := res.JSON().Path("$.item.id").String().Raw()
	res.Status(201)
	upstream = &v1.Upstream{
		Name: "bar",
	}
	res = c.POST("/v1/upstreams").WithJSON(upstream).Expect()
	id2 := res.JSON().Path("$.item.id").String().Raw()
	res.Status(201)

	t.Run("list returns multiple upstreams", func(t *testing.T) {
		body := c.GET("/v1/upstreams").Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(2)
		var gotIDs []string
		for _, item := range items.Iter() {
			gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
		}
		require.ElementsMatch(t, []string{id1, id2}, gotIDs)
	})
}