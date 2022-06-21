package resource

import (
	"context"
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
	err := c.ProcessDefaults(context.Background())
	require.Nil(t, err)
	require.True(t, validUUID(c.ID()))
}

func goodConsumer() Consumer {
	c := NewConsumer()
	_ = c.ProcessDefaults(context.Background())
	c.Consumer.Username = "my-company"
	c.Consumer.CustomId = "my-company-ID"
	return c
}

func TestConsumer_Validate(t *testing.T) {
	t.Run("empty consumer must fail", func(t *testing.T) {
		c := NewConsumer()
		err := c.Validate(context.Background())
		require.Error(t, err)
		verr, ok := err.(validation.Error)
		require.True(t, ok)
		e := []*model.ErrorDetail{
			{
				Type: model.ErrorType_ERROR_TYPE_ENTITY,
				Messages: []string{
					"missing properties: 'id'",
					"at least one of custom_id or username must be set",
				},
			},
		}
		require.ElementsMatch(t, verr.Errs, e)
	})
	t.Run("default consumer must fail", func(t *testing.T) {
		c := NewConsumer()
		_ = c.ProcessDefaults(context.Background())
		err := c.Validate(context.Background())
		require.Error(t, err)
		verr, ok := err.(validation.Error)
		require.True(t, ok)
		e := []*model.ErrorDetail{
			{
				Type: model.ErrorType_ERROR_TYPE_ENTITY,
				Messages: []string{
					"at least one of custom_id or username must be set",
				},
			},
		}
		require.ElementsMatch(t, verr.Errs, e)
	})
	t.Run("good consumer with username and no custom_id must pass", func(t *testing.T) {
		c := NewConsumer()
		_ = c.ProcessDefaults(context.Background())
		c.Consumer.Username = "john.doe+koko@example.org"
		err := c.Validate(context.Background())
		require.NoError(t, err)
	})
	t.Run("good consumer with username containing whitespaces must pass", func(t *testing.T) {
		c := NewConsumer()
		_ = c.ProcessDefaults(context.Background())
		c.Consumer.Username = "John Doe 09/02"
		err := c.Validate(context.Background())
		require.NoError(t, err)
	})
	t.Run("good consumer with custom_id and no username must pass", func(t *testing.T) {
		c := NewConsumer()
		_ = c.ProcessDefaults(context.Background())
		c.Consumer.CustomId = "my-company-ID"
		err := c.Validate(context.Background())
		require.NoError(t, err)
	})
	t.Run("good consumer with custom_id and username must pass", func(t *testing.T) {
		c := goodConsumer()
		err := c.Validate(context.Background())
		require.NoError(t, err)
	})
	t.Run("good consumer with custom_id containing spaces must pass", func(t *testing.T) {
		c := goodConsumer()
		c.Consumer.CustomId = "my company ID"
		err := c.Validate(context.Background())
		require.NoError(t, err)
	})
	t.Run("good consumer with custom_id containing allowed special characters must pass", func(t *testing.T) {
		c := goodConsumer()
		c.Consumer.CustomId = "my company ID #%@|.-_~()+"
		err := c.Validate(context.Background())
		require.NoError(t, err)
	})
	t.Run("custom_id beginning with a space must fail", func(t *testing.T) {
		c := goodConsumer()
		c.Consumer.CustomId = " my company ID"
		err := c.Validate(context.Background())
		require.Error(t, err)
		verr, ok := err.(validation.Error)
		require.True(t, ok)
		e := []*model.ErrorDetail{
			{
				Type:  model.ErrorType_ERROR_TYPE_FIELD,
				Field: "custom_id",
				Messages: []string{
					`must match pattern '^[0-9a-zA-Z.\-_~\(\)#%@|+]+(?: [0-9a-zA-Z.\-_~\(\)#%@|+]+)*$'`,
				},
			},
		}
		require.ElementsMatch(t, verr.Errs, e)
	})
	t.Run("custom_id ending with a space must fail", func(t *testing.T) {
		c := goodConsumer()
		c.Consumer.CustomId = "my company ID "
		err := c.Validate(context.Background())
		require.Error(t, err)
		verr, ok := err.(validation.Error)
		require.True(t, ok)
		e := []*model.ErrorDetail{
			{
				Type:  model.ErrorType_ERROR_TYPE_FIELD,
				Field: "custom_id",
				Messages: []string{
					`must match pattern '^[0-9a-zA-Z.\-_~\(\)#%@|+]+(?: [0-9a-zA-Z.\-_~\(\)#%@|+]+)*$'`,
				},
			},
		}
		require.ElementsMatch(t, verr.Errs, e)
	})
	t.Run("invalid custom_id must fail", func(t *testing.T) {
		c := goodConsumer()
		c.Consumer.CustomId = "my company ID!"
		err := c.Validate(context.Background())
		require.Error(t, err)
		verr, ok := err.(validation.Error)
		require.True(t, ok)
		e := []*model.ErrorDetail{
			{
				Type:  model.ErrorType_ERROR_TYPE_FIELD,
				Field: "custom_id",
				Messages: []string{
					`must match pattern '^[0-9a-zA-Z.\-_~\(\)#%@|+]+(?: [0-9a-zA-Z.\-_~\(\)#%@|+]+)*$'`,
				},
			},
		}
		require.ElementsMatch(t, verr.Errs, e)
	})
}
