package vpc

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceAWSOrgVPCPeeringConnection() *schema.Resource {
	return &schema.Resource{
		ReadContext: common.WithGenClientDiag(dataSourceAWSOrgVPCPeeringConnectionRead),
		Description: "Gets information about an AWS VPC peering connection.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenAWSOrgVPCPeeringConnectionSchema,
			"organization_id", "organization_vpc_id", "aws_account_id", "aws_vpc_id", "aws_vpc_region"),
	}
}

func dataSourceAWSOrgVPCPeeringConnectionRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) diag.Diagnostics {
	var (
		orgID        = d.Get("organization_id").(string)
		orgVpcID     = d.Get("organization_vpc_id").(string)
		awsAccountID = d.Get("aws_account_id").(string)
		awsVpcID     = d.Get("aws_vpc_id").(string)
		awsRegion    = d.Get("aws_vpc_region").(string)
	)

	d.SetId(schemautil.BuildResourceID(orgID, orgVpcID, awsAccountID, awsVpcID, awsRegion))

	return resourceAWSOrgVPCPeeringConnectionRead(ctx, d, client)
}
