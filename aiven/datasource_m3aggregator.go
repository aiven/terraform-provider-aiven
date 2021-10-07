package aiven

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceM3Aggregator() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceServiceRead,
		Description: "The M3 Aggregator data source provides information about the existing Aiven M3 Aggregator.",
		Schema:      resourceSchemaAsDatasourceSchema(aivenM3AggregatorSchema(), "project", "service_name"),
	}
}
