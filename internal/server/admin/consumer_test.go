package admin

import (
	"testing"

	"github.com/gavv/httpexpect/v2"
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
		upstream := &v1.Consumer{}
		res := c.POST("/v1/consumers").WithJSON(upstream).Expect()
		res.Status(400)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		err := body.Value("details").Array().Element(0).Object()
		err.ValueEqual("type", v1.ErrorType_ERROR_TYPE_ENTITY.String())
		err.Value("messages").Array().Element(0).String().
			Equal("at-least one of custom_id or username must be set")
	})
}
