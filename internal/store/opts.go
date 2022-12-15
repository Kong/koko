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
	ReferenceType          model.Type
	ReferenceID            string
	ReferenceReverseLookup bool

	PageSize int
	Page     int

	// CEL expression used for filtering.
	// Read more: https://github.com/google/cel-spec
	Filter *exprpb.Expr
}

func (o *ListOpts) validate() error {
	// TODO(tjasko): Implement proper support for combining both ListFor() & ListWithFilter().
	if o.Filter != nil && o.ReferenceType != "" && o.ReferenceID != "" {
		return ErrUnsupportedListOpts{
			"listing resources scoped to a resource while applying a filter are not yet supported",
		}
	}

	return nil
}

type ListOptsFunc func(*ListOpts)

// ErrUnsupportedListOpts is used when the provided combination of list options is not supported.
type ErrUnsupportedListOpts struct{ message string }

func (e ErrUnsupportedListOpts) Error() string {
	return e.message
}

const (
	DefaultPage     = 1
	DefaultPageSize = 100
	MaxPageSize     = 1000
)

// NewListOpts executes the provided list option functions, validates that all set
// options are compatible with each other, and returns the generated ListOpts.
func NewListOpts(fns ...ListOptsFunc) (*ListOpts, error) {
	res := &ListOpts{PageSize: DefaultPageSize, Page: DefaultPage}
	for _, fn := range fns {
		fn(res)
	}

	if err := res.validate(); err != nil {
		return nil, err
	}

	return res, nil
}

// ListFor allows a list call to return the objects passed into resource.NewList() and
// filter the results based on the passed in foreign relation.
//
// As such, this is used to list the resource(s) that have the associated foreign key index.
//
// Given the following routes:
//   - id: route-1
//     service: {id: service-1}
//   - id: route-2
//
// When calling `ListFor(resource.TypeService, "service-1")` with
// `resource.NewList(resource.TypeRoute)`
//
// Then this would return the route with ID `route-1`.
func ListFor(typ model.Type, id string) ListOptsFunc {
	return func(opt *ListOpts) {
		opt.ReferenceType = typ
		opt.ReferenceID = id
	}
}

// ListReverseFor is like ListFor(), however the only difference
// is the foreign key indexes reside on the passed in type instead
// of the type passed into resource.NewList().
//
// As such, this is used to list the foreign resources themselves.
//
// Given the following routes:
//   - id: route-1
//     service: {id: service-1}
//   - id: route-2
//     service: {id: service-1}
//   - id: route-3
//
// When calling `ListReverseFor(resource.TypeRoute, "route-1")` with
// `resource.NewList(resource.TypeService)`.
//
// Then this would return the service with ID `service-1`.
func ListReverseFor(typ model.Type, id string) ListOptsFunc {
	return func(opt *ListOpts) {
		opt.ReferenceType = typ
		opt.ReferenceID = id
		opt.ReferenceReverseLookup = true
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
