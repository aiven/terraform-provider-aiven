package m3db

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceM3DB() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceRead,
		Description: "The M3 DB data source provides information about the existing Aiven M3 services.",
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(aivenM3DBSchema(), "project", "service_name"),
	}
}
