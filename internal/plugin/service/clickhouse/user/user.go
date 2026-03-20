package user

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

func init() {
	ResourceOptions.Update = updateViewPasswordOnly
}

func updateViewPasswordOnly(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	if d.HasChange("password") || d.HasChange("password_wo_version") {
		// We don't want to run updateView unless the password has changed,
		// since running it unnecessarily would reset the password when it was not provided by the user.
		// Currently, the password is the only attribute that can be updated,
		// so this function should only be called in this case.
		// However, in the future, new fields may be added that can be updated,
		// so this "if" statement becomes a safety net.
		return updateView(ctx, client, d)
	}

	return nil
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
