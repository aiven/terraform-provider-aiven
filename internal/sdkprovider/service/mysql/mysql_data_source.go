package mysql

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceMySQL() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceRead,
		Description: "Gets information about an Aiven for MySQL® service.",
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(aivenMySQLSchema(), "project", "service_name"),
	}
}
