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

var (
	Marshal   = Marshaller.Marshal
	Unmarshal = Marshaller.Unmarshal
)
