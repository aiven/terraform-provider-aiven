package user

import (
	"context"
	"strings"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/serviceuser"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

// valkeyACLPrefix marks the ACL attributes that the API nests under
// access_control but which the resource exposes as top-level list attributes.
const valkeyACLPrefix = "valkey_acl_"

func createView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	return serviceuser.Create(ctx, client, d, expandModifier(ctx, client))
}

func updateView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	aclChanged := false
	accessControl := make(map[string]any)
	for key := range d.Schema().Properties {
		if strings.HasPrefix(key, valkeyACLPrefix) {
			accessControl[key] = forceSlice(d.Get(key))
			if d.HasChange(key) {
				aclChanged = true
			}
		}
	}

	if aclChanged {
		var ac service.AccessControlIn
		err := schemautil.Remarshal(accessControl, &ac)
		if err != nil {
			return err
		}

		req := &service.ServiceUserCredentialsModifyIn{
			Operation:     service.ServiceUserCredentialsModifyOperationTypeSetAccessControl,
			AccessControl: &ac,
		}

		_, err = client.ServiceUserCredentialsModify(ctx, d.Get("project").(string), d.Get("service_name").(string), d.Get("username").(string), req)
		if err != nil {
			return err
		}
	}

	return serviceuser.ResetPassword(ctx, client, d)
}

// expandModifier moves the top-level valkey_acl_* attributes into the
// access_control object expected by the create request.
func expandModifier(_ context.Context, _ avngen.Client) adapter.MapModifier {
	return func(d adapter.ResourceData, dto map[string]any) error {
		accessControl := make(map[string]any)
		for key := range d.Schema().Properties {
			if strings.HasPrefix(key, valkeyACLPrefix) {
				accessControl[key] = forceSlice(d.Get(key))
				delete(dto, key)
			}
		}
		if len(accessControl) > 0 {
			dto["access_control"] = accessControl
		}
		return nil
	}
}

// flattenModifier lifts the valkey_acl_* lists out of access_control, then runs
// the shared password reconciliation.
func flattenModifier(_ context.Context, _ avngen.Client) adapter.MapModifier {
	return adapter.ComposeMapModifiers(flattenACL, serviceuser.PasswordFlatten)
}

func flattenACL(d adapter.ResourceData, dto map[string]any) error {
	accessControl, _ := dto["access_control"].(map[string]any)
	delete(dto, "access_control")

	// The API omits the ACLs it has no values for, so each one is set explicitly.
	// Otherwise an ACL dropped in Aiven would keep its stale value instead of drifting.
	for key := range d.Schema().Properties {
		if strings.HasPrefix(key, valkeyACLPrefix) {
			dto[key] = forceSlice(accessControl[key])
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
