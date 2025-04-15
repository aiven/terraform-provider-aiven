package opensearch

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceOpenSearch() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceRead,
		Description: "Gets information about an Aiven for OpenSearch® service.",
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(opensearchSchema(), "project", "service_name"),
	}
}
