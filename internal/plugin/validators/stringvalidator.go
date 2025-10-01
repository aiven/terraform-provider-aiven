// Copyright (c) 2025 Aiven, Helsinki, Finland. https://aiven.io/

package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = (*stringValidator)(nil)

// NewStringValidator creates a string validator with custom validation logic.
// Example: NewStringValidator("must be a valid time", validateTimeString)
func NewStringValidator(message string, validate func(v string) error) validator.String {
	return &stringValidator{
		message:  message,
		validate: validate,
	}
}

type stringValidator struct {
	message  string
	validate func(v string) error
}

func (s *stringValidator) Description(ctx context.Context) string {
	return s.message
}

func (s *stringValidator) MarkdownDescription(ctx context.Context) string {
	return s.message
}

func (s *stringValidator) ValidateString(ctx context.Context, req validator.StringRequest, rsp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	v := req.ConfigValue.ValueString()
	err := s.validate(v)
	if err != nil {
		rsp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			req.Path,
			"Invalid String Attribute",
			fmt.Sprintf("%q %s", v, s.message),
		))
	}
}
