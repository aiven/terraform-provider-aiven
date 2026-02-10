package database

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func init() {
	ResourceOptions.Create = create
	ResourceOptions.Read = read
	ResourceOptions.Delete = delete
	DataSourceOptions.Read = read
}

func create(ctx context.Context, client avngen.Client, plan, config *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics

	err := schemautil.CheckDbConflict(ctx, client, plan.Project.ValueString(), plan.ServiceName.ValueString(), plan.DatabaseName.ValueString())
	if err != nil {
		diags.Append(errmsg.FromError("Database conflict check error", err))
		return diags
	}

	diags.Append(createView(ctx, client, plan, config)...)
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
