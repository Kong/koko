package validation

import (
	"fmt"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model/json/schema"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var jsonpb = runtime.JSONPb{
	MarshalOptions: protojson.MarshalOptions{UseProtoNames: true},
}

func Validate(typ string, message proto.Message) error {
	var v interface{}
	var err error
	js, err := jsonpb.Marshal(message)
	if err != nil {
		return err
	}
	err = jsonpb.Unmarshal(js, &v)
	if err != nil {
		return err
	}
	schema, err := schema.Get(typ)
	if err != nil {
		panic(err)
	}
	err = schema.Validate(v)
	if err != nil {
		ve, ok := err.(*jsonschema.ValidationError)
		if !ok {
			panic(fmt.Sprintf("unexpected type: %T", err))
		}
		return renderErrs(ve.DetailedOutput(), schema)
	}
	return nil
}

func renderErrs(schemaErr jsonschema.Detailed,
	schema *jsonschema.Schema) Error {
	t := ErrorTranslator{
		errs: map[string]*model.ErrorDetail{},
	}
	t.renderErrs(schemaErr, schema)
	return t.result()
}
