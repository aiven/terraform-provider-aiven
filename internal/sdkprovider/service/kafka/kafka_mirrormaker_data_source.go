package kafka

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceKafkaMirrormaker() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceRead,
		Description: "The Kafka MirrorMaker data source provides information about the existing Aiven Kafka MirrorMaker 2 service.",
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(aivenKafkaMirrormakerSchema(), "project", "service_name"),
	}
}
