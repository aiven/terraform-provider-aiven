package vpc

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceTransitGatewayVPCAttachment() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVPCPeeringConnectionRead,
		Description: "Gets information about an Aiven VPC attachment to external cloud provider networks.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenTransitGatewayVPCAttachmentSchema,
			"vpc_id", "peer_cloud_account", "peer_vpc"),
	}
}
