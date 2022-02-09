package validation

import (
	"bytes"
	"fmt"
	"net/url"
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
		var message string
		switch hint {
		case "properties":
			fallthrough
		case "if":
			fallthrough
		case "then":
			fallthrough
		case "not":
			fallthrough
		case "items":
			fallthrough
		case "allOf":
			fallthrough
		case "oneOf":
			fallthrough
		case "patternProperties":
			fallthrough
		case "anyOf":
			if schema != nil {
				message = schema.Description
			} else {
				message = schemaErr.Error
			}
		case "pattern":
			message = fmt.Sprintf("must match pattern '%v'",
				schema.Pattern.String())
		case "required":
			fallthrough
		case "enum":
			fallthrough
		case "exclusiveMinimum":
			fallthrough
		case "minimum":
			fallthrough
		case "maximum":
			fallthrough
		case "maxItems":
			fallthrough
		case "additionalProperties":
			fallthrough
		case "minLength":
			fallthrough
		case "maxLength":
			message = schemaErr.Error
		default:
			panic("unexpected hint: " + hint)
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
		hint := fragment
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
		case "patternProperties":
			i++
			if i == len(fragments) {
				return
			}
			key, err := url.PathUnescape(fragments[i])
			if err != nil {
				panic(err)
			}
			for pattern, patternSchema := range schema.PatternProperties {
				if key == pattern.String() {
					schema = patternSchema
					break
				}
			}
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
		default:
		}
		ok := fn(schema, hint)
		if !ok {
			return
		}
	}
}
