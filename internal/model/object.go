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

// ObjectWithResourceDTO defines a model object that may not store the exact JSON-encoded representation of the
// underlining Protobuf resource in the persistence store. All callers will still work with the underlining Protobuf
// resource returned by Object.Resource(), however, when it comes time to store/fetch the resource from the persistence
// store, it will be translated to/from automatically.
//
// A practical use-case of this interface is to ensure commonality across JSON keys, in the event such key needs to be
// indexed. For example, there could be `$.name` key that is an indexed value in the datastore, however on one resource,
// instead of calling it `name`, it's called `description`. These (un)marhsal methods can then be used to re-write that
// key to `name` only within the datastore, in order to prevent the need of creating another index.
//
// Given the above example, another use-case is such key may be called the same, however, it may be within a nested
// object, e.g.: `$.metadata.name`. In that case, these (un)marhsal methods can then be used to rewrite that nested
// `name` field to be stored at the root of the JSON object, in order to take advantage of an already existing index.
type ObjectWithResourceDTO interface {
	// MarshalResourceJSON is just like json.Marshaler, but specifically
	// for marshalling the resource for use in the persistence store.
	MarshalResourceJSON() ([]byte, error)

	// UnmarshalResourceJSON is just like json.Unmarshaler, but specifically for unmarshalling
	// a resource's JSON representation outputted by MarshalResourceJSON().
	UnmarshalResourceJSON([]byte) error
}

// ObjectWithOptions is used to define an object that has options that differ from DefaultObjectOptions.
//
// TODO(tjasko): Eventually we'll likely want to enforce this to be used on all objects, however
// given the current limited use-case, it feels appropriate to not require it.
type ObjectWithOptions interface {
	// Options returns the options that represent this model object. Any options must be
	// fixed values and must never dynamically change based on the object's contents.
	Options() ObjectOptions
}

// ObjectOptions defines various options that can influence how this object is handled & stored.
type ObjectOptions struct {
	// CascadeOnDelete refers to that in the event this object is used in foreign key relations
	// (either one-to-one or one-to-many), this setting dictates whether the object should be
	// automatically cascaded when its foreign key relation is deleted.
	//
	// For example, when a consumer is deleted that is associated to a consumer group, we do not
	// want the consumer group to be deleted, as it's perfectly valid to have a consumer group
	// with no consumers associated to it. Likewise, when consumer is deleted that is associated
	// to specific route(s), we do indeed want it to delete those route(s) as well.
	CascadeOnDelete bool
}

// DefaultObjectOptions defines the configuration of a model object
// that does not implement the ObjectWithOptions interface.
var DefaultObjectOptions = ObjectOptions{
	CascadeOnDelete: true,
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
