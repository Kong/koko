package store

import "github.com/kong/koko/internal/model"

type CreateOpts struct{}

type CreateOptsFunc func(*CreateOpts)

type ReadOpts struct {
	id   string
	name string

	idxName, idxValue string
}

type ReadOptsFunc func(*ReadOpts)

func NewReadOpts(fns ...ReadOptsFunc) *ReadOpts {
	res := &ReadOpts{}
	for _, fn := range fns {
		fn(res)
	}
	return res
}

func GetByID(id string) ReadOptsFunc {
	return func(opt *ReadOpts) {
		opt.id = id
	}
}

func GetByName(name string) ReadOptsFunc {
	return func(opt *ReadOpts) {
		opt.name = name
	}
}

func GetByIndex(indexName, indexValue string) ReadOptsFunc {
	return func(opt *ReadOpts) {
		opt.idxName = indexName
		opt.idxValue = indexValue
	}
}

type DeleteOpts struct {
	id  string
	typ model.Type
}

type DeleteOptsFunc func(*DeleteOpts)

func NewDeleteOpts(fns ...DeleteOptsFunc) *DeleteOpts {
	res := &DeleteOpts{}
	for _, fn := range fns {
		fn(res)
	}
	return res
}

func DeleteByID(id string) DeleteOptsFunc {
	return func(opt *DeleteOpts) {
		opt.id = id
	}
}

func DeleteByType(typ model.Type) DeleteOptsFunc {
	return func(opt *DeleteOpts) {
		opt.typ = typ
	}
}

type ListOpts struct {
	ReferenceType model.Type
	ReferenceID   string
	PageSize      int
	Offset        int
}

type ListOptsFunc func(*ListOpts)

func NewListOpts(fns ...ListOptsFunc) *ListOpts {
	res := &ListOpts{}
	for _, fn := range fns {
		fn(res)
	}
	return res
}

func ListFor(typ model.Type, id string) ListOptsFunc {
	return func(opt *ListOpts) {
		opt.ReferenceType = typ
		opt.ReferenceID = id
	}
}

func ListWithPaging(pagesize int, offset int) ListOptsFunc {
	return func(opt *ListOpts) {
		opt.PageSize = pagesize
		opt.Offset = offset
	}
}
