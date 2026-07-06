// Copyright (c) 2026 Aiven, Helsinki, Finland. https://aiven.io/

package validators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestCIDR(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		value       types.String
		diagnostics diag.Diagnostics
		name        string
	}{
		{
			name:        "valid IPv4 CIDR",
			value:       types.StringValue("10.0.0.0/24"),
			diagnostics: nil,
		},
		{
			name:        "valid IPv6 CIDR",
			value:       types.StringValue("2001:db8::/32"),
			diagnostics: nil,
		},
		{
			name:  "invalid CIDR",
			value: types.StringValue("256.256.256.256/24"),
			diagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid String Attribute",
					`"256.256.256.256/24" must be a valid CIDR Value`,
				),
			},
		},
		{
			name:        "null value",
			value:       types.StringNull(),
			diagnostics: nil,
		},
		{
			name:        "unknown value",
			value:       types.StringUnknown(),
			diagnostics: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := validator.StringRequest{
				Path:        path.Root("test"),
				ConfigValue: tc.value,
			}
			res := &validator.StringResponse{}
			CIDR().ValidateString(context.Background(), req, res)
			assert.Equal(t, tc.diagnostics, res.Diagnostics)
		})
	}
}
