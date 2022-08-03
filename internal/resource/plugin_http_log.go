package resource

import (
	"fmt"
	"strings"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

const pluginNameHTTPLog = "http-log"

// rewriteHTTPLogHeadersFromListToString takes in a given HTTP log plugin & transforms its header values from an
// array of strings to a single string. In the event the header values are of a different type (a string), this
// will no-op and no error will be returned.
//
// This is to support the header value type change in Kong 3.0, going from an array of strings to a single string.
func rewriteHTTPLogHeadersFromListToString(p *v1.Plugin) error {
	headersPbStruct, err := getHTTPLogHeaders(p)
	if err != nil {
		return err
	}

	for header, pbValue := range headersPbStruct.Fields {
		headersPbList := pbValue.GetListValue()
		if headersPbList == nil {
			// Header value is not a list, so no need to re-write.
			break
		}

		// Re-write the header from a list to a single string value. This matches
		// the behavior of the 2.8 -> 3.0 migration happening on the data plane.
		//
		// Read more: https://github.com/Kong/kong/pull/9162
		headersIface := headersPbList.AsSlice()
		headers := make([]string, len(headersIface))
		for i, valIface := range headersIface {
			var ok bool
			if headers[i], ok = valIface.(string); !ok {
				return fmt.Errorf("unexpected header value type for %q, got: %T, expected: string", header, valIface)
			}
		}
		headersPbStruct.Fields[header] = structpb.NewStringValue(strings.Join(headers, ", "))
	}

	return nil
}

// getHTTPLogHeaders extracts the `headers` field defined on an `http-log` plugin. When
// no headers are defined, this will return an empty Protobuf struct with no error.
func getHTTPLogHeaders(p *v1.Plugin) (*structpb.Struct, error) {
	pbValue, ok := p.Config.Fields["headers"]
	if !ok {
		// This happens when the `headers` key is not provided on the object.
		return &structpb.Struct{}, nil
	}

	switch v := pbValue.Kind.(type) {
	case *structpb.Value_StructValue:
		// Headers have been set.
		return v.StructValue, nil
	case *structpb.Value_NullValue:
		// This happens when the headers are set to null: `{"headers": null}`.
		return &structpb.Struct{}, nil
	case *structpb.Value_ListValue:
		// This happens when an empty list has been provided: `{"headers": []}`.
		return &structpb.Struct{}, nil
	}

	// Should never happen, but just a sanity check.
	return nil, fmt.Errorf("unexpected headers type: %T", pbValue)
}
