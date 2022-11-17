package admin

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/google/uuid"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
)

func goodKey() *v1.Key {
	return &v1.Key{
		Id:   uuid.NewString(),
		Kid:  uuid.NewString(),
		Name: "simpleKey-" + uuid.NewString(),
	}
}

func validateGoodKey(body *httpexpect.Object) {
	body.ContainsKey("id")
	body.ContainsKey("created_at")
	body.ContainsKey("updated_at")
	body.ContainsKey("kid")
}

func TestKeyCreate(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	t.Run("creates a valid key", func(_ *testing.T) {
		res := c.POST("/v1/keys").WithJSON(goodKey()).Expect()
		res.Status(http.StatusCreated)
		res.Header("grpc-metadata-koko-status-code").Empty()
		body := res.JSON().Path("$.item").Object()
		validateGoodKey(body)
	})
	t.Run("recreating the same key fails", func(_ *testing.T) {
		k := goodKey()
		c.POST("/v1/keys").WithJSON(k).Expect().Status(http.StatusCreated)
		c.POST("/v1/keys").WithJSON(k).Expect().Status(http.StatusBadRequest)
	})
}

func TestKeyUpsert(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	t.Run("upsert a valid key", func(_ *testing.T) {
		res := c.PUT("/v1/keys/" + uuid.NewString()).
			WithJSON(goodKey()).
			Expect()
		res.Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		validateGoodKey(body)
	})
	t.Run("re-upserting the same route with different id fails",
		func(_ *testing.T) {
			k := goodKey()
			c.PUT("/v1/keys/" + k.Id).WithJSON(k).Expect().Status(http.StatusOK)
			c.PUT("/v1/keys/" + uuid.NewString()).WithJSON(k).Expect().Status(http.StatusBadRequest)
		})
	t.Run("upsert correctly updates a route", func(_ *testing.T) {
		k := goodKey()
		k.Jwk = "first_key_content"

		res := c.POST("/v1/keys").WithJSON(k).Expect()
		res.Status(http.StatusCreated)

		res = c.GET("/v1/keys/" + k.Id).Expect()
		res.Status(http.StatusOK)
		res.JSON().Path("$.item.jwk").Equal("first_key_content")

		k.Jwk = "content_of_second_key"
		res = c.PUT("/v1/keys/" + k.Id).WithJSON(k).Expect()
		res.Status(http.StatusOK)

		res = c.GET("/v1/keys/" + k.Id).Expect()
		res.Status(http.StatusOK)
		res.JSON().Path("$.item.jwk").Equal("content_of_second_key")
	})
}

func TestKeyRead(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	k := goodKey()
	res := c.POST("/v1/keys").WithJSON(k).Expect()
	res.Status(http.StatusCreated)
	id := res.JSON().Path("$.item.id").String().Raw()
	t.Run("reading a non-existent key returns 404", func(_ *testing.T) {
		randomID := uuid.NewString()
		c.GET("/v1/keys/" + randomID).Expect().Status(http.StatusNotFound)
	})
	t.Run("read key with no name match returns 404", func(_ *testing.T) {
		res := c.GET("/v1/keys/somename").Expect()
		res.Status(http.StatusNotFound)
	})
	t.Run("reading a key return 200", func(_ *testing.T) {
		res := c.GET("/v1/keys/" + id).Expect().Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		validateGoodKey(body)
	})
	t.Run("reading a key by name return 200", func(_ *testing.T) {
		res := c.GET("/v1/keys/" + k.Name).Expect().Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		validateGoodKey(body)
	})
	t.Run("read request without an ID returns 400", func(_ *testing.T) {
		res := c.GET("/v1/keys/").Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "required ID is missing")
	})
	t.Run("read request with invalid name or ID match returns 400", func(_ *testing.T) {
		invalidRef := "234wabc?!@"
		res = c.GET("/v1/keys/" + invalidRef).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", fmt.Sprintf("invalid ID:'%s'", invalidRef))
	})
}

func TestKeyDelete(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	k := goodKey()
	res := c.POST("/v1/keys").WithJSON(k).Expect()
	id := res.JSON().Path("$.item.id").String().Raw()
	res.Status(http.StatusCreated)
	t.Run("deleting a non-existent key returns 404", func(_ *testing.T) {
		randomID := "071f5040-3e4a-46df-9d98-451e79e318fd"
		c.DELETE("/v1/keys/" + randomID).Expect().Status(http.StatusNotFound)
	})
	t.Run("deleting a key return 204", func(_ *testing.T) {
		c.DELETE("/v1/keys/" + id).Expect().Status(http.StatusNoContent)
	})
	t.Run("delete request without an ID returns 400", func(_ *testing.T) {
		res := c.DELETE("/v1/keys/").Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", " '' is not a valid uuid")
	})
	t.Run("delete request with an invalid ID returns 400", func(_ *testing.T) {
		res := c.DELETE("/v1/keys/" + "Not-Valid").Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", " 'Not-Valid' is not a valid uuid")
	})
}

func TestKeyList(t *testing.T) {
}
