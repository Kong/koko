package persistence

import (
	"fmt"

	"github.com/google/cel-go/common/operators"
	"github.com/samber/lo"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

// GetTagsFromExpression takes in a pre-validated CEL expression and extracts the tags to filter upon.
// The CEL logical operator is also returned (`operators.LogicalAnd` or `operators.LogicalOr`). As a
// sanity check, an error will be returned when the expression does not conform to what is expected.
func GetTagsFromExpression(expr *exprpb.Expr) (tags []*exprpb.Constant, exprFunction string, err error) {
	switch exprKind := expr.GetExprKind().(type) {
	case *exprpb.Expr_CallExpr:
		// Support simple logic and/or queries, e.g.:
		// - `"tag1" in tags && "tag2" in tags`
		// - `"tag1" in tags || "tag2" in tags`
		callExpr := exprKind.CallExpr

		// Extract whether this is a logical `and` or `or` expression.
		exprFunction = callExpr.Function

		// Should not happen as long as we're validating expressions properly.
		if !lo.Contains([]string{operators.LogicalAnd, operators.LogicalOr, operators.In}, callExpr.Function) {
			return nil, "", fmt.Errorf("operator %q is not supported", callExpr.Function)
		}

		// Extract the tags the expression is asking to filter on. We're purposefully ignoring the
		// field names to filter upon, as right now, the only supported identifier is `tags` anyway.
		for _, arg := range callExpr.Args {
			switch exprKind := arg.GetExprKind().(type) {
			case *exprpb.Expr_IdentExpr:
				// No-op as we don't need to do anything with identifiers right now.
				continue
			case *exprpb.Expr_CallExpr:
				if exprKind.CallExpr.Function != operators.In {
					return nil, "", fmt.Errorf("unexpected operator %q for call expression", exprKind.CallExpr.Function)
				}

				// Should never happen, just a sanity check.
				if len(exprKind.CallExpr.Args) != 2 { //nolint:gomnd
					return nil, "", fmt.Errorf(
						"expecting both constant & identifier expression, but got %d call expression arguments",
						len(exprKind.CallExpr.Args),
					)
				}

				constExpr := exprKind.CallExpr.Args[0].GetConstExpr()
				if constExpr == nil {
					return nil, "", fmt.Errorf(
						"expected constant expression, got %T",
						exprKind.CallExpr.Args[0].GetExprKind(),
					)
				}
				tags = append(tags, constExpr)
			case *exprpb.Expr_ConstExpr:
				// Has a single tag to filter on, e.g: `"tag1" in tags`.
				tags = append(tags, exprKind.ConstExpr)
			default:
				return nil, "", fmt.Errorf("unexpected expression kind: %T", exprKind)
			}
		}
	case *exprpb.Expr_ComprehensionExpr:
		// Support `exists()` & `all()` macros, e.g.:
		// - `["tag1", "tag2"].all(x, x in tags)`
		// - `["tag1", "tag2"].exists(x, x in tags)`
		compExpr := exprKind.ComprehensionExpr

		// Extract whether this is a logical `and` or `or` expression.
		loopCallExpr := compExpr.LoopStep.GetCallExpr()
		if loopCallExpr == nil {
			return nil, "", fmt.Errorf("unexpected loop step expression kind: %T", compExpr.LoopStep.GetExprKind())
		}
		exprFunction = loopCallExpr.Function

		listExpr := compExpr.IterRange.GetListExpr()
		if listExpr == nil {
			return nil, "", fmt.Errorf("unexpected iterative range kind: %T", compExpr.IterRange.GetExprKind())
		}

		for _, element := range listExpr.Elements {
			constExpr := element.GetConstExpr()
			if constExpr == nil {
				return nil, "", fmt.Errorf("unexpected list expression kind: %T", element.GetExprKind())
			}
			tags = append(tags, constExpr)
		}
	default:
		return nil, "", fmt.Errorf("unsupported expression kind: %T", exprKind)
	}

	return tags, exprFunction, nil
}

// GetQueryArgsFromExprConstants is a simple helper function to convert user-provided constant
// values to their string interface counterparts, which is helpful for use as placeholder values
// when writing DB queries. This will also filter out any duplicate tag values. An error will
// only be returned when a constant is not of the `exprpb.Constant_StringValue` type.
func GetQueryArgsFromExprConstants(constants []*exprpb.Constant) ([]interface{}, error) {
	args, seenTags := make([]interface{}, 0, len(constants)), make(map[string]bool, len(constants))
	for _, c := range constants {
		strConst, ok := c.GetConstantKind().(*exprpb.Constant_StringValue)
		if !ok {
			return nil, fmt.Errorf("unexpected constant kind: %T", c.GetConstantKind())
		}

		// If for whatever reason the user provided duplicate tags, we'll want to drop them,
		// as the length of the returned tags array can be used in `HAVING` clauses, in order
		// to assert that all tags have been returned (when exprFunction = `operators.Equals`).
		if !seenTags[strConst.StringValue] {
			args, seenTags[strConst.StringValue] = append(args, strConst.StringValue), true
		}
	}

	return args, nil
}
