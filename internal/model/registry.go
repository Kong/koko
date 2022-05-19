package model

import (
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Registry interface{}

var (
	types = map[Type]func() Object{}

	// Map proto message's full name descriptor -> Resource type.
	// Example `kong.admin.model.v1.Consumer` -> `resource.TypeConsumer`.
	protoToType = map[string]Type{}
)

// RegisterType handles mapping an object's type name to its model & generated Protobuf
// message. An error will be returned when the type is already registered.
func RegisterType(typ Type, p proto.Message, fn func() Object) error {
	if _, ok := types[typ]; ok {
		return fmt.Errorf("type already registered: %v", typ)
	}
	var pr protoreflect.Message
	if p == nil {
		return errors.New("must not provide empty Protobuf message")
	} else if pr = p.ProtoReflect(); !pr.IsValid() {
		return errors.New("must not provide invalid Protobuf message")
	}
	protoName := string(pr.Descriptor().FullName())
	if _, ok := protoToType[protoName]; ok {
		return fmt.Errorf("protobuf message already registered: %s", protoName)
	}
	types[typ], protoToType[protoName] = fn, typ
	return nil
}

func NewObject(typ Type) (Object, error) {
	fn, ok := types[typ]
	if !ok {
		return nil, fmt.Errorf("type not register: %v", typ)
	}
	return fn(), nil
}

func AllTypes() []Type {
	var res []Type
	for t := range types {
		res = append(res, t)
	}
	return res
}

// ObjectFromProto returns the relevant model object from the given Protobuf message, while retaining
// all provided fields that are set. When the relevant object cannot be found, an error is returned.
func ObjectFromProto(p proto.Message) (Object, error) {
	var pr protoreflect.Message
	if p == nil {
		return nil, errors.New("cannot resolve empty Protobuf message to object")
	} else if pr = p.ProtoReflect(); !pr.IsValid() {
		return nil, errors.New("cannot resolve invalid Protobuf message to object")
	}

	// Check to see if we've mapped this Protobuf message to a resource.
	expectedFullName := string(pr.Descriptor().FullName())
	if typ, ok := protoToType[expectedFullName]; ok {
		obj, err := NewObject(typ)
		if err != nil {
			return nil, err
		}

		// Replace the model object's current values with those on the passed in Protobuf message.
		// This should never error as we've already validated the descriptor's match.
		return obj, obj.SetResource(Resource(p))
	}

	// This would happen when the provided Protobuf message does not
	// match any underlining resource on our model objects.
	return nil, fmt.Errorf("cannot find type from Protobuf message %s", expectedFullName)
}
