package user

import (
	"context"
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
	// The adapter's RefreshState retries 404s, but won't retry on empty password.
	// For custom passwords, wait for the exact value to propagate.
	// For auto-generated passwords, wait for any non-empty value.
	if plan.PasswordWoVersion.ValueInt64() == 0 {
		diags.Append(waitForState(ctx, client, plan, passwordCheck(plan.Password.ValueString()))...)
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

		// Wait for the access control change to propagate.
		// The adapter's RefreshState Read may return stale data without this.
		expected := plan.PgAllowReplication.ValueBool()
		diags.Append(waitForState(ctx, client, plan, func(u *service.ServiceUserGetOut) error {
			if u.AccessControl == nil || u.AccessControl.PgAllowReplication == nil || *u.AccessControl.PgAllowReplication != expected {
				return fmt.Errorf("pg_allow_replication not yet propagated")
			}
			return nil
		})...)
		if diags.HasError() {
			return diags
		}
	}

	if passwordChanged {
		diags.Append(resetPassword(ctx, client, plan, config)...)
		if diags.HasError() {
			return diags
		}

		diags.Append(waitForState(ctx, client, plan, passwordCheck(plan.Password.ValueString()))...)
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

// waitForState retries ServiceUserGet until the check function returns nil.
// Handles API eventual consistency where a mutation returns 200 but subsequent reads may return stale data.
func waitForState(ctx context.Context, client avngen.Client, plan *tfModel, check func(*service.ServiceUserGetOut) error) diag.Diagnostics {
	var diags diag.Diagnostics
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
			return check(user)
		},
		retry.Context(ctx),
		retry.Delay(time.Second),
		retry.Attempts(10),
	)
	if err != nil {
		diags.Append(errmsg.FromError("Error waiting for expected state", err))
	}
	return diags
}

// passwordCheck returns a check function for waitForState.
// For custom passwords, it waits until the exact password propagates.
// For auto-generated passwords, it waits until any non-empty password appears.
func passwordCheck(expected string) func(*service.ServiceUserGetOut) error {
	return func(u *service.ServiceUserGetOut) error {
		if u.Password == "" {
			return fmt.Errorf("password not received from API")
		}
		if expected != "" && u.Password != expected {
			return fmt.Errorf("custom password not yet propagated")
		}
		return nil
	}
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
