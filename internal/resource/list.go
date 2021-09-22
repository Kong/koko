package resource

import "github.com/kong/koko/internal/model"

type List struct {
	typ     model.Type
	objects []model.Object
}

func NewList(typ model.Type) model.ObjectList {
	return &List{typ: typ}
}

func (l *List) Type() model.Type {
	return l.typ
}

func (l *List) Add(object model.Object) {
	l.objects = append(l.objects, object)
}

func (l *List) GetAll() []model.Object {
	return l.objects
}
