package cmd

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/store"
	"github.com/kong/koko/internal/test/util"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRun_RegisterInstallation(t *testing.T) {
	l, err := zap.NewDevelopment()
	require.NoError(t, err)
	st := setupTestDB(t, l)

	t.Run("RegisterInstallation returns a valid UUID", func(t *testing.T) {
		id, err := registerInstallation(context.Background(), st, l)
		require.NoError(t, err)
		require.True(t, validUUID(id))
	})

	t.Run("does not overwrite existing ID", func(t *testing.T) {
		id, err := registerInstallation(context.Background(), st, l)
		require.NoError(t, err)
		id2, err := registerInstallation(context.Background(), st, l)
		require.NoError(t, err)
		require.Equal(t, id, id2)
	})
}

func TestRun_GetInstallationID(t *testing.T) {
	l, err := zap.NewDevelopment()
	require.NoError(t, err)
	st := setupTestDB(t, l)
	inst := resource.NewInstallation()

	t.Run("returns error not found when ID is not in the store", func(t *testing.T) {
		id, err := getInstallationID(context.Background(), st, inst)
		require.ErrorContains(t, err, "not found")
		require.Equal(t, "", id)
	})

	t.Run("returns valid UUID", func(t *testing.T) {
		_, err := setInstallationID(context.Background(), st, inst, l)
		require.NoError(t, err)
		id, err := getInstallationID(context.Background(), st, inst)
		require.NoError(t, err)
		require.True(t, validUUID(id))
	})
}

func TestRun_SetInstallationID(t *testing.T) {
	l, err := zap.NewDevelopment()
	require.NoError(t, err)
	st := setupTestDB(t, l)
	inst := resource.NewInstallation()

	t.Run("sets valid UUID", func(t *testing.T) {
		id, err := setInstallationID(context.Background(), st, inst, l)
		require.NoError(t, err)
		require.True(t, validUUID(id))
	})

	t.Run("does not overwrite existing ID", func(t *testing.T) {
		id, err := setInstallationID(context.Background(), st, inst, l)
		require.NoError(t, err)
		id2, err := setInstallationID(context.Background(), st, inst, l)
		require.NoError(t, err)
		require.Equal(t, id, id2)
	})
}

func setupTestDB(t *testing.T, logger *zap.Logger) store.Store {
	p, err := util.GetPersister(t)
	require.NoError(t, err)
	return store.New(p, logger.With(zap.String("component", "test-store"))).ForCluster("default")
}

func validUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}
