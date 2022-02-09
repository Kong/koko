package admin

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/google/uuid"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/stretchr/testify/require"
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
		body := res.JSON().Path("$.item").Object()
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
			err.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_FIELD.
				String())
			fields = append(fields, err.Object().Value("field").String().Raw())
		}
		require.ElementsMatch(t, []string{"path", "connect_timeout"}, fields)
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
			err.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_REFERENCE.
				String())
			err.Object().ValueEqual("field", "name")
		})
	t.Run("service with a '-' in name can be created", func(t *testing.T) {
		svc := goodService()
		svc.Name = "foo-with-dash"
		res := c.POST("/v1/services").WithJSON(svc).Expect()
		res.Status(201)
	})
}

func TestServiceUpsert(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	t.Run("upserts a valid service", func(t *testing.T) {
		res := c.PUT("/v1/services/" + uuid.NewString()).
			WithJSON(goodService()).
			Expect()
		res.Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		validateGoodService(body)
	})
	t.Run("upserting an invalid service fails with 400", func(t *testing.T) {
		svc := &v1.Service{
			Name:           "foo",
			Host:           "example.com",
			Path:           "//foo", // invalid '//' sequence
			ConnectTimeout: -2,      // invalid timeout
		}
		res := c.PUT("/v1/services/" + uuid.NewString()).
			WithJSON(svc).
			Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(2)
		errs := body.Value("details").Array()
		var fields []string
		for _, err := range errs.Iter() {
			err.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_FIELD.
				String())
			fields = append(fields, err.Object().Value("field").String().Raw())
		}
		require.ElementsMatch(t, []string{"path", "connect_timeout"}, fields)
	})
	t.Run("recreating the service with the same name but different id fails",
		func(t *testing.T) {
			svc := goodService()
			res := c.PUT("/v1/services/" + uuid.NewString()).
				WithJSON(svc).
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
	t.Run("upsert correctly updates a route", func(t *testing.T) {
		sid := uuid.NewString()
		svc := &v1.Service{
			Id:   sid,
			Name: "r1",
			Host: "example.com",
			Path: "/bar",
		}
		res := c.POST("/v1/services").
			WithJSON(svc).
			Expect()
		res.Status(http.StatusCreated)

		svc = &v1.Service{
			Id:   sid,
			Name: "r1",
			Host: "new.example.com",
			Path: "/bar-new",
		}
		res = c.PUT("/v1/services/" + sid).
			WithJSON(svc).
			Expect()
		res.Status(http.StatusOK)

		res = c.GET("/v1/services/" + sid).Expect()
		res.Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		body.Value("host").Equal("new.example.com")
		body.Value("path").Equal("/bar-new")
	})
}

func TestServiceDelete(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	svc := goodService()
	res := c.POST("/v1/services").WithJSON(svc).Expect()
	id := res.JSON().Path("$.item.id").String().Raw()
	res.Status(201)
	t.Run("deleting a non-existent service returns 404", func(t *testing.T) {
		randomID := "071f5040-3e4a-46df-9d98-451e79e318fd"
		c.DELETE("/v1/services/" + randomID).Expect().Status(404)
	})
	t.Run("deleting a service return 204", func(t *testing.T) {
		c.DELETE("/v1/services/" + id).Expect().Status(204)
	})
	t.Run("delete request without an ID returns 400", func(t *testing.T) {
		c.DELETE("/v1/services/").Expect().Status(400)
	})
}

func TestServiceRead(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	svc := goodService()
	res := c.POST("/v1/services").WithJSON(svc).Expect()
	id := res.JSON().Path("$.item.id").String().Raw()
	res.Status(201)
	t.Run("reading a non-existent service returns 404", func(t *testing.T) {
		randomID := "071f5040-3e4a-46df-9d98-451e79e318fd"
		c.GET("/v1/services/" + randomID).Expect().Status(404)
	})
	t.Run("reading a service return 200", func(t *testing.T) {
		res := c.GET("/v1/services/" + id).Expect().Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		validateGoodService(body)
	})
	t.Run("read request without an ID returns 400", func(t *testing.T) {
		c.GET("/v1/services/").Expect().Status(400)
	})
}

func TestServiceList(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	svc := goodService()
	res := c.POST("/v1/services").WithJSON(svc).Expect()
	id1 := res.JSON().Path("$.item.id").String().Raw()
	res.Status(201)
	svc = &v1.Service{
		Name: "bar",
		Host: "bar.com",
		Path: "/bar",
	}
	res = c.POST("/v1/services").WithJSON(svc).Expect()
	id2 := res.JSON().Path("$.item.id").String().Raw()
	res.Status(201)

	t.Run("list returns multiple services", func(t *testing.T) {
		body := c.GET("/v1/services").Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(2)
		var gotIDs []string
		for _, item := range items.Iter() {
			gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
		}
		require.ElementsMatch(t, []string{id1, id2}, gotIDs)
	})
}

func TestServiceListPagination(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	// Create ten services
	svcName := "myservice-%d"
	svc := goodService()
	for i := 0; i < 10; i++ {
		svc.Name = fmt.Sprintf(svcName, i)
		res := c.POST("/v1/services").WithJSON(svc).Expect()
		res.JSON().Path("$.item.id").String()
		res.Status(201)
	}
	var tailID string
	var headID string
	body := c.GET("/v1/services").Expect().Status(http.StatusOK).JSON().Object()
	items := body.Value("items").Array()
	items.Length().Equal(10)
	// Get the head's id so that we can make sure that it is consistent
	headID = items.Element(0).Object().Value("id").String().Raw()
	require.NotEmpty(t, headID)
	// Get the tail's id so that we can make sure that it is consistent
	tailID = items.Element(9).Object().Value("id").String().Raw()
	require.NotEmpty(t, tailID)
	body.Value("pagination").Object().Value("total_count").Number().Equal(10)
	body.Value("pagination").Object().NotContainsKey("next_page")

	t.Run("list size 1 page 10 returns 1 service total_count=10", func(t *testing.T) {
		// Get First Page
		body := c.GET("/v1/services").
			WithQuery("cluster.id", "default").
			WithQuery("pagination.size", "1").
			WithQuery("pagination.page", "1").
			Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(1)

		body.Value("pagination").Object().Value("total_count").Number().Equal(10)
		body.Value("pagination").Object().Value("next_page").Number().Equal(2)
		firstID := items.Element(0).Object().Value("id").String().Raw()
		require.NotEmpty(t, firstID)
		require.Equal(t, headID, firstID)
		// Go to last page and get the last element
		body = c.GET("/v1/services").
			WithQuery("cluster.id", "default").
			WithQuery("pagination.size", "1").
			WithQuery("pagination.page", "10").
			Expect().Status(http.StatusOK).JSON().Object()
		items = body.Value("items").Array()
		items.Length().Equal(1)
		body.Value("pagination").Object().Value("total_count").Number().Equal(10)
		body.Value("pagination").Object().NotContainsKey("next_page")

		lastID := items.Element(0).Object().Value("id").String().Raw()
		require.NotEmpty(t, lastID)
		require.Equal(t, tailID, lastID)
	})
	t.Run("list page_size 2 and page 10 returns 2 services with total_count=10", func(t *testing.T) {
		// Get First Page
		body := c.GET("/v1/services").
			WithQuery("cluster.id", "default").
			WithQuery("pagination.size", "2").
			WithQuery("pagination.page", "1").
			Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(2)
		body.Value("pagination").Object().Value("total_count").Number().Equal(10)
		body.Value("pagination").Object().Value("next_page").Number().Equal(2)

		firstID := items.Element(0).Object().Value("id").String().Raw()
		require.NotEmpty(t, firstID)
		require.Equal(t, headID, firstID)
		// Go to last page and get the last element
		body = c.GET("/v1/services").
			WithQuery("cluster.id", "default").
			WithQuery("pagination.size", "2").
			WithQuery("pagination.page", "5").
			Expect().Status(http.StatusOK).JSON().Object()
		items = body.Value("items").Array()
		items.Length().Equal(2)
		body.Value("pagination").Object().Value("total_count").Number().Equal(10)
		body.Value("pagination").Object().NotContainsKey("next_page")
		lastID := items.Element(1).Object().Value("id").String().Raw()
		require.NotEmpty(t, lastID)
		require.Equal(t, tailID, lastID)
	})
	t.Run("list page_size 3 and page 10 returns 3 services with total_count=10", func(t *testing.T) {
		// Get First Page
		body := c.GET("/v1/services").
			WithQuery("cluster.id", "default").
			WithQuery("pagination.size", "3").
			WithQuery("pagination.page", "1").
			Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(3)
		body.Value("pagination").Object().Value("total_count").Number().Equal(10)
		body.Value("pagination").Object().Value("next_page").Number().Equal(2)
		firstID := items.Element(0).Object().Value("id").String().Raw()
		require.NotEmpty(t, firstID)
		require.Equal(t, headID, firstID)
		// Go to last page and get the last element
		body = c.GET("/v1/services").
			WithQuery("cluster.id", "default").
			WithQuery("pagination.size", "3").
			WithQuery("pagination.page", "4").
			Expect().Status(http.StatusOK).JSON().Object()
		items = body.Value("items").Array()
		items.Length().Equal(1)
		body.Value("pagination").Object().Value("total_count").Number().Equal(10)
		body.Value("pagination").Object().NotContainsKey("next_page")
		lastID := items.Element(0).Object().Value("id").String().Raw()
		require.NotEmpty(t, lastID)
		require.Equal(t, tailID, lastID)
	})
	t.Run("list page_size 4 and page 10 returns 4 services with total_count=10", func(t *testing.T) {
		// Get First Page
		body := c.GET("/v1/services").
			WithQuery("cluster.id", "default").
			WithQuery("pagination.size", "4").
			WithQuery("pagination.page", "1").
			Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(4)
		body.Value("pagination").Object().Value("total_count").Number().Equal(10)
		body.Value("pagination").Object().Value("next_page").Number().Equal(2)
		firstID := items.Element(0).Object().Value("id").String().Raw()
		require.NotEmpty(t, firstID)
		require.Equal(t, headID, firstID)
		// Go to last page and get the last element
		body = c.GET("/v1/services").
			WithQuery("cluster.id", "default").
			WithQuery("pagination.size", "4").
			WithQuery("pagination.page", "3").
			Expect().Status(http.StatusOK).JSON().Object()
		items = body.Value("items").Array()
		items.Length().Equal(2)
		body.Value("pagination").Object().Value("total_count").Number().Equal(10)
		body.Value("pagination").Object().NotContainsKey("next_page")
		lastID := items.Element(1).Object().Value("id").String().Raw()
		require.NotEmpty(t, lastID)
		require.Equal(t, tailID, lastID)
	})
	t.Run("list page_size 10 and page 1 returns 10 services with total_count=10", func(t *testing.T) {
		// Get First Page
		body := c.GET("/v1/services").
			WithQuery("cluster.id", "default").
			WithQuery("pagination.size", "10").
			WithQuery("pagination.page", "1").
			Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(10)
		body.Value("pagination").Object().Value("total_count").Number().Equal(10)
		body.Value("pagination").Object().NotContainsKey("next_page")
		firstID := items.Element(0).Object().Value("id").String().Raw()
		require.NotEmpty(t, firstID)
		require.Equal(t, headID, firstID)

		lastID := items.Element(9).Object().Value("id").String().Raw()
		require.NotEmpty(t, lastID)
		require.Equal(t, tailID, lastID)
	})
	t.Run("list page_size 10 and page 10 returns no services", func(t *testing.T) {
		body := c.GET("/v1/services").
			WithQuery("cluster.id", "default").
			WithQuery("pagination.size", "10").
			WithQuery("pagination.page", "10").
			Expect().Status(http.StatusOK).JSON().Object()
		body.NotContainsKey("items")
	})
	t.Run("list page_size 10 and no Page returns 10 services with total_count=10", func(t *testing.T) {
		// Get First Page
		body := c.GET("/v1/services").
			WithQuery("cluster.id", "default").
			WithQuery("pagination.size", "10").
			Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(10)
		body.Value("pagination").Object().Value("total_count").Number().Equal(10)
		body.Value("pagination").Object().NotContainsKey("next_page")
		firstID := items.Element(0).Object().Value("id").String().Raw()
		require.NotEmpty(t, firstID)
		require.Equal(t, headID, firstID)

		lastID := items.Element(9).Object().Value("id").String().Raw()
		require.NotEmpty(t, lastID)
		require.Equal(t, tailID, lastID)
	})
	t.Run("list no page_size and Page 1 returns 10 services with total_count=10", func(t *testing.T) {
		// Get First Page
		body := c.GET("/v1/services").
			WithQuery("cluster.id", "default").
			WithQuery("pagination.page", "1").
			Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(10)
		body.Value("pagination").Object().Value("total_count").Number().Equal(10)
		body.Value("pagination").Object().NotContainsKey("next_page")
		firstID := items.Element(0).Object().Value("id").String().Raw()
		require.NotEmpty(t, firstID)
		require.Equal(t, headID, firstID)

		lastID := items.Element(9).Object().Value("id").String().Raw()
		require.NotEmpty(t, lastID)
		require.Equal(t, tailID, lastID)
	})
	t.Run("list page_size 11 and page 1 returns 10 services with total_count=10", func(t *testing.T) {
		// Get First Page
		body := c.GET("/v1/services").
			WithQuery("cluster.id", "default").
			WithQuery("pagination.size", "11").
			WithQuery("pagination.page", "1").
			Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(10)
		body.Value("pagination").Object().Value("total_count").Number().Equal(10)
		body.Value("pagination").Object().NotContainsKey("next_page")
		firstID := items.Element(0).Object().Value("id").String().Raw()
		require.NotEmpty(t, firstID)
		require.Equal(t, headID, firstID)

		lastID := items.Element(9).Object().Value("id").String().Raw()
		require.NotEmpty(t, lastID)
		require.Equal(t, tailID, lastID)
	})
	t.Run("list > 1001 page size and page 10 returns error", func(t *testing.T) {
		// Get First Page
		body := c.GET("/v1/services").
			WithQuery("cluster.id", "default").
			WithQuery("pagination.size", "1001").
			WithQuery("pagination.page", "1").
			Expect().Status(http.StatusBadRequest).JSON().Object()
		body.Value("code").Number().Equal(3)
		body.Value("message").String().Equal("invalid page_size '1001', must be within range [1 - 1000]")
	})
	t.Run("list page_size 10 and page < 0 returns error", func(t *testing.T) {
		// Get First Page
		body := c.GET("/v1/services").
			WithQuery("cluster.id", "default").
			WithQuery("pagination.size", "10").
			WithQuery("pagination.page", "-1").
			Expect().Status(http.StatusBadRequest).JSON().Object()
		body.Value("code").Number().Equal(3)
		body.Value("message").String().Equal("invalid page '-1', page must be > 0")
	})
}
