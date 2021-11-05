// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func datasourceTransitGatewayVPCAttachment() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVPCPeeringConnectionRead,
		Description: "The Transit Gateway VPC Attachment resource allows the creation and management Transit Gateway VPC Attachment VPC peering connection between Aiven and AWS.",
		Schema: resourceSchemaAsDatasourceSchema(aivenTransitGatewayVPCAttachmentSchema,
			"vpc_id", "peer_cloud_account", "peer_vpc"),
	}
}
