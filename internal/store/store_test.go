package store

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/log"
	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/store/event"
	"github.com/kong/koko/internal/test/util"
	"github.com/stretchr/testify/require"
)

func TestService(t *testing.T) {
	// TODO improve these tests
	persister, err := util.GetPersister()
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
	persister, err := util.GetPersister()
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
	persister, err := util.GetPersister()
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

func TestStoredValue(t *testing.T) {
	t.Run("store value is a JSON string", func(t *testing.T) {
		ctx := context.Background()
		persister, err := util.GetPersister()
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
		require.Nil(t, s.Create(ctx, svc))
		key, err := s.genID(resource.TypeService, id)
		require.Nil(t, err)
		value, err := persister.Get(ctx, key)
		require.Nil(t, err)
		var v map[string]interface{}
		err = json.Marshaller.Unmarshal(value, &v)
		require.Nil(t, err)

		require.Equal(t, "bar", v["name"])
		require.Equal(t, id, v["id"])
		require.Equal(t, "foo.com", v["host"])
		require.Equal(t, "/bar", v["path"])
	})
}

func TestIndexForeign(t *testing.T) {
	t.Run("object with foreign references cannot be deleted", func(t *testing.T) {
		ctx := context.Background()
		persister, err := util.GetPersister()
		require.Nil(t, err)
		s := New(persister, log.Logger).ForCluster("default")
		svc := resource.NewService()
		serviceID := uuid.NewString()
		svc.Service = &v1.Service{
			Id:   serviceID,
			Name: "bar",
			Host: "foo.com",
			Path: "/bar",
		}
		require.Nil(t, s.Create(ctx, svc))

		route := resource.NewRoute()
		routeID := uuid.NewString()
		route.Route = &v1.Route{
			Id:    routeID,
			Name:  "foo",
			Hosts: []string{"example.com"},
			Service: &v1.Service{
				Id: serviceID,
			},
		}
		require.Nil(t, s.Create(ctx, route))

		err = s.Delete(ctx, DeleteByType(resource.TypeService),
			DeleteByID(serviceID))
		require.NotNil(t, err)
		require.True(t, errors.As(err, &ErrConstraint{}))
	})
}
