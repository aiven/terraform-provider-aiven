package access

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/organizationgovernance"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
)

func NewResource() resource.Resource {
	return adapter.NewResource(adapter.ResourceOptions[*resourceModel, tfModel]{
		TypeName: aivenName,
		IDFields: composeID(),
		Schema:   newResourceSchema,
		Read:     readAccess,
		Create:   createAccess,
		Delete:   deleteAccess,
	})
}

func createAccess(ctx context.Context, client avngen.Client, plan *tfModel) diag.Diagnostics {
	var req organizationgovernance.OrganizationGovernanceAccessCreateIn
	diags := expandData(ctx, plan, nil, &req)
	if diags.HasError() {
		return diags
	}

	rsp, err := client.OrganizationGovernanceAccessCreate(ctx, plan.OrganizationID.ValueString(), &req)
	if err != nil {
		diags.AddError(errmsg.SummaryErrorCreatingResource, err.Error())
		return diags
	}

	// Sets ID fields to Read() the resource
	plan.SetID(plan.OrganizationID.ValueString(), rsp.AccessId)
	return readAccess(ctx, client, plan)
}

func readAccess(ctx context.Context, client avngen.Client, state *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics
	rsp, err := client.OrganizationGovernanceAccessGet(ctx, state.OrganizationID.ValueString(), state.SusbcriptionID.ValueString())
	if err != nil {
		diags.AddError(errmsg.SummaryErrorReadingResource, err.Error())
		return diags
	}

	return flattenData(ctx, state, rsp)
}

func deleteAccess(ctx context.Context, client avngen.Client, state *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics
	_, err := client.OrganizationGovernanceAccessDelete(ctx, state.OrganizationID.ValueString(), state.SusbcriptionID.ValueString())
	if err != nil {
		diags.AddError(errmsg.SummaryErrorDeletingResource, err.Error())
		return diags
	}

	return nil
}
