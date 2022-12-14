package model

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/samber/lo"
	"github.com/yalp/jsonpath"
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

// IndexAction determines how the index is managed by the persistence store.
type IndexAction int

const (
	// IndexActionManaged is the default index action, which automatically removes the previously defined indexes (found
	// on the currently stored object in the DB) when applicable & inserts the new indexes defined on the resource.
	//
	// The automated removal is useful in most use cases, as if the related entity ID/name/etc. changes, we no longer
	// want to keep that index around. However, in the event that we're storing multiple indexes for a list of foreign
	// resource IDs, it's impossible to remove a single index without fetching all existing objects (which has
	// scalability & performance concerns when there's a large number of foreign keys).
	//
	// Cannot be combined with any other actions.
	IndexActionManaged IndexAction = iota

	// IndexActionAdd explicitly defines that this index should
	// be created without deleting any prior existing indexes.
	//
	// This should be used when foreign key relations should be added,
	// and it's not ideal to look up all existing indexes on the object.
	IndexActionAdd

	// IndexActionRemove explicitly defines that this index should
	// be removed without deleting any prior existing indexes.
	//
	// This should be used when foreign key relations should be removed,
	// and it's not ideal to look up all existing indexes on the object.
	IndexActionRemove
)

// Index defines an index/foreign key based on the defined Index.Type
// & its applicable fields. Unless noted, all fields are required.
//
// NOTE: If any pointer values are introduced here, Indexes.Validate()
// must be updated to better handle duplicate indexes.
type Index struct {
	// Name of the Index.
	Name string
	// FieldName is the name of the field this constraint applies on.
	// This is used for annotating errors.
	//
	// Use JSON path notation (foo.bar.baz) for nested fields.
	// The `$.` prefix is automatically assumed.
	//
	// TODO(tjasko): This is not required when calling Index.Validate(), should it be? It'll require some
	//  code changes, as it's currently not specified everywhere (mostly for certain unique indexes).
	FieldName string
	// Type is the type of the Index.
	Type TypeIndex
	// ForeignType denotes the type of the foreign object.
	// This must be populated for IndexForeign and otherwise must be empty.
	ForeignType Type
	// Value is the value of the field for this Index.
	Value string
	// Action defines the way this index will be handled. Defaults to IndexActionManaged.
	Action IndexAction
}

// Indexes is a list of Index objects.
type Indexes []Index

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

// Actions returns all unique index actions used within the slice of Index objects.
func (i Indexes) Actions() map[IndexAction]bool {
	return lo.Associate(i, func(idx Index) (IndexAction, bool) { return idx.Action, true })
}

// Validate ensures the provided Index object is a valid configuration.
func (i *Index) Validate() error {
	for field, friendlyName := range map[string]string{
		i.Name:         "name",
		string(i.Type): "type",
		i.Value:        "value",
	} {
		if field == "" {
			return fmt.Errorf("index %s is not set", friendlyName)
		}
	}

	if i.FieldName != "" {
		const jsonPathPrefix = "$."
		if strings.HasPrefix(i.FieldName, jsonPathPrefix) {
			// Technically we could automatically remove the prefix, but
			// this is done to enforce consistency across the codebase.
			return fmt.Errorf("must not include JSONPath prefix (%q) in field name", jsonPathPrefix)
		}
		if _, err := jsonpath.Prepare(jsonPathPrefix + i.FieldName); err != nil {
			return fmt.Errorf("invalid JSONPath field name: %w", err)
		}
	}

	if i.Type == IndexForeign && i.ForeignType == "" {
		return errors.New("index foreign type is not set")
	}

	return nil
}

// Validate ensures the provided Index objects are a valid configuration.
func (i Indexes) Validate() error {
	if actions := i.Actions(); len(actions) >= 2 && actions[IndexActionManaged] {
		return errors.New("the IndexActionManaged action cannot be used in conjunction with any other actions")
	}

	// Helper function to attempt to describe the specific index object that failed validation.
	errFunc := func(idx Index, err error) error {
		var descriptor string
		var descriptorParts []interface{}
		if idx.Name != "" {
			descriptorParts = []interface{}{"name", idx.Name}
		} else if idx.FieldName != "" {
			descriptorParts = []interface{}{"field_name", idx.FieldName}
		}
		if len(descriptorParts) > 0 {
			descriptor = fmt.Sprintf(" (%s = %q)", descriptorParts...)
		}
		return fmt.Errorf("invalid index%s in list: %w", descriptor, err)
	}

	for _, idx := range i {
		if err := idx.Validate(); err != nil {
			return errFunc(idx, err)
		}
	}

	if _, diff := lo.Difference(lo.FindUniques(i), i); len(diff) > 0 {
		return errFunc(diff[0], errors.New("duplicate index contents"))
	}

	return nil
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
