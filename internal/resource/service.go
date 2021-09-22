package resource

import (
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"

	"github.com/kong/koko/internal/model"
)

type Type string

const (
	TypeService = model.Type("service")
)

type Service struct {
	Service *v1.Service
}

func (r Service) ID() string {
	return r.Service.Id
}

func (r Service) Type() model.Type {
	return TypeService
}

func (r Service) Resource() model.Resource {
	return r.Service
}

func init() {
	err := model.RegisterType(TypeService, func() model.Object {
		return NewService()
	})
	if err != nil {
		panic(err)
	}
}

func NewService() Service {
	return Service{
		Service: &v1.Service{},
	}
}
