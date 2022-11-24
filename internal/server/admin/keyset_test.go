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

func goodKeySet() *v1.KeySet {
	return &v1.KeySet{
		Id:   uuid.NewString(),
		Name: "set_of_keys" + uuid.NewString(),
	}
}

func validateGoodKeySet(body *httpexpect.Object) {
	body.ContainsKey("id")
	body.ContainsKey("name")
}

func TestKeySetCreate(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	t.Run("creates a valid key", func(_ *testing.T) {
		res := c.POST("/v1/key-sets").WithJSON(goodKeySet()).Expect()
		res.Status(http.StatusCreated)
		res.Header("grpc-metadata-koko-status-code").Empty()
		body := res.JSON().Path("$.item").Object()
		validateGoodKeySet(body)
	})
	t.Run("recreating the same key fails", func(_ *testing.T) {
		ks := goodKeySet()
		c.POST("/v1/key-sets").WithJSON(ks).Expect().Status(http.StatusCreated)
		c.POST("/v1/key-sets").WithJSON(ks).Expect().Status(http.StatusBadRequest)
	})
}

func TestKeySetUpsert(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	t.Run("upsert a valid key", func(_ *testing.T) {
		res := c.PUT("/v1/key-sets/" + uuid.NewString()).
			WithJSON(goodKeySet()).
			Expect()
		res.Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		validateGoodKeySet(body)
	})
	t.Run("re-upserting the same key with different id fails",
		func(_ *testing.T) {
			ks := goodKeySet()
			c.PUT("/v1/key-sets/" + ks.Id).WithJSON(ks).Expect().Status(http.StatusOK)
			c.PUT("/v1/key-sets/" + uuid.NewString()).WithJSON(ks).Expect().Status(http.StatusBadRequest)
		})
	t.Run("upsert correctly updates a route", func(_ *testing.T) {
		ks := goodKeySet()

		res := c.POST("/v1/key-sets").WithJSON(ks).Expect()
		res.Status(http.StatusCreated)

		res = c.GET("/v1/key-sets/" + ks.Id).Expect()
		res.Status(http.StatusOK)
		res.JSON().Path("$.item.name").Equal(ks.Name)

		ks.Name = "notSameKeys"
		res = c.PUT("/v1/key-sets/" + ks.Id).WithJSON(ks).Expect()
		res.Status(http.StatusOK)

		res = c.GET("/v1/key-sets/" + ks.Id).Expect()
		res.Status(http.StatusOK)
		res.JSON().Path("$.item.name").Equal("notSameKeys")
	})
}

func TestKeySetRead(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	ks := goodKeySet()
	res := c.POST("/v1/key-sets").WithJSON(ks).Expect()
	res.Status(http.StatusCreated)
	id := res.JSON().Path("$.item.id").String().Raw()
	t.Run("reading a non-existent keyset returns 404", func(_ *testing.T) {
		randomID := uuid.NewString()
		c.GET("/v1/key-sets/" + randomID).Expect().Status(http.StatusNotFound)
	})
	t.Run("read keyset with no name match returns 404", func(_ *testing.T) {
		res := c.GET("/v1/key-sets/somename").Expect()
		res.Status(http.StatusNotFound)
	})
	t.Run("reading a keyset return 200", func(_ *testing.T) {
		res := c.GET("/v1/key-sets/" + id).Expect().Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		validateGoodKeySet(body)
	})
	t.Run("reading a keyset by name return 200", func(_ *testing.T) {
		res := c.GET("/v1/key-sets/" + ks.Name).Expect().Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		validateGoodKeySet(body)
	})
	t.Run("read request without an ID returns 400", func(_ *testing.T) {
		res := c.GET("/v1/key-sets/").Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "required ID is missing")
	})
	t.Run("read request with invalid name or ID match returns 400", func(_ *testing.T) {
		invalidRef := "234wabc?!@"
		res = c.GET("/v1/key-sets/" + invalidRef).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", fmt.Sprintf("invalid ID:'%s'", invalidRef))
	})
}

func TestKeySetDelete(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	ks := goodKeySet()
	res := c.POST("/v1/key-sets").WithJSON(ks).Expect()
	id := res.JSON().Path("$.item.id").String().Raw()
	res.Status(http.StatusCreated)
	t.Run("deleting a non-existent keyset returns 404", func(_ *testing.T) {
		randomID := "071f5040-3e4a-46df-9d98-451e79e318fd"
		c.DELETE("/v1/key-sets/" + randomID).Expect().Status(http.StatusNotFound)
	})
	t.Run("deleting a keyset return 204", func(_ *testing.T) {
		c.DELETE("/v1/key-sets/" + id).Expect().Status(http.StatusNoContent)
	})
	t.Run("delete request without an ID returns 400", func(_ *testing.T) {
		res := c.DELETE("/v1/key-sets/").Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", " '' is not a valid uuid")
	})
	t.Run("delete request with an invalid ID returns 400", func(_ *testing.T) {
		res := c.DELETE("/v1/key-sets/" + "Not-Valid").Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", " 'Not-Valid' is not a valid uuid")
	})
}

func TestKeySetList(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()

	c := httpexpect.New(t, s.URL)

	ks1 := goodKeySet()
	res := c.POST("/v1/key-sets").WithJSON(ks1).Expect()
	res.Status(http.StatusCreated)
	id1 := res.JSON().Path("$.item.id").String().Raw()

	ks2 := goodKeySet()
	res = c.POST("/v1/key-sets").WithJSON(ks2).Expect()
	res.Status(http.StatusCreated)
	id2 := res.JSON().Path("$.item.id").String().Raw()

	ks3 := goodKeySet()
	res = c.POST("/v1/key-sets").WithJSON(ks3).Expect()
	res.Status(http.StatusCreated)
	id3 := res.JSON().Path("$.item.id").String().Raw()

	ks4 := goodKeySet()
	res = c.POST("/v1/key-sets").WithJSON(ks4).Expect()
	res.Status(http.StatusCreated)
	id4 := res.JSON().Path("$.item.id").String().Raw()

	ks5 := goodKeySet()
	res = c.POST("/v1/key-sets").WithJSON(ks5).Expect()
	res.Status(http.StatusCreated)
	id5 := res.JSON().Path("$.item.id").String().Raw()

	t.Run("list all key sets", func(t *testing.T) {
		body := c.GET("/v1/key-sets").Expect().Status(http.StatusOK).JSON()
		ids := body.Path("$..id").Array().Raw()
		require.ElementsMatch(t, []string{id1, id2, id3, id4, id5}, ids)
	})

	t.Run("list all key sets with paging", func(t *testing.T) {
		// first page
		body := c.GET("/v1/key-sets").
			WithQuery("page.size", 2).
			WithQuery("page.number", 1).
			Expect().Status(http.StatusOK).JSON().Object()
		body.Path("$.page.total_count").Number().Equal(5)
		body.Path("$.page.next_page_num").Number().Equal(2)
		ids := body.Path("$.items..id").Array().Raw()
		require.Equal(t, 2, len(ids))

		// second page.
		body = c.GET("/v1/key-sets").
			WithQuery("page.size", 2).
			WithQuery("page.number", 2).
			Expect().Status(http.StatusOK).JSON().Object()
		body.Path("$.page.total_count").Number().Equal(5)
		body.Path("$.page.next_page_num").Number().Equal(3)
		ids = append(ids, body.Path("$.items..id").Array().Raw()...)
		require.Equal(t, 4, len(ids))

		// last page.
		body = c.GET("/v1/key-sets").
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
