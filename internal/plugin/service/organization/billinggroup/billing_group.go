package billinggroup

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/organizationbilling"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
)

const resourceName = "aiven_organization_billing_group"

func NewOrganizationBillingGroupResource() resource.Resource {
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

func NewOrganizationBillingGroupDatasource() datasource.DataSource {
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
	var req organizationbilling.OrganizationBillingGroupCreateIn
	diags := expandData(ctx, plan, &req, emailsToMap)
	if diags.HasError() {
		return diags
	}

	rsp, err := c.client.OrganizationBillingGroupCreate(ctx, plan.OrganizationID.ValueString(), &req)
	if err != nil {
		diags.AddError(errmsg.SummaryErrorCreatingResource, err.Error())
		return diags
	}

	// Read() reads the remote state using these two IDs.
	plan.OrganizationID = types.StringValue(rsp.OrganizationId)
	plan.BillingGroupID = types.StringValue(rsp.BillingGroupId)
	return c.Read(ctx, plan)
}

func (c *view) Update(ctx context.Context, plan, state *dataModel) diag.Diagnostics {
	var req organizationbilling.OrganizationBillingGroupUpdateIn
	diags := expandData(ctx, plan, &req, emailsToMap)
	if diags.HasError() {
		return diags
	}

	_, err := c.client.OrganizationBillingGroupUpdate(ctx, state.OrganizationID.ValueString(), state.BillingGroupID.ValueString(), &req)
	if err != nil {
		diags.AddError(errmsg.SummaryErrorUpdatingResource, err.Error())
		return diags
	}

	return c.Read(ctx, plan)
}

func (c *view) Delete(ctx context.Context, state *dataModel) diag.Diagnostics {
	var diags diag.Diagnostics
	err := c.client.OrganizationBillingGroupDelete(ctx, state.OrganizationID.ValueString(), state.BillingGroupID.ValueString())
	if err != nil {
		diags.AddError(errmsg.SummaryErrorDeletingResource, err.Error())
		return diags
	}

	return nil
}

func (c *view) Read(ctx context.Context, state *dataModel) diag.Diagnostics {
	var diags diag.Diagnostics
	rsp, err := c.client.OrganizationBillingGroupGet(ctx, state.OrganizationID.ValueString(), state.BillingGroupID.ValueString())
	if err != nil {
		diags.AddError(errmsg.SummaryErrorReadingResource, err.Error())
		return diags
	}

	return flattenData(ctx, state, rsp, emailsToStr)
}

func emailsToMap(req map[string]any, in *dtoModel) error {
	if in.BillingEmails != nil {
		emails := make([]map[string]any, 0)
		for _, v := range *in.BillingEmails {
			emails = append(emails, map[string]any{"email": v})
		}
		req["billing_emails"] = emails
	}

	if in.BillingContactEmails != nil {
		emails := make([]map[string]any, 0)
		for _, v := range *in.BillingContactEmails {
			emails = append(emails, map[string]any{"email": v})
		}
		req["billing_contact_emails"] = emails
	}
	return nil
}

func emailsToStr(rsp map[string]any, in *organizationbilling.OrganizationBillingGroupGetOut) error {
	emails := make([]string, 0)
	for _, v := range in.BillingEmails {
		emails = append(emails, v.Email)
	}

	contactEmails := make([]string, 0)
	for _, v := range in.BillingContactEmails {
		contactEmails = append(contactEmails, v.Email)
	}

	// It is super important to not set nil slices.
	// They must have zero lengths or shouldn't be set at all.
	// Otherwise, Terraform will try to cast nil to set and fail:
	// > types.SetType[!!! MISSING TYPE!!!] / underlying type: tftypes.Set[tftypes.DynamicPseudoType]
	if len(emails) > 0 {
		rsp["billing_emails"] = emails
	}
	if len(contactEmails) > 0 {
		rsp["billing_contact_emails"] = contactEmails
	}
	return nil
}
