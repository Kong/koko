package admin

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/google/uuid"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model/json/validation/typedefs"
	"github.com/kong/koko/internal/test/util"
	"github.com/stretchr/testify/require"
)

func goodRoute() *v1.Route {
	return &v1.Route{
		Name:  "foo",
		Paths: []string{"/foo"},
	}
}

func validateGoodRoute(body *httpexpect.Object) {
	body.ContainsKey("id")
	body.ContainsKey("created_at")
	body.ContainsKey("updated_at")
	body.ValueEqual("protocols", []string{"http", "https"})
	body.ValueEqual("request_buffering", true)
	body.ValueEqual("response_buffering", true)
	body.ValueEqual("preserve_host", false)
	body.ValueEqual("strip_path", true)
	body.ValueEqual("path_handling", "v0")
	body.ValueEqual("https_redirect_status_code", 426)
}

func TestRouteCreate(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	t.Run("creates a valid route", func(t *testing.T) {
		res := c.POST("/v1/routes").WithJSON(goodRoute()).Expect()
		res.Status(http.StatusCreated)
		res.Header("grpc-metadata-koko-status-code").Empty()
		body := res.JSON().Path("$.item").Object()
		validateGoodRoute(body)
	})
	t.Run("creating a route with ws protocol fails", func(t *testing.T) {
		util.SkipTestIfEnterpriseTesting(t, true)
		route := goodRoute()
		route.Protocols = []string{typedefs.ProtocolWS}
		res := c.POST("/v1/routes").WithJSON(route).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		resErr := body.Value("details").Array().Element(0)
		resErr.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_ENTITY.String())
		resErr.Object().ValueEqual("messages", []string{
			"'ws' and 'wss' protocols are Kong Enterprise-only features. " +
				"Please upgrade to Kong Enterprise to use this feature.",
		})
	})
	t.Run("creating a route with wss protocol fails", func(t *testing.T) {
		util.SkipTestIfEnterpriseTesting(t, true)
		route := goodRoute()
		route.Protocols = []string{typedefs.ProtocolWSS}
		res := c.POST("/v1/routes").WithJSON(route).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		resErr := body.Value("details").Array().Element(0)
		resErr.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_ENTITY.String())
		resErr.Object().ValueEqual("messages", []string{
			"'ws' and 'wss' protocols are Kong Enterprise-only features. " +
				"Please upgrade to Kong Enterprise to use this feature.",
		})
	})
	t.Run("recreating the same route fails", func(t *testing.T) {
		route := goodRoute()
		res := c.POST("/v1/routes").WithJSON(route).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "data constraint error")
		body.Value("details").Array().Length().Equal(1)
		err := body.Value("details").Array().Element(0)
		err.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_REFERENCE.String())
		err.Object().ValueEqual("field", "name")
	})
	t.Run("creating a route with a non-existent service fails", func(t *testing.T) {
		route := &v1.Route{
			Name:  "bar",
			Paths: []string{"/"},
			Service: &v1.Service{
				Id: uuid.NewString(),
			},
		}
		res := c.POST("/v1/routes").WithJSON(route).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "data constraint error")
		body.Value("details").Array().Length().Equal(1)
		err := body.Value("details").Array().Element(0)
		err.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_REFERENCE.String())
		err.Object().ValueEqual("field", "service.id")
	})
	t.Run("creating a route with a valid service.id succeeds", func(t *testing.T) {
		service := goodService()
		service.Id = uuid.NewString()
		res := c.POST("/v1/services").WithJSON(service).Expect()
		res.Status(http.StatusCreated)
		route := &v1.Route{
			Name:  "bar",
			Paths: []string{"/"},
			Service: &v1.Service{
				Id: service.Id,
			},
		}
		res = c.POST("/v1/routes").WithJSON(route).Expect()
		res.Status(http.StatusCreated)
	})
	t.Run("creates a valid route specifying the ID using POST", func(t *testing.T) {
		route := goodRoute()
		route.Name = "with-id"
		route.Id = uuid.NewString()
		res := c.POST("/v1/routes").WithJSON(route).Expect()
		res.Status(http.StatusCreated)
		body := res.JSON().Path("$.item").Object()
		body.Value("id").Equal(route.Id)
	})
	t.Run("creates a route with destinations and sources succeeds", func(t *testing.T) {
		route := &v1.Route{
			Name:      "quz",
			Protocols: []string{"tcp"},
			Destinations: []*v1.CIDRPort{
				{
					Ip:   "192.0.2.0/24",
					Port: int32(80),
				},
				{
					Ip: "198.51.100.0/24",
				},
				{
					Port: 8080,
				},
			},
			Sources: []*v1.CIDRPort{
				{
					Ip:   "203.0.113.0/24",
					Port: int32(80),
				},
			},
		}
		res := c.POST("/v1/routes").WithJSON(route).Expect()
		res.Status(http.StatusCreated)
		body := res.JSON().Path("$.item").Object()
		body.Value("destinations").Array().Length().Equal(3)
		body.Value("sources").Array().Length().Equal(1)
	})
}

func TestRouteUpsert(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	t.Run("upsert a valid route", func(t *testing.T) {
		res := c.PUT("/v1/routes/" + uuid.NewString()).
			WithJSON(goodRoute()).
			Expect()
		res.Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		validateGoodRoute(body)
	})
	t.Run("re-upserting the same route with different id fails",
		func(t *testing.T) {
			route := goodRoute()
			res := c.PUT("/v1/routes/" + uuid.NewString()).
				WithJSON(route).
				Expect()
			res.Status(http.StatusBadRequest)
			body := res.JSON().Object()
			body.ValueEqual("message", "data constraint error")
			body.Value("details").Array().Length().Equal(1)
			err := body.Value("details").Array().Element(0)
			err.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_REFERENCE.String())
			err.Object().ValueEqual("field", "name")
		})
	t.Run("upserting a route with a non-existent service fails", func(t *testing.T) {
		route := &v1.Route{
			Name:  "bar",
			Paths: []string{"/"},
			Service: &v1.Service{
				Id: uuid.NewString(),
			},
		}
		res := c.PUT("/v1/routes/" + uuid.NewString()).
			WithJSON(route).
			Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "data constraint error")
		body.Value("details").Array().Length().Equal(1)
		err := body.Value("details").Array().Element(0)
		err.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_REFERENCE.String())
		err.Object().ValueEqual("field", "service.id")
	})
	t.Run("upserting a route with a valid service.id succeeds", func(t *testing.T) {
		service := goodService()
		sid := uuid.NewString()
		res := c.PUT("/v1/services/" + sid).
			WithJSON(service).
			Expect()
		res.Status(http.StatusOK)
		route := &v1.Route{
			Name:  "bar",
			Paths: []string{"/"},
			Service: &v1.Service{
				Id: sid,
			},
		}
		res = c.PUT("/v1/routes/" + uuid.NewString()).
			WithJSON(route).
			Expect()
		res.Status(http.StatusOK)
	})
	t.Run("upsert correctly updates a route", func(t *testing.T) {
		rid := uuid.NewString()
		route := &v1.Route{
			Id:    rid,
			Name:  "r1",
			Paths: []string{"/"},
		}
		res := c.POST("/v1/routes").
			WithJSON(route).
			Expect()
		res.Status(http.StatusCreated)

		route = &v1.Route{
			Name:  "r1",
			Paths: []string{"/new-value"},
		}
		res = c.PUT("/v1/routes/" + rid).
			WithJSON(route).
			Expect()
		res.Status(http.StatusOK)

		res = c.GET("/v1/routes/" + rid).Expect()
		res.Status(http.StatusOK)
		paths := res.JSON().Path("$.item.paths").Array()
		paths.Length().Equal(1)
		paths.Element(0).String().Equal("/new-value")
	})
	t.Run("route's service can be updated correctly", func(t *testing.T) {
		svc0 := goodService()
		svc0.Name = "s0"
		sid0 := uuid.NewString()

		res := c.PUT("/v1/services/" + sid0).WithJSON(svc0).Expect()
		res.Status(http.StatusOK)

		svc1 := goodService()
		svc1.Name = "s1"
		sid1 := uuid.NewString()

		res = c.PUT("/v1/services/" + sid1).WithJSON(svc1).Expect()
		res.Status(http.StatusOK)

		rid := uuid.NewString()
		route := &v1.Route{
			Id:    rid,
			Name:  "route-for-update",
			Paths: []string{"/"},
			Service: &v1.Service{
				Id: sid0,
			},
		}
		res = c.POST("/v1/routes").
			WithJSON(route).
			Expect()
		res.Status(http.StatusCreated)

		// update the route to point to new service
		route.Service.Id = sid1
		res = c.PUT("/v1/routes/" + rid).
			WithJSON(route).
			Expect()
		res.Status(http.StatusOK)

		res = c.GET("/v1/routes/" + rid).Expect()
		res.Status(http.StatusOK)
		newServiceID := res.JSON().Path("$.item.service.id").String().Raw()
		require.Equal(t, sid1, newServiceID)
	})
	t.Run("upsert route without id fails", func(t *testing.T) {
		route := goodRoute()
		res := c.PUT("/v1/routes/").
			WithJSON(route).
			Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", " '' is not a valid uuid")
	})
}

func TestRouteDelete(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	svc := goodRoute()
	res := c.POST("/v1/routes").WithJSON(svc).Expect()
	id := res.JSON().Path("$.item.id").String().Raw()
	res.Status(http.StatusCreated)
	t.Run("deleting a non-existent route returns 404", func(t *testing.T) {
		randomID := "071f5040-3e4a-46df-9d98-451e79e318fd"
		c.DELETE("/v1/routes/" + randomID).Expect().Status(http.StatusNotFound)
	})
	t.Run("deleting a route return 204", func(t *testing.T) {
		c.DELETE("/v1/routes/" + id).Expect().Status(http.StatusNoContent)
	})
	t.Run("delete request without an ID returns 400", func(t *testing.T) {
		res := c.DELETE("/v1/routes/").Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", " '' is not a valid uuid")
	})
	t.Run("delete request with an invalid ID returns 400", func(t *testing.T) {
		res := c.DELETE("/v1/routes/" + "Not-Valid").Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", " 'Not-Valid' is not a valid uuid")
	})
}

func TestRouteRead(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	svc := goodRoute()
	res := c.POST("/v1/routes").WithJSON(svc).Expect()
	id := res.JSON().Path("$.item.id").String().Raw()
	res.Status(http.StatusCreated)
	t.Run("reading a non-existent route returns 404", func(t *testing.T) {
		randomID := uuid.NewString()
		c.GET("/v1/routes/" + randomID).Expect().Status(http.StatusNotFound)
	})
	t.Run("read route with no name match returns 404", func(t *testing.T) {
		res := c.GET("/v1/routes/somename").Expect()
		res.Status(http.StatusNotFound)
	})
	t.Run("reading a route return 200", func(t *testing.T) {
		res := c.GET("/v1/routes/" + id).Expect().Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		validateGoodRoute(body)
	})
	t.Run("reading a route by name return 200", func(t *testing.T) {
		res := c.GET("/v1/routes/" + svc.Name).Expect().Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		validateGoodRoute(body)
	})
	t.Run("read request without an ID returns 400", func(t *testing.T) {
		res := c.GET("/v1/routes/").Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "required ID is missing")
	})
	t.Run("read request with invalid name or ID match returns 400", func(t *testing.T) {
		invalidKey := "234wabc?!@"
		res = c.GET("/v1/routes/" + invalidKey).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", fmt.Sprintf("invalid ID:'%s'", invalidKey))
	})
}

func TestRouteList(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)

	svc := &v1.Service{
		Name: "foo",
		Host: "example.com",
		Path: "/foo",
	}
	res := c.POST("/v1/services").WithJSON(svc).Expect()
	res.Status(http.StatusCreated)
	serviceID1 := res.JSON().Path("$.item.id").String().Raw()
	svc = &v1.Service{
		Name: "bar",
		Host: "example.com",
		Path: "/bar",
	}
	res = c.POST("/v1/services").WithJSON(svc).Expect()
	res.Status(http.StatusCreated)
	serviceID2 := res.JSON().Path("$.item.id").String().Raw()
	svc = &v1.Service{
		Name: "baz",
		Host: "example.com",
		Path: "/baz",
	}
	res = c.POST("/v1/services").WithJSON(svc).Expect()
	res.Status(http.StatusCreated)
	serviceID3 := res.JSON().Path("$.item.id").String().Raw()

	rte := goodRoute()
	res = c.POST("/v1/routes").WithJSON(rte).Expect()
	res.Status(http.StatusCreated)
	routeID1 := res.JSON().Path("$.item.id").String().Raw()
	rte = &v1.Route{
		Name:  "bar",
		Paths: []string{"/foo"},
	}
	res = c.POST("/v1/routes").WithJSON(rte).Expect()
	res.Status(http.StatusCreated)
	routeID2 := res.JSON().Path("$.item.id").String().Raw()
	rte = &v1.Route{
		Name:  "qux",
		Paths: []string{"/qux"},
		Service: &v1.Service{
			Id: serviceID1,
		},
	}
	res = c.POST("/v1/routes").WithJSON(rte).Expect()
	res.Status(http.StatusCreated)
	routeID3 := res.JSON().Path("$.item.id").String().Raw()
	rte = &v1.Route{
		Name:  "quux",
		Paths: []string{"/quux"},
		Service: &v1.Service{
			Id: serviceID1,
		},
	}
	res = c.POST("/v1/routes").WithJSON(rte).Expect()
	res.Status(http.StatusCreated)
	routeID4 := res.JSON().Path("$.item.id").String().Raw()
	rte = &v1.Route{
		Name:  "quuz",
		Paths: []string{"/quuz"},
		Service: &v1.Service{
			Id: serviceID2,
		},
	}
	res = c.POST("/v1/routes").WithJSON(rte).Expect()
	res.Status(http.StatusCreated)
	routeID5 := res.JSON().Path("$.item.id").String().Raw()

	t.Run("list all routes", func(t *testing.T) {
		body := c.GET("/v1/routes").Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(5)
		var gotIDs []string
		for _, item := range items.Iter() {
			gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
		}
		require.ElementsMatch(t, []string{
			routeID1,
			routeID2,
			routeID3,
			routeID4,
			routeID5,
		}, gotIDs)
	})

	t.Run("list all routes with paging", func(t *testing.T) {
		// Get First Page
		body := c.GET("/v1/routes").
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
		body = c.GET("/v1/routes").
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
		body = c.GET("/v1/routes").
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
			routeID1,
			routeID2,
			routeID3,
			routeID4,
			routeID5,
		}, gotIDs)
	})

	t.Run("list routes by service", func(t *testing.T) {
		body := c.GET("/v1/routes").WithQuery("service_id", serviceID1).
			Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(2)
		var gotIDs []string
		for _, item := range items.Iter() {
			gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
		}
		require.ElementsMatch(t, []string{
			routeID3,
			routeID4,
		}, gotIDs)

		body = c.GET("/v1/routes").WithQuery("service_id", serviceID2).
			Expect().Status(http.StatusOK).JSON().Object()
		items = body.Value("items").Array()
		items.Length().Equal(1)
		gotIDs = nil
		for _, item := range items.Iter() {
			gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
		}
		require.ElementsMatch(t, []string{routeID5}, gotIDs)
	})
	t.Run("list routes by service with paging", func(t *testing.T) {
		body := c.GET("/v1/routes").
			WithQuery("service_id", serviceID1).
			WithQuery("page.size", "1").
			WithQuery("page.number", "1").
			Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(1)
		body.Value("page").Object().Value("total_count").Number().Equal(2)
		body.Value("page").Object().Value("next_page_num").Number().Equal(2)
		id1Got := items.Element(0).Object().Value("id").String().Raw()
		// Next
		body = c.GET("/v1/routes").
			WithQuery("service_id", serviceID1).
			WithQuery("page.size", "1").
			WithQuery("page.number", "2").
			Expect().Status(http.StatusOK).JSON().Object()
		items = body.Value("items").Array()
		items.Length().Equal(1)
		body.Value("page").Object().Value("total_count").Number().Equal(2)
		body.Value("page").Object().NotContainsKey("next_page_num")
		id2Got := items.Element(0).Object().Value("id").String().Raw()
		require.ElementsMatch(t, []string{routeID3, routeID4}, []string{id1Got, id2Got})
	})

	t.Run("list routes by service - no routes associated with service", func(t *testing.T) {
		body := c.GET("/v1/routes").WithQuery("service_id", serviceID3).
			Expect().Status(http.StatusOK).JSON().Object()
		body.Empty()
	})

	t.Run("list routes by service - invalid service UUID", func(t *testing.T) {
		body := c.GET("/v1/routes").WithQuery("service_id", "invalid-uuid").
			Expect().Status(http.StatusBadRequest).JSON().Object()
		body.Keys().Length().Equal(2)
		body.ValueEqual("code", 3)
		body.ValueEqual("message", "service_id 'invalid-uuid' is not a UUID")
	})
}
