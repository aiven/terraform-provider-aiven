package user

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
)

func init() {
	ResourceOptions.Create = createView
	ResourceOptions.Update = updateView
}

func createView(ctx context.Context, client avngen.Client, plan *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics
	var req service.ServiceUserCreateIn
	diags.Append(expandData(ctx, plan, nil, &req)...)
	if diags.HasError() {
		return diags
	}

	_, err := client.ServiceUserCreate(ctx, plan.Project.ValueString(), plan.ServiceName.ValueString(), &req)
	if err != nil {
		diags.Append(errmsg.FromError("ServiceUserCreate Error", err))
		return diags
	}

	diags.Append(resetPassword(ctx, client, plan, plan.PasswordWo.String())...)
	return diags
}

func updateView(ctx context.Context, client avngen.Client, plan, state, config *tfModel) diag.Diagnostics {
	hasChanged := plan.Password.ValueString() != state.Password.ValueString() ||
		plan.PasswordWoVersion.ValueInt64() != state.PasswordWoVersion.ValueInt64() ||
		plan.Authentication.ValueString() != state.Authentication.ValueString()

	if !hasChanged {
		return nil
	}

	// PasswordWO is available only in config, not in plan.
	return resetPassword(ctx, client, plan, config.PasswordWo.ValueString())
}

func resetPassword(ctx context.Context, client avngen.Client, plan *tfModel, passwordWO string) diag.Diagnostics {
	var diags diag.Diagnostics
	req := service.ServiceUserCredentialsModifyIn{
		Operation:      service.ServiceUserCredentialsModifyOperationTypeResetCredentials,
		Authentication: service.AuthenticationType(plan.Authentication.ValueString()),
		NewPassword:    util.NilIfZero(plan.Password.ValueString(), passwordWO),
	}

	_, err := client.ServiceUserCredentialsModify(ctx, plan.Project.ValueString(), plan.ServiceName.ValueString(), plan.Username.ValueString(), &req)
	if err != nil {
		diags.Append(errmsg.FromError("ServiceUserCredentialsModify Error", err))
	}
	return diags
}

func flattenModifier(ctx context.Context, client avngen.Client) util.MapModifier[tfModel] {
	return func(r util.RawMap, plan *tfModel) error {
		if plan.PasswordWoVersion.ValueInt64() != 0 {
			// Clears previous password if PasswordWo is used
			return r.Delete("password")
		}
		return nil
	}
}
