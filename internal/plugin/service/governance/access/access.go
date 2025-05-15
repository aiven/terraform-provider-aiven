package access

import (
	"context"

	"github.com/aiven/go-client-codegen/handler/organizationgovernance"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
)

func NewResource() resource.Resource {
	return adapter.NewResource(aivenName, new(view), newResourceSchema, newResourceModel, composeID())
}

type view struct{ adapter.View }

func (vw *view) Create(ctx context.Context, plan *tfModel) diag.Diagnostics {
	var req organizationgovernance.OrganizationGovernanceAccessCreateIn
	diags := expandData(ctx, plan, nil, &req)
	if diags.HasError() {
		return diags
	}

	rsp, err := vw.Client.OrganizationGovernanceAccessCreate(ctx, plan.OrganizationID.ValueString(), &req)
	if err != nil {
		diags.AddError(errmsg.SummaryErrorCreatingResource, err.Error())
		return diags
	}

	// Sets ID fields to Read() the resource
	plan.SetID(plan.OrganizationID.ValueString(), rsp.AccessId)
	return vw.Read(ctx, plan)
}

func (vw *view) Update(_ context.Context, _, _ *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.AddError(errmsg.SummaryErrorUpdatingResource, "Update is not supported for this resource")
	return diags
}

func (vw *view) Read(ctx context.Context, state *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics
	rsp, err := vw.Client.OrganizationGovernanceAccessGet(ctx, state.OrganizationID.ValueString(), state.SusbcriptionID.ValueString())
	if err != nil {
		diags.AddError(errmsg.SummaryErrorReadingResource, err.Error())
		return diags
	}

	return flattenData(ctx, state, rsp)
}

func (vw *view) Delete(ctx context.Context, state *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics
	_, err := vw.Client.OrganizationGovernanceAccessDelete(ctx, state.OrganizationID.ValueString(), state.SusbcriptionID.ValueString())
	if err != nil {
		diags.AddError(errmsg.SummaryErrorDeletingResource, err.Error())
		return diags
	}

	return nil
}
