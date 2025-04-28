package billinggrouplist

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/organizationbilling"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
)

const resourceName = "aiven_organization_billing_group_list"

func NewOrganizationBillingGroupListDatasource() datasource.DataSource {
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

func (c *view) Read(ctx context.Context, state *dataModel) diag.Diagnostics {
	var diags diag.Diagnostics
	rsp, err := c.client.OrganizationBillingGroupList(ctx, state.OrganizationID.ValueString())
	if err != nil {
		diags.AddError(errmsg.SummaryErrorReadingResource, err.Error())
		return diags
	}

	// The client returns a list, but the schema is an object.
	list := organizationBillingGroupList{BillingGroups: rsp}
	return flattenData(ctx, state, &list, emailsToStr)
}

type organizationBillingGroupList struct {
	BillingGroups []organizationbilling.BillingGroupOut `json:"billing_groups"`
}

func emailsToStr(rsp map[string]any, in *organizationBillingGroupList) error {
	if len(in.BillingGroups) == 0 {
		return nil
	}

	items := rsp["billing_groups"].([]any)
	for i, item := range items {
		dto := in.BillingGroups[i]
		emails := make([]string, 0)

		for _, v := range dto.BillingEmails {
			emails = append(emails, v.Email)
		}

		contactEmails := make([]string, 0)
		for _, v := range dto.BillingContactEmails {
			contactEmails = append(contactEmails, v.Email)
		}

		// It is super important to not set nil slices.
		// They must have zero lengths or shouldn't be set at all.
		// Otherwise, Terraform will try to cast nil to set and fail:
		// > types.SetType[!!! MISSING TYPE!!!] / underlying type: tftypes.Set[tftypes.DynamicPseudoType]
		m := item.(map[string]any)
		if len(emails) > 0 {
			m["billing_emails"] = emails
		}
		if len(contactEmails) > 0 {
			m["billing_contact_emails"] = contactEmails
		}
	}

	return nil
}
