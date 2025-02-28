// Package validation provides validation functionality for resources and data sources.
package validation

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/diagnostics"
)

// ValidateRequiredStringField validates that a string field is set.
func ValidateRequiredStringField(
	field types.String,
	fieldName string,
	diags *diag.Diagnostics,
	diagHelper *diagnostics.DiagnosticsHelper,
) {
	if field.IsNull() || field.IsUnknown() {
		diagHelper.AddRequiredFieldError(diags, path.Root(fieldName), fieldName)
	}
}

// ValidateRequiredListField validates that a list field is set and not empty.
func ValidateRequiredListField(
	ctx context.Context,
	field types.List,
	fieldName string,
	diags *diag.Diagnostics,
	diagHelper *diagnostics.DiagnosticsHelper,
) {
	if field.IsNull() || field.IsUnknown() {
		diagHelper.AddRequiredFieldError(diags, path.Root(fieldName), fieldName)
		return
	}

	// Check if the list is empty for string lists
	if field.ElementType(ctx) == types.StringType {
		var elements []string
		elemDiags := field.ElementsAs(ctx, &elements, false)
		diags.Append(elemDiags...)
		if len(elements) == 0 {
			diagHelper.AddEmptyFieldError(diags, path.Root(fieldName), fieldName)
		}
	}
}

// ValidateImportID validates the import ID against the expected format.
// It returns the split parts if the validation passes, or an error if it fails.
// The expectedFormat parameter should describe the expected format (e.g. "project_name/service_name").
// The function infers the expected number of parts from the expectedFormat string.
func ValidateImportID(
	id string,
	expectedFormat string,
) ([]string, error) {
	parts := strings.Split(id, "/")

	// Check that we have at least one part
	if len(parts) < 1 {
		return nil, fmt.Errorf("invalid ID format. Expected %s", expectedFormat)
	}

	// Verify no part is empty
	for i, part := range parts {
		if part == "" {
			return nil, fmt.Errorf("invalid ID format. Part %d cannot be empty. Expected %s", i+1, expectedFormat)
		}
	}

	// Infer expected parts count from format string (count slashes + 1)
	expectedParts := strings.Count(expectedFormat, "/") + 1

	// Validate the number of parts
	if len(parts) != expectedParts {
		return nil, fmt.Errorf("invalid ID format. Expected %d parts (%s) but got %d parts",
			expectedParts, expectedFormat, len(parts))
	}

	return parts, nil
}
