package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/avast/retry-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
)

func init() {
	ResourceOptions.Create = createView
	ResourceOptions.Update = updateView
}

// errPasswordNotReceived is returned when the API has not yet delivered the password.
var errPasswordNotReceived = fmt.Errorf("password not received from API")

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

	// Wait until the password is delivered by the API.
	// The adapter's RefreshState already retries 404s, but won't retry on empty password.
	diags.Append(waitForPassword(ctx, client, plan)...)
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

		diags.Append(waitForPassword(ctx, client, plan)...)
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

// waitForPassword retries until the API delivers a non-empty password.
// Skip when using write-only password (password is not stored in state).
func waitForPassword(ctx context.Context, client avngen.Client, plan *tfModel) diag.Diagnostics {
	if plan.PasswordWoVersion.ValueInt64() != 0 {
		return nil
	}

	var diags diag.Diagnostics
	err := retry.Do(
		func() error {
			user, err := client.ServiceUserGet(ctx, plan.Project.ValueString(), plan.ServiceName.ValueString(), plan.Username.ValueString())
			if err != nil {
				return retry.Unrecoverable(err)
			}
			if user.Password == "" {
				return errPasswordNotReceived
			}
			return nil
		},
		retry.Context(ctx),
		retry.Delay(time.Second),
		retry.Attempts(10),
		retry.RetryIf(func(err error) bool {
			return errors.Is(err, errPasswordNotReceived)
		}),
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

func flattenModifier(_ context.Context, _ avngen.Client) util.MapModifier[tfModel] {
	return func(r util.RawMap, plan *tfModel) error {
		// Extract pg_allow_replication from access_control in API response
		if val, ok := r.GetBool("access_control", "pg_allow_replication"); ok {
			if err := r.Set(val, "pg_allow_replication"); err != nil {
				return err
			}
		}

		// Clear password if using write-only password
		if plan.PasswordWoVersion.ValueInt64() != 0 {
			if err := r.Delete("password"); err != nil {
				return err
			}
		}
		return nil
	}
}
