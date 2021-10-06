package store

import (
	"context"
	"testing"

	"github.com/google/uuid"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/log"
	"github.com/kong/koko/internal/persistence"
	"github.com/kong/koko/internal/resource"
	"github.com/stretchr/testify/assert"
)

func TestService(t *testing.T) {
	// TODO improve these tests
	s := New(&persistence.Memory{}, log.Logger)
	svc := resource.NewService()
	id := uuid.NewString()
	svc.Service = &v1.Service{
		Id:   id,
		Name: "bar",
		Host: "foo.com",
		Path: "/bar",
	}
	assert.Nil(t, s.Create(context.Background(), svc))
	svc = resource.NewService()
	assert.Nil(t, s.Read(context.Background(), svc, GetByID(id)))
	assert.Equal(t, "bar", svc.Service.Name)
	assert.Nil(t, s.Delete(context.Background(),
		DeleteByType(resource.TypeService), DeleteByID(svc.ID())))
	assert.NotNil(t, s.Read(context.Background(), svc, GetByID(id)))

	svc = resource.NewService()
	svc.Service = &v1.Service{
		Id:   uuid.NewString(),
		Name: "bar",
		Host: "foo.com",
		Path: "/bar",
	}
	assert.Nil(t, s.Create(context.Background(), svc))

	svc = resource.NewService()
	svc.Service = &v1.Service{
		Id:   uuid.NewString(),
		Name: "foo",
		Host: "foo.com",
		Path: "/bar",
	}
	assert.Nil(t, s.Create(context.Background(), svc))
	svcs := resource.NewList(resource.TypeService)
	err := s.List(context.Background(), svcs)
	assert.Nil(t, err)
	assert.Len(t, svcs.GetAll(), 2)
}
