package vpc

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceAzureOrgVPCPeeringConnection() *schema.Resource {
	return &schema.Resource{
		ReadContext: common.WithGenClientDiag(datasourceAzureOrgVPCPeeringConnectionRead),
		Description: "Gets information about about an Azure VPC peering connection.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenAzureOrgVPCPeeringConnectionSchema,
			"organization_id", "organization_vpc_id", "azure_subscription_id",
			"peer_resource_group", "vnet_name"),
	}
}

func datasourceAzureOrgVPCPeeringConnectionRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) diag.Diagnostics {
	var (
		orgID = d.Get("organization_id").(string)
		vpcID = d.Get("organization_vpc_id").(string)
		subID = d.Get("azure_subscription_id").(string)
		vnet  = d.Get("vnet_name").(string)
		rg    = d.Get("peer_resource_group").(string)
	)

	d.SetId(schemautil.BuildResourceID(orgID, vpcID, subID, vnet, rg))

	return resourceAzureOrgVPCPeeringConnectionRead(ctx, d, client)
}
