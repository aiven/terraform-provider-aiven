package opensearch

import (
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceOpensearch() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceRead,
		Description: "The Opensearch data source provides information about the existing Aiven Opensearch service.",
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(opensearchSchema(), "project", "service_name"),
	}
}
