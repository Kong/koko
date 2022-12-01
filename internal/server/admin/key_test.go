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

func goodKey() *v1.Key {
	return &v1.Key{
		Id:   uuid.NewString(),
		Jwk:  &v1.JwkKey{Kid: uuid.NewString()},
		Name: "simpleKey-" + uuid.NewString(),
	}
}

func validateGoodKey(body *httpexpect.Object) {
	body.ContainsKey("id")
	body.ContainsKey("created_at")
	body.ContainsKey("updated_at")
	body.ContainsKey("jwk")
	body.Path("$.jwk").Object().ContainsKey("kid")
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
	t.Run("re-upserting the same key with different id fails",
		func(_ *testing.T) {
			k := goodKey()
			c.PUT("/v1/keys/" + k.Id).WithJSON(k).Expect().Status(http.StatusOK)
			c.PUT("/v1/keys/" + uuid.NewString()).WithJSON(k).Expect().Status(http.StatusBadRequest)
		})
	t.Run("upsert correctly updates a key", func(_ *testing.T) {
		k := goodKey()
		k.Name = "first_key"

		res := c.POST("/v1/keys").WithJSON(k).Expect()
		res.Status(http.StatusCreated)

		res = c.GET("/v1/keys/" + k.Id).Expect()
		res.Status(http.StatusOK)
		res.JSON().Path("$.item.name").Equal("first_key")

		k.Name = "second_key"
		res = c.PUT("/v1/keys/" + k.Id).WithJSON(k).Expect()
		res.Status(http.StatusOK)

		res = c.GET("/v1/keys/" + k.Id).Expect()
		res.Status(http.StatusOK)
		res.JSON().Path("$.item.name").Equal("second_key")
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
	s, cleanup := setup(t)
	defer cleanup()

	c := httpexpect.New(t, s.URL)

	k1 := goodKey()
	res := c.POST("/v1/keys").WithJSON(k1).Expect()
	res.Status(http.StatusCreated)
	id1 := res.JSON().Path("$.item.id").String().Raw()

	k2 := goodKey()
	res = c.POST("/v1/keys").WithJSON(k2).Expect()
	res.Status(http.StatusCreated)
	id2 := res.JSON().Path("$.item.id").String().Raw()

	k3 := goodKey()
	res = c.POST("/v1/keys").WithJSON(k3).Expect()
	res.Status(http.StatusCreated)
	id3 := res.JSON().Path("$.item.id").String().Raw()

	k4 := goodKey()
	res = c.POST("/v1/keys").WithJSON(k4).Expect()
	res.Status(http.StatusCreated)
	id4 := res.JSON().Path("$.item.id").String().Raw()

	k5 := goodKey()
	res = c.POST("/v1/keys").WithJSON(k5).Expect()
	res.Status(http.StatusCreated)
	id5 := res.JSON().Path("$.item.id").String().Raw()

	t.Run("list all keys", func(t *testing.T) {
		body := c.GET("/v1/keys").Expect().Status(http.StatusOK).JSON()
		ids := body.Path("$..id").Array().Raw()
		require.ElementsMatch(t, []string{id1, id2, id3, id4, id5}, ids)
	})

	t.Run("list all keys with paging", func(t *testing.T) {
		// first page
		body := c.GET("/v1/keys").
			WithQuery("page.size", 2).
			WithQuery("page.number", 1).
			Expect().Status(http.StatusOK).JSON().Object()
		body.Path("$.page.total_count").Number().Equal(5)
		body.Path("$.page.next_page_num").Number().Equal(2)
		ids := body.Path("$.items..id").Array().Raw()
		require.Equal(t, 2, len(ids))

		// second page.
		body = c.GET("/v1/keys").
			WithQuery("page.size", 2).
			WithQuery("page.number", 2).
			Expect().Status(http.StatusOK).JSON().Object()
		body.Path("$.page.total_count").Number().Equal(5)
		body.Path("$.page.next_page_num").Number().Equal(3)
		ids = append(ids, body.Path("$.items..id").Array().Raw()...)
		require.Equal(t, 4, len(ids))

		// last page.
		body = c.GET("/v1/keys").
			WithQuery("page.size", 2).
			WithQuery("page.number", 3).
			Expect().Status(http.StatusOK).JSON().Object()
		body.Path("$.page.total_count").Number().Equal(5)
		body.Value("page").Object().NotContainsKey("next_page_num")
		ids = append(ids, body.Path("$.items..id").Array().Raw()...)
		require.Equal(t, 5, len(ids))

		// they're all there
		require.ElementsMatch(t, []string{id1, id2, id3, id4, id5}, ids)
	})
}

func TestKeyAndKeyset(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()

	c := httpexpect.New(t, s.URL)

	// create a keyset
	ks := &v1.KeySet{
		Id:   uuid.NewString(),
		Name: "set-of-keys",
	}
	c.POST("/v1/key-sets").WithJSON(ks).Expect().Status(http.StatusCreated)

	t.Run("create a key belonging to a keyset", func(t *testing.T) {
		k := goodKey()
		k.Set = &v1.KeySet{Id: ks.Id}
		res := c.POST("/v1/keys").WithJSON(k).Expect()
		res.Status(http.StatusCreated)
	})

	t.Run("create a key without set and update to it", func(t *testing.T) {
		k := goodKey()
		res := c.POST("/v1/keys").WithJSON(k).Expect()
		res.Status(http.StatusCreated)

		k.Set = &v1.KeySet{Id: ks.Id}
		res = c.PUT("/v1/keys/" + k.Id).WithJSON(k).Expect()
		res.Status(http.StatusOK)
	})

	nonKeysetId := uuid.NewString()

	t.Run("fails to create a key belonging to a non-existent keyset", func(t *testing.T) {
		k := goodKey()
		k.Set = &v1.KeySet{Id: nonKeysetId}
		res := c.POST("/v1/keys").WithJSON(k).Expect()
		res.Status(http.StatusBadRequest)
	})

	t.Run("create a key without set and fails to update to a non-existent keyset", func(t *testing.T) {
		k := goodKey()
		res := c.POST("/v1/keys").WithJSON(k).Expect()
		res.Status(http.StatusCreated)

		k.Set = &v1.KeySet{Id: nonKeysetId}
		res = c.PUT("/v1/keys/" + k.Id).WithJSON(k).Expect()
		res.Status(http.StatusBadRequest)
	})
}
