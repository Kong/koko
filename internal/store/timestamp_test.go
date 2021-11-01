package store

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/log"
	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/test/util"
	"github.com/stretchr/testify/require"
)

func TestAddTS(t *testing.T) {
	t.Run("timestamps are added by addTS()", func(t *testing.T) {
		svc := resource.NewService()
		svc.Service = &v1.Service{
			Id:   uuid.NewString(),
			Name: "bar",
			Host: "foo.com",
			Path: "/bar",
		}
		addTS(svc.Resource())
		require.NotEmpty(t, svc.Service.CreatedAt)
		require.NotEmpty(t, svc.Service.UpdatedAt)
		createdAt := time.Unix(int64(svc.Service.CreatedAt), 0)
		updatedAt := time.Unix(int64(svc.Service.UpdatedAt), 0)
		now := time.Now()
		// reasonably be sure that current time was used
		require.True(t, now.Sub(createdAt) < 1*time.Second)
		require.True(t, now.Sub(updatedAt) < 1*time.Second)
	})
	t.Run("timestamps are added to persisted resource", func(t *testing.T) {
		persister := util.GetPersister(t)
		s := New(persister, log.Logger).ForCluster("default")
		svc := resource.NewService()
		id := uuid.NewString()
		svc.Service = &v1.Service{
			Id:   id,
			Name: "bar",
			Host: "foo.com",
			Path: "/bar",
		}
		require.Nil(t, s.Create(context.Background(), svc))
		require.NotEmpty(t, svc.Service.CreatedAt)
		require.NotEmpty(t, svc.Service.UpdatedAt)
	})
	t.Run("timestamps provided in input are overridden", func(t *testing.T) {
		persister := util.GetPersister(t)
		s := New(persister, log.Logger).ForCluster("default")
		svc := resource.NewService()
		id := uuid.NewString()
		svc.Service = &v1.Service{
			Id:        id,
			Name:      "bar",
			Host:      "foo.com",
			Path:      "/bar",
			CreatedAt: 42,
			UpdatedAt: 42,
		}
		require.Nil(t, s.Create(context.Background(), svc))
		require.NotEqual(t, 42, svc.Service.CreatedAt)
		require.NotEqual(t, 42, svc.Service.UpdatedAt)
		createdAt := time.Unix(int64(svc.Service.CreatedAt), 0)
		updatedAt := time.Unix(int64(svc.Service.UpdatedAt), 0)
		now := time.Now()
		// reasonably be sure that current time was used
		require.True(t, now.Sub(createdAt) < 1*time.Second)
		require.True(t, now.Sub(updatedAt) < 1*time.Second)
	})
}
