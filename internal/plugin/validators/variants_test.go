// Copyright (c) 2026 Aiven, Helsinki, Finland. https://aiven.io/

package validators

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVariants(t *testing.T) {
	// Like OpenSearch custom_repos settings: base_path is required by the schema
	repos := []Variant{
		{Title: "Variant 1", Attributes: []string{"account", "base_path", "container"}, Required: []string{"account", "container"}},
		{Title: "Variant 2", Attributes: []string{"base_path", "bucket", "credentials"}, Required: []string{"bucket", "credentials"}},
		{Title: "Variant 3", Attributes: []string{"access_key", "base_path", "bucket", "region", "secret_key"}, Required: []string{"access_key", "bucket", "region", "secret_key"}},
	}
	ambiguous := []Variant{
		{Title: "A", Attributes: []string{"a", "b"}, Required: []string{"a"}},
		{Title: "B", Attributes: []string{"a", "c"}, Required: []string{"a"}},
	}
	pairwise := []Variant{
		{Title: "A", Attributes: []string{"a", "b"}},
		{Title: "B", Attributes: []string{"b", "c"}},
		{Title: "C", Attributes: []string{"a", "c"}},
	}

	str := types.StringValue
	credentials := func(v ...attr.Value) types.List { return types.ListValueMust(types.StringType, v) }
	attrTypes := map[string]attr.Type{"credentials": types.ListType{ElemType: types.StringType}}
	for _, k := range []string{"a", "b", "c", "access_key", "account", "base_path", "bucket", "container", "other", "region", "secret_key"} {
		attrTypes[k] = types.StringType
	}

	cases := []struct {
		name      string
		validator validator.Object
		config    map[string]attr.Value // Null when nil
		path      string
		summary   string
		detail    string
	}{
		{
			name:      "null",
			validator: ExactlyOneVariant(repos...),
		},
		{
			name:      "valid",
			validator: ExactlyOneVariant(repos...),
			config:    map[string]attr.Value{"account": str("a"), "base_path": str("p"), "container": str("c"), "other": str("o")},
		},
		{
			name:      "nothing set",
			validator: ExactlyOneVariant(repos...),
			config:    map[string]attr.Value{},
			path:      "settings",
			summary:   "Missing Attribute Configuration",
			detail:    "Set the attributes of one variant: Variant 1 requires `account` and `container`; Variant 2 requires `bucket` and `credentials`; Variant 3 requires `access_key`, `bucket`, `region` and `secret_key`.",
		},
		{
			name:      "conflict",
			validator: ExactlyOneVariant(repos...),
			config:    map[string]attr.Value{"bucket": str("b"), "container": str("c")},
			path:      "settings.container",
			summary:   "Invalid Attribute Combination",
			detail:    "`bucket` and `container` can't be set together: `bucket` is in Variant 2 and Variant 3, `container` is in Variant 1.",
		},
		{
			name:      "no variant has all",
			validator: AtLeastOneVariant(pairwise...),
			config:    map[string]attr.Value{"a": str("a"), "b": str("b"), "c": str("c")},
			path:      "settings",
			summary:   "Invalid Attribute Combination",
			detail:    "`a`, `b` and `c` can't be set together: no variant has all of them.",
		},
		{
			name:      "missing in one variant",
			validator: ExactlyOneVariant(repos...),
			config:    map[string]attr.Value{"credentials": credentials(str("c"))},
			path:      "settings",
			summary:   "Missing Attribute Configuration",
			detail:    "Variant 2 requires `bucket`.",
		},
		{
			name:      "missing in several variants",
			validator: ExactlyOneVariant(repos...),
			config:    map[string]attr.Value{"bucket": str("b")},
			path:      "settings",
			summary:   "Missing Attribute Configuration",
			detail:    "Set the attributes of one variant: Variant 2 requires `credentials`; Variant 3 requires `access_key`, `region` and `secret_key`.",
		},
		{
			name:      "unknown is set for required",
			validator: ExactlyOneVariant(repos...),
			config:    map[string]attr.Value{"bucket": str("b"), "credentials": types.ListUnknown(types.StringType)},
		},
		{
			name:      "unknown doesn't conflict",
			validator: ExactlyOneVariant(repos...),
			config:    map[string]attr.Value{"account": str("a"), "container": str("c"), "credentials": types.ListUnknown(types.StringType)},
		},
		{
			name:      "empty block is unset",
			validator: ExactlyOneVariant(repos...),
			config:    map[string]attr.Value{"account": str("a"), "container": str("c"), "credentials": credentials()},
		},
		{
			name:      "fits several",
			validator: ExactlyOneVariant(ambiguous...),
			config:    map[string]attr.Value{"a": str("a")},
			path:      "settings",
			summary:   "Invalid Attribute Combination",
			detail:    "The attributes fit A and B: set the attributes of exactly one.",
		},
		{
			name:      "fits several of anyOf",
			validator: AtLeastOneVariant(ambiguous...),
			config:    map[string]attr.Value{"a": str("a")},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			config := types.ObjectNull(attrTypes)
			if c.config != nil {
				values := make(map[string]attr.Value, len(attrTypes))
				for k, v := range attrTypes {
					values[k] = types.StringNull()
					if v != types.StringType {
						values[k] = types.ListNull(types.StringType)
					}
					if cv, ok := c.config[k]; ok {
						values[k] = cv
					}
				}
				config = types.ObjectValueMust(attrTypes, values)
			}

			rsp := new(validator.ObjectResponse)
			c.validator.ValidateObject(t.Context(), validator.ObjectRequest{Path: path.Root("settings"), ConfigValue: config}, rsp)
			if c.detail == "" {
				assert.Empty(t, rsp.Diagnostics)
				return
			}

			require.Len(t, rsp.Diagnostics, 1)
			d := rsp.Diagnostics[0]
			assert.Equal(t, c.summary, d.Summary())
			assert.Equal(t, c.detail, d.Detail())
			require.Implements(t, (*diag.DiagnosticWithPath)(nil), d)
			assert.Equal(t, c.path, d.(diag.DiagnosticWithPath).Path().String())
		})
	}
}
