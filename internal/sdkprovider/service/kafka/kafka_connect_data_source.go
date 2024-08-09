package kafka

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceKafkaConnect() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceRead,
		Description: "Gets information about an Aiven for Apache Kafka® Connect service.",
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(aivenKafkaConnectSchema(), "project", "service_name"),
	}
}
