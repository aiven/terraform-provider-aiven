package vpc

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DataSourceOrganizationVPC() *schema.Resource {
	return &schema.Resource{
		Description: "Gets information about an existing VPC in an Aiven organization.",
		ReadContext: common.WithGenClient(datasourceOrganizationVPCRead),
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenOrganizationVPCSchema,
			"organization_id", "organization_vpc_id"),
	}
}

func datasourceOrganizationVPCRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	var (
		orgID    = d.Get("organization_id").(string)
		orgVpcID = d.Get("organization_vpc_id").(string)
	)

	d.SetId(schemautil.BuildResourceID(orgID, orgVpcID))

	return resourceOrganizationVPCRead(ctx, d, client)
}
