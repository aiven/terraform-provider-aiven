package project

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.org/x/exp/maps"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

func DatasourceBillingGroup() *schema.Resource {
	s := schemautil.ResourceSchemaAsDatasourceSchema(aivenBillingGroupSchema)
	maps.Copy(s, map[string]*schema.Schema{
		"billing_group_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: userconfig.Desc("The ID of the billing group.").Referenced().Build(),
		},
	})

	return &schema.Resource{
		ReadContext: datasourceBillingGroupRead,
		Description: "Gets information about a billing group.",
		Schema:      s,
	}
}

func datasourceBillingGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.SetId(d.Get("billing_group_id").(string))
	return resourceBillingGroupRead(ctx, d, m)
}
