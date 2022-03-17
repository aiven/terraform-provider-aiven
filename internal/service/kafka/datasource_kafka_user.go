package kafka

import (
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceKafkaUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceUserRead,
		Description: "The Kafka User data source provides information about the existing Aiven Kafka User.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenKafkaUserSchema,
			"project", "service_name", "username"),
	}
}
