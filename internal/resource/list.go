package resource

import "github.com/kong/koko/internal/model"

type List struct {
	typ     model.Type
	objects []model.Object
	count   int
}

func (l *List) SetCount(count int) {
	l.count = count
}

func (l *List) GetCount() int {
	return l.count
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
