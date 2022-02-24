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

func goodConsumer() *v1.Consumer {
	return &v1.Consumer{
		Username: "consumerA",
		CustomId: "customIDA",
	}
}

func TestConsumerCreate(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	t.Run("creates a valid consumer", func(t *testing.T) {
		res := c.POST("/v1/consumers").WithJSON(goodConsumer()).Expect()
		res.Status(http.StatusCreated)
		res.Header("grpc-metadata-koko-status-code").Empty()
		body := res.JSON().Path("$.item").Object()
		body.Value("username").String().Equal("consumerA")
		body.Value("custom_id").String().Equal("customIDA")
		body.Value("id").String().NotEmpty()
		body.Value("created_at").Number().Gt(0)
		body.Value("updated_at").Number().Gt(0)
	})
	t.Run("creating a empty consumer fails with 400", func(t *testing.T) {
		consumer := &v1.Consumer{}
		res := c.POST("/v1/consumers").WithJSON(consumer).Expect()
		res.Status(400)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		err := body.Value("details").Array().Element(0).Object()
		err.ValueEqual("type", v1.ErrorType_ERROR_TYPE_ENTITY.String())
		err.Value("messages").Array().Element(0).String().
			Equal("at least one of custom_id or username must be set")
	})
	t.Run("recreating the consumer with the same username but different id fails",
		func(t *testing.T) {
			consumer := &v1.Consumer{}
			// Change the name to something that does not exist in the DB
			consumer.Username = "duplicateUserName"
			consumer.CustomId = ""
			res := c.POST("/v1/consumers").
				WithJSON(consumer).
				Expect()
			res.Status(201)
			res.Header("grpc-metadata-koko-status-code").Empty()
			// Now try to create a new consumer with same username
			res = c.POST("/v1/consumers").
				WithJSON(consumer).
				Expect()
			res.Status(http.StatusBadRequest)
			body := res.JSON().Object()
			body.ValueEqual("message", "data constraint error")
			body.Value("details").Array().Length().Equal(1)
			err := body.Value("details").Array().Element(0)
			err.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_REFERENCE.
				String())
			err.Object().ValueEqual("field", "username")
		})
	t.Run("recreating the consumer with the same customId but different id fails",
		func(t *testing.T) {
			consumer := goodConsumer()
			// Change the name to something that does not exist in the DB
			consumer.CustomId = "duplicateCustomID"
			consumer.Username = ""
			res := c.POST("/v1/consumers").
				WithJSON(consumer).
				Expect()
			res.Status(201)
			res.Header("grpc-metadata-koko-status-code").Empty()
			// Now try to create a new consumer with same CustomID
			res = c.POST("/v1/consumers").
				WithJSON(consumer).
				Expect()
			res.Status(http.StatusBadRequest)
			body := res.JSON().Object()
			body.ValueEqual("message", "data constraint error")
			body.Value("details").Array().Length().Equal(1)
			err := body.Value("details").Array().Element(0)
			err.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_REFERENCE.
				String())
			err.Object().ValueEqual("field", "custom_id")
		})
	t.Run("creates a valid consumer specifying the ID", func(t *testing.T) {
		consumer := goodConsumer()
		consumer.Username = "withID"
		consumer.CustomId = "withCustomID"
		consumer.Id = uuid.NewString()
		res := c.POST("/v1/consumers").WithJSON(consumer).Expect()
		res.Status(201)
		body := res.JSON().Path("$.item").Object()
		body.Value("id").Equal(consumer.Id)
	})
}

func TestConsumerUpsert(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	t.Run("upserts a valid consumer", func(t *testing.T) {
		res := c.PUT("/v1/consumers/" + uuid.NewString()).
			WithJSON(goodConsumer()).
			Expect()
		res.Status(http.StatusOK) // Should this be 201 since uuid is new
		res.Header("grpc-metadata-koko-status-code").Empty()
		body := res.JSON().Path("$.item").Object()
		body.Value("username").String().Equal("consumerA")
		body.Value("custom_id").String().Equal("customIDA")
		body.Value("id").String().NotEmpty()
		body.Value("created_at").Number().Gt(0)
		body.Value("updated_at").Number().Gt(0)
	})
	t.Run("upserting an invalid consumer fails with 400", func(t *testing.T) {
		consumer := &v1.Consumer{}
		res := c.PUT("/v1/consumers/" + uuid.NewString()).
			WithJSON(consumer).
			Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		err := body.Value("details").Array().Element(0).Object()
		err.ValueEqual("type", v1.ErrorType_ERROR_TYPE_ENTITY.String())
		err.Value("messages").Array().Element(0).String().
			Equal("at least one of custom_id or username must be set")
	})
	t.Run("recreating the consumer with the same username but different id fails",
		func(t *testing.T) {
			consumer := goodConsumer()
			consumer.Username = "foo"
			consumer.CustomId = ""
			res := c.PUT("/v1/consumers/" + uuid.NewString()).
				WithJSON(consumer).
				Expect()
			res.Status(http.StatusOK)
			res.Header("grpc-metadata-koko-status-code").Empty()
			body := res.JSON().Path("$.item").Object()
			body.Value("username").String().Equal("foo")
			body.Value("id").String().NotEmpty()
			body.Value("created_at").Number().Gt(0)
			body.Value("updated_at").Number().Gt(0)

			// Now upsert the same consumer with new ID
			res = c.PUT("/v1/consumers/" + uuid.NewString()).
				WithJSON(consumer).
				Expect()
			res.Status(http.StatusBadRequest)
			body = res.JSON().Object()
			body.ValueEqual("message", "data constraint error")
			body.Value("details").Array().Length().Equal(1)
			err := body.Value("details").Array().Element(0)
			err.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_REFERENCE.
				String())
			err.Object().ValueEqual("field", "username")
		})
	t.Run("recreating the consumer with the same customID but different id fails",
		func(t *testing.T) {
			consumer := goodConsumer()
			consumer.Username = ""
			consumer.CustomId = "InitialCustomID"
			res := c.PUT("/v1/consumers/" + uuid.NewString()).
				WithJSON(consumer).
				Expect()
			res.Status(http.StatusOK)
			res.Header("grpc-metadata-koko-status-code").Empty()
			body := res.JSON().Path("$.item").Object()
			body.Value("custom_id").String().Equal("InitialCustomID")
			body.Value("id").String().NotEmpty()
			body.Value("created_at").Number().Gt(0)
			body.Value("updated_at").Number().Gt(0)

			// Now upsert the same consumer with new ID
			res = c.PUT("/v1/consumers/" + uuid.NewString()).
				WithJSON(consumer).
				Expect()
			res.Status(http.StatusBadRequest)
			body = res.JSON().Object()
			body.ValueEqual("message", "data constraint error")
			body.Value("details").Array().Length().Equal(1)
			err := body.Value("details").Array().Element(0)
			err.Object().ValueEqual("type", v1.ErrorType_ERROR_TYPE_REFERENCE.
				String())
			err.Object().ValueEqual("field", "custom_id")
		})
	t.Run("upsert consumer without id fails", func(t *testing.T) {
		consumer := goodConsumer()
		res := c.PUT("/v1/consumers/").
			WithJSON(consumer).
			Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", " '' is not a valid uuid")
	})
	t.Run("upsert consumer with not a uuid id fails", func(t *testing.T) {
		consumer := goodConsumer()
		res := c.PUT("/v1/consumers/baduuid").
			WithJSON(consumer).
			Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", " 'baduuid' is not a valid uuid")
	})
}

func TestConsumerRead(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	consumer := goodConsumer()
	res := c.POST("/v1/consumers").WithJSON(consumer).Expect()
	res.Status(http.StatusCreated)
	res.Header("grpc-metadata-koko-status-code").Empty()
	body := res.JSON().Path("$.item").Object()
	body.Value("username").String().Equal("consumerA")
	body.Value("custom_id").String().Equal("customIDA")
	id := body.Value("id").String().Raw()
	require.NotEmpty(t, id)
	body.Value("created_at").Number().Gt(0)
	body.Value("updated_at").Number().Gt(0)
	t.Run("looking up a consumer with valid if succeeds", func(t *testing.T) {
		res := c.GET("/v1/consumers/" + id).WithJSON(consumer).Expect()
		res.Status(http.StatusOK)
		res.Header("grpc-metadata-koko-status-code").Empty()
		body := res.JSON().Path("$.item").Object()
		body.Value("username").String().Equal("consumerA")
		body.Value("custom_id").String().Equal("customIDA")
		gotID := body.Value("id").String().Raw()
		require.Equal(t, id, gotID)
		body.Value("created_at").Number().Gt(0)
		body.Value("updated_at").Number().Gt(0)
	})
	t.Run("reading a non-existent consumer returns not found", func(t *testing.T) {
		randomID := "071f5040-3e4a-46df-9d98-451e79e318fd"
		c.GET("/v1/consumers/" + randomID).Expect().Status(http.StatusNotFound)
	})
}

func TestConsumerList(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	consumer := goodConsumer()
	idList := make([]string, 0, 3)
	customIDList := make([]string, 0, 3)
	userNameList := make([]string, 0, 3)
	// Create 3 consumers
	for i := 0; i < 3; i++ {
		customID := fmt.Sprintf("CustomId-%d", i)
		customIDList = append(customIDList, customID)
		userName := fmt.Sprintf("UserName-%d", i)
		userNameList = append(userNameList, userName)
		consumer.CustomId = customID
		consumer.Username = userName
		res := c.POST("/v1/consumers").WithJSON(consumer).Expect()
		res.Status(http.StatusCreated)
		res.Header("grpc-metadata-koko-status-code").Empty()
		body := res.JSON().Path("$.item").Object()
		body.Value("username").String().Equal(userName)
		body.Value("custom_id").String().Equal(customID)
		id := body.Value("id").String().Raw()
		require.NotEmpty(t, id)
		idList = append(idList, id)
		body.Value("created_at").Number().Gt(0)
		body.Value("updated_at").Number().Gt(0)
	}
	t.Run("list consumers with default succeeds", func(t *testing.T) {
		res := c.GET("/v1/consumers").Expect()
		res.Status(http.StatusOK)
		body := res.JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(3)
		var gotIDs []string
		var gotUserNames []string
		var gotCustomIds []string
		for _, item := range items.Iter() {
			gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
			gotUserNames = append(gotUserNames, item.Object().Value("username").String().Raw())
			gotCustomIds = append(gotCustomIds, item.Object().Value("custom_id").String().Raw())
		}
		require.ElementsMatch(t, idList, gotIDs)
		require.ElementsMatch(t, userNameList, gotUserNames)
		require.ElementsMatch(t, customIDList, gotCustomIds)
	})
	t.Run("list consumers with paging succeeds", func(t *testing.T) {
		var gotIDs []string
		var gotUserNames []string
		var gotCustomIds []string
		// Get First Page
		body := c.GET("/v1/consumers").
			WithQuery("page.size", "1").
			WithQuery("page.number", "1").
			Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(1)
		gotIDOne := items.Element(0).Object().Value("id").String().Raw()
		require.NotEmpty(t, gotIDOne)
		gotIDs = append(gotIDs, gotIDOne)
		gotUserNameOne := items.Element(0).Object().Value("username").String().Raw()
		require.NotEmpty(t, gotUserNameOne)
		gotUserNames = append(gotUserNames, gotUserNameOne)
		gotCustomIDOne := items.Element(0).Object().Value("custom_id").String().Raw()
		require.NotEmpty(t, gotCustomIDOne)
		gotCustomIds = append(gotCustomIds, gotCustomIDOne)
		body.Value("page").Object().Value("total_count").Number().Equal(3)
		body.Value("page").Object().Value("next_page_num").Number().Equal(2)

		// Second Page
		body = c.GET("/v1/consumers").
			WithQuery("page.size", "1").
			WithQuery("page.number", "2").
			Expect().Status(http.StatusOK).JSON().Object()
		items = body.Value("items").Array()
		items.Length().Equal(1)
		body.Value("page").Object().Value("total_count").Number().Equal(3)
		body.Value("page").Object().Value("next_page_num").Number().Equal(3)
		gotIDTwo := items.Element(0).Object().Value("id").String().Raw()
		require.NotEmpty(t, gotIDTwo)
		gotIDs = append(gotIDs, gotIDTwo)
		gotUserNameTwo := items.Element(0).Object().Value("username").String().Raw()
		require.NotEmpty(t, gotUserNameTwo)
		gotUserNames = append(gotUserNames, gotUserNameTwo)
		gotCustomIDTwo := items.Element(0).Object().Value("custom_id").String().Raw()
		require.NotEmpty(t, gotCustomIDTwo)
		gotCustomIds = append(gotCustomIds, gotCustomIDTwo)

		// Last Page
		body = c.GET("/v1/consumers").
			WithQuery("page.size", "1").
			WithQuery("page.number", "3").
			Expect().Status(http.StatusOK).JSON().Object()
		items = body.Value("items").Array()
		items.Length().Equal(1)
		gotIDThree := items.Element(0).Object().Value("id").String().Raw()
		require.NotEmpty(t, gotIDThree)
		gotIDs = append(gotIDs, gotIDThree)
		gotUserNameThree := items.Element(0).Object().Value("username").String().Raw()
		require.NotEmpty(t, gotUserNameThree)
		gotUserNames = append(gotUserNames, gotUserNameThree)
		gotCustomIDThree := items.Element(0).Object().Value("custom_id").String().Raw()
		require.NotEmpty(t, gotCustomIDThree)
		gotCustomIds = append(gotCustomIds, gotCustomIDThree)

		body.Value("page").Object().Value("total_count").Number().Equal(3)
		body.Value("page").Object().NotContainsKey("next_page_num")

		require.ElementsMatch(t, idList, gotIDs)
		require.ElementsMatch(t, userNameList, gotUserNames)
		require.ElementsMatch(t, customIDList, gotCustomIds)
	})
	t.Run("read request on resource with slash returns bad request", func(t *testing.T) {
		res := c.GET("/v1/consumers/").Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "required ID is missing")
	})
}
