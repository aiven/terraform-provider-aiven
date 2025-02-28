// Package diagnostics provides diagnostic functionality for resources and data sources.
package diagnostics

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
)

// DiagnosticsHelper provides helper methods for working with diagnostics.
type DiagnosticsHelper struct {
	typeName string
}

// NewDiagnosticsHelper creates a new DiagnosticsHelper.
func NewDiagnosticsHelper(typeName string) *DiagnosticsHelper {
	return &DiagnosticsHelper{
		typeName: typeName,
	}
}

// AddError adds a resource error diagnostic.
func (h *DiagnosticsHelper) AddError(diags *diag.Diagnostics, operation string, err error) {
	var summary, detail string

	switch operation {
	case "creating":
		summary = errmsg.SummaryErrorCreatingResource
		detail = fmt.Sprintf(errmsg.DetailErrorCreatingResource, h.typeName, err)
	case "reading":
		summary = errmsg.SummaryErrorReadingResource
		detail = fmt.Sprintf(errmsg.DetailErrorReadingResource, h.typeName, err)
	case "updating":
		summary = errmsg.SummaryErrorUpdatingResource
		detail = fmt.Sprintf(errmsg.DetailErrorUpdatingResource, h.typeName, err)
	case "deleting":
		summary = errmsg.SummaryErrorDeletingResource
		detail = fmt.Sprintf(errmsg.DetailErrorDeletingResource, h.typeName, err)
	case "importing":
		summary = errmsg.SummaryErrorImportingResource
		detail = fmt.Sprintf(errmsg.DetailResourceNotFound, err, err)
	default:
		summary = errmsg.SummaryUnexpectedError
		detail = fmt.Sprintf(errmsg.DetailUnexpectedError, err)
	}

	diags.Append(diag.NewErrorDiagnostic(summary, detail))
}

// AddAttributeError adds an attribute error diagnostic.
func (h *DiagnosticsHelper) AddAttributeError(diags *diag.Diagnostics, attributePath path.Path, summary, detail string) {
	diags.Append(diag.NewAttributeErrorDiagnostic(
		attributePath,
		summary,
		detail,
	))
}

// AddRequiredFieldError adds a validation error for a required field.
func (h *DiagnosticsHelper) AddRequiredFieldError(diags *diag.Diagnostics, attributePath path.Path, fieldName string) {
	diags.Append(diag.NewAttributeErrorDiagnostic(
		attributePath,
		errmsg.SummaryMissingRequiredField,
		fmt.Sprintf(errmsg.DetailMissingRequiredField, fieldName),
	))
}

// AddEmptyFieldError adds a validation error for an empty field.
func (h *DiagnosticsHelper) AddEmptyFieldError(diags *diag.Diagnostics, attributePath path.Path, fieldName string) {
	diags.Append(diag.NewAttributeErrorDiagnostic(
		attributePath,
		errmsg.SummaryEmptyRequiredField,
		fmt.Sprintf(errmsg.DetailEmptyRequiredField, fieldName),
	))
}

// AddInvalidFormatError adds a validation error for a field with an invalid format.
func (h *DiagnosticsHelper) AddInvalidFormatError(diags *diag.Diagnostics, attributePath path.Path, fieldName, format string) {
	diags.Append(diag.NewAttributeErrorDiagnostic(
		attributePath,
		errmsg.SummaryInvalidFieldFormat,
		fmt.Sprintf(errmsg.DetailInvalidFieldFormat, fieldName, format),
	))
}

// AddInvalidImportIDError adds a validation error for an invalid import ID.
func (h *DiagnosticsHelper) AddInvalidImportIDError(diags *diag.Diagnostics, expectedFormat, actualID string) {
	diags.Append(diag.NewErrorDiagnostic(
		errmsg.SummaryErrorImportingResource,
		fmt.Sprintf(errmsg.DetailInvalidImportIDFormat, expectedFormat, actualID),
	))
}

// AddWarning adds a warning diagnostic.
func (h *DiagnosticsHelper) AddWarning(diags *diag.Diagnostics, summary, detail string) {
	diags.Append(diag.NewWarningDiagnostic(
		summary,
		detail,
	))
}

// AddAttributeWarning adds an attribute warning diagnostic.
func (h *DiagnosticsHelper) AddAttributeWarning(diags *diag.Diagnostics, attributePath path.Path, summary, detail string) {
	diags.Append(diag.NewAttributeWarningDiagnostic(
		attributePath,
		summary,
		detail,
	))
}

// AddAPIError adds an API error diagnostic.
func (h *DiagnosticsHelper) AddAPIError(diags *diag.Diagnostics, operation string, err error) {
	var summary string

	switch operation {
	case "creating":
		summary = errmsg.SummaryErrorCreatingResource
	case "reading":
		summary = errmsg.SummaryErrorReadingResource
	case "updating":
		summary = errmsg.SummaryErrorUpdatingResource
	case "deleting":
		summary = errmsg.SummaryErrorDeletingResource
	default:
		summary = errmsg.SummaryUnexpectedError
	}

	diags.Append(diag.NewErrorDiagnostic(
		summary,
		fmt.Sprintf("Could not %s %s: %s", operation, h.typeName, err),
	))
}

// AddValidationError adds a validation error diagnostic.
func (h *DiagnosticsHelper) AddValidationError(diags *diag.Diagnostics, attributePath path.Path, msg string) {
	diags.Append(diag.NewAttributeErrorDiagnostic(
		attributePath,
		errmsg.SummaryInvalidConfiguration,
		msg,
	))
}
