// Copyright (c) 2026 Aiven, Helsinki, Finland. https://aiven.io/

package validators

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestSensitiveRegexMatches(t *testing.T) {
	t.Parallel()

	pattern := `^[a-z]+$`
	message := fmt.Sprintf(`must match pattern %q`, pattern)
	re := regexp.MustCompile(pattern)
	badValue := "INVALID123"

	invalid := diag.Diagnostics{
		diag.NewAttributeErrorDiagnostic(
			path.Root("test"),
			"Invalid Attribute Value Match",
			fmt.Sprintf("Attribute %s %s", path.Root("test"), message),
		),
	}

	testCases := []struct {
		value       types.String
		diagnostics diag.Diagnostics
		name        string
	}{
		{
			name:        "matching value",
			value:       types.StringValue("abc"),
			diagnostics: nil,
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
		{
			name:        "mismatching value",
			value:       types.StringValue(badValue),
			diagnostics: invalid,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := validator.StringRequest{
				Path:        path.Root("test"),
				ConfigValue: tc.value,
			}
			res := &validator.StringResponse{}
			SensitiveRegexMatches(re, message).ValidateString(context.Background(), req, res)
			assert.Equal(t, tc.diagnostics, res.Diagnostics)
			if tc.diagnostics != nil {
				assert.NotContains(t, res.Diagnostics.Errors()[0].Detail(), badValue)
			}
		})
	}
}
