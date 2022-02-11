package store

import (
	"context"
	encodingJSON "encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	protoJSON "github.com/kong/koko/internal/json"
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
		err = s.Create(ctx, route)
		require.Nil(t, err)
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

	rid := uuid.NewString()
	status := resource.NewStatus()
	status.Status = &v1.Status{
		ContextReference: &v1.EntityReference{
			Type: string(resource.TypeRoute),
			Id:   rid,
		},
		Conditions: []*v1.Condition{
			{
				Code:     "R0023",
				Message:  "foo bar",
				Severity: resource.SeverityError,
			},
		},
	}
	err = s.Create(context.Background(), status)
	require.Nil(t, err)

	t.Run("reading an object by id succeeds", func(t *testing.T) {
		svc := resource.NewService()
		err := s.Read(context.Background(), svc, GetByID(sid))
		require.Nil(t, err)
		require.Equal(t, "s0", svc.Service.Name)
		require.Equal(t, "foo.com", svc.Service.Host)
		require.Equal(t, "/bar", svc.Service.Path)
	})
	t.Run("reading an object by name succeeds", func(t *testing.T) {
		svc := resource.NewService()
		err := s.Read(context.Background(), svc, GetByName("s0"))
		require.Nil(t, err)
		require.Equal(t, sid, svc.Service.Id)
		require.Equal(t, "s0", svc.Service.Name)
		require.Equal(t, "foo.com", svc.Service.Host)
		require.Equal(t, "/bar", svc.Service.Path)
	})
	t.Run("reading a non-existent object by id returns ErrNotFound", func(t *testing.T) {
		svc := resource.NewService()
		err := s.Read(context.Background(), svc, GetByID(uuid.NewString()))
		require.IsType(t, ErrNotFound, err)
	})
	t.Run("reading a non-existent object by name returns ErrNotFound", func(t *testing.T) {
		svc := resource.NewService()
		err := s.Read(context.Background(), svc, GetByName("does-not-exist"))
		require.IsType(t, ErrNotFound, err)
	})
	t.Run("reading via an index returns the resource", func(t *testing.T) {
		status := resource.NewStatus()
		err := s.Read(context.Background(), status, GetByIndex("ctx_ref",
			model.MultiValueIndex("route", rid)))
		require.Nil(t, err)
		require.Equal(t, "R0023", status.Status.Conditions[0].Code)
	})
	t.Run("reading a non-existent index returns ErrNotFound", func(t *testing.T) {
		status := resource.NewStatus()
		err := s.Read(context.Background(), status, GetByIndex("borked",
			model.MultiValueIndex("route", rid)))
		require.IsType(t, ErrNotFound, err)
	})
	t.Run("reading a non-existent  value in an index returns ErrNotFound", func(t *testing.T) {
		status := resource.NewStatus()
		err := s.Read(context.Background(), status, GetByIndex("ctx_ref",
			model.MultiValueIndex("route", uuid.NewString())))
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
		require.Nil(t, s.Read(ctx, svc, GetByID(id)))

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

func TestPagination(t *testing.T) {
	persister, err := util.GetPersister()
	require.Nil(t, err)
	store := New(persister, log.Logger).ForCluster("default")
	ctx := context.Background()
	// Create 10 services and perform the pagination
	svcName := "myservice-%d"
	svc := resource.NewService()
	svc.Service = &v1.Service{
		Name: "foo",
		Host: "example.com",
		Path: "/",
	}
	idList := make([]string, 0, 10)
	for i := 0; i < 10; i++ {
		svc.Service.Name = fmt.Sprintf(svcName, i)
		svc.Service.Id = uuid.NewString()
		idList = append(idList, svc.Service.Id)
		require.Nil(t, store.Create(ctx, svc))
	}
	// Retrieve the List to get head and tail
	svcs := resource.NewList(resource.TypeService)
	err = store.List(context.Background(), svcs)
	require.Nil(t, err)
	require.Equal(t, svcs.GetTotalCount(), 10)
	head := svcs.GetAll()[0].ID()
	require.NotEmpty(t, head)
	require.True(t, contains(idList, head))
	tail := svcs.GetAll()[9].ID()
	require.NotEmpty(t, tail)
	require.True(t, contains(idList, tail))
	// Now Test each pagination scenario
	t.Run("Page 1, Size 1 success", func(t *testing.T) {
		pageSize := ListWithPageSize(1)
		pageNum := ListWithPageNum(1)
		svcs = resource.NewList(resource.TypeService)
		err = store.List(ctx, svcs, pageSize, pageNum)
		require.NoError(t, err)
		require.Equal(t, 10, svcs.GetTotalCount())
		require.Equal(t, 2, svcs.GetNextPage())
		require.Equal(t, head, svcs.GetAll()[0].ID())
		// Get the last Page and Element
		pageNum = ListWithPageNum(10)
		require.NoError(t, err)
		svcs = resource.NewList(resource.TypeService)
		err = store.List(ctx, svcs, pageSize, pageNum)
		require.NoError(t, err)
		require.Equal(t, 10, svcs.GetTotalCount())
		require.Equal(t, 0, svcs.GetNextPage())
		require.Equal(t, tail, svcs.GetAll()[0].ID())
	})
	t.Run("Page 1, Size 2 success", func(t *testing.T) {
		pageSize := ListWithPageSize(2)

		pageNum := ListWithPageNum(1)
		require.NoError(t, err)
		svcs = resource.NewList(resource.TypeService)
		err = store.List(ctx, svcs, pageSize, pageNum)
		require.NoError(t, err)
		require.Equal(t, 10, svcs.GetTotalCount())
		require.Equal(t, 2, svcs.GetNextPage())
		require.Equal(t, head, svcs.GetAll()[0].ID())
		// Get the last Page and Element
		pageNum = ListWithPageNum(5)
		svcs = resource.NewList(resource.TypeService)
		err = store.List(ctx, svcs, pageSize, pageNum)
		require.NoError(t, err)
		require.Equal(t, 10, svcs.GetTotalCount())
		require.Equal(t, 0, svcs.GetNextPage())
		require.Equal(t, tail, svcs.GetAll()[1].ID())
	})
	t.Run("Page 1, Size 3 success", func(t *testing.T) {
		pageSize := ListWithPageSize(3)
		pageNum := ListWithPageNum(1)
		svcs = resource.NewList(resource.TypeService)
		err = store.List(ctx, svcs, pageSize, pageNum)
		require.NoError(t, err)
		require.Equal(t, 10, svcs.GetTotalCount())
		require.Equal(t, 2, svcs.GetNextPage())
		require.Equal(t, head, svcs.GetAll()[0].ID())
		// Get the last Page and Element
		pageNum = ListWithPageNum(4)
		svcs = resource.NewList(resource.TypeService)
		err = store.List(ctx, svcs, pageSize, pageNum)
		require.NoError(t, err)
		require.Equal(t, 10, svcs.GetTotalCount())
		require.Equal(t, 0, svcs.GetNextPage())
		require.Equal(t, tail, svcs.GetAll()[0].ID())
	})
	t.Run("Page 1, Size 10 success", func(t *testing.T) {
		pageSize := ListWithPageSize(10)
		pageNum := ListWithPageNum(1)
		svcs = resource.NewList(resource.TypeService)
		err = store.List(ctx, svcs, pageSize, pageNum)
		require.NoError(t, err)
		require.Equal(t, 10, svcs.GetTotalCount())
		require.Equal(t, 0, svcs.GetNextPage())
		require.Equal(t, head, svcs.GetAll()[0].ID())
		require.Equal(t, tail, svcs.GetAll()[9].ID())
	})
	t.Run("Page 1, Size 11 success", func(t *testing.T) {
		pageSize := ListWithPageSize(11)
		pageNum := ListWithPageNum(1)
		svcs = resource.NewList(resource.TypeService)
		err = store.List(ctx, svcs, pageSize, pageNum)
		require.NoError(t, err)
		require.Equal(t, 10, svcs.GetTotalCount())
		require.Equal(t, 0, svcs.GetNextPage())
		require.Len(t, svcs.GetAll(), 10)
		require.Equal(t, head, svcs.GetAll()[0].ID())
		require.Equal(t, tail, svcs.GetAll()[9].ID())
	})
	t.Run("Page 11, Size 1 empty", func(t *testing.T) {
		pageSize := ListWithPageSize(1)
		pageNum := ListWithPageNum(11)
		svcs = resource.NewList(resource.TypeService)
		err = store.List(ctx, svcs, pageSize, pageNum)
		require.NoError(t, err)
		require.Equal(t, 0, svcs.GetTotalCount())
		require.Equal(t, 0, svcs.GetNextPage())
		require.Len(t, svcs.GetAll(), 0)
	})
}

type jsonWrapper struct {
	Value string `json:"value"`
}

func json(value string) []byte {
	res, err := protoJSON.Marshal(jsonWrapper{value})
	if err != nil {
		panic(fmt.Sprintf("marshal json: %v", err))
	}
	return res
}

func TestFullListPaging(t *testing.T) {
	p, err := util.GetPersister()
	require.Nil(t, err)
	t.Run("we retrieve full list despite default pagination", func(t *testing.T) {
		ctx := context.Background()
		tx, err := p.Tx(ctx)
		require.Nil(t, err)
		var expectedValuesBatchOne, expectedKeysBatchOne []string
		for i := 0; i < 1001; i++ {
			value := json(fmt.Sprintf("prefix-value-%06d", i))
			key := fmt.Sprintf("myprefix/key%06d", i)
			err = tx.Put(ctx, key, value)
			require.Nil(t, err)
			expectedKeysBatchOne = append(expectedKeysBatchOne, key)
			expectedValuesBatchOne = append(expectedValuesBatchOne, string(value))
		}

		listResult, err := getFullList(ctx, tx, "myprefix/")
		require.Nil(t, err)
		require.Equal(t, 1001, listResult.TotalCount)
		var valuesAsStrings []string
		var keysAsStrings []string
		for _, kv := range listResult.KVList {
			key := string(kv.Key)
			value := string(kv.Value)
			keysAsStrings = append(keysAsStrings, key)
			value = strings.ReplaceAll(value, " ", "")
			valuesAsStrings = append(valuesAsStrings, value)
		}
		sort.Strings(keysAsStrings)
		sort.Strings(expectedKeysBatchOne)
		sort.Strings(valuesAsStrings)
		sort.Strings(expectedValuesBatchOne)
		require.Equal(t, expectedKeysBatchOne, keysAsStrings)
		tx.Rollback()
	})
}

func contains(repo []string, search string) bool {
	for _, v := range repo {
		if v == search {
			return true
		}
	}
	return false
}
