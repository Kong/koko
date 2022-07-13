package resource

import (
	"context"
	"testing"

	"github.com/google/uuid"
	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/stretchr/testify/require"
)

func TestNewInstallation(t *testing.T) {
	i := NewInstallation()
	require.NotNil(t, i)
	require.NotNil(t, i.Installation)
}

func TestInstallation_ID(t *testing.T) {
	i := NewInstallation()
	require.Equal(t, "installation_id", i.ID())
}

func TestInstallation_Type(t *testing.T) {
	require.Equal(t, TypeInstallation, NewInstallation().Type())
}

func TestInstallation_Validate(t *testing.T) {
	t.Run("empty installation must fail", func(t *testing.T) {
		i := NewInstallation()
		err := i.Validate(context.Background())
		require.Error(t, err)
		verr, ok := err.(validation.Error)

		require.True(t, ok)
		e := []*model.ErrorDetail{
			{
				Type: model.ErrorType_ERROR_TYPE_ENTITY,
				Messages: []string{
					"missing properties: 'value'",
				},
			},
		}
		require.ElementsMatch(t, verr.Errs, e)
	})
	t.Run("valid installation must pass", func(t *testing.T) {
		i := NewInstallation()
		i.Installation.Value = uuid.NewString()
		err := i.Validate(context.Background())
		require.NoError(t, err)
	})
}
