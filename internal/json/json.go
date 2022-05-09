package json

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/protobuf/encoding/protojson"
)

var Marshaller runtime.Marshaler = &runtime.JSONPb{
	MarshalOptions: protojson.MarshalOptions{
		UseProtoNames: true,
	},
}

// MarshallerWithDiscard discards unknown fields.
// When in doubt, use Marshaller. This should be used very carefully and mostly
// for temporary purposes.
var MarshallerWithDiscard runtime.Marshaler = &runtime.JSONPb{
	MarshalOptions: protojson.MarshalOptions{
		UseProtoNames: true,
	},
	UnmarshalOptions: protojson.UnmarshalOptions{
		DiscardUnknown: true,
	},
}

var (
	ProtoJSONMarshal   = Marshaller.Marshal
	ProtoJSONUnmarshal = Marshaller.Unmarshal
)

// RawMessage is a raw encoded JSON value.
// It implements Marshaler and Unmarshaler and can
// be used to delay JSON decoding or precompute a JSON encoding.
type RawMessage json.RawMessage

// MarshalJSON returns m as the JSON encoding of m.
func (m RawMessage) MarshalJSON() ([]byte, error) {
	return json.RawMessage(m).MarshalJSON()
}

// UnmarshalJSON sets *m to a copy of data.
func (m *RawMessage) UnmarshalJSON(data []byte) error {
	return (*json.RawMessage)(m).UnmarshalJSON(data)
}

var (
	_ json.Marshaler   = (*RawMessage)(nil)
	_ json.Unmarshaler = (*RawMessage)(nil)
)

// Marshal returns the JSON encoding of v.
func Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

// Unmarshal parses the JSON-encoded data and stores the result
// in the value pointed to by v. If v is nil or not a pointer,
// Unmarshal returns an InvalidUnmarshalError.
func Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *json.Encoder {
	return json.NewEncoder(w)
}

// Compact appends to dst the JSON-encoded src with
// insignificant space characters elided.
func Compact(dst *bytes.Buffer, src []byte) error {
	return json.Compact(dst, src)
}

// MarshalIndent is like Marshal but applies Indent to format the output.
// Each JSON element in the output will begin on a new line beginning with prefix
// followed by one or more copies of indent according to the indentation nesting.
func MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	return json.MarshalIndent(v, prefix, indent)
}
