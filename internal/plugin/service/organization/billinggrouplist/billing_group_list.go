package billinggrouplist

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/organizationbilling"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
)

func NewDataSource() datasource.DataSource {
	return adapter.NewDataSource(adapter.DataSourceOptions[*datasourceModel, tfModel]{
		TypeName: typeName,
		Schema:   datasourceSchema,
		Read:     readBillingGroupList,
	})
}

func readBillingGroupList(ctx context.Context, client avngen.Client, state *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics
	rsp, err := client.OrganizationBillingGroupList(ctx, state.OrganizationID.ValueString())
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

func emailsToStr(rsp util.RawMap, in *organizationBillingGroupList) error {
	if len(in.BillingGroups) == 0 {
		return nil
	}

	for i, dto := range in.BillingGroups {
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
		index := fmt.Sprintf("[%d]", i)
		if len(emails) > 0 {
			err := rsp.Set(emails, "billing_groups", index, "billing_emails")
			if err != nil {
				return err
			}
		}
		if len(contactEmails) > 0 {
			err := rsp.Set(contactEmails, "billing_groups", index, "billing_contact_emails")
			if err != nil {
				return err
			}
		}
	}

	return nil
}
