package billinggrouplist

import (
	"context"

	"github.com/aiven/go-client-codegen/handler/organizationbilling"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
)

func NewDatasource() datasource.DataSource {
	return adapter.NewDatasource(aivenName, new(view), newDatasourceSchema, newDatasourceModel)
}

type view struct{ adapter.View }

func (vw *view) Read(ctx context.Context, state *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics
	rsp, err := vw.Client.OrganizationBillingGroupList(ctx, state.OrganizationID.ValueString())
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
