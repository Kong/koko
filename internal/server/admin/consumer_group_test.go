package admin

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/google/uuid"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/test/util"
)

func TestConsumerGroupIsNotExposed(t *testing.T) {
	util.SkipTestIfEnterpriseTesting(t, true)

	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.Default(t, s.URL)

	t.Run("cannot POST consumer-group", func(t *testing.T) {
		cg := &v1.ConsumerGroup{
			Name: "foo",
		}
		res := c.POST("/v1/consumer-groups").WithJSON(cg).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		err := body.Value("details").Array().Element(0).Object()
		err.ValueEqual("type", v1.ErrorType_ERROR_TYPE_ENTITY.String())
		err.Value("messages").Array().Element(0).String().
			Equal(`Consumer Groups are a Kong Enterprise-only feature. ` +
				`Please upgrade to Kong Enterprise to use this feature.`)
	})
	t.Run("cannot PUT consumer-group", func(t *testing.T) {
		cg := &v1.ConsumerGroup{
			Name: "foo",
		}
		res := c.PUT("/v1/consumer-groups/" + uuid.NewString()).WithJSON(cg).Expect()
		res.Status(http.StatusBadRequest)
		body := res.JSON().Object()
		body.ValueEqual("message", "validation error")
		body.Value("details").Array().Length().Equal(1)
		err := body.Value("details").Array().Element(0).Object()
		err.ValueEqual("type", v1.ErrorType_ERROR_TYPE_ENTITY.String())
		err.Value("messages").Array().Element(0).String().
			Equal(`Consumer Groups are a Kong Enterprise-only feature. ` +
				`Please upgrade to Kong Enterprise to use this feature.`)
	})
}
