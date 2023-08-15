package clickhouse

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceClickhouse() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceRead,
		Description: "The Clickhouse data source provides information about the existing Aiven Clickhouse service.",
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(clickhouseSchema(), "project", "service_name"),
	}
}
