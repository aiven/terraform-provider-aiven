package kafka

import (
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceKafkaConnect() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceRead,
		Description: "The Kafka Connect data source provides information about the existing Aiven Kafka Connect service.",
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(aivenKafkaConnectSchema(), "project", "service_name"),
	}
}
