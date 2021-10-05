package resource

import (
	"fmt"
	"reflect"
	"time"

	ozzo "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
	"github.com/kong/koko/internal/model/validation"
	"github.com/kong/koko/internal/model/validation/typedefs"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func notHTTPProtocol(protocol string) bool {
	return protocol != typedefs.ProtocolHTTP && protocol != typedefs.ProtocolHTTPS
}

func mergeRules(rules ...interface{}) []ozzo.Rule {
	var res []ozzo.Rule
	for _, rule := range rules {
		switch v := rule.(type) {
		case ozzo.Rule:
			res = append(res, v)
		case []ozzo.Rule:
			res = append(res, v...)
		default:
			panic(fmt.Sprintf("unexpected type: %T", rule))
		}
	}
	return res
}

func validUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}

func validationErr(err error) error {
	verr, ok := err.(ozzo.Errors)
	if !ok {
		panic("unexpected type")
	}
	var res validation.Error
	for k, v := range verr {
		res.Fields = append(res.Fields, validation.FieldError{
			Name:    k,
			Message: v.Error(),
		})
	}
	return res
}

func setTS(obj interface{}, field string, override bool) {
	ptr := reflect.ValueOf(obj)
	if ptr.Kind() != reflect.Ptr {
		return
	}
	v := reflect.Indirect(ptr)
	ts := v.FieldByName(field)
	if ts.Int() == 0 || override {
		ts.Set(reflect.ValueOf(int32(time.Now().Unix())))
		return
	}
}

func addTZ(obj interface{}) {
	setTS(obj, "CreatedAt", false)
	setTS(obj, "UpdatedAt", true)
}

func defaultID(id *string) {
	if id == nil || *id == "" {
		*id = uuid.NewString()
	}
}

type wrappersPBTransformer struct{}

func (t wrappersPBTransformer) Transformer(typ reflect.Type) func(dst, src reflect.Value) error {
	var b *wrapperspb.BoolValue
	switch typ {
	case reflect.TypeOf(b):
		return func(dst, src reflect.Value) error {
			if !dst.IsNil() {
				return nil
			}
			return nil
		}
	default:

		return nil
	}
}
