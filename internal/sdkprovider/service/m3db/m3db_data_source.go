package m3db

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceM3DB() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceRead,
		Description: "Gets information about an Aiven for M3DB service.",
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(aivenM3DBSchema(), "project", "service_name"),
	}
}
