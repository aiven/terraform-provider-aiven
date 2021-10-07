package aiven

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceElasticsearch() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceServiceRead,
		Description: "The Elasticsearch data source provides information about the existing Aiven Elasticsearch service.",
		Schema:      resourceSchemaAsDatasourceSchema(elasticsearchSchema(), "project", "service_name"),
	}
}
