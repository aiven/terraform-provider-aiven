package aiven

import "github.com/hashicorp/terraform/helper/schema"

func datasourceTransitGatewayVPCAttachment() *schema.Resource {
	return &schema.Resource{
		Read: datasourceVPCPeeringConnectionRead,
		Schema: resourceSchemaAsDatasourceSchema(aivenTransitGatewayVPCAttachmentSchema,
			"vpc_id", "peer_cloud_account", "peer_vpc"),
	}
}
