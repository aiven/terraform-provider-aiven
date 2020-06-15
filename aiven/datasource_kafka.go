package aiven

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func datasourceKafka() *schema.Resource {
	return &schema.Resource{
		Read:   datasourceServiceRead,
		Schema: resourceSchemaAsDatasourceSchema(aivenKafkaSchema(), "project", "service_name"),
	}
}
