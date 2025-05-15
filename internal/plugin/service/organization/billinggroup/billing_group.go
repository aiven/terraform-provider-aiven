package billinggroup

import (
	"context"

	"github.com/aiven/go-client-codegen/handler/organizationbilling"
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
	var req organizationbilling.OrganizationBillingGroupCreateIn
	diags := expandData(ctx, plan, nil, &req, emailsToMap)
	if diags.HasError() {
		return diags
	}

	rsp, err := vw.Client.OrganizationBillingGroupCreate(ctx, plan.OrganizationID.ValueString(), &req)
	if err != nil {
		diags.AddError(errmsg.SummaryErrorCreatingResource, err.Error())
		return diags
	}

	// Sets ID fields to Read() the resource
	plan.SetID(rsp.OrganizationId, rsp.BillingGroupId)
	return vw.Read(ctx, plan)
}

func (vw *view) Update(ctx context.Context, plan, state *tfModel) diag.Diagnostics {
	var req organizationbilling.OrganizationBillingGroupUpdateIn
	diags := expandData(ctx, plan, state, &req, emailsToMap)
	if diags.HasError() {
		return diags
	}

	rsp, err := vw.Client.OrganizationBillingGroupUpdate(ctx, state.OrganizationID.ValueString(), state.BillingGroupID.ValueString(), &req)
	if err != nil {
		diags.AddError(errmsg.SummaryErrorUpdatingResource, err.Error())
		return diags
	}

	// Sets ID fields to Read() the resource
	plan.SetID(rsp.OrganizationId, rsp.BillingGroupId)
	return vw.Read(ctx, plan)
}

func (vw *view) Delete(ctx context.Context, state *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics
	err := vw.Client.OrganizationBillingGroupDelete(ctx, state.OrganizationID.ValueString(), state.BillingGroupID.ValueString())
	if err != nil {
		diags.AddError(errmsg.SummaryErrorDeletingResource, err.Error())
		return diags
	}

	return nil
}

func (vw *view) Read(ctx context.Context, state *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics
	rsp, err := vw.Client.OrganizationBillingGroupGet(ctx, state.OrganizationID.ValueString(), state.BillingGroupID.ValueString())
	if err != nil {
		diags.AddError(errmsg.SummaryErrorReadingResource, err.Error())
		return diags
	}

	return flattenData(ctx, state, rsp, emailsToStr)
}

func emailsToMap(req map[string]any, in *apiModel) error {
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
