package kafka

import (
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceKafka() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceRead,
		Description: "The Kafka data source provides information about the existing Aiven Kafka services.",
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(aivenKafkaSchema(), "project", "service_name"),
	}
}
