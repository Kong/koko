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
	Page          int
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

func ListWithPaging(pagesize int, page int) ListOptsFunc {
	return func(opt *ListOpts) {
		pagesize, page = getPagingDefaults(pagesize, page)
		opt.PageSize = pagesize
		opt.Page = page
	}
}

func getPagingDefaults(pagesize int, page int) (int, int) {
	if pagesize <= 0 {
		pagesize = 100
	}
	if page <= 0 {
		page = 1
	}
	return pagesize, page
}
