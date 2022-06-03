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

var (
	defaultSchemaErrHandleFunc = func(schemaErr jsonschema.Detailed, _ *jsonschema.Schema) string {
		return schemaErr.Error
	}
	descriptionOrErrHandleFunc = func(schemaErr jsonschema.Detailed, schema *jsonschema.Schema) string {
		if schema != nil {
			return schema.Description
		}
		return schemaErr.Error
	}

	// hintErrHandleFuncMap contains all supported JSON schema hints that are supported,
	// and the relevant function to translate the schema error to a friendly error.
	hintErrHandleFuncMap = map[string]func(schemaErr jsonschema.Detailed, schema *jsonschema.Schema) string{
		"additionalProperties": defaultSchemaErrHandleFunc,
		"dependencies":         defaultSchemaErrHandleFunc,
		"enum":                 defaultSchemaErrHandleFunc,
		"exclusiveMinimum":     defaultSchemaErrHandleFunc,
		"format":               defaultSchemaErrHandleFunc,
		"maxItems":             defaultSchemaErrHandleFunc,
		"maxLength":            defaultSchemaErrHandleFunc,
		"maximum":              defaultSchemaErrHandleFunc,
		"minLength":            defaultSchemaErrHandleFunc,
		"minimum":              defaultSchemaErrHandleFunc,
		"required":             defaultSchemaErrHandleFunc,
		"uniqueItems":          defaultSchemaErrHandleFunc,

		"allOf":             descriptionOrErrHandleFunc,
		"anyOf":             descriptionOrErrHandleFunc,
		"if":                descriptionOrErrHandleFunc,
		"items":             descriptionOrErrHandleFunc,
		"not":               descriptionOrErrHandleFunc,
		"oneOf":             descriptionOrErrHandleFunc,
		"patternProperties": descriptionOrErrHandleFunc,
		"properties":        descriptionOrErrHandleFunc,
		"then":              descriptionOrErrHandleFunc,

		"pattern": func(_ jsonschema.Detailed, schema *jsonschema.Schema) string {
			return fmt.Sprintf("must match pattern '%v'", schema.Pattern.String())
		},
	}
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
	// Sorting errors for predictability.
	SortErrorDetails(res)
	return Error{Errs: res}
}

func (t ErrorTranslator) addErr(field string, errorType model.ErrorType,
	message string,
) {
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
	schema *jsonschema.Schema,
) {
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
	schema *jsonschema.Schema,
) bool {
	var (
		ok    bool
		field = schemaErr.InstanceLocation
	)
	walk(schemaErr.KeywordLocation, schema, func(schema *jsonschema.Schema,
		hint string,
	) bool {
		var message string
		if f := hintErrHandleFuncMap[hint]; f != nil {
			message = f(schemaErr, schema)
		} else {
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
	fn func(*jsonschema.Schema, string) bool,
) {
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
