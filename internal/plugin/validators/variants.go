// Copyright (c) 2026 Aiven, Helsinki, Finland. https://aiven.io/

package validators

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/samber/lo"
)

// Variant is an object schema of a "oneOf" or "anyOf".
type Variant struct {
	Title      string
	Attributes []string
	Required   []string // Besides the attributes the schema requires
}

// ExactlyOneVariant fails unless the configured attributes fit exactly one variant ("oneOf").
func ExactlyOneVariant(variants ...Variant) validator.Object {
	return variantsValidator{variants: variants, exactlyOne: true}
}

// AtLeastOneVariant fails unless the configured attributes fit a variant ("anyOf").
func AtLeastOneVariant(variants ...Variant) validator.Object {
	return variantsValidator{variants: variants}
}

type variantsValidator struct {
	variants   []Variant
	exactlyOne bool
}

func (v variantsValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v variantsValidator) MarkdownDescription(context.Context) string {
	return fmt.Sprintf("the attributes must fit %s of %s", lo.Ternary(v.exactlyOne, "exactly one", "at least one"), joinWords(lo.Map(v.variants, title)))
}

func (v variantsValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, rsp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	// An unknown value can be anything, so it neither conflicts nor misses
	var set, unknown []string
	attrs := req.ConfigValue.Attributes()
	all := lo.FlatMap(v.variants, func(x Variant, _ int) []string { return x.Attributes })
	for _, k := range slices.Sorted(maps.Keys(attrs)) {
		switch val := attrs[k]; {
		case !slices.Contains(all, k):
		case val.IsUnknown():
			unknown = append(unknown, k)
		case !isEmpty(val):
			set = append(set, k)
		}
	}

	fit := lo.Filter(v.variants, func(x Variant, _ int) bool { return lo.Every(x.Attributes, set) })
	if len(fit) == 0 {
		rsp.Diagnostics.Append(v.conflict(req.Path, set))
		return
	}

	var match []Variant
	var missing []string
	for _, x := range fit {
		m := lo.Without(lo.Without(x.Required, set...), unknown...)
		if len(m) == 0 {
			match = append(match, x)
		} else {
			missing = append(missing, fmt.Sprintf("%s requires %s", x.Title, joinWords(quote(m))))
		}
	}

	switch {
	case len(match) == 0 && len(missing) == 1:
		rsp.Diagnostics.AddAttributeError(req.Path, "Missing Attribute Configuration", missing[0]+".")
	case len(match) == 0:
		rsp.Diagnostics.AddAttributeError(req.Path, "Missing Attribute Configuration",
			"Set the attributes of one variant: "+strings.Join(missing, "; ")+".")
	case v.exactlyOne && len(match) > 1 && len(unknown) == 0:
		rsp.Diagnostics.AddAttributeError(req.Path, "Invalid Attribute Combination",
			fmt.Sprintf("The attributes fit %s: set the attributes of exactly one.", joinWords(lo.Map(match, title))))
	}
}

// conflict reports the first two attributes no variant has together.
func (v variantsValidator) conflict(p path.Path, set []string) diag.Diagnostic {
	in := func(k string) []string {
		return lo.FilterMap(v.variants, func(x Variant, _ int) (string, bool) {
			return x.Title, slices.Contains(x.Attributes, k)
		})
	}

	for i, a := range set {
		for _, b := range set[i+1:] {
			if len(lo.Intersect(in(a), in(b))) == 0 {
				return diag.NewAttributeErrorDiagnostic(p.AtName(b), "Invalid Attribute Combination", fmt.Sprintf(
					"`%s` and `%s` can't be set together: `%s` is in %s, `%s` is in %s.",
					a, b, a, joinWords(in(a)), b, joinWords(in(b)),
				))
			}
		}
	}
	return diag.NewAttributeErrorDiagnostic(p, "Invalid Attribute Combination",
		joinWords(quote(set))+" can't be set together: no variant has all of them.")
}

// isEmpty tells if the value is null or an empty collection, e.g., a nested block that isn't configured.
func isEmpty(v attr.Value) bool {
	switch c := v.(type) {
	case basetypes.ListValue:
		return len(c.Elements()) == 0
	case basetypes.SetValue:
		return len(c.Elements()) == 0
	case basetypes.MapValue:
		return len(c.Elements()) == 0
	}
	return v.IsNull()
}

func title(x Variant, _ int) string {
	return x.Title
}

func quote(names []string) []string {
	return lo.Map(names, func(s string, _ int) string { return "`" + s + "`" })
}

// joinWords lists the words as "a, b and c".
func joinWords(words []string) string {
	if len(words) < 2 {
		return strings.Join(words, "")
	}
	return strings.Join(words[:len(words)-1], ", ") + " and " + words[len(words)-1]
}
