package aiven

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceM3Aggregator() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceServiceRead,
		Schema:      resourceSchemaAsDatasourceSchema(aivenM3AggregatorSchema(), "project", "service_name"),
	}
}
