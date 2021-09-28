package model

import "fmt"

type Registry interface{}

var types = map[Type]func() Object{}

func RegisterType(typ Type, fn func() Object) error {
	if _, ok := types[typ]; ok {
		return fmt.Errorf("type already registered: %v", typ)
	}
	types[typ] = fn
	return nil
}

func NewObject(typ Type) (Object, error) {
	fn, ok := types[typ]
	if !ok {
		return nil, fmt.Errorf("type not register: %v", typ)
	}
	return fn(), nil
}
