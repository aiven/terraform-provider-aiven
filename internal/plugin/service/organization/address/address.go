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

const resourceName = "aiven_organization_address"

func NewOrganizationAddressResource() resource.Resource {
	return adapter.NewResource(
		resourceName,
		new(view),
		resourceSchema,
		func() adapter.DataModel[dataModel] {
			return new(resourceDataModel)
		},
		idFields(),
	)
}

func NewOrganizationAddressDatasource() datasource.DataSource {
	return adapter.NewDatasource(
		resourceName,
		new(view),
		datasourceSchema,
		func() adapter.DataModel[dataModel] {
			return new(datasourceDataModel)
		},
	)
}

type view struct {
	client avngen.Client
}

func (c *view) Configure(client avngen.Client) {
	c.client = client
}

func (c *view) Create(ctx context.Context, plan *dataModel) diag.Diagnostics {
	var req organization.OrganizationAddressCreateIn
	diags := expandData(ctx, plan, nil, &req)
	if diags.HasError() {
		return diags
	}

	rsp, err := c.client.OrganizationAddressCreate(ctx, plan.OrganizationID.ValueString(), &req)
	if err != nil {
		diags.AddError(errmsg.SummaryErrorCreatingResource, err.Error())
		return diags
	}

	// Sets ID fields to Read() the resource
	plan.SetID(rsp.OrganizationId, rsp.AddressId)
	return c.Read(ctx, plan)
}

func (c *view) Update(ctx context.Context, plan, state *dataModel) diag.Diagnostics {
	var req organization.OrganizationAddressUpdateIn
	diags := expandData(ctx, plan, state, &req)
	if diags.HasError() {
		return diags
	}

	rsp, err := c.client.OrganizationAddressUpdate(ctx, state.OrganizationID.ValueString(), state.AddressID.ValueString(), &req)
	if err != nil {
		diags.AddError(errmsg.SummaryErrorUpdatingResource, err.Error())
		return diags
	}

	// Sets ID fields to Read() the resource
	plan.SetID(rsp.OrganizationId, rsp.AddressId)
	return c.Read(ctx, plan)
}

func (c *view) Read(ctx context.Context, state *dataModel) diag.Diagnostics {
	var diags diag.Diagnostics
	rsp, err := c.client.OrganizationAddressGet(ctx, state.OrganizationID.ValueString(), state.AddressID.ValueString())
	if err != nil {
		diags.AddError(errmsg.SummaryErrorReadingResource, err.Error())
		return diags
	}

	return flattenData(ctx, state, rsp)
}

func (c *view) Delete(ctx context.Context, state *dataModel) diag.Diagnostics {
	var diags diag.Diagnostics
	err := c.client.OrganizationAddressDelete(ctx, state.OrganizationID.ValueString(), state.AddressID.ValueString())
	if err != nil {
		diags.AddError(errmsg.SummaryErrorDeletingResource, err.Error())
		return diags
	}

	return nil
}
