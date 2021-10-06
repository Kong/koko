package admin_test

import (
	"testing"

	"github.com/gavv/httpexpect/v2"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/stretchr/testify/assert"
)

func goodService() *v1.Service {
	return &v1.Service{
		Name: "foo",
		Host: "example.com",
		Path: "/",
	}
}

func validateGoodService(body *httpexpect.Object) {
	body.ContainsKey("id")
	body.ContainsKey("created_at")
	body.ContainsKey("updated_at")
	body.ValueEqual("write_timeout", 60000)
	body.ValueEqual("read_timeout", 60000)
	body.ValueEqual("connect_timeout", 60000)
	body.ValueEqual("name", "foo")
	body.ValueEqual("path", "/")
	body.ValueEqual("host", "example.com")
	body.ValueEqual("port", 80)
	body.ValueEqual("retries", 5)
	body.ValueEqual("protocol", "http")
}

func TestServiceCreate(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	t.Run("creates a valid service", func(t *testing.T) {
		res := c.POST("/v1/services").WithJSON(goodService()).Expect()
		res.Status(201)
		res.Header("grpc-metadata-koko-status-code").Empty()
		body := res.JSON().Object()
		validateGoodService(body)
	})
	t.Run("creating invalid service fails with 400", func(t *testing.T) {
		svc := &v1.Service{
			Name:           "foo",
			Host:           "example.com",
			Path:           "//foo", // invalid '//' sequence
			ConnectTimeout: -2,      // invalid timeout
		}
		res := c.POST("/v1/services").WithJSON(svc).Expect()
		res.Status(400)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(2)
		errs := body.Value("details").Array()
		var fields []string
		for _, err := range errs.Iter() {
			err.Object().ValueEqual("type", "validation")
			fields = append(fields, err.Object().Value("field").String().Raw())
		}
		assert.ElementsMatch(t, []string{"path", "connect_timeout"}, fields)
	})
	t.Run("recreating the service with the same name fails",
		func(t *testing.T) {
			svc := goodService()
			res := c.POST("/v1/services").WithJSON(svc).Expect()
			res.Status(400)
			body := res.JSON().Object()
			body.ValueEqual("message", "data constraint error")
			body.Value("details").Array().Length().Equal(1)
			err := body.Value("details").Array().Element(0)
			err.Object().ValueEqual("type", "constraint")
			err.Object().ValueEqual("field", "name")
		})
}

func TestServiceDelete(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	svc := goodService()
	res := c.POST("/v1/services").WithJSON(svc).Expect()
	id := res.JSON().Object().Value("id").String().Raw()
	res.Status(201)
	t.Run("deleting a non-existent service returns 404", func(t *testing.T) {
		randomID := "071f5040-3e4a-46df-9d98-451e79e318fd"
		c.DELETE("/v1/services/" + randomID).Expect().Status(404)
	})
	t.Run("deleting a service return 204", func(t *testing.T) {
		c.DELETE("/v1/services/" + id).Expect().Status(204)
	})
}

func TestServiceRead(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	svc := goodService()
	res := c.POST("/v1/services").WithJSON(svc).Expect()
	id := res.JSON().Object().Value("id").String().Raw()
	res.Status(201)
	t.Run("reading a non-existent service returns 404", func(t *testing.T) {
		randomID := "071f5040-3e4a-46df-9d98-451e79e318fd"
		c.GET("/v1/services/" + randomID).Expect().Status(404)
	})
	t.Run("reading a service return 200", func(t *testing.T) {
		body := c.GET("/v1/services/" + id).Expect().Status(200).JSON().Object()
		validateGoodService(body)
	})
}

func TestServiceList(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	svc := goodService()
	res := c.POST("/v1/services").WithJSON(svc).Expect()
	id1 := res.JSON().Object().Value("id").String().Raw()
	res.Status(201)
	svc = &v1.Service{
		Name: "bar",
		Host: "bar.com",
		Path: "/bar",
	}
	res = c.POST("/v1/services").WithJSON(svc).Expect()
	id2 := res.JSON().Object().Value("id").String().Raw()
	res.Status(201)

	t.Run("list returns multiple services", func(t *testing.T) {
		body := c.GET("/v1/services").Expect().Status(200).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(2)
		var gotIDs []string
		for _, item := range items.Iter() {
			gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
		}
		assert.ElementsMatch(t, []string{id1, id2}, gotIDs)
	})
}
