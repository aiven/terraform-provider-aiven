package flink

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceFlink() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceRead,
		Description: "The Flink data source provides information about the existing Aiven Flink service.",
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(aivenFlinkSchema(), "project", "service_name"),
	}
}
