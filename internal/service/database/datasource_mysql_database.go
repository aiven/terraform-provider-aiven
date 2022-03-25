package database

import (
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceMySQLDatabase() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceDatabaseRead,
		Description: "The Database data source provides information about the existing Aiven Database.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenDatabaseMySQLSchema,
			"project", "service_name", "database_name"),
	}
}
