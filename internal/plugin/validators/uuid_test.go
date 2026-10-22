// Copyright (c) 2026 Aiven, Helsinki, Finland. https://aiven.io/

package validators

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestUUID(t *testing.T) {
	t.Parallel()

	invalid := func(v string) diag.Diagnostics {
		return diag.Diagnostics{
			diag.NewAttributeErrorDiagnostic(
				path.Root("test"),
				"Invalid String Attribute",
				fmt.Sprintf("%q must be a valid UUID", v),
			),
		}
	}

	testCases := []struct {
		value       types.String
		diagnostics diag.Diagnostics
		name        string
	}{
		{
			name:        "valid lowercase UUID",
			value:       types.StringValue("1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d"),
			diagnostics: nil,
		},
		{
			name:        "valid uppercase UUID",
			value:       types.StringValue("1A2B3C4D-5E6F-4A8B-9C0D-1E2F3A4B5C6D"),
			diagnostics: nil,
		},
		{
			name:        "valid nil UUID",
			value:       types.StringValue("00000000-0000-0000-0000-000000000000"),
			diagnostics: nil,
		},
		{
			name:        "missing hyphens",
			value:       types.StringValue("1a2b3c4d5e6f4a8b9c0d1e2f3a4b5c6d"),
			diagnostics: invalid("1a2b3c4d5e6f4a8b9c0d1e2f3a4b5c6d"),
		},
		{
			name:        "braced UUID",
			value:       types.StringValue("{1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d}"),
			diagnostics: invalid("{1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d}"),
		},
		{
			name:        "non-hex character",
			value:       types.StringValue("1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6g"),
			diagnostics: invalid("1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6g"),
		},
		{
			name:        "empty string",
			value:       types.StringValue(""),
			diagnostics: invalid(""),
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
			UUID().ValidateString(context.Background(), req, res)
			assert.Equal(t, tc.diagnostics, res.Diagnostics)
		})
	}
}
