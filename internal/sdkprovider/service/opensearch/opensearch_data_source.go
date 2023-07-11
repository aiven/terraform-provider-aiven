package opensearch

import (
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceOpenSearch() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceRead,
		Description: "The OpenSearch data source provides information about the existing Aiven OpenSearch service.",
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(opensearchSchema(), "project", "service_name"),
	}
}
