package admin

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	pbModel "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_validateFilter(t *testing.T) {
	// Extending CEL environment so that we can run proper tests.
	testEnv, err := celEnv.Extend(
		cel.Types(&pbModel.Consumer{}),
		cel.Declarations(
			decls.NewVar("mapData", decls.NewMapType(decls.String, decls.String)),
			decls.NewVar("consumer", decls.NewObjectType("kong.admin.model.v1.Consumer")),
		),
	)
	require.NoError(t, err)

	tests := []struct{ name, expression, expectedError string }{
		// Valid expressions.
		{name: "single tag match", expression: `"tag1" in tags`},
		{name: "single tag match containing spaces", expression: `"tag1 with spaces" in tags`},
		{name: "logical and", expression: `"tag1" in tags && "tag2" in tags`},
		{name: "logical or", expression: `"tag1" in tags || "tag2" in tags`},
		{name: "redundant parenthesis", expression: `("tag1" in tags && "tag2" in tags) && "tag3" in tags`},
		// Functionally equivalent to the "logical and" test.
		{name: "list.all()", expression: `["tag1", "tag2"].all(x, x in tags)`},
		{name: "list.all() containing spaces", expression: `["tag1 with spaces"].all(x, x in tags)`},
		// Functionally equivalent to the "logical or" test.
		{name: "list.exists()", expression: `["tag1", "tag2"].exists(x, x in tags)`},
		{name: "list.exists() containing spaces", expression: `["tag1 with spaces"].exists(x, x in tags)`},
		{
			name:       "exactly max length",
			expression: fmt.Sprintf(`%q in tags`, strings.Repeat("x", 2038)),
		},
		{
			name:          "over max length",
			expression:    fmt.Sprintf(`%q in tags`, strings.Repeat("x", 2039)),
			expectedError: "length must be <= 2048, but got 2049",
		},

		// Invalid expressions.
		{
			name:          "unknown field",
			expression:    "unknownField",
			expectedError: "invalid filter expression: undeclared reference to 'unknownField'",
		},
		{
			name:          "unknown field in `logical and` expression",
			expression:    `"tag1" in unknownField && "tag2" in unknownField`,
			expectedError: "invalid filter expression: undeclared reference to 'unknownField'",
		},
		{
			name:          "invalid expression",
			expression:    `tags.all(x, x)`,
			expectedError: "invalid filter expression: found no matching overload for '_&&_' applied to '(bool, string)'",
		},

		// Unsupported expressions.
		{
			name:          "macro ranging on identifier",
			expression:    `tags.all(x, x in ["tag1", "tag2"])`,
			expectedError: "macros must range upon a provided list value, not a variable",
		},
		{
			name:          "mixed operators: #1",
			expression:    `("tag1" in tags && "tag2" in tags) || "tag3" in tags`,
			expectedError: "multiple logical operators are not supported in expressions",
		},
		{
			name:          "mixed operators: #2",
			expression:    `("tag1" in tags && "tag2" in tags) && ("tag3" in tags || "tag4" in tags)`,
			expectedError: "multiple logical operators are not supported in expressions",
		},
		{
			name:          "list indexing",
			expression:    `tags[0] == "tag1"`,
			expectedError: "invalid filter expression: undeclared reference to '_[_]'",
		},
		{
			name:          "string concatenation",
			expression:    `"tag" + "1" in tags`,
			expectedError: "invalid filter expression: undeclared reference to '_+_'",
		},
		{
			name:          "maps",
			expression:    `{'key': 'value'}`,
			expectedError: `unsupported expression: map`,
		},
		{
			name:          "messages",
			expression:    `google.protobuf.Int32Value{value: 1}`,
			expectedError: `unsupported expression: message (google.protobuf.Int32Value)`,
		},
		{
			name:          "field selection",
			expression:    `consumer.id`,
			expectedError: `unsupported expression: field selection`,
		},
		{
			name:          "list creation",
			expression:    `["value"]`,
			expectedError: `unsupported expression: list`,
		},

		// Unsupported operators.
		{
			name:          "operator: conditional",
			expression:    `(("tag1" in tags) ? 1 : 2)`,
			expectedError: "invalid filter expression: undeclared reference to '_?_:_'",
		},
		{
			name:          "operator: negate",
			expression:    `!("tag1" in tags)`,
			expectedError: "invalid filter expression: undeclared reference to '!_'",
		},
		{
			name:          "operator: equals",
			expression:    `1 == 2`,
			expectedError: "invalid filter expression: undeclared reference to '_==_'",
		},
		{
			name:          "operator: not equals",
			expression:    "1 != 2",
			expectedError: "invalid filter expression: undeclared reference to '_!=_'",
		},
		{
			name:          "operator: less than",
			expression:    "1 < 2",
			expectedError: "invalid filter expression: undeclared reference to '_<_'",
		},
		{
			name:          "operator: less than or equal",
			expression:    "1 <= 2",
			expectedError: "invalid filter expression: undeclared reference to '_<=_'",
		},
		{
			name:          "operator: greater than",
			expression:    "1 > 2",
			expectedError: "invalid filter expression: undeclared reference to '_>_'",
		},
		{
			name:          "operator: greater than or equals",
			expression:    "1 >= 2",
			expectedError: "invalid filter expression: undeclared reference to '_>=_'",
		},

		// All arithmetic is unsupported.
		{
			name:          "arithmetic: addition",
			expression:    "1 + 1",
			expectedError: "invalid filter expression: undeclared reference to '_+_'",
		},
		{
			name:          "arithmetic: subtraction",
			expression:    "1 - 2",
			expectedError: "invalid filter expression: undeclared reference to '_-_'",
		},
		{
			name:          "arithmetic: multiplication",
			expression:    "1 * 2",
			expectedError: "invalid filter expression: undeclared reference to '_*_'",
		},
		{
			name:          "arithmetic: division",
			expression:    "1 / 2",
			expectedError: "invalid filter expression: undeclared reference to '_/_'",
		},
		{
			name:          "arithmetic: modulus",
			expression:    "1 % 2",
			expectedError: "invalid filter expression: undeclared reference to '_%_'",
		},

		// Unsupported macros.
		{
			name:          "macro: has",
			expression:    "has(mapData.key1)",
			expectedError: "invalid filter expression: undeclared reference to 'has'",
		},
		{
			name:          "macro: exists_one",
			expression:    `tags.exists_one(x, "tag1" in tags)`,
			expectedError: "invalid filter expression: undeclared reference to 'exists_one'",
		},
		{
			name:          "macro: map",
			expression:    "tags.map(x, x)",
			expectedError: "invalid filter expression: undeclared reference to 'map'",
		},
		{
			name:          "macro: filter",
			expression:    `tags.filter(x, "tag1" in tags)`,
			expectedError: "invalid filter expression: undeclared reference to 'filter'",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := validateFilter(testEnv, tt.expression)
			// Modify the error to allow for easier assertion.
			var errPrefix string
			if err != nil {
				if validationErr, ok := err.(validation.Error); ok {
					errs := make([]string, len(validationErr.Errs))
					for i, validationErr := range validationErr.Errs {
						for _, msg := range validationErr.Messages {
							errs[i] = msg
						}
					}
					err = errors.New(strings.Join(errs, " - "))
				} else {
					errPrefix = "rpc error: code = InvalidArgument desc = "
				}
			}

			if tt.expectedError != "" {
				assert.EqualError(t, err, errPrefix+tt.expectedError)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, expr)
		})
	}
}
