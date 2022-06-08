package admin

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/operators"
	"github.com/google/cel-go/common/overloads"
	"github.com/google/cel-go/parser"
	pbModel "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/samber/lo"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

var (
	// Exposes a shared CEL environment, used for filtering based upon tags.
	//
	// TODO(tjasko): Yes, this implementation is rather simple, and was done
	//  purposefully to reduce complexity. In the event we were to introduce
	//  more advanced filtering, we'd likely want to create a CEL environment
	//  for each resource (as certain fields won't exist on all resources).
	celEnv *cel.Env

	celUndeclaredReferencePattern = regexp.MustCompile("undeclared reference to '(.*)'")
)

func init() {
	var err error
	if celEnv, err = cel.NewCustomEnv(
		// Only support all() & exists() macros.
		cel.Macros(parser.AllMacros[1], parser.AllMacros[2]),
		cel.Declarations(getCELDeclarations()...),
		cel.EagerlyValidateDeclarations(true),
	); err != nil {
		panic(fmt.Errorf("unable to establish CEL environment: %w", err))
	}
}

// validateFilter attempts to extract a filter expression from the pagination request.
// When an expression is invalid/unsupported, a validation.Error will be returned.
func validateFilter(celEnv *cel.Env, filter string) (*exprpb.Expr, error) {
	// We're limiting the max filter length to 2048 characters, in order to comply with the rest of Koko's
	// API, as we enforce limits on nearly all user input. 2048 characters is a seemingly generous amount
	// that attempts to not hinder the user, and allows us to impose some sort of upper limit.
	const maxFilterLength = 2048
	if filterLen := len(filter); filterLen > maxFilterLength {
		return nil, validation.Error{Errs: []*pbModel.ErrorDetail{{
			Type:  pbModel.ErrorType_ERROR_TYPE_FIELD,
			Field: "page.filter",
			// This error message is written in a way to copy that of the JSON schema max
			// length error message, so that we're consistent with our error messaging.
			Messages: []string{fmt.Sprintf("length must be <= %d, but got %d", maxFilterLength, filterLen)},
		}}}
	}

	// Attempt to parse & type check the provided CEL expression filter.
	ast, issues := celEnv.Compile(filter)
	if issues != nil {
		validationErr := validation.Error{}
		for _, err := range issues.Errors() {
			// Removing the irrelevant "in container..." messages from the errors, as it's
			// not helpful. e.g.: "undeclared reference to 'exists_one' (in container '')
			msg := strings.TrimSpace(strings.ReplaceAll(err.Message, `(in container '')`, ""))

			errDetail := &pbModel.ErrorDetail{
				Type:     pbModel.ErrorType_ERROR_TYPE_FIELD,
				Field:    "page.filter",
				Messages: []string{"invalid filter expression: " + msg},
			}

			// In the event this error is about an undeclared reference, only return that error.
			matches := celUndeclaredReferencePattern.FindStringSubmatch(msg)
			if len(matches) == 2 { //nolint:gomnd
				return nil, validation.Error{Errs: []*pbModel.ErrorDetail{errDetail}}
			}

			validationErr.Errs = append(validationErr.Errs, errDetail)
		}

		return nil, validationErr
	}

	// Now that we know we have a syntactically correct CEL expression, walk the AST tree & validate
	// the entirety of the expression, against what we're supporting in the CEL specification.
	expr := ast.Expr()
	if err := validateExpression(expr, nil, nil); err != nil {
		return nil, validation.Error{Errs: []*pbModel.ErrorDetail{{
			Type:     pbModel.ErrorType_ERROR_TYPE_FIELD,
			Field:    "page.filter",
			Messages: []string{err.Error()},
		}}}
	}

	return expr, nil
}

// validateExpression is a recursive function used to validate a user-inputted CEL expression. When called
// outside this function, the `parentExpr` expression & `usedOperators` map should not be provided.
//
// In the event the expression is invalid, a friendly error will be returned, that then can
// then be safely converted into a validation.Error if you wish.
func validateExpression(expr *exprpb.Expr, parentExpr *exprpb.Expr, usedOperators map[string]bool) error {
	if usedOperators == nil {
		usedOperators = make(map[string]bool)
	}

	// Ensure the expression that was parsed is supported.
	if err := validateExpressionKind(expr, parentExpr, usedOperators); err != nil {
		return err
	}

	// Validate calls to predefined functions & operators.
	if ce := expr.GetCallExpr(); ce != nil {
		// Keep track of the operators that we do not support combining. This prevents queries like:
		// `("tag1" in tags && "tag2" in tags) || ("tag3" in tags && "tag4" in tags)`.
		if lo.Contains([]string{operators.LogicalAnd, operators.LogicalOr}, ce.Function) {
			usedOperators[ce.Function] = true
		}

		// CEL prevents infinite recursion (checked during expression
		// compilation), so we don't have to worry about that here.
		for _, arg := range ce.Args {
			if err := validateExpression(arg, expr, usedOperators); err != nil {
				return err
			}
		}
	}

	// Not that is impossible to support having multiple logical operators, it just adds additional complexity.
	if len(usedOperators) > 1 {
		return errors.New("multiple logical operators are not supported in expressions")
	}

	return nil
}

// validateExpressionKind contains specific business logic to our particular CEL implementation,
// as we currently limit what expressions are supported. This will also recursively validate other
// expressions, that are generated when utilizing comprehension features (aka, macros).
func validateExpressionKind(expr *exprpb.Expr, parentExpr *exprpb.Expr, usedOperators map[string]bool) error {
	var unsupportedExpression string
	switch exprKind := expr.GetExprKind().(type) {
	case *exprpb.Expr_ConstExpr, *exprpb.Expr_IdentExpr:
		// No-op as both constant expressions (e.g.: `tag1`) & identifiers (`tags`) are supported.
		break
	case *exprpb.Expr_CallExpr:
		// Only allow "logical not" on standard overloads (which are used internally by CEL during loop conditions).
		if parentExpr != nil {
			if parentExprKind, ok := parentExpr.GetExprKind().(*exprpb.Expr_CallExpr); ok {
				if parentExprKind.CallExpr.Function == operators.NotStrictlyFalse {
					break
				}
			}
		}
		if exprKind.CallExpr.Function == operators.LogicalNot {
			return fmt.Errorf("invalid filter expression: undeclared reference to '%s'", operators.LogicalNot)
		}
	case *exprpb.Expr_ComprehensionExpr:
		if exprKind.ComprehensionExpr.IterRange != nil {
			// Comprehension is only partially supported, as we force the user to provide a list when
			// writing an expression that uses a macro, e.g.: `["tag1", "tag2"].all(x, x in tags)`.
			if _, ok := exprKind.ComprehensionExpr.IterRange.GetExprKind().(*exprpb.Expr_ListExpr); !ok {
				return errors.New("macros must range upon a provided list value, not a variable")
			}

			// Validate that the range to iterate on is of the proper type (a list).
			if err := validateExpression(exprKind.ComprehensionExpr.IterRange, expr, usedOperators); err != nil {
				return err
			}
		}

		// Validate the conditional logic used (which is internally generated). This specifically
		// allows us to support negations (which is used by the auto-generated `__result__` identifier),
		// however return an error when they're used for anything else, e.g.: `!("tag1" in tags)`.
		if exprKind.ComprehensionExpr.LoopCondition != nil {
			if err := validateExpression(exprKind.ComprehensionExpr.LoopCondition, expr, usedOperators); err != nil {
				return err
			}
		}
	case *exprpb.Expr_SelectExpr:
		// As there is no current use-case for field selection, e.g.: `field.key`, it's not supported.
		unsupportedExpression = "field selection"
	case *exprpb.Expr_ListExpr:
		// List expressions are supported with comprehension (the `list()` macro), however otherwise, they are not.
		if _, ok := parentExpr.GetExprKind().(*exprpb.Expr_ComprehensionExpr); !ok {
			unsupportedExpression = "list"
		}
	case *exprpb.Expr_StructExpr:
		// Per the docs of `google.api.expr.v1alpha1.CreateStruct.message_name`,
		// the message name is only provided when it's a message/object.
		if exprKind.StructExpr.MessageName != "" {
			unsupportedExpression = fmt.Sprintf("message (%s)", exprKind.StructExpr.MessageName)
		} else {
			unsupportedExpression = "map"
		}
	default:
		// Should never happen, but just a sanity check.
		unsupportedExpression = "unknown"
	}
	if unsupportedExpression != "" {
		return fmt.Errorf("unsupported expression: %s", unsupportedExpression)
	}

	return nil
}

// Modified from cel-go's checker.init() function, which sets up CEL's default declarations.
func getCELDeclarations() []*exprpb.Decl {
	// Some shortcuts we use when building declarations.
	paramA := decls.NewTypeParamType("A")
	typeParamAList := []string{"A"}
	listOfA := decls.NewListType(paramA)

	return []*exprpb.Decl{
		// Our own variables that are supported.
		decls.NewVar("tags", decls.NewListType(decls.String)),

		// Only string constants are supported right now.
		decls.NewVar(checker.FormatCheckedType(decls.String), decls.NewTypeType(decls.String)),

		// Users are able to provide a list, which is helpful when using the `all()` or `exists()` macros.
		decls.NewVar("list", decls.NewTypeType(listOfA)),

		// Booleans.
		decls.NewFunction(operators.LogicalAnd, decls.NewOverload(
			overloads.LogicalAnd,
			[]*exprpb.Type{decls.Bool, decls.Bool},
			decls.Bool,
		)),
		decls.NewFunction(operators.LogicalOr, decls.NewOverload(
			overloads.LogicalOr,
			[]*exprpb.Type{decls.Bool, decls.Bool},
			decls.Bool,
		)),
		decls.NewFunction(operators.NotStrictlyFalse, decls.NewOverload(
			overloads.NotStrictlyFalse,
			[]*exprpb.Type{decls.Bool},
			decls.Bool,
		)),
		// "Logical not" is only supported for the `exists()` macro.
		decls.NewFunction(operators.LogicalNot, decls.NewOverload(
			overloads.LogicalNot,
			[]*exprpb.Type{decls.Bool},
			decls.Bool,
		)),

		// Collections.
		decls.NewFunction(operators.In, decls.NewParameterizedOverload(
			overloads.InList,
			[]*exprpb.Type{paramA, listOfA},
			decls.Bool,
			typeParamAList,
		)),
	}
}
