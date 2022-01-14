package store

import (
	"context"
	encodingJSON "encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/log"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/store/event"
	"github.com/kong/koko/internal/test/util"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestCreate(t *testing.T) {
	persister, err := util.GetPersister()
	require.Nil(t, err)
	ctx := context.Background()
	s := New(persister, log.Logger).ForCluster("default")
	t.Run("creating a nil object fails", func(t *testing.T) {
		err := s.Create(ctx, nil)
		require.Equal(t, errNoObject, err)
	})
	t.Run("creating an object that fails validation fails", func(t *testing.T) {
		svc := resource.NewService()
		id := uuid.NewString()
		svc.Service = &v1.Service{
			Id:   id,
			Name: "%bar",
			Host: "foo.com",
			Path: "/bar",
		}
		err := s.Create(ctx, svc)
		require.IsType(t, validation.Error{}, err)
	})
	t.Run("creating an object without ID generates an ID", func(t *testing.T) {
		svc := resource.NewService()
		svc.Service = &v1.Service{
			Name: "s0",
			Host: "foo.com",
			Path: "/bar",
		}
		err := s.Create(ctx, svc)
		require.Nil(t, err)
		require.NotPanics(t, func() {
			uuid.MustParse(svc.Service.Id)
		})
	})
	t.Run("creating an object with unique index violation fails", func(t *testing.T) {
		svc := resource.NewService()
		svc.Service = &v1.Service{
			Name: "s0",
			Host: "foo.com",
			Path: "/bar",
		}
		err := s.Create(ctx, svc)
		require.IsType(t, ErrConstraint{}, err)
		constraintErr, ok := err.(ErrConstraint)
		require.True(t, ok)
		require.Equal(t,
			model.Index{
				Name:      "name",
				FieldName: "name",
				Type:      "unique",
				Value:     "s0",
			},
			constraintErr.Index)
		require.Equal(t, "", constraintErr.Message)
	})
	t.Run("creating an object with valid foreign reference succeeds", func(t *testing.T) {
		svc := resource.NewService()
		sid := uuid.NewString()
		svc.Service = &v1.Service{
			Id:   sid,
			Name: "s1",
			Host: "foo.com",
			Path: "/bar",
		}
		err := s.Create(ctx, svc)
		require.Nil(t, err)

		route := resource.NewRoute()
		route.Route = &v1.Route{
			Name:  "r0",
			Hosts: []string{"example.com"},
			Service: &v1.Service{
				Id: sid,
			},
		}
		require.Nil(t, s.Create(ctx, route))
	})
	t.Run("creating an object with invalid foreign reference fails", func(t *testing.T) {
		route := resource.NewRoute()
		sid := uuid.NewString()
		route.Route = &v1.Route{
			Name:  "r1",
			Hosts: []string{"example.com"},
			Service: &v1.Service{
				Id: sid,
			},
		}
		err := s.Create(ctx, route)
		require.IsType(t, ErrConstraint{}, err)
		constraintErr, ok := err.(ErrConstraint)
		require.True(t, ok)
		require.Equal(t,
			model.Index{
				Name:        "svc_id",
				FieldName:   "service.id",
				Type:        "foreign",
				ForeignType: "service",
				Value:       sid,
			},
			constraintErr.Index)
		require.Equal(t, "", constraintErr.Message)
	})
}

func TestRead(t *testing.T) {
	persister, err := util.GetPersister()
	require.Nil(t, err)
	ctx := context.Background()
	s := New(persister, log.Logger).ForCluster("default")
	svc := resource.NewService()
	sid := uuid.NewString()
	svc.Service = &v1.Service{
		Id:   sid,
		Name: "s0",
		Host: "foo.com",
		Path: "/bar",
	}
	err = s.Create(ctx, svc)
	require.Nil(t, err)
	t.Run("reading an  object succeeds", func(t *testing.T) {
		svc := resource.NewService()
		err := s.Read(context.Background(), svc, GetByID(sid))
		require.Nil(t, err)
		require.Equal(t, "s0", svc.Service.Name)
		require.Equal(t, "foo.com", svc.Service.Host)
		require.Equal(t, "/bar", svc.Service.Path)
	})
	t.Run("deleting a non-existent object fails", func(t *testing.T) {
		svc := resource.NewService()
		err := s.Read(context.Background(), svc, GetByID(uuid.NewString()))
		require.IsType(t, ErrNotFound, err)
	})
}

func TestDelete(t *testing.T) {
	persister, err := util.GetPersister()
	require.Nil(t, err)
	s := New(persister, log.Logger).ForCluster("default")
	t.Run("deleting an existing object succeeds", func(t *testing.T) {
		svc := resource.NewService()
		id := uuid.NewString()
		svc.Service = &v1.Service{
			Id:   id,
			Name: "s0",
			Host: "foo.com",
			Path: "/bar",
		}
		require.Nil(t, s.Create(context.Background(), svc))

		require.Nil(t, s.Delete(context.Background(),
			DeleteByType(resource.TypeService), DeleteByID(svc.ID())))

		// read again to ensure object was actually deleted
		require.NotNil(t, s.Read(context.Background(), svc, GetByID(id)))
	})
	t.Run("deleting a non-existent object fails", func(t *testing.T) {
		err := s.Delete(context.Background(),
			DeleteByType(resource.TypeService),
			DeleteByID(uuid.NewString()),
		)
		require.IsType(t, ErrNotFound, err)
	})
	t.Run("deleting an object with foreign-references fails", func(t *testing.T) {
		ctx := context.Background()
		persister, err := util.GetPersister()
		require.Nil(t, err)
		s := New(persister, log.Logger).ForCluster("default")
		svc := resource.NewService()
		sid := uuid.NewString()
		svc.Service = &v1.Service{
			Id:   sid,
			Name: "s1",
			Host: "foo.com",
			Path: "/bar",
		}
		require.Nil(t, s.Create(ctx, svc))

		route := resource.NewRoute()
		rid := uuid.NewString()
		route.Route = &v1.Route{
			Id:    rid,
			Name:  "r0",
			Hosts: []string{"example.com"},
			Service: &v1.Service{
				Id: sid,
			},
		}
		require.Nil(t, s.Create(ctx, route))

		err = s.Delete(ctx, DeleteByType(resource.TypeService),
			DeleteByID(sid))
		require.NotNil(t, err)
		constraintErr, ok := err.(ErrConstraint)
		require.True(t, ok)
		errMessage := fmt.Sprintf("foreign reference exist: %s (id: %s)",
			"route", rid)
		require.Equal(t, errMessage, constraintErr.Message)
	})
}

func TestUpsert(t *testing.T) {
	persister, err := util.GetPersister()
	require.Nil(t, err)
	s := New(persister, log.Logger).ForCluster("default")
	ctx := context.Background()

	t.Run("upsert a nil object fails", func(t *testing.T) {
		err := s.Upsert(ctx, nil)
		require.Equal(t, errNoObject, err)
	})
	t.Run("upsert an object that fails validation fails",
		func(t *testing.T) {
			svc := resource.NewService()
			id := uuid.NewString()
			svc.Service = &v1.Service{
				Id:   id,
				Name: "%bar",
				Host: "foo.com",
				Path: "/bar",
			}
			err := s.Upsert(ctx, svc)
			require.IsType(t, validation.Error{}, err)
		})
	t.Run("upsert an object without ID generates an ID", func(t *testing.T) {
		svc := resource.NewService()
		svc.Service = &v1.Service{
			Name: "s0",
			Host: "foo.com",
			Path: "/bar",
		}
		err := s.Upsert(ctx, svc)
		require.Nil(t, err)
		require.NotPanics(t, func() {
			uuid.MustParse(svc.Service.Id)
		})
	})
	t.Run("uspert an object with unique index violation fails",
		func(t *testing.T) {
			svc := resource.NewService()
			svc.Service = &v1.Service{
				Name: "s0",
				Host: "foo.com",
				Path: "/bar",
			}
			err := s.Upsert(ctx, svc)
			require.IsType(t, ErrConstraint{}, err)
			constraintErr, ok := err.(ErrConstraint)
			require.True(t, ok)
			require.Equal(t,
				model.Index{
					Name:      "name",
					FieldName: "name",
					Type:      "unique",
					Value:     "s0",
				},
				constraintErr.Index)
			require.Equal(t, "", constraintErr.Message)
		})
	t.Run("upsert an object with valid foreign reference succeeds",
		func(t *testing.T) {
			svc := resource.NewService()
			sid := uuid.NewString()
			svc.Service = &v1.Service{
				Id:   sid,
				Name: "s1",
				Host: "foo.com",
				Path: "/bar",
			}
			err := s.Upsert(ctx, svc)
			require.Nil(t, err)

			route := resource.NewRoute()
			route.Route = &v1.Route{
				Name:  "r0",
				Hosts: []string{"example.com"},
				Service: &v1.Service{
					Id: sid,
				},
			}
			require.Nil(t, s.Upsert(ctx, route))
		})
	t.Run("upsert an object with invalid foreign reference fails",
		func(t *testing.T) {
			route := resource.NewRoute()
			sid := uuid.NewString()
			route.Route = &v1.Route{
				Name:  "r1",
				Hosts: []string{"example.com"},
				Service: &v1.Service{
					Id: sid,
				},
			}
			err := s.Upsert(ctx, route)
			require.IsType(t, ErrConstraint{}, err)
			constraintErr, ok := err.(ErrConstraint)
			require.True(t, ok)
			require.Equal(t,
				model.Index{
					Name:        "svc_id",
					FieldName:   "service.id",
					Type:        "foreign",
					ForeignType: "service",
					Value:       sid,
				},
				constraintErr.Index)
			require.Equal(t, "", constraintErr.Message)
		})
	t.Run("object can be upserted without any side-effect", func(t *testing.T) {
		svc := resource.NewService()
		sid1 := uuid.NewString()
		svc.Service = &v1.Service{
			Id:   sid1,
			Name: "s2",
			Host: "foo.com",
			Path: "/bar",
		}
		require.Nil(t, s.Upsert(ctx, svc))

		// update the id
		svc.Service.Name = "s2new"
		require.Nil(t, s.Upsert(ctx, svc))

		// same id can be used by another entity now
		sid2 := uuid.NewString()
		svc.Service = &v1.Service{
			Id:   sid2,
			Name: "s2",
			Host: "foo.com",
			Path: "/bar",
		}
		require.Nil(t, s.Upsert(ctx, svc))

		svc = resource.NewService()
		require.Nil(t, s.Read(ctx, svc, GetByID(sid1)))
		require.Equal(t, "s2new", svc.Service.Name)

		svc = resource.NewService()
		require.Nil(t, s.Read(ctx, svc, GetByID(sid1)))
		require.Nil(t, s.Read(ctx, svc, GetByID(sid2)))
		require.Equal(t, "s2", svc.Service.Name)
	})
	t.Run("object with foreign references are updated fine",
		func(t *testing.T) {
			ctx := context.Background()
			persister, err := util.GetPersister()
			require.Nil(t, err)
			s := New(persister, log.Logger).ForCluster("default")
			svc := resource.NewService()
			serviceID := uuid.NewString()
			svc.Service = &v1.Service{
				Id:   serviceID,
				Name: "foo",
				Host: "foo.com",
				Path: "/bar",
			}
			require.Nil(t, s.Create(ctx, svc))

			serviceID2 := uuid.NewString()
			svc.Service = &v1.Service{
				Id:   serviceID2,
				Name: "bar",
				Host: "foo.com",
				Path: "/bar",
			}
			require.Nil(t, s.Upsert(ctx, svc))

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

			route.Route = &v1.Route{
				Id:    routeID,
				Name:  "foo",
				Hosts: []string{"example.com"},
				Service: &v1.Service{
					Id: serviceID2,
				},
			}

			require.Nil(t, s.Upsert(ctx, route))

			svc.Service = &v1.Service{
				Id:   serviceID2,
				Name: "fubar",
				Host: "foo.com",
				Path: "/bar",
			}
			require.Nil(t, s.Upsert(ctx, svc))

			err = s.Delete(ctx, DeleteByType(resource.TypeService),
				DeleteByID(serviceID))
			require.Nil(t, err)

			svc = resource.NewService()
			require.Nil(t, s.Read(ctx, svc, GetByID(serviceID2)))
			require.Equal(t, "fubar", svc.Service.Name)

			err = s.Delete(ctx, DeleteByType(resource.TypeService),
				DeleteByID(serviceID2))
			require.NotNil(t, err)
			require.True(t, errors.As(err, &ErrConstraint{}))
		})
}

func TestList(t *testing.T) {
	persister, err := util.GetPersister()
	require.Nil(t, err)
	s := New(persister, log.Logger).ForCluster("default")
	t.Run("list returns zero results without an error", func(t *testing.T) {
		svcs := resource.NewList(resource.TypeService)
		err = s.List(context.Background(), svcs)
		require.Nil(t, err)
		require.Len(t, svcs.GetAll(), 0)
	})
	t.Run("list returns the necessary items", func(t *testing.T) {
		svc := resource.NewService()
		svc.Service = &v1.Service{
			Name: "s0",
			Host: "foo.com",
			Path: "/bar",
		}
		require.Nil(t, s.Create(context.Background(), svc))

		svc = resource.NewService()
		svc.Service = &v1.Service{
			Name: "s1",
			Host: "foo.com",
			Path: "/bar",
		}
		require.Nil(t, s.Create(context.Background(), svc))

		svcs := resource.NewList(resource.TypeService)
		err = s.List(context.Background(), svcs)
		require.Nil(t, err)
		require.Len(t, svcs.GetAll(), 2)
	})
	t.Run("list returns elements referenced via foreign index", func(t *testing.T) {
		ctx := context.Background()
		svc := resource.NewService()
		sid := uuid.NewString()
		svc.Service = &v1.Service{
			Id:   sid,
			Name: "s2",
			Host: "foo.com",
			Path: "/bar",
		}
		err := s.Create(ctx, svc)
		require.Nil(t, err)

		route := resource.NewRoute()
		route.Route = &v1.Route{
			Name:  "r0",
			Hosts: []string{"example.com"},
			Service: &v1.Service{
				Id: sid,
			},
		}
		require.Nil(t, s.Create(ctx, route))

		route = resource.NewRoute()
		route.Route = &v1.Route{
			Name:  "r1",
			Hosts: []string{"example.com"},
			Service: &v1.Service{
				Id: sid,
			},
		}
		require.Nil(t, s.Create(ctx, route))

		routesForService := resource.NewList(resource.TypeRoute)
		err = s.List(ctx, routesForService, ListFor(resource.TypeService, sid))
		require.Nil(t, err)
		require.Len(t, routesForService.GetAll(), 2)
	})
	t.Run("list returns no error when no foreign resources exists", func(t *testing.T) {
		ctx := context.Background()
		svc := resource.NewService()
		sid := uuid.NewString()
		svc.Service = &v1.Service{
			Id:   sid,
			Name: "s3",
			Host: "foo.com",
			Path: "/bar",
		}
		err := s.Create(ctx, svc)
		require.Nil(t, err)

		routesForService := resource.NewList(resource.TypeRoute)
		err = s.List(ctx, routesForService, ListFor(resource.TypeService, sid))
		require.Nil(t, err)
		require.Len(t, routesForService.GetAll(), 0)
	})
}

func TestNew(t *testing.T) {
	persister, err := util.GetPersister()
	require.Nil(t, err)

	t.Run("panics if no logger is provided", func(t *testing.T) {
		require.Panics(t, func() {
			New(persister, nil)
		})
	})
	t.Run("panics if no persister is provided", func(t *testing.T) {
		require.Panics(t, func() {
			New(nil, log.Logger)
		})
	})
	t.Run("does not panic when both provided", func(t *testing.T) {
		require.NotPanics(t, func() {
			require.NotNil(t, New(persister, log.Logger))
		})
	})
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
	t.Run("update event value is a uuid", func(t *testing.T) {
		svc := resource.NewService()
		id := uuid.NewString()
		svc.Service = &v1.Service{
			Id:   id,
			Name: "s0",
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
	t.Run("create creates an update event", func(t *testing.T) {
		svc := resource.NewService()
		id := uuid.NewString()
		svc.Service = &v1.Service{
			Id:   id,
			Name: "s1",
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
			Name: "s2",
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
	t.Run("read does not create a new update event", func(t *testing.T) {
		svc := resource.NewService()
		id := uuid.NewString()
		svc.Service = &v1.Service{
			Id:   id,
			Name: "s3",
			Host: "foo.com",
			Path: "/bar",
		}
		require.Nil(t, s.Create(ctx, svc))

		e := event.New()
		err := s.Read(ctx, e, GetByID(event.ID))
		require.Nil(t, err)
		firstEvent := e.StoreEvent.Value

		svc = resource.NewService()
		require.Nil(t, s.Read(context.Background(), svc, GetByID(id)))

		e = event.New()
		err = s.Read(ctx, e, GetByID(event.ID))
		require.Nil(t, err)
		secondEvent := e.StoreEvent.Value

		require.Equal(t, firstEvent, secondEvent)
	})
}

func TestUpdateEventForNode(t *testing.T) {
	persister, err := util.GetPersister()
	require.Nil(t, err)
	s := New(persister, log.Logger).ForCluster("default")
	ctx := context.Background()
	t.Run("creating node doesn't create an update event", func(t *testing.T) {
		node := resource.NewNode()
		id := uuid.NewString()
		node.Node = &v1.Node{
			Id:       id,
			Hostname: "foo",
			Version:  "bar",
			Type:     resource.NodeTypeKongProxy,
			LastPing: 42,
		}
		require.Nil(t, s.Create(ctx, node))
		e := event.New()
		err := s.Read(ctx, e, GetByID(event.ID))
		require.Equal(t, ErrNotFound, err)
	})
	t.Run("upserting node doesn't create an update event", func(t *testing.T) {
		node := resource.NewNode()
		id := uuid.NewString()
		node.Node = &v1.Node{
			Id:       id,
			Hostname: "foo",
			Version:  "bar",
			Type:     resource.NodeTypeKongProxy,
			LastPing: 42,
		}
		require.Nil(t, s.Upsert(ctx, node))
		e := event.New()
		err := s.Read(ctx, e, GetByID(event.ID))
		require.Equal(t, ErrNotFound, err)
	})
	t.Run("deleting node doesn't create an update event", func(t *testing.T) {
		node := resource.NewNode()
		id := uuid.NewString()
		node.Node = &v1.Node{
			Id:       id,
			Hostname: "foo",
			Version:  "bar",
			Type:     resource.NodeTypeKongProxy,
			LastPing: 42,
		}
		require.Nil(t, s.Upsert(ctx, node))
		e := event.New()
		err := s.Read(ctx, e, GetByID(event.ID))
		require.Equal(t, ErrNotFound, err)

		require.Nil(t, s.Delete(ctx, DeleteByID(id),
			DeleteByType(resource.TypeNode)))
		err = s.Read(ctx, e, GetByID(event.ID))
		require.Equal(t, ErrNotFound, err)
	})
}

func TestStoredValue(t *testing.T) {
	t.Run("stored value is a protobuf aware JSON", func(t *testing.T) {
		ctx := context.Background()
		persister, err := util.GetPersister()
		require.Nil(t, err)
		s := New(persister, log.Logger).ForCluster("default")
		route := resource.NewRoute()
		id := uuid.NewString()
		route.Route = &v1.Route{
			Id:               id,
			Name:             "bar",
			Paths:            []string{"/"},
			RequestBuffering: wrapperspb.Bool(true),
		}
		require.Nil(t, s.Create(ctx, route))
		key, err := s.genID(resource.TypeRoute, id)
		require.Nil(t, err)
		value, err := persister.Get(ctx, key)
		require.Nil(t, err)
		var v struct {
			Object map[string]interface{} `json:"object"`
		}
		err = encodingJSON.Unmarshal(value, &v)
		require.Nil(t, err)
		// check a wrapper type was rendered correctly,
		// this fails if a protobuf unaware parser is used
		require.True(t, v.Object["request_buffering"].(bool))
	})
}
