package admin

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/google/uuid"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
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
		res.Status(201)
		res.Header("grpc-metadata-koko-status-code").Empty()
		body := res.JSON().Path("$.item").Object()
		validateGoodRoute(body)
	})
	t.Run("recreating the same route fails", func(t *testing.T) {
		route := goodRoute()
		res := c.POST("/v1/routes").WithJSON(route).Expect()
		res.Status(400)
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
		res.Status(400)
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
		res.Status(201)
		route := &v1.Route{
			Name:  "bar",
			Paths: []string{"/"},
			Service: &v1.Service{
				Id: service.Id,
			},
		}
		res = c.POST("/v1/routes").WithJSON(route).Expect()
		res.Status(201)
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
		res.Status(400)
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
}

func TestRouteDelete(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	svc := goodRoute()
	res := c.POST("/v1/routes").WithJSON(svc).Expect()
	id := res.JSON().Path("$.item.id").String().Raw()
	res.Status(201)
	t.Run("deleting a non-existent route returns 404", func(t *testing.T) {
		randomID := "071f5040-3e4a-46df-9d98-451e79e318fd"
		c.DELETE("/v1/routes/" + randomID).Expect().Status(404)
	})
	t.Run("deleting a route return 204", func(t *testing.T) {
		c.DELETE("/v1/routes/" + id).Expect().Status(204)
	})
	t.Run("delete request without an ID returns 400", func(t *testing.T) {
		c.DELETE("/v1/routes/").Expect().Status(400)
	})
}

func TestRouteRead(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	svc := goodRoute()
	res := c.POST("/v1/routes").WithJSON(svc).Expect()
	id := res.JSON().Path("$.item.id").String().Raw()
	res.Status(201)
	t.Run("reading a non-existent route returns 404", func(t *testing.T) {
		randomID := "071f5040-3e4a-46df-9d98-451e79e318fd"
		c.GET("/v1/routes/" + randomID).Expect().Status(404)
	})
	t.Run("reading a route return 200", func(t *testing.T) {
		res := c.GET("/v1/routes/" + id).Expect().Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		validateGoodRoute(body)
	})
	t.Run("read request without an ID returns 400", func(t *testing.T) {
		c.GET("/v1/routes/").Expect().Status(400)
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
	res.Status(201)
	serviceID1 := res.JSON().Path("$.item.id").String().Raw()
	svc = &v1.Service{
		Name: "bar",
		Host: "example.com",
		Path: "/bar",
	}
	res = c.POST("/v1/services").WithJSON(svc).Expect()
	res.Status(201)
	serviceID2 := res.JSON().Path("$.item.id").String().Raw()
	svc = &v1.Service{
		Name: "baz",
		Host: "example.com",
		Path: "/baz",
	}
	res = c.POST("/v1/services").WithJSON(svc).Expect()
	res.Status(201)
	serviceID3 := res.JSON().Path("$.item.id").String().Raw()

	rte := goodRoute()
	res = c.POST("/v1/routes").WithJSON(rte).Expect()
	res.Status(201)
	routeID1 := res.JSON().Path("$.item.id").String().Raw()
	rte = &v1.Route{
		Name:  "bar",
		Paths: []string{"/foo"},
	}
	res = c.POST("/v1/routes").WithJSON(rte).Expect()
	res.Status(201)
	routeID2 := res.JSON().Path("$.item.id").String().Raw()
	rte = &v1.Route{
		Name:  "qux",
		Paths: []string{"/qux"},
		Service: &v1.Service{
			Id: serviceID1,
		},
	}
	res = c.POST("/v1/routes").WithJSON(rte).Expect()
	res.Status(201)
	routeID3 := res.JSON().Path("$.item.id").String().Raw()
	rte = &v1.Route{
		Name:  "quux",
		Paths: []string{"/quux"},
		Service: &v1.Service{
			Id: serviceID1,
		},
	}
	res = c.POST("/v1/routes").WithJSON(rte).Expect()
	res.Status(201)
	routeID4 := res.JSON().Path("$.item.id").String().Raw()
	rte = &v1.Route{
		Name:  "quuz",
		Paths: []string{"/quuz"},
		Service: &v1.Service{
			Id: serviceID2,
		},
	}
	res = c.POST("/v1/routes").WithJSON(rte).Expect()
	res.Status(201)
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
			WithQuery("pagination.size", "2").
			WithQuery("pagination.page", "1").
			Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(2)
		body.Value("pagination").Object().Value("total_count").Number().Equal(5)
		body.Value("pagination").Object().Value("next_page").Number().Equal(2)

		var gotIDs []string
		for _, item := range items.Iter() {
			gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
		}

		// Next Page
		body = c.GET("/v1/routes").
			WithQuery("pagination.size", "2").
			WithQuery("pagination.page", "2").
			Expect().Status(http.StatusOK).JSON().Object()
		items = body.Value("items").Array()
		items.Length().Equal(2)
		body.Value("pagination").Object().Value("total_count").Number().Equal(5)
		body.Value("pagination").Object().Value("next_page").Number().Equal(3)
		for _, item := range items.Iter() {
			gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
		}
		// Last Page
		body = c.GET("/v1/routes").
			WithQuery("pagination.size", "2").
			WithQuery("pagination.page", "3").
			Expect().Status(http.StatusOK).JSON().Object()
		items = body.Value("items").Array()
		items.Length().Equal(1)
		body.Value("pagination").Object().Value("total_count").Number().Equal(5)
		body.Value("pagination").Object().NotContainsKey("next_page")
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
			WithQuery("pagination.size", "1").
			WithQuery("pagination.page", "1").
			Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(1)
		body.Value("pagination").Object().Value("total_count").Number().Equal(2)
		body.Value("pagination").Object().Value("next_page").Number().Equal(2)
		id1Got := items.Element(0).Object().Value("id").String().Raw()
		// Next
		body = c.GET("/v1/routes").
			WithQuery("service_id", serviceID1).
			WithQuery("pagination.size", "1").
			WithQuery("pagination.page", "2").
			Expect().Status(http.StatusOK).JSON().Object()
		items = body.Value("items").Array()
		items.Length().Equal(1)
		body.Value("pagination").Object().Value("total_count").Number().Equal(2)
		body.Value("pagination").Object().NotContainsKey("next_page")
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
