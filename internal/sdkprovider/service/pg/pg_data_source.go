package pg

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourcePG() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceRead,
		Description: "Gets information about an Aiven for PostgreSQL® service.",
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(aivenPGSchema(), "project", "service_name"),
	}
}
