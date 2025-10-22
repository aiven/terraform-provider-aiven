package address

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/organization"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
)

func NewResource() resource.Resource {
	return adapter.NewResource(adapter.ResourceOptions[*resourceModel, tfModel]{
		TypeName: aivenName,
		IDFields: idFields(),
		Schema:   newResourceSchema,
		Read:     readAddress,
		Create:   createAddress,
		Update:   updateAddress,
		Delete:   deleteAddress,
	})
}

func NewDatasource() datasource.DataSource {
	return adapter.NewDatasource(adapter.DatasourceOptions[*datasourceModel, tfModel]{
		TypeName: aivenName,
		Schema:   newDatasourceSchema,
		Read:     readAddress,
	})
}

func createAddress(ctx context.Context, client avngen.Client, plan *tfModel) diag.Diagnostics {
	var req organization.OrganizationAddressCreateIn
	diags := expandData(ctx, plan, nil, &req)
	if diags.HasError() {
		return diags
	}

	rsp, err := client.OrganizationAddressCreate(ctx, plan.OrganizationID.ValueString(), &req)
	if err != nil {
		diags.AddError(errmsg.SummaryErrorCreatingResource, err.Error())
		return diags
	}

	// Sets ID fields to Read() the resource
	plan.SetID(rsp.OrganizationId, rsp.AddressId)
	return readAddress(ctx, client, plan)
}

func updateAddress(ctx context.Context, client avngen.Client, plan, state, _ *tfModel) diag.Diagnostics {
	var req organization.OrganizationAddressUpdateIn
	diags := expandData(ctx, plan, state, &req)
	if diags.HasError() {
		return diags
	}

	rsp, err := client.OrganizationAddressUpdate(ctx, state.OrganizationID.ValueString(), state.AddressID.ValueString(), &req)
	if err != nil {
		diags.AddError(errmsg.SummaryErrorUpdatingResource, err.Error())
		return diags
	}

	// Sets ID fields to Read() the resource
	plan.SetID(rsp.OrganizationId, rsp.AddressId)
	return readAddress(ctx, client, plan)
}

func readAddress(ctx context.Context, client avngen.Client, state *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics
	rsp, err := client.OrganizationAddressGet(ctx, state.OrganizationID.ValueString(), state.AddressID.ValueString())
	if err != nil {
		diags.AddError(errmsg.SummaryErrorReadingResource, err.Error())
		return diags
	}

	return flattenData(ctx, state, rsp)
}

func deleteAddress(ctx context.Context, client avngen.Client, state *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics
	err := client.OrganizationAddressDelete(ctx, state.OrganizationID.ValueString(), state.AddressID.ValueString())
	if err != nil {
		diags.AddError(errmsg.SummaryErrorDeletingResource, err.Error())
		return diags
	}

	return nil
}
