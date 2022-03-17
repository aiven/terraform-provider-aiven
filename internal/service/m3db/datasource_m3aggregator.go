package m3db

import (
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceM3Aggregator() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceRead,
		Description: "The M3 Aggregator data source provides information about the existing Aiven M3 Aggregator.",
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(aivenM3AggregatorSchema(), "project", "service_name"),
	}
}
