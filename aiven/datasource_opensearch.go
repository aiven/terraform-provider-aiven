package aiven

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceOpensearch() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceServiceRead,
		Description: "The Opensearch data source provides information about the existing Aiven Opensearch service.",
		Schema:      resourceSchemaAsDatasourceSchema(opensearchSchema(), "project", "service_name"),
	}
}
