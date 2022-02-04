package resource

import (
	"testing"

	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/stretchr/testify/require"
)

func TestConsumer_ID(t *testing.T) {
	var c Consumer
	id := c.ID()
	require.Empty(t, id)
	c = NewConsumer()
	id = c.ID()
	require.Empty(t, id)
}

func TestConsumer_Type(t *testing.T) {
	require.Equal(t, TypeConsumer, NewConsumer().Type())
}

func TestConsumer_Defaults(t *testing.T) {
	c := NewConsumer()
	err := c.ProcessDefaults()
	require.Nil(t, err)
	require.True(t, validUUID(c.ID()))
}

func goodConsumer() Consumer {
	c := NewConsumer()
	_ = c.ProcessDefaults()
	c.Consumer.Username = "my-company"
	c.Consumer.CustomId = "my-company-ID"
	return c
}

func TestConsumer_Validate(t *testing.T) {
	t.Run("empty consumer must fail", func(t *testing.T) {
		c := NewConsumer()
		err := c.Validate()
		require.Error(t, err)
		verr, ok := err.(validation.Error)
		require.True(t, ok)
		e := []*model.ErrorDetail{
			{
				Type: model.ErrorType_ERROR_TYPE_ENTITY,
				Messages: []string{
					"missing properties: 'id'",
					"missing properties: 'custom_id'",
					"missing properties: 'username'",
				},
			},
		}
		require.ElementsMatch(t, verr.Errs, e)
	})
	t.Run("default consumer must fail", func(t *testing.T) {
		c := NewConsumer()
		_ = c.ProcessDefaults()
		err := c.Validate()
		require.Error(t, err)
		verr, ok := err.(validation.Error)
		require.True(t, ok)
		e := []*model.ErrorDetail{
			{
				Type: model.ErrorType_ERROR_TYPE_ENTITY,
				Messages: []string{
					"missing properties: 'custom_id'",
					"missing properties: 'username'",
				},
			},
		}
		require.ElementsMatch(t, verr.Errs, e)
	})
	t.Run("good consumer with username and no custom_id must pass", func(t *testing.T) {
		c := NewConsumer()
		_ = c.ProcessDefaults()
		c.Consumer.Username = "my-company-name"
		err := c.Validate()
		require.NoError(t, err)
	})
	t.Run("good consumer with custom_id and no username must pass", func(t *testing.T) {
		c := NewConsumer()
		_ = c.ProcessDefaults()
		c.Consumer.CustomId = "my-company-ID"
		err := c.Validate()
		require.NoError(t, err)
	})
	t.Run("good consumer with custom_id and username must pass", func(t *testing.T) {
		c := goodConsumer()
		err := c.Validate()
		require.NoError(t, err)
	})
}
