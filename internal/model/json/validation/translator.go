package validation

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// ErrorTranslator translates JSON Schema errors into Error.
type ErrorTranslator struct {
	errs map[string]*model.ErrorDetail
}

func (t ErrorTranslator) result() Error {
	res := make([]*model.ErrorDetail, 0, len(t.errs))
	for _, err := range t.errs {
		res = append(res, err)
	}
	return Error{Errs: res}
}

func (t ErrorTranslator) addErr(field string, errorType model.ErrorType,
	message string) {
	err, ok := t.errs[field]
	if !ok {
		err = &model.ErrorDetail{
			Type:  errorType,
			Field: field,
		}
		t.errs[field] = err
	}
	for _, m := range err.Messages {
		if message == m {
			return
		}
	}
	err.Messages = append(err.Messages, message)
}

func (t ErrorTranslator) renderErrs(schemaErr jsonschema.Detailed,
	schema *jsonschema.Schema) {
	ok := t.getErr(schemaErr, schema)
	if ok {
		return
	}
	for _, err := range schemaErr.Errors {
		t.renderErrs(err, schema)
	}
}

func pretty(input string) string {
	if input == "" {
		return ""
	}
	if input[0] == '/' {
		input = input[1:]
	}
	fragments := strings.Split(input, "/")
	var (
		buf bytes.Buffer
		i   = 0
	)
	for {
		fragment := fragments[i]
		// fmt.Println(fragment)
		pos, err := strconv.Atoi(fragment)
		if err == nil {
			buf.Truncate(buf.Len() - 1)
			buf.WriteString(fmt.Sprintf("[%d]", pos))
		} else {
			buf.WriteString(fragment)
		}
		i++
		if i == len(fragments) {
			break
		}
		buf.WriteString(".")
	}
	return buf.String()
}

func (t ErrorTranslator) getErr(schemaErr jsonschema.Detailed,
	schema *jsonschema.Schema) bool {
	var (
		ok    bool
		field = schemaErr.InstanceLocation
	)
	walk(schemaErr.KeywordLocation, schema, func(schema *jsonschema.Schema,
		hint string) bool {
		message := ""
		switch hint {
		case "":
			message = schema.Description
		case "Pattern":
			message = "must match pattern" + schema.Pattern.String()
		case "Required":
			message = schemaErr.Error
		case "Enum":
			message = fmt.Sprintf("must be one of %v", schema.Enum)
		case "Minimum":
			message = schemaErr.Error
		case "Maximum":
			message = schemaErr.Error
		case "MaxItems":
			message = schemaErr.Error
		default:
			panic("unexpected hint")
		}
		if message != "" {
			ok = true
			var errorType model.ErrorType
			if field != "" {
				field = pretty(field)
				errorType = model.ErrorType_ERROR_TYPE_FIELD
			} else {
				errorType = model.ErrorType_ERROR_TYPE_ENTITY
			}
			t.addErr(field, errorType, message)
			// stop walking
			return false
		}
		return true
	})
	return ok
}

func walk(location string, schema *jsonschema.Schema,
	fn func(*jsonschema.Schema, string) bool) {
	if location == "" {
		return
	}
	fragments := strings.Split(location, "/")
	// fmt.Println(fragments)
	for i := 0; i < len(fragments); i++ {
		fragment := fragments[i]
		if fragment == "" {
			continue
		}
		hint := ""
		switch fragment {
		case "properties":
			i++
			schema = schema.Properties[fragments[i]]
		case "allOf":
			i++
			if i == len(fragments) {
				return
			}
			pos, err := strconv.Atoi(fragments[i])
			if err != nil {
				panic(err)
			}
			schema = schema.AllOf[pos]
		case "oneOf":
			i++
			if i == len(fragments) {
				return
			}
			pos, err := strconv.Atoi(fragments[i])
			if err != nil {
				panic(err)
			}
			schema = schema.OneOf[pos]
		case "anyOf":
			i++
			if i == len(fragments) {
				return
			}
			pos, err := strconv.Atoi(fragments[i])
			if err != nil {
				panic(err)
			}
			schema = schema.AnyOf[pos]
		case "items":
			item := schema.Items2020
			schema = item
		case "then":
			schema = schema.Then
		case "if":
			schema = schema.If
		case "not":
			schema = schema.Not
		case "required":
			hint = "Required"
		case "minimum":
			hint = "Minimum"
		case "maximum":
			hint = "Maximum"
		case "maxItems":
			hint = "MaxItems"
		case "pattern":
			hint = "Pattern"
		case "enum":
			hint = "Enum"
		default:
			panic("unexpected fragment: " + fragment)
		}
		ok := fn(schema, hint)
		if !ok {
			return
		}
	}
}
