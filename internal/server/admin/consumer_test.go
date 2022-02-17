package admin

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/google/uuid"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
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
		res.Status(201)
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
			Equal("at-least one of custom_id or username must be set")
	})
	t.Run("recreating the consumer with the same username but different id fails",
		func(t *testing.T) {
			consumer := goodConsumer()
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
}

func TestConsumerUpsert(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.New(t, s.URL)
	t.Run("upserts a valid consumer", func(t *testing.T) {
		res := c.PUT("/v1/consumers/" + uuid.NewString()).
			WithJSON(goodConsumer()).
			Expect()
		res.Status(http.StatusOK)
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
			Equal("at-least one of custom_id or username must be set")
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
	t.Run("upsert consumer without id fails", func(t *testing.T) {
		consumer := goodConsumer()
		res := c.PUT("/v1/consumers/").
			WithJSON(consumer).
			Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", " '' is not a valid uuid")
	})
}
