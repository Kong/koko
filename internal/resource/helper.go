package resource

import (
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
	"github.com/kong/koko/internal/model/validation/typedefs"
)

func notHTTPProtocol(protocol string) bool {
	return protocol != typedefs.ProtocolHTTP && protocol != typedefs.ProtocolHTTPS
}

func mergeRules(rules ...interface{}) []validation.Rule {
	var res []validation.Rule
	for _, rule := range rules {
		switch v := rule.(type) {
		case validation.Rule:
			res = append(res, v)
		case []validation.Rule:
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
