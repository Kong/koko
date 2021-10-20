package store

import (
	"context"
	"testing"

	"github.com/google/uuid"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/log"
	"github.com/kong/koko/internal/persistence"
	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/store/event"
	"github.com/stretchr/testify/require"
)

func TestService(t *testing.T) {
	// TODO improve these tests
	persister, err := persistence.NewSQLite(":memory:")
	require.Nil(t, err)
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
	svc = resource.NewService()
	require.Nil(t, s.Read(context.Background(), svc, GetByID(id)))
	require.Equal(t, "bar", svc.Service.Name)
	require.Nil(t, s.Delete(context.Background(),
		DeleteByType(resource.TypeService), DeleteByID(svc.ID())))
	require.NotNil(t, s.Read(context.Background(), svc, GetByID(id)))

	svc = resource.NewService()
	svc.Service = &v1.Service{
		Id:   uuid.NewString(),
		Name: "bar",
		Host: "foo.com",
		Path: "/bar",
	}
	require.Nil(t, s.Create(context.Background(), svc))

	svc = resource.NewService()
	svc.Service = &v1.Service{
		Id:   uuid.NewString(),
		Name: "foo",
		Host: "foo.com",
		Path: "/bar",
	}
	require.Nil(t, s.Create(context.Background(), svc))
	svcs := resource.NewList(resource.TypeService)
	err = s.List(context.Background(), svcs)
	require.Nil(t, err)
	require.Len(t, svcs.GetAll(), 2)
}

func TestForCluster(t *testing.T) {
	persister, err := persistence.NewSQLite(":memory:")
	require.Nil(t, err)
	s := New(persister, log.Logger)
	require.Empty(t, s.cluster)
	// validate clusterRegex is being used
	require.Panics(t, func() {
		s.ForCluster("borked.on.purpose")
	})
	// validate clusterKey panics correctly
	require.Panics(t, func() {
		s.clusterKey("foo")
	})
}

func TestUpdateEvent(t *testing.T) {
	persister, err := persistence.NewMemory()
	require.Nil(t, err)
	s := New(persister, log.Logger).ForCluster("default")
	ctx := context.Background()
	t.Run("empty cluster has no update event", func(t *testing.T) {
		e := event.New()
		err := s.Read(ctx, e, GetByID(event.ID))
		require.Equal(t, ErrNotFound, err)
	})
	t.Run("create creates an update event", func(t *testing.T) {
		svc := resource.NewService()
		id := uuid.NewString()
		svc.Service = &v1.Service{
			Id:   id,
			Name: "bar",
			Host: "foo.com",
			Path: "/bar",
		}
		require.Nil(t, s.Create(ctx, svc))
		e := event.New()
		err := s.Read(ctx, e, GetByID(event.ID))
		require.Nil(t, err)
		require.NotPanics(t, func() {
			uuid.MustParse(e.StoreEvent.Value)
		})
	})
	t.Run("delete creates a new update event", func(t *testing.T) {
		svc := resource.NewService()
		id := uuid.NewString()
		svc.Service = &v1.Service{
			Id:   id,
			Name: "bar0",
			Host: "foo.com",
			Path: "/bar",
		}
		require.Nil(t, s.Create(ctx, svc))
		e := event.New()
		err := s.Read(ctx, e, GetByID(event.ID))
		require.Nil(t, err)
		createEventID := e.StoreEvent.Value
		// verify value is a UUID
		require.NotPanics(t, func() {
			uuid.MustParse(createEventID)
		})

		require.Nil(t, s.Delete(ctx, DeleteByID(id),
			DeleteByType(resource.TypeService)))
		e = event.New()
		err = s.Read(ctx, e, GetByID(event.ID))
		require.Nil(t, err)
		deleteEventID := e.StoreEvent.Value
		require.NotPanics(t, func() {
			uuid.MustParse(deleteEventID)
		})
		require.NotEqual(t, createEventID, deleteEventID)
	})
}
