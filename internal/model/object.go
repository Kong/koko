package model

import (
	"google.golang.org/protobuf/proto"
)

type Type string

type Resource interface {
	proto.Message
}

type Object interface {
	ID() string
	Type() Type
	Resource() Resource
	Validate() error
	ProcessDefaults() error
}

type ObjectList interface {
	Type() Type
	Add(Object)
	GetAll() []Object
}
