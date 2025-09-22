package common

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// CheckDeprecatedTimeoutDefault adds a warning diagnostic if timeouts.default is configured
func CheckDeprecatedTimeoutDefault(d *schema.ResourceData) diag.Diagnostic {
	if timeout := d.Timeout(schema.TimeoutDefault); timeout > 0 {
		return diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Deprecated timeout configuration",
			Detail: "The 'timeouts.default' field is deprecated and will be removed in a future version.\n\n" +
				"Use specific CRUD timeouts (create, read, update, delete) instead.\n\n" +
				"See: https://github.com/aiven/terraform-provider-aiven/blob/main/docs/guides/update-deprecated-resources.md",
		}
	}
	return diag.Diagnostic{}
}
