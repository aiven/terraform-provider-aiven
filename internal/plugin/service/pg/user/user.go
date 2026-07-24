package user

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/serviceuser"
)

func init() {
	ResourceOptions.Update = updateView
}

func createView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	return serviceuser.Create(ctx, client, d, expandModifier(ctx, client))
}

func updateView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	if d.HasChange("pg_allow_replication") {
		req := &service.ServiceUserCredentialsModifyIn{
			Operation: service.ServiceUserCredentialsModifyOperationTypeSetAccessControl,
			AccessControl: &service.AccessControlIn{
				PGAllowReplication: new(d.Get("pg_allow_replication").(bool)),
			},
		}
		if _, err := client.ServiceUserCredentialsModify(ctx, d.Get("project").(string), d.Get("service_name").(string), d.Get("username").(string), req); err != nil {
			return err
		}
	}

	return serviceuser.ResetPassword(ctx, client, d)
}

// expandModifier moves pg_allow_replication into access_control for the API request.
func expandModifier(_ context.Context, _ avngen.Client) adapter.MapModifier {
	return func(d adapter.ResourceData, dto map[string]any) error {
		if v, ok := d.GetOk("pg_allow_replication"); ok {
			dto["access_control"] = map[string]any{"pg_allow_replication": v}
			delete(dto, "pg_allow_replication")
		}
		return nil
	}
}

// flattenModifier extracts pg_allow_replication out of access_control, then runs
// the shared password reconciliation.
func flattenModifier(_ context.Context, _ avngen.Client) adapter.MapModifier {
	return adapter.ComposeMapModifiers(flattenAllowReplication, serviceuser.PasswordFlatten)
}

func flattenAllowReplication(_ adapter.ResourceData, dto map[string]any) error {
	if v, ok := dto["access_control"]; ok {
		dto["pg_allow_replication"] = v.(map[string]any)["pg_allow_replication"]
		delete(dto, "access_control")
	}
	return nil
}
