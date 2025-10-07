package applicationuser

import (
	"context"

	"github.com/aiven/go-client-codegen/handler/applicationuser"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
)

func NewResource() resource.Resource {
	return adapter.NewResource(aivenName, new(view), newResourceSchema, newResourceModel, composeID())
}

func NewDatasource() datasource.DataSource {
	return adapter.NewDatasource(aivenName, new(view), newDatasourceSchema, newDatasourceModel)
}

type view struct{ adapter.View }

func (vw *view) Create(ctx context.Context, plan *tfModel) diag.Diagnostics {
	var req applicationuser.ApplicationUserCreateIn
	diags := expandData(ctx, plan, nil, &req)
	if diags.HasError() {
		return diags
	}

	rsp, err := vw.Client.ApplicationUserCreate(ctx, plan.OrganizationID.ValueString(), &req)
	if err != nil {
		diags.AddError(errmsg.SummaryErrorCreatingResource, err.Error())
		return diags
	}

	// Sets ID fields to Read() the resource
	plan.SetID(plan.OrganizationID.ValueString(), rsp.UserId)
	return vw.Read(ctx, plan)
}

func (vw *view) Update(ctx context.Context, plan, state, _ *tfModel) diag.Diagnostics {
	var req applicationuser.ApplicationUserUpdateIn
	diags := expandData(ctx, plan, state, &req)
	if diags.HasError() {
		return diags
	}

	rsp, err := vw.Client.ApplicationUserUpdate(ctx, state.OrganizationID.ValueString(), state.UserID.ValueString(), &req)
	if err != nil {
		diags.AddError(errmsg.SummaryErrorUpdatingResource, err.Error())
		return diags
	}

	// Sets ID fields to Read() the resource
	plan.SetID(plan.OrganizationID.ValueString(), rsp.UserId)
	return vw.Read(ctx, plan)
}

func (vw *view) Read(ctx context.Context, state *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics
	rsp, err := vw.Client.ApplicationUserGet(ctx, state.OrganizationID.ValueString(), state.UserID.ValueString())
	if err != nil {
		diags.AddError(errmsg.SummaryErrorReadingResource, err.Error())
		return diags
	}

	return flattenData(ctx, state, rsp)
}

func (vw *view) Delete(ctx context.Context, state *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics
	err := vw.Client.ApplicationUserDelete(ctx, state.OrganizationID.ValueString(), state.UserID.ValueString())
	if err != nil {
		diags.AddError(errmsg.SummaryErrorDeletingResource, err.Error())
		return diags
	}

	return nil
}
