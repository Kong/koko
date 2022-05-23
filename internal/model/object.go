package model

import (
	"bytes"
	"context"
	"errors"
	"fmt"

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
	Validate(ctx context.Context) error
	Indexes() []Index
	ProcessDefaults(ctx context.Context) error

	// SetResource replaces the object's underlining resource with the provided resource.
	SetResource(Resource) error
}

type TypeIndex string

const (
	IndexUnique  TypeIndex = "unique"
	IndexForeign TypeIndex = "foreign"
)

type Index struct {
	// Name of the Index.
	Name string
	// FieldName is the name of the field this constraint applies on.
	// This is used for annotating errors.
	// Use JSON path notation (foo.bar.baz) for nested fields.
	FieldName string
	// Type is the type of the Index.
	Type TypeIndex
	// ForeignType denotes the type of the foreign object.
	// This must be populated for IndexForeign and otherwise must be empty.
	ForeignType Type
	// Value is the value of the field for this Index.
	Value string
}

type ObjectList interface {
	Type() Type
	Add(Object)
	GetAll() []Object
	// GetTotalCount returns the count of objects in the underlying store across all pages.
	GetTotalCount() int
	SetTotalCount(count int)
	SetNextPage(pageNum int)
	GetNextPage() int
}

func MultiValueIndex(values ...string) string {
	switch len(values) {
	case 0:
		return ""
	case 1:
		return values[0]
	case 2: //nolint:gomnd
		return fmt.Sprintf("%s:%s", values[0], values[1])
	default:
		var buf bytes.Buffer
		l := len(values)
		for i, value := range values {
			buf.WriteString(value)
			if i+1 != l {
				buf.WriteRune(':')
			}
		}
		return buf.String()
	}
}

// SetResource replaces the object's underlining resource with the provided resource.
func SetResource(o Object, r Resource) error {
	expected, actual := o.Resource().ProtoReflect().Descriptor(), r.ProtoReflect().Descriptor()
	if expected != actual {
		return fmt.Errorf("unable to set resource: expected %q but got %q", expected.FullName(), actual.FullName())
	}
	dst := o.Resource()
	if !dst.ProtoReflect().IsValid() {
		return errors.New("unable to set resource: got invalid destination resource")
	}
	proto.Reset(dst)
	proto.Merge(dst, r)
	return nil
}
