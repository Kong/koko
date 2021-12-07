package json

import (
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
	Marshal   = Marshaller.Marshal
	Unmarshal = Marshaller.Unmarshal
)
