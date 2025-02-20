package vpc

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceGCPOrgVPCPeeringConnection() *schema.Resource {
	return &schema.Resource{
		ReadContext: common.WithGenClientDiag(datasourceGCPOrgVPCPeeringConnectionRead),
		Description: "The GCP VPC Peering Connection data source provides information about the existing Aiven VPC Peering Connection.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenGCPOrgVPCPeeringConnectionSchema,
			"organization_id", "organization_vpc_id", "gcp_project_id", "peer_vpc"),
	}
}

func datasourceGCPOrgVPCPeeringConnectionRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) diag.Diagnostics {
	var (
		orgID        = d.Get("organization_id").(string)
		vpcID        = d.Get("organization_vpc_id").(string)
		gcpProjectID = d.Get("gcp_project_id").(string)
		peerVPC      = d.Get("peer_vpc").(string)
	)

	d.SetId(schemautil.BuildResourceID(orgID, vpcID, gcpProjectID, peerVPC))

	return resourceGCPOrgVPCPeeringConnectionRead(ctx, d, client)
}
