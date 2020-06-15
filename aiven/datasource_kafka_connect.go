package aiven

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func datasourceKafkaConnect() *schema.Resource {
	return &schema.Resource{
		Read:   datasourceServiceRead,
		Schema: resourceSchemaAsDatasourceSchema(aivenKafkaConnectSchema, "project", "service_name"),
	}
}
