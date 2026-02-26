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

func createView(ctx context.Context, client avngen.Client, plan, config *tfModel) diag.Diagnostics {
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

	diags.Append(resetPassword(ctx, client, plan, config)...)
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
	return resetPassword(ctx, client, plan, config)
}

func resetPassword(ctx context.Context, client avngen.Client, plan, config *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics
	req := service.ServiceUserCredentialsModifyIn{
		Operation:      service.ServiceUserCredentialsModifyOperationTypeResetCredentials,
		Authentication: service.AuthenticationType(plan.Authentication.ValueString()),
		NewPassword:    util.NilIfZero(plan.Password.ValueString(), config.PasswordWo.ValueString()),
	}

	_, err := client.ServiceUserCredentialsModify(ctx, plan.Project.ValueString(), plan.ServiceName.ValueString(), plan.Username.ValueString(), &req)
	if err != nil {
		diags.Append(errmsg.FromError("ServiceUserCredentialsModify Error", err))
	}
	return diags
}

// flattenModifier adjusts the API response before it is unmarshalled into the state.
// After Create/Update, the adapter's refreshState calls ServiceUserGet which may return
// stale data due to API eventual consistency.
// When the plan already has a known value for a field, we use it instead of the API response to prevent
// inconsistent result after apply errors.
// On import or auto-generated passwords, the plan value is null/unknown,
// so the API response passes through unchanged.
func flattenModifier(ctx context.Context, client avngen.Client) util.MapModifier[tfModel] {
	return func(r util.RawMap, plan *tfModel) error {
		// password: use plan value if known (custom password), otherwise let the API value through (auto-generated).
		if !plan.Password.IsNull() && !plan.Password.IsUnknown() {
			if err := r.Set(plan.Password.ValueString(), "password"); err != nil {
				return err
			}
		}

		// authentication: use plan value if known, otherwise let the API value through.
		if !plan.Authentication.IsNull() && !plan.Authentication.IsUnknown() {
			if err := r.Set(plan.Authentication.ValueString(), "authentication"); err != nil {
				return err
			}
		}

		// Clear password from state when using write-only password.
		if plan.PasswordWoVersion.ValueInt64() != 0 {
			return r.Delete("password")
		}
		return nil
	}
}
