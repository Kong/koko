package store

import (
	"context"
	"testing"

	"github.com/kong/koko/internal/persistence"
	"github.com/kong/koko/internal/resource"
	"github.com/stretchr/testify/assert"
)

func TestService(t *testing.T) {
	s := New(&persistence.Memory{})
	svc := resource.NewService()
	svc.Service.Id = "foo"
	svc.Service.Name = "bar"
	assert.Nil(t, s.Create(context.Background(), svc))
	svc = resource.NewService()
	assert.Nil(t, s.Read(context.Background(), svc, GetByID("foo")))
	assert.Equal(t, "bar", svc.Service.Name)
	assert.Nil(t, s.Delete(context.Background(),
		DeleteByType(resource.TypeService), DeleteByID(svc.ID())))
	assert.NotNil(t, s.Read(context.Background(), svc, GetByID("foo")))

	svc = resource.NewService()
	svc.Service.Id = "bar1"
	svc.Service.Name = "bar1"
	assert.Nil(t, s.Create(context.Background(), svc))

	svc = resource.NewService()
	svc.Service.Id = "bar2"
	svc.Service.Name = "bar2"
	assert.Nil(t, s.Create(context.Background(), svc))
	svcs := resource.NewList(resource.TypeService)
	err := s.List(context.Background(), svcs)
	assert.Nil(t, err)
	assert.Len(t, svcs.GetAll(), 2)
}
