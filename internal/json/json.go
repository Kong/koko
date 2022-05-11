package json

import (
	"bytes"
	stdjson "encoding/json"
	"io"
)

// RawMessage is a raw encoded JSON value.
// It implements Marshaler and Unmarshaler and can
// be used to delay JSON decoding or precompute a JSON encoding.
type RawMessage stdjson.RawMessage

// MarshalJSON returns m as the JSON encoding of m.
func (m RawMessage) MarshalJSON() ([]byte, error) {
	return stdjson.RawMessage(m).MarshalJSON()
}

// UnmarshalJSON sets *m to a copy of data.
func (m *RawMessage) UnmarshalJSON(data []byte) error {
	return (*stdjson.RawMessage)(m).UnmarshalJSON(data)
}

var (
	_ stdjson.Marshaler   = (*RawMessage)(nil)
	_ stdjson.Unmarshaler = (*RawMessage)(nil)
)

// Marshal returns the JSON encoding of v.
func Marshal(v any) ([]byte, error) {
	return stdjson.Marshal(v)
}

// Unmarshal parses the JSON-encoded data and stores the result
// in the value pointed to by v. If v is nil or not a pointer,
// Unmarshal returns an InvalidUnmarshalError.
func Unmarshal(data []byte, v any) error {
	return stdjson.Unmarshal(data, v)
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *stdjson.Encoder {
	return stdjson.NewEncoder(w)
}

// Compact appends to dst the JSON-encoded src with
// insignificant space characters elided.
func Compact(dst *bytes.Buffer, src []byte) error {
	return stdjson.Compact(dst, src)
}

// MarshalIndent is like Marshal but applies Indent to format the output.
// Each JSON element in the output will begin on a new line beginning with prefix
// followed by one or more copies of indent according to the indentation nesting.
func MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	return stdjson.MarshalIndent(v, prefix, indent)
}
