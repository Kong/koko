package model

import (
	"bytes"
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
	Validate() error
	Indexes() []Index
	ProcessDefaults() error
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
	GetCount() int
	SetCount(count int)
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
