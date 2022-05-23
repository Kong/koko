package resource

import (
	"context"
	"testing"

	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/stretchr/testify/require"
)

func TestHash_ID(t *testing.T) {
	h := NewHash()
	require.Equal(t, "config-hash-id", h.ID())

	var emptyHash Hash
	require.Equal(t, "", emptyHash.ID())
}

func TestHash_Type(t *testing.T) {
	require.Equal(t, TypeHash, NewHash().Type())
}

func TestHash_Validate(t *testing.T) {
	t.Run("empty hash must fail", func(t *testing.T) {
		h := NewHash()
		err := h.Validate(context.Background())
		require.Error(t, err)
		verr, ok := err.(validation.Error)
		require.True(t, ok)
		e := []*model.ErrorDetail{
			{
				Type: model.ErrorType_ERROR_TYPE_ENTITY,
				Messages: []string{
					"missing properties: 'expected_hash'",
				},
			},
		}
		require.ElementsMatch(t, verr.Errs, e)
	})
	t.Run("default hash must fail", func(t *testing.T) {
		h := NewHash()
		_ = h.ProcessDefaults(context.Background())
		err := h.Validate(context.Background())
		require.Error(t, err)
		verr, ok := err.(validation.Error)
		require.True(t, ok)
		e := []*model.ErrorDetail{
			{
				Type: model.ErrorType_ERROR_TYPE_ENTITY,
				Messages: []string{
					"missing properties: 'expected_hash'",
				},
			},
		}
		require.ElementsMatch(t, verr.Errs, e)
	})
	t.Run("good hash with a valid hash must pass", func(t *testing.T) {
		h := NewHash()
		_ = h.ProcessDefaults(context.Background())
		h.Hash.ExpectedHash = "37f525e2b6fc3cb4abd882f708ab80eb"
		err := h.Validate(context.Background())
		require.NoError(t, err)
	})
	t.Run("bad hash with invalid hash fails", func(t *testing.T) {
		h := NewHash()
		_ = h.ProcessDefaults(context.Background())
		h.Hash.ExpectedHash = "37f525e2b6fc3cb4abd882f708ab80ex"
		err := h.Validate(context.Background())
		verr, ok := err.(validation.Error)
		require.True(t, ok)
		e := []*model.ErrorDetail{
			{
				Type:  model.ErrorType_ERROR_TYPE_FIELD,
				Field: "expected_hash",
				Messages: []string{
					"must match pattern '^[0-9a-f]{32}$'",
				},
			},
		}
		require.ElementsMatch(t, verr.Errs, e)
	})
	t.Run("bad hash with invalid length fails", func(t *testing.T) {
		h := NewHash()
		_ = h.ProcessDefaults(context.Background())
		h.Hash.ExpectedHash = "37f525e2b6fc3cb4abd882f708ab80eff"
		err := h.Validate(context.Background())
		verr, ok := err.(validation.Error)
		require.True(t, ok)
		e := []*model.ErrorDetail{
			{
				Type:  model.ErrorType_ERROR_TYPE_FIELD,
				Field: "expected_hash",
				Messages: []string{
					"must match pattern '^[0-9a-f]{32}$'",
				},
			},
		}
		require.ElementsMatch(t, verr.Errs, e)
	})
}
