package valkeyuser

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/serviceuser"
)

func init() {
	ResourceOptions.RefreshStateCheck = serviceuser.PasswordIsReady
}

func createView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	return serviceuser.Create(ctx, client, d, expandModifier(ctx, client))
}

func updateView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	return serviceuser.Update(ctx, client, d, expandModifier(ctx, client))
}

// expandModifier sends [] for the unset ACLs rather than omitting them.
// Expand nests them under access_control afterwards.
func expandModifier(_ context.Context, _ avngen.Client) adapter.MapModifier {
	return forceACLSlices
}

// flattenModifier sets the ACLs the API omits to [].
// Otherwise an ACL dropped in Aiven would keep its stale value instead of drifting.
// Flatten lifts them out of access_control beforehand.
func flattenModifier(_ context.Context, _ avngen.Client) adapter.MapModifier {
	return adapter.ComposeMapModifiers(forceACLSlices, serviceuser.PasswordFlatten)
}

// forceACLSlices applies forceSlice to the list ACLs hoisted from access_control.
// On expand, the API applies the default policy to an omitted ACL, and none to [].
func forceACLSlices(d adapter.ResourceData, dto map[string]any) error {
	for key, sch := range d.Schema().Properties {
		isList := sch.Type == adapter.SchemaTypeList || sch.Type == adapter.SchemaTypeSet
		if isList && sch.HoistedFrom == serviceuser.AccessControlField {
			dto[key] = forceSlice(dto[key])
		}
	}
	return nil
}

// forceSlice ensures that the value is a non-nil empty slice, not null.
func forceSlice(v any) []any {
	s, ok := v.([]any)
	if !ok || len(s) == 0 {
		return make([]any, 0)
	}
	return s
}
