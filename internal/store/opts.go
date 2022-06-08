package store

import (
	"github.com/kong/koko/internal/model"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

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

	// CEL expression used for filtering.
	// Read more: https://github.com/google/cel-spec
	Filter *exprpb.Expr
}

type ListOptsFunc func(*ListOpts)

const (
	DefaultPage     = 1
	DefaultPageSize = 100
	MaxPageSize     = 1000
)

func NewListOpts(fns ...ListOptsFunc) *ListOpts {
	res := &ListOpts{PageSize: DefaultPageSize, Page: DefaultPage}
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

func ListWithPageNum(page int) ListOptsFunc {
	return func(opt *ListOpts) {
		if page == 0 {
			opt.Page = DefaultPage
		} else {
			opt.Page = page
		}
	}
}

func ListWithPageSize(pageSize int) ListOptsFunc {
	return func(opt *ListOpts) {
		if pageSize == 0 {
			opt.PageSize = DefaultPageSize
		} else {
			opt.PageSize = pageSize
		}
	}
}

// ListWithFilter associates the passed in CEL expression with the current list pagination options.
func ListWithFilter(expr *exprpb.Expr) ListOptsFunc {
	return func(opt *ListOpts) {
		opt.Filter = expr
	}
}
