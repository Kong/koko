package mold

import (
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
)

type GrpcContent struct {
	Services []*v1.Service
	Routes   []*v1.Route
}
