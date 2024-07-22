package kafka

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceKafkaUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceUserRead,
		Description: "Gets information about an Aiven for Apache Kafka® service user.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenKafkaUserSchema,
			"project", "service_name", "username"),
	}
}
