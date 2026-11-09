package main

import (
	"fmt"
	"go/token"
	"net/http"
	"reflect"
	"slices"

	"github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

// validateSteps checks the operations a view calls in order: API operations and Go functions.
func validateSteps(ops Operations) error {
	for i, op := range ops {
		if (op.ID == "") == (op.Fn == "") {
			return fmt.Errorf("operation %d: set either id or fn", i)
		}
		if op.Fn != "" {
			if !token.IsIdentifier(op.Fn) {
				return fmt.Errorf("fn %q is not a Go identifier", op.Fn)
			}
			if !reflect.DeepEqual(*op, Operation{Type: op.Type, Fn: op.Fn}) {
				return fmt.Errorf("fn %q takes only type", op.Fn)
			}
		}
		if op.Pagination != nil && len(op.Errors) > 0 {
			return fmt.Errorf("%s: errors can't be combined with pagination", op.ID)
		}
		for _, r := range op.Errors {
			if !slices.Contains(errorAsValues, r.As) {
				values := lo.Map(errorAsValues, func(v ErrorAs, _ int) string { return string(v) })
				return fmt.Errorf("%s: errors.as must be %s, got %q", op.ID, humanizeCodeList(values, " or "), r.As)
			}
			if r.As != ErrorAsIgnore && r.Status == 0 && r.Contains == "" {
				return fmt.Errorf("%s: errors.as %s requires status or contains", op.ID, r.As)
			}
			if r.Status != 0 && http.StatusText(r.Status) == "" {
				return fmt.Errorf("%s: errors.status %d is not an HTTP status", op.ID, r.Status)
			}
		}
	}
	return nil
}

// errorAsSentinels names the adapter error that wraps a matched error, an empty name drops it.
var errorAsSentinels = map[ErrorAs]string{
	ErrorAsIgnore:   "",
	ErrorAsNotFound: "ErrNotFound",
	ErrorAsPending:  "ErrRefreshStateDesired",
}

// genMapError renders adapter.MapError(err, rules...).
func genMapError(rules []ErrorRule) *jen.Statement {
	return jen.Qual(adapterPackage, "MapError").CallFunc(func(g *jen.Group) {
		g.Err()
		for _, r := range rules {
			fields := jen.Dict{}
			if r.Status != 0 {
				fields[jen.Id("Status")] = jen.Lit(r.Status)
			}
			if r.Contains != "" {
				fields[jen.Id("Contains")] = jen.Lit(r.Contains)
			}
			as := jen.Nil()
			if name := errorAsSentinels[r.As]; name != "" {
				as = jen.Qual(adapterPackage, name)
			}
			fields[jen.Id("As")] = as
			g.Qual(adapterPackage, "ErrorRule").Values(fields)
		}
	})
}
