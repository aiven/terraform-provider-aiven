// Copyright (c) 2025 Aiven, Helsinki, Finland. https://aiven.io/

package legacytimeouts

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestTimeDuration(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		value       types.String
		diagnostics diag.Diagnostics
		name        string
	}{
		{
			name:        "valid duration - seconds",
			value:       types.StringValue("30s"),
			diagnostics: nil,
		},
		{
			name:        "valid duration - minutes",
			value:       types.StringValue("5m"),
			diagnostics: nil,
		},
		{
			name:        "valid duration - hours",
			value:       types.StringValue("1h"),
			diagnostics: nil,
		},
		{
			name:        "valid duration - mixed",
			value:       types.StringValue("1h2m3s"),
			diagnostics: nil,
		},
		{
			name:  "invalid duration",
			value: types.StringValue("invalid"),
			diagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid String Attribute",
					`"invalid" must be a valid time duration string, e.g. "30s", "5m", "1h"`,
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
			TimeDuration().ValidateString(context.Background(), req, res)
			assert.Equal(t, tc.diagnostics, res.Diagnostics)
		})
	}
}
