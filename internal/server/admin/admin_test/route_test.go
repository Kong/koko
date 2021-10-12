package admin_test

import (
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/google/uuid"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/stretchr/testify/assert"
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
		body := res.JSON().Object()
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
		err.Object().ValueEqual("type", "constraint")
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
		err.Object().ValueEqual("type", "constraint")
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

func TestRouteDelete(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	svc := goodRoute()
	res := c.POST("/v1/routes").WithJSON(svc).Expect()
	id := res.JSON().Object().Value("id").String().Raw()
	res.Status(201)
	t.Run("deleting a non-existent route returns 404", func(t *testing.T) {
		randomID := "071f5040-3e4a-46df-9d98-451e79e318fd"
		c.DELETE("/v1/routes/" + randomID).Expect().Status(404)
	})
	t.Run("deleting a route return 204", func(t *testing.T) {
		c.DELETE("/v1/routes/" + id).Expect().Status(204)
	})
}

func TestRouteRead(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	svc := goodRoute()
	res := c.POST("/v1/routes").WithJSON(svc).Expect()
	id := res.JSON().Object().Value("id").String().Raw()
	res.Status(201)
	t.Run("reading a non-existent route returns 404", func(t *testing.T) {
		randomID := "071f5040-3e4a-46df-9d98-451e79e318fd"
		c.GET("/v1/routes/" + randomID).Expect().Status(404)
	})
	t.Run("reading a route return 200", func(t *testing.T) {
		body := c.GET("/v1/routes/" + id).Expect().Status(200).JSON().Object()
		validateGoodRoute(body)
	})
}

func TestRouteList(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	svc := goodRoute()
	res := c.POST("/v1/routes").WithJSON(svc).Expect()
	id1 := res.JSON().Object().Value("id").String().Raw()
	res.Status(201)
	svc = &v1.Route{
		Name:  "bar",
		Paths: []string{"/foo"},
	}
	res = c.POST("/v1/routes").WithJSON(svc).Expect()
	id2 := res.JSON().Object().Value("id").String().Raw()
	res.Status(201)

	t.Run("list returns multiple routes", func(t *testing.T) {
		body := c.GET("/v1/routes").Expect().Status(200).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(2)
		var gotIDs []string
		for _, item := range items.Iter() {
			gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
		}
		assert.ElementsMatch(t, []string{id1, id2}, gotIDs)
	})
}
