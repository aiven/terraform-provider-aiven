package aiven

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func datasourceTransitGatewayVPCAttachment() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVPCPeeringConnectionRead,
		Schema: resourceSchemaAsDatasourceSchema(aivenTransitGatewayVPCAttachmentSchema,
			"vpc_id", "peer_cloud_account", "peer_vpc"),
	}
}
