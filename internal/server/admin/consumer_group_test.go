package admin

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/google/uuid"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
)

func TestConsumerGroupIsNotExposed(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()
	c := httpexpect.Default(t, s.URL)

	t.Run("cannot GET consumer-groups", func(t *testing.T) {
		res := c.GET("/v1/consumer-groups").Expect()
		res.Status(http.StatusNotFound)
	})
	t.Run("cannot POST consumer-group", func(t *testing.T) {
		cg := &v1.ConsumerGroup{
			Name: "foo",
		}
		res := c.POST("/v1/consumer-groups").WithJSON(cg).Expect()
		res.Status(http.StatusNotFound)
	})
	t.Run("cannot PUT consumer-group", func(t *testing.T) {
		cg := &v1.ConsumerGroup{
			Name: "foo",
		}
		res := c.PUT("/v1/consumer-groups/" + uuid.NewString()).WithJSON(cg).Expect()
		res.Status(http.StatusNotFound)
	})
}
