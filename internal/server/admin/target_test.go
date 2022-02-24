package admin

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/google/uuid"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/stretchr/testify/require"
)

func TestTargetCreate(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)

	upstream := goodUpstream()
	upstream.Id = uuid.NewString()
	res := c.POST("/v1/upstreams").WithJSON(upstream).Expect()
	res.Status(201)

	t.Run("creating a target with a non-existent upstream fails",
		func(t *testing.T) {
			target := &v1.Target{
				Id:     uuid.NewString(),
				Target: "10.0.24.7",
				Upstream: &v1.Upstream{
					Id: uuid.NewString(),
				},
			}

			res := c.PUT("/v1/targets/" + target.Id).WithJSON(target).Expect()
			res.Status(400)
			body := res.JSON().Object()
			body.ValueEqual("message", "data constraint error")
			body.Value("details").Array().Length().Equal(1)
			err := body.Value("details").Array().Element(0)
			err.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_REFERENCE.String())
			err.Object().ValueEqual("field", "upstream.id")
		})
	t.Run("creating a target with a valid upstream.id succeeds", func(t *testing.T) {
		target := &v1.Target{
			Target: "10.42.42.42",
			Upstream: &v1.Upstream{
				Id: upstream.Id,
			},
		}
		res = c.POST("/v1/targets").WithJSON(target).Expect()
		res.Status(201)
	})
	t.Run("recreating the same target on the same upstream fails", func(t *testing.T) {
		target := &v1.Target{
			Id:     uuid.NewString(),
			Target: "10.42.42.42",
			Upstream: &v1.Upstream{
				Id: upstream.Id,
			},
		}
		res := c.PUT("/v1/targets/" + target.Id).WithJSON(target).Expect()
		res.Status(400)
		body := res.JSON().Object()
		body.ValueEqual("message", "data constraint error")
		body.Value("details").Array().Length().Equal(1)
		err := body.Value("details").Array().Element(0)
		err.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_REFERENCE.String())
		err.Object().ValueEqual("field", "target")
		err.Object().Path("$.messages[0]").String().Contains("constraint failed")
	})
	t.Run("recreating the same target on a different upstream succeeds", func(t *testing.T) {
		upstream := goodUpstream()
		upstream.Name = "bar"
		upstream.Id = uuid.NewString()
		res := c.PUT("/v1/upstreams/" + upstream.Id).WithJSON(upstream).Expect()
		res.Status(200)

		target := &v1.Target{
			Target: "10.42.42.42",
			Upstream: &v1.Upstream{
				Id: upstream.Id,
			},
		}
		res = c.POST("/v1/targets").WithJSON(target).Expect()
		res.Status(201)
	})
	t.Run("creating a target with an invalid target fails", func(t *testing.T) {
		target := &v1.Target{
			Target: "10.42.42.42.42",
			Upstream: &v1.Upstream{
				Id: upstream.Id,
			},
		}
		res = c.POST("/v1/targets").WithJSON(target).Expect()
		res.Status(400)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		err := body.Value("details").Array().Element(0)
		err.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_FIELD.String())
		err.Object().ValueEqual("field", "target")
	})
	t.Run("creating a target with an invalid target fails", func(t *testing.T) {
		target := &v1.Target{
			Target: "invalid.domain:80:80",
			Upstream: &v1.Upstream{
				Id: upstream.Id,
			},
		}
		res = c.POST("/v1/targets").WithJSON(target).Expect()
		res.Status(400)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		err := body.Value("details").Array().Element(0)
		err.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_FIELD.String())
		err.Object().ValueEqual("field", "target")
	})
	t.Run("creating a target with an invalid target fails", func(t *testing.T) {
		target := &v1.Target{
			Target: "invalid.domain:99999999999",
			Upstream: &v1.Upstream{
				Id: upstream.Id,
			},
		}
		res = c.POST("/v1/targets").WithJSON(target).Expect()
		res.Status(400)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		err := body.Value("details").Array().Element(0)
		err.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_FIELD.String())
		err.Object().ValueEqual("field", "target")
	})
	t.Run("creating a target with an invalid target fails", func(t *testing.T) {
		target := &v1.Target{
			Target: "1000.42.42.42",
			Upstream: &v1.Upstream{
				Id: upstream.Id,
			},
		}
		res = c.POST("/v1/targets").WithJSON(target).Expect()
		res.Status(400)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		err := body.Value("details").Array().Element(0)
		err.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_FIELD.String())
		err.Object().ValueEqual("field", "target")
	})
	t.Run("creates a valid target specifying the ID using POST", func(t *testing.T) {
		target := &v1.Target{
			Target: "192.0.2.1",
			Upstream: &v1.Upstream{
				Id: upstream.Id,
			},
			Id: uuid.NewString(),
		}
		res := c.POST("/v1/targets").WithJSON(target).Expect()
		res.Status(201)
		body := res.JSON().Path("$.item").Object()
		body.Value("id").Equal(target.Id)
	})
}

func TestTargetUpsert(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)

	upstream := goodUpstream()
	upstream.Id = uuid.NewString()
	res := c.POST("/v1/upstreams").WithJSON(upstream).Expect()
	res.Status(201)

	t.Run("creating a target with a non-existent upstream fails", func(t *testing.T) {
		target := &v1.Target{
			Target: "10.0.24.7",
			Upstream: &v1.Upstream{
				Id: uuid.NewString(),
			},
		}
		targetID := uuid.NewString()
		res := c.PUT("/v1/targets/" + targetID).WithJSON(target).Expect()
		res.Status(400)
		body := res.JSON().Object()
		body.ValueEqual("message", "data constraint error")
		body.Value("details").Array().Length().Equal(1)
		err := body.Value("details").Array().Element(0)
		err.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_REFERENCE.String())
		err.Object().ValueEqual("field", "upstream.id")
	})
	t.Run("creating a target with a valid upstream.id succeeds", func(t *testing.T) {
		target := &v1.Target{
			Target: "10.42.42.42",
			Upstream: &v1.Upstream{
				Id: upstream.Id,
			},
		}
		targetID := uuid.NewString()
		res = c.PUT("/v1/targets/" + targetID).WithJSON(target).Expect()
		res.Status(200)
		res.JSON().Object().Path("$.item.id").Equal(targetID)
	})
	t.Run("recreating the same target on the same upstream fails", func(t *testing.T) {
		target := &v1.Target{
			Target: "10.42.42.42",
			Upstream: &v1.Upstream{
				Id: upstream.Id,
			},
		}
		targetID := uuid.NewString()
		res := c.PUT("/v1/targets/" + targetID).WithJSON(target).Expect()
		res.Status(400)
		body := res.JSON().Object()
		body.ValueEqual("message", "data constraint error")
		body.Value("details").Array().Length().Equal(1)
		err := body.Value("details").Array().Element(0)
		err.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_REFERENCE.String())
		err.Object().ValueEqual("field", "target")
		err.Object().Path("$.messages[0]").String().Contains("constraint failed")
	})
	t.Run("changing the upstream ID in a PUT works", func(t *testing.T) {
		target := &v1.Target{
			Target: "10.60.24.7",
			Upstream: &v1.Upstream{
				Id: upstream.Id,
			},
		}
		targetID := uuid.NewString()
		res = c.PUT("/v1/targets/" + targetID).WithJSON(target).Expect()
		res.Status(200)
		res.JSON().Object().Path("$.item.id").Equal(targetID)
		res.JSON().Object().Path("$.item.upstream.id").Equal(upstream.Id)

		newUpstream := goodUpstream()
		newUpstream.Name = "baz"
		res := c.POST("/v1/upstreams").WithJSON(newUpstream).Expect()
		res.Status(201)
		newUpstreamID := res.JSON().Object().Path("$.item.id").String().Raw()
		newTarget := &v1.Target{
			Target: "10.60.24.7",
			Upstream: &v1.Upstream{
				Id: newUpstreamID,
			},
		}
		res = c.PUT("/v1/targets/" + targetID).WithJSON(newTarget).Expect()
		res.Status(200)
		object := res.JSON().Object().Path("$.item").Object()
		object.Value("id").Equal(targetID)
		object.Value("target").Equal("10.60.24.7:8000")
		object.Path("$.upstream.id").Equal(newUpstreamID)
	})
	t.Run("upsert target without id fails", func(t *testing.T) {
		target := &v1.Target{
			Target: "10.60.24.7",
			Upstream: &v1.Upstream{
				Id: upstream.Id,
			},
		}
		res := c.PUT("/v1/targets/").
			WithJSON(target).
			Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", " '' is not a valid uuid")
	})
}

func TestTargetDelete(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	upstream := goodUpstream()
	res := c.POST("/v1/upstreams").WithJSON(upstream).Expect()
	res.Status(201)
	upstreamID := res.JSON().Object().Path("$.item.id").String().Raw()

	target := &v1.Target{
		Target: "10.42.42.42",
		Upstream: &v1.Upstream{
			Id: upstreamID,
		},
	}
	res = c.POST("/v1/targets").WithJSON(target).Expect()
	res.Status(201)
	targetID := res.JSON().Object().Path("$.item.id").String().Raw()

	t.Run("deleting a non-existent target returns 404", func(t *testing.T) {
		randomID := "071f5040-3e4a-46df-9d98-451e79e318fd"
		c.DELETE("/v1/targets/" + randomID).Expect().Status(404)
	})
	t.Run("deleting a target return 204", func(t *testing.T) {
		c.DELETE("/v1/targets/" + targetID).Expect().Status(204)
	})
	t.Run("delete request without an ID returns 400", func(t *testing.T) {
		res := c.DELETE("/v1/targets/").Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", " '' is not a valid uuid")
	})
	t.Run("delete request with an invalid ID returns 400", func(t *testing.T) {
		res := c.DELETE("/v1/targets/" + "Not-Valid").Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", " 'Not-Valid' is not a valid uuid")
	})
}

func TestTargetRead(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	upstream := goodUpstream()
	res := c.POST("/v1/upstreams").WithJSON(upstream).Expect()
	res.Status(201)
	upstreamID := res.JSON().Object().Path("$.item.id").String().Raw()

	target := &v1.Target{
		Target: "10.42.42.42",
		Upstream: &v1.Upstream{
			Id: upstreamID,
		},
	}
	res = c.POST("/v1/targets").WithJSON(target).Expect()
	res.Status(201)
	targetID := res.JSON().Object().Path("$.item.id").String().Raw()

	t.Run("reading a non-existent target returns 404", func(t *testing.T) {
		randomID := "071f5040-3e4a-46df-9d98-451e79e318fd"
		c.GET("/v1/targets/" + randomID).Expect().Status(404)
	})
	t.Run("reading a target return 200", func(t *testing.T) {
		res := c.GET("/v1/targets/" + targetID).Expect().Status(http.StatusOK)
		object := res.JSON().Path("$.item").Object()
		object.Value("target").Equal("10.42.42.42:8000")
		object.Value("id").Equal(targetID)
		object.Path("$.upstream.id").Equal(upstreamID)
	})
	t.Run("read request without an ID returns 400", func(t *testing.T) {
		c.GET("/v1/targets/").Expect().Status(400)
	})
}

func TestTargetList(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)

	upstream := &v1.Upstream{
		Name: "foo",
	}
	res := c.POST("/v1/upstreams").WithJSON(upstream).Expect()
	res.Status(201)

	upstreamID1 := res.JSON().Path("$.item.id").String().Raw()
	upstream = &v1.Upstream{
		Name: "bar",
	}
	res = c.POST("/v1/upstreams").WithJSON(upstream).Expect()
	res.Status(201)
	upstreamID2 := res.JSON().Path("$.item.id").String().Raw()

	upstream = &v1.Upstream{
		Name: "baz",
	}
	res = c.POST("/v1/upstreams").WithJSON(upstream).Expect()
	res.Status(201)
	upstreamID3 := res.JSON().Path("$.item.id").String().Raw()

	target := &v1.Target{
		Upstream: &v1.Upstream{
			Id: upstreamID1,
		},
		Target: "10.0.42.1",
	}
	res = c.POST("/v1/targets").WithJSON(target).Expect()
	res.Status(201)
	targetID1 := res.JSON().Path("$.item.id").String().Raw()
	target = &v1.Target{
		Upstream: &v1.Upstream{
			Id: upstreamID1,
		},
		Target: "10.0.42.2",
	}
	res = c.POST("/v1/targets").WithJSON(target).Expect()
	res.Status(201)
	targetID2 := res.JSON().Path("$.item.id").String().Raw()

	target = &v1.Target{
		Upstream: &v1.Upstream{
			Id: upstreamID2,
		},
		Target: "10.0.42.3",
	}
	res = c.POST("/v1/targets").WithJSON(target).Expect()
	res.Status(201)
	targetID3 := res.JSON().Path("$.item.id").String().Raw()

	target = &v1.Target{
		Upstream: &v1.Upstream{
			Id: upstreamID2,
		},
		Target: "10.0.42.4",
	}
	res = c.POST("/v1/targets").WithJSON(target).Expect()
	res.Status(201)
	targetID4 := res.JSON().Path("$.item.id").String().Raw()

	target = &v1.Target{
		Upstream: &v1.Upstream{
			Id: upstreamID2,
		},
		Target: "10.0.42.5",
	}
	res = c.POST("/v1/targets").WithJSON(target).Expect()
	res.Status(201)
	targetID5 := res.JSON().Path("$.item.id").String().Raw()

	t.Run("list all targets", func(t *testing.T) {
		body := c.GET("/v1/targets").Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(5)
		var gotIDs []string
		for _, item := range items.Iter() {
			gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
		}
		require.ElementsMatch(t, []string{
			targetID1,
			targetID2,
			targetID3,
			targetID4,
			targetID5,
		}, gotIDs)
	})

	t.Run("list all targets with paging", func(t *testing.T) {
		// Get First Page
		body := c.GET("/v1/targets").
			WithQuery("page.size", "2").
			WithQuery("page.number", "1").
			Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(2)
		body.Value("page").Object().Value("total_count").Number().Equal(5)
		body.Value("page").Object().Value("next_page_num").Number().Equal(2)

		var gotIDs []string
		for _, item := range items.Iter() {
			gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
		}

		// Next Page
		body = c.GET("/v1/targets").
			WithQuery("page.size", "2").
			WithQuery("page.number", "2").
			Expect().Status(http.StatusOK).JSON().Object()
		items = body.Value("items").Array()
		items.Length().Equal(2)
		body.Value("page").Object().Value("total_count").Number().Equal(5)
		body.Value("page").Object().Value("next_page_num").Number().Equal(3)
		for _, item := range items.Iter() {
			gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
		}
		// Last Page
		body = c.GET("/v1/targets").
			WithQuery("page.size", "2").
			WithQuery("page.number", "3").
			Expect().Status(http.StatusOK).JSON().Object()
		items = body.Value("items").Array()
		items.Length().Equal(1)
		body.Value("page").Object().Value("total_count").Number().Equal(5)
		body.Value("page").Object().NotContainsKey("next_page_num")
		for _, item := range items.Iter() {
			gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
		}
		require.ElementsMatch(t, []string{
			targetID1,
			targetID2,
			targetID3,
			targetID4,
			targetID5,
		}, gotIDs)
	})

	t.Run("list targets by upstream", func(t *testing.T) {
		body := c.GET("/v1/targets").WithQuery("upstream_id", upstreamID1).
			Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(2)
		var gotIDs []string
		for _, item := range items.Iter() {
			gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
		}
		require.ElementsMatch(t, []string{
			targetID1,
			targetID2,
		}, gotIDs)

		body = c.GET("/v1/targets").WithQuery("upstream_id", upstreamID2).
			Expect().Status(http.StatusOK).JSON().Object()
		items = body.Value("items").Array()
		items.Length().Equal(3)
		gotIDs = nil
		for _, item := range items.Iter() {
			gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
		}
		require.ElementsMatch(t, []string{targetID3, targetID4, targetID5}, gotIDs)
	})
	t.Run("list targets by upstream with paging", func(t *testing.T) {
		body := c.GET("/v1/targets").
			WithQuery("upstream_id", upstreamID1).
			WithQuery("page.size", "1").
			WithQuery("page.number", "1").
			Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(1)
		body.Value("page").Object().Value("total_count").Number().Equal(2)
		body.Value("page").Object().Value("next_page_num").Number().Equal(2)
		gotID1 := items.Element(0).Object().Value("id").String().Raw()
		// Next
		body = c.GET("/v1/targets").
			WithQuery("upstream_id", upstreamID1).
			WithQuery("page.size", "1").
			WithQuery("page.number", "2").
			Expect().Status(http.StatusOK).JSON().Object()
		items = body.Value("items").Array()
		items.Length().Equal(1)
		body.Value("page").Object().Value("total_count").Number().Equal(2)
		body.Value("page").Object().NotContainsKey("next_page_num")
		gotID2 := items.Element(0).Object().Value("id").String().Raw()
		require.ElementsMatch(t, []string{targetID1, targetID2}, []string{gotID1, gotID2})
	})

	t.Run("list targets by upstream - no targets associated with upstream", func(t *testing.T) {
		body := c.GET("/v1/targets").WithQuery("upstream_id", upstreamID3).
			Expect().Status(http.StatusOK).JSON().Object()
		body.Empty()
	})

	t.Run("list targets by upstream - invalid upstream UUID", func(t *testing.T) {
		body := c.GET("/v1/targets").WithQuery("upstream_id", "invalid-uuid").
			Expect().Status(http.StatusBadRequest).JSON().Object()
		body.Keys().Length().Equal(2)
		body.ValueEqual("code", 3)
		body.ValueEqual("message", "upstream_id 'invalid-uuid' is not a UUID")
	})
}
