// Copyright (c) 2026 Aiven, Helsinki, Finland. https://aiven.io/

package validators

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = sensitiveRegexMatchesValidator{}

type sensitiveRegexMatchesValidator struct {
	regexp  *regexp.Regexp
	message string
}

// SensitiveRegexMatches is stringvalidator.RegexMatches without the value in the error.
func SensitiveRegexMatches(re *regexp.Regexp, message string) validator.String {
	return sensitiveRegexMatchesValidator{regexp: re, message: message}
}

func (v sensitiveRegexMatchesValidator) Description(_ context.Context) string {
	return v.message
}

func (v sensitiveRegexMatchesValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v sensitiveRegexMatchesValidator) ValidateString(_ context.Context, req validator.StringRequest, rsp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	if !v.regexp.MatchString(req.ConfigValue.ValueString()) {
		rsp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			req.Path,
			"Invalid Attribute Value Match",
			fmt.Sprintf("Attribute %s %s", req.Path, v.message),
		))
	}
}
