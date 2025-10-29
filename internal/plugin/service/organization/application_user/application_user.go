package applicationuser

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/applicationuser"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
)

func NewResource() resource.Resource {
	return adapter.NewResource(adapter.ResourceOptions[*resourceModel, tfModel]{
		TypeName: typeName,
		IDFields: idFields(),
		Schema:   resourceSchema,
		Read:     readApplicationUser,
		Create:   createApplicationUser,
		Update:   updateApplicationUser,
		Delete:   deleteApplicationUser,
	})
}

func NewDataSource() datasource.DataSource {
	return adapter.NewDataSource(adapter.DataSourceOptions[*datasourceModel, tfModel]{
		TypeName: typeName,
		Schema:   datasourceSchema,
		Read:     readApplicationUser,
	})
}

func createApplicationUser(ctx context.Context, client avngen.Client, plan *tfModel) diag.Diagnostics {
	var req applicationuser.ApplicationUserCreateIn
	diags := expandData(ctx, plan, nil, &req)
	if diags.HasError() {
		return diags
	}

	rsp, err := client.ApplicationUserCreate(ctx, plan.OrganizationID.ValueString(), &req)
	if err != nil {
		diags.AddError(errmsg.SummaryErrorCreatingResource, err.Error())
		return diags
	}

	// Sets ID fields to Read() the resource
	plan.SetID(plan.OrganizationID.ValueString(), rsp.UserId)
	return readApplicationUser(ctx, client, plan)
}

func updateApplicationUser(ctx context.Context, client avngen.Client, plan, state, _ *tfModel) diag.Diagnostics {
	var req applicationuser.ApplicationUserUpdateIn
	diags := expandData(ctx, plan, state, &req)
	if diags.HasError() {
		return diags
	}

	rsp, err := client.ApplicationUserUpdate(ctx, state.OrganizationID.ValueString(), state.UserID.ValueString(), &req)
	if err != nil {
		diags.AddError(errmsg.SummaryErrorUpdatingResource, err.Error())
		return diags
	}

	// Sets ID fields to Read() the resource
	plan.SetID(plan.OrganizationID.ValueString(), rsp.UserId)
	return readApplicationUser(ctx, client, plan)
}

func readApplicationUser(ctx context.Context, client avngen.Client, state *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics
	rsp, err := client.ApplicationUserGet(ctx, state.OrganizationID.ValueString(), state.UserID.ValueString())
	if err != nil {
		diags.AddError(errmsg.SummaryErrorReadingResource, err.Error())
		return diags
	}

	return flattenData(ctx, state, rsp)
}

func deleteApplicationUser(ctx context.Context, client avngen.Client, state *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics
	err := client.ApplicationUserDelete(ctx, state.OrganizationID.ValueString(), state.UserID.ValueString())
	if err != nil {
		diags.AddError(errmsg.SummaryErrorDeletingResource, err.Error())
		return diags
	}

	return nil
}
