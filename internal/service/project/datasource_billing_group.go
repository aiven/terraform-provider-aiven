package project

import (
	"context"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceBillingGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceBillingGroupRead,
		Description: "The Billing Group data source provides information about the existing Aiven Account.",
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(aivenBillingGroupSchema, "name"),
	}
}

func datasourceBillingGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	name := d.Get("name").(string)

	list, err := client.BillingGroup.ListAll()
	if err != nil {
		return diag.FromErr(err)
	}

	for _, bg := range list {
		if bg.BillingGroupName == name {
			d.SetId(bg.Id)
			return resourceBillingGroupRead(ctx, d, m)
		}
	}

	return diag.Errorf("billing group %s not found", name)
}
