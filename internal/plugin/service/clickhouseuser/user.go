package clickhouseuser

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

// requirePasswordChange skips the update unless the password changed: the reset would replace a password
// the user didn't provide. The password is the only attribute that can be updated now, so this is a
// safety net for attributes added later.
func requirePasswordChange(_ context.Context, _ avngen.Client, d adapter.ResourceData) error {
	if d.HasChange("password") || d.HasChange("password_wo_version") {
		return nil
	}
	return adapter.ErrNoChanges
}

func expandModifier(_ context.Context, _ avngen.Client) adapter.MapModifier {
	return func(d adapter.ResourceData, dto map[string]any) error {
		if v, ok := d.GetOk("password_wo"); ok {
			// Sets Write-only password to the password field in the API request
			dto["password"] = v
		}
		return nil
	}
}

func flattenModifier(_ context.Context, _ avngen.Client) adapter.MapModifier {
	return func(d adapter.ResourceData, dto map[string]any) error {
		if _, ok := d.Schema().Properties["password_wo_version"]; ok && d.Get("password_wo_version").(int) != 0 {
			// Clear password from state when using write-only password.
			delete(dto, "password")
		}

		return nil
	}
}
