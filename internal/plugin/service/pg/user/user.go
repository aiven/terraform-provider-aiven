package user

import (
	"context"
	"fmt"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/avast/retry-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

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
	diags.Append(expandData(ctx, plan, nil, &req, expandModifier(ctx, client))...)
	if diags.HasError() {
		return diags
	}

	_, err := client.ServiceUserCreate(ctx, plan.Project.ValueString(), plan.ServiceName.ValueString(), &req)
	if err != nil {
		diags.Append(errmsg.FromError("ServiceUserCreate Error", err))
		return diags
	}

	diags.Append(resetPassword(ctx, client, plan, config)...)
	if diags.HasError() {
		return diags
	}

	// Wait until the password is delivered by the API and capture it on the plan.
	// For auto-generated passwords, plan.Password is unknown — capturing it here
	// lets flattenModifier protect it from stale reads during refreshState.
	if plan.PasswordWoVersion.ValueInt64() == 0 {
		diags.Append(waitAndCapturePassword(ctx, client, plan)...)
	}
	return diags
}

func updateView(ctx context.Context, client avngen.Client, plan, state, config *tfModel) diag.Diagnostics {
	passwordChanged := plan.Password.ValueString() != state.Password.ValueString() ||
		plan.PasswordWoVersion.ValueInt64() != state.PasswordWoVersion.ValueInt64()
	replicationChanged := !plan.PgAllowReplication.Equal(state.PgAllowReplication)

	if !passwordChanged && !replicationChanged {
		return nil
	}

	var diags diag.Diagnostics

	if replicationChanged {
		req := &service.ServiceUserCredentialsModifyIn{
			Operation: service.ServiceUserCredentialsModifyOperationTypeSetAccessControl,
			AccessControl: &service.AccessControlIn{
				PgAllowReplication: util.ToPtr(plan.PgAllowReplication.ValueBool()),
			},
		}
		_, err := client.ServiceUserCredentialsModify(ctx, plan.Project.ValueString(), plan.ServiceName.ValueString(), plan.Username.ValueString(), req)
		if err != nil {
			diags.Append(errmsg.FromError("ServiceUserCredentialsModify Error", err))
			return diags
		}
	}

	if passwordChanged {
		diags.Append(resetPassword(ctx, client, plan, config)...)
		if diags.HasError() {
			return diags
		}

		if plan.PasswordWoVersion.ValueInt64() == 0 {
			diags.Append(waitAndCapturePassword(ctx, client, plan)...)
		}
	}

	return diags
}

func resetPassword(ctx context.Context, client avngen.Client, plan, config *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics
	req := &service.ServiceUserCredentialsModifyIn{
		Operation:   service.ServiceUserCredentialsModifyOperationTypeResetCredentials,
		NewPassword: util.NilIfZero(plan.Password.ValueString(), config.PasswordWo.ValueString()),
	}

	_, err := client.ServiceUserCredentialsModify(ctx, plan.Project.ValueString(), plan.ServiceName.ValueString(), plan.Username.ValueString(), req)
	if err != nil {
		diags.Append(errmsg.FromError("ServiceUserCredentialsModify Error", err))
	}
	return diags
}

// waitAndCapturePassword waits for the password to appear in the API and sets it on the plan.
// For custom passwords, it waits for the exact value. For auto-generated, any non-empty value.
// Capturing the password on the plan lets flattenModifier protect it from stale reads during refreshState.
func waitAndCapturePassword(ctx context.Context, client avngen.Client, plan *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics
	expected := plan.Password.ValueString()
	err := retry.Do(
		func() error {
			user, err := client.ServiceUserGet(ctx, plan.Project.ValueString(), plan.ServiceName.ValueString(), plan.Username.ValueString())
			if err != nil {
				// 404 is retryable: user may not have propagated yet after Create
				if avngen.IsNotFound(err) {
					return err
				}
				return retry.Unrecoverable(err)
			}
			if user.Password == "" {
				return fmt.Errorf("password not received from API")
			}
			if expected != "" && user.Password != expected {
				return fmt.Errorf("custom password not yet propagated")
			}
			plan.Password = types.StringValue(user.Password)
			return nil
		},
		retry.Context(ctx),
		retry.Delay(time.Second),
		retry.Attempts(10),
	)
	if err != nil {
		diags.Append(errmsg.FromError("Error waiting for password", err))
	}

	return diags
}

func expandModifier(_ context.Context, _ avngen.Client) util.MapModifier[tfModel] {
	return func(r util.RawMap, plan *tfModel) error {
		// Move pg_allow_replication into access_control for the API request
		if !plan.PgAllowReplication.IsNull() && !plan.PgAllowReplication.IsUnknown() {
			if err := r.Set(plan.PgAllowReplication.ValueBool(), "access_control", "pg_allow_replication"); err != nil {
				return err
			}
		}
		return r.Delete("pg_allow_replication")
	}
}

// flattenModifier adjusts the API response before it is unmarshalled into the state.
// After Create/Update, the adapter's refreshState calls ServiceUserGet which may return
// stale data due to API eventual consistency.
// When the plan already has a known value for a field, we use it instead of the API response to prevent
// inconsistent result after apply errors.
// On import the plan value is null/unknown, so the API response passes through unchanged.
func flattenModifier(_ context.Context, _ avngen.Client) util.MapModifier[tfModel] {
	return func(r util.RawMap, plan *tfModel) error {
		// pg_allow_replication: use plan value if known, otherwise extract from access_control.
		if !plan.PgAllowReplication.IsNull() && !plan.PgAllowReplication.IsUnknown() {
			if err := r.Set(plan.PgAllowReplication.ValueBool(), "pg_allow_replication"); err != nil {
				return err
			}
		} else if val, ok := r.GetBool("access_control", "pg_allow_replication"); ok {
			if err := r.Set(val, "pg_allow_replication"); err != nil {
				return err
			}
		}

		// password: use plan value if known (custom password), otherwise let the API value through.
		if !plan.Password.IsNull() && !plan.Password.IsUnknown() {
			if err := r.Set(plan.Password.ValueString(), "password"); err != nil {
				return err
			}
		}

		// Clear password from state when using write-only password.
		if plan.PasswordWoVersion.ValueInt64() != 0 {
			if err := r.Delete("password"); err != nil {
				return err
			}
		}
		return nil
	}
}
