package validation

import (
	"fmt"

	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/model/json/schema"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"google.golang.org/protobuf/proto"
)

func Validate(typ string, message proto.Message) error {
	var v interface{}
	var err error
	js, err := json.Marshal(message)
	if err != nil {
		return err
	}
	err = json.Unmarshal(js, &v)
	if err != nil {
		return err
	}
	schema, err := schema.GetEntity(typ)
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
