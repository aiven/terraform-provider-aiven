package user

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
)

func init() {
	ResourceOptions.Create = createView
	ResourceOptions.Update = updateView
}

func createView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	var req service.ServiceUserCreateIn
	err := d.Expand(&req, expandModifier(ctx, client))
	if err != nil {
		return err
	}

	_, err = client.ServiceUserCreate(ctx, d.Get("project").(string), d.Get("service_name").(string), &req)
	if err != nil {
		return err
	}

	return resetPassword(ctx, client, d)
}

func updateView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	if d.HasChange("pg_allow_replication") {
		req := &service.ServiceUserCredentialsModifyIn{
			Operation: service.ServiceUserCredentialsModifyOperationTypeSetAccessControl,
			AccessControl: &service.AccessControlIn{
				PgAllowReplication: util.ToPtr(d.Get("pg_allow_replication").(bool)),
			},
		}
		_, err := client.ServiceUserCredentialsModify(ctx, d.Get("project").(string), d.Get("service_name").(string), d.Get("username").(string), req)
		if err != nil {
			return err
		}
	}

	if d.HasChange("password") || d.HasChange("password_wo_version") {
		return resetPassword(ctx, client, d)
	}
	return nil
}

func resetPassword(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	password := util.NilIfZero(d.Get("password_wo").(string), d.Get("password").(string))
	if password == nil {
		return nil
	}

	req := &service.ServiceUserCredentialsModifyIn{
		Operation:   service.ServiceUserCredentialsModifyOperationTypeResetCredentials,
		NewPassword: password,
	}

	_, err := client.ServiceUserCredentialsModify(ctx, d.Get("project").(string), d.Get("service_name").(string), d.Get("username").(string), req)
	return err
}

func expandModifier(_ context.Context, _ avngen.Client) adapter.MapModifier {
	return func(d adapter.ResourceData, dto map[string]any) error {
		// Move pg_allow_replication into access_control for the API request
		if v, ok := d.GetOk("pg_allow_replication"); ok {
			dto["access_control"] = map[string]any{"pg_allow_replication": v}
			delete(dto, "pg_allow_replication")
		}
		return nil
	}
}

func flattenModifier(_ context.Context, _ avngen.Client) adapter.MapModifier {
	return func(d adapter.ResourceData, dto map[string]any) error {
		// pg_allow_replication: use plan value if known, otherwise extract from access_control.
		if v, ok := dto["access_control"]; ok {
			dto["pg_allow_replication"] = v.(map[string]any)["pg_allow_replication"]
			delete(dto, "access_control")
		}

		// Clear password from state when using write-only password.
		if d.Get("password_wo_version").(int) != 0 {
			delete(dto, "password")
		}
		return nil
	}
}
