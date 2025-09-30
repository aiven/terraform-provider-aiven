package common

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// CheckDeprecatedTimeoutDefault adds a warning diagnostic if timeouts.default is configured during apply operations.
// During plan refresh, GetRawConfig returns null, so warning won't be shown.
func CheckDeprecatedTimeoutDefault(d *schema.ResourceData) diag.Diagnostic {
	rawConfig := d.GetRawConfig()
	if rawConfig.IsNull() {
		return diag.Diagnostic{}
	}

	timeoutsVal := rawConfig.GetAttr(schema.TimeoutsConfigKey)
	if timeoutsVal.IsNull() {
		return diag.Diagnostic{}
	}

	if !timeoutsVal.Type().HasAttribute(schema.TimeoutDefault) {
		return diag.Diagnostic{}
	}

	defaultVal := timeoutsVal.GetAttr(schema.TimeoutDefault)
	if defaultVal.IsNull() {
		return diag.Diagnostic{}
	}

	return diag.Diagnostic{
		Severity: diag.Warning,
		Summary:  "Deprecated timeout configuration",
		Detail: "The 'timeouts.default' field is deprecated and will be removed in a future version.\n\n" +
			"Use specific CRUD timeouts (create, read, update, delete) instead.\n\n" +
			"See: https://github.com/aiven/terraform-provider-aiven/blob/main/docs/guides/update-deprecated-resources.md",
	}
}
