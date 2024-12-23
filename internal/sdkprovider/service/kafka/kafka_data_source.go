package kafka

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceKafka() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceRead,
		Description: "Gets information about an Aiven for Apache Kafka® service.",
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(aivenKafkaSchema(), "project", "service_name"),
	}
}
