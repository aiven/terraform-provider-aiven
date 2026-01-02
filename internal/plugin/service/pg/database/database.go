package database

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

const defaultLC = "en_US.UTF-8"

func init() {
	ResourceOptions.Create = create
	ResourceOptions.Read = read
	ResourceOptions.Delete = delete
	ResourceOptions.ModifyPlan = modifyPlan
	DataSourceOptions.Read = read
}

func create(ctx context.Context, client avngen.Client, plan *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics

	err := schemautil.CheckDbConflict(ctx, client, plan.Project.ValueString(), plan.ServiceName.ValueString(), plan.DatabaseName.ValueString())
	if err != nil {
		diags.Append(errmsg.FromError("Database conflict check error", err))
		return diags
	}

	diags.Append(createView(ctx, client, plan)...)
	if !diags.HasError() {
		return diags
	}

	// We have already checked for the existence of the database.
	// Getting a conflict here means the client retried the request.
	return errmsg.DropDiagError(diags, avngen.IsAlreadyExists)
}

func read(ctx context.Context, client avngen.Client, state *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics

	err := schemautil.CheckServiceIsPowered(ctx, client, state.Project.ValueString(), state.ServiceName.ValueString())
	if err != nil {
		diags.Append(errmsg.FromError("Service is powered off", err))
		return diags
	}

	diags.Append(readView(ctx, client, state)...)

	if errmsg.HasDiagError(diags, avngen.IsNotFound) {
		schemautil.ForgetDatabase(state.Project.ValueString(), state.ServiceName.ValueString(), state.DatabaseName.ValueString())
	}

	return diags
}

func delete(ctx context.Context, client avngen.Client, state *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(deleteView(ctx, client, state)...)
	if errmsg.HasDiagError(diags, avngen.IsNotFound) {
		// The resource is already gone.
		schemautil.ForgetDatabase(state.Project.ValueString(), state.ServiceName.ValueString(), state.DatabaseName.ValueString())
		return errmsg.DropDiagError(diags, avngen.IsNotFound)
	}
	if diags.HasError() {
		return diags
	}

	// Waits until database is deleted.
	err := schemautil.WaitUntilNotFound(ctx, func() error {
		_, err := findDatabaseByName(ctx, client, state.Project.ValueString(), state.ServiceName.ValueString(), state.DatabaseName.ValueString())
		if err == nil {
			return fmt.Errorf("database %q still exists", state.DatabaseName.ValueString())
		}
		return err
	})
	if err != nil {
		diags.Append(errmsg.FromError("Waiting for database deletion failed", err))
		return diags
	}

	schemautil.ForgetDatabase(state.Project.ValueString(), state.ServiceName.ValueString(), state.DatabaseName.ValueString())
	return diags
}

// expandModifier makes `""` equivalent to "unset" for lc_* fields.
func expandModifier(_ context.Context, _ avngen.Client) util.MapModifier[tfModel] {
	return func(r util.RawMap, plan *tfModel) error {
		if plan.LcCollate.ValueString() == "" {
			// TODO: should we take "lc_collate" key from field tags?
			if err := r.Delete("lc_collate"); err != nil {
				return err
			}
		}

		if plan.LcCtype.ValueString() == "" {
			// TODO: should we take "lc_ctype" key from field tags?
			if err := r.Delete("lc_ctype"); err != nil {
				return err
			}
		}

		return nil
	}
}

// flattenModifier normalizes missing/empty lc_* values to the default to avoid drift.
func flattenModifier(_ context.Context, _ avngen.Client) util.MapModifier[tfModel] {
	return func(r util.RawMap, _ *tfModel) error {
		// TODO: should we take "lc_collate" key from field tags?
		if v, ok := r.GetString("lc_collate"); !ok || v == "" {
			if err := r.Set(defaultLC, "lc_collate"); err != nil {
				return err
			}
		}

		// TODO: should we take "lc_ctype" key from field tags?
		if v, ok := r.GetString("lc_ctype"); !ok || v == "" {
			if err := r.Set(defaultLC, "lc_ctype"); err != nil {
				return err
			}
		}

		return nil
	}
}

func modifyPlan(_ context.Context, _ avngen.Client, plan, state, _ *tfModel) diag.Diagnostics {
	if state == nil || state.ID.IsNull() || state.ID.ValueString() == "" {
		// Resource doesn't exist yet. Don't change create plans.
		return nil
	}

	normalize := func(planV, stateV types.String) (types.String, bool) {
		if planV.IsUnknown() || planV.IsNull() {
			return planV, false
		}

		p := planV.ValueString()
		s := stateV.ValueString()

		switch {
		case p == "":
			// Empty string is treated as "no change".
			return stateV, true
		case p == defaultLC && s == "":
			// Legacy state may store empty lc_* even if DB was created with defaults.
			return types.StringValue(""), true
		default:
			return planV, false
		}
	}

	if v, changed := normalize(plan.LcCollate, state.LcCollate); changed {
		plan.LcCollate = v
	}
	if v, changed := normalize(plan.LcCtype, state.LcCtype); changed {
		plan.LcCtype = v
	}

	return nil
}

func findDatabaseByName(ctx context.Context, client avngen.Client, project, serviceName, dbName string) (*service.DatabaseOut, error) {
	err := schemautil.CheckServiceIsPowered(ctx, client, project, serviceName)
	if err != nil {
		return nil, err
	}

	list, err := client.ServiceDatabaseList(ctx, project, serviceName)
	if err != nil {
		return nil, err
	}

	for _, db := range list {
		if db.DatabaseName == dbName {
			return &db, nil
		}
	}

	return nil, avngen.Error{
		Message:     fmt.Sprintf("`%s` with given `database_name` not found", typeName),
		OperationID: "ServiceDatabaseList",
		Status:      404,
	}
}
