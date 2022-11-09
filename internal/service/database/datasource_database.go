package database

import (
	"context"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// DatasourceDatabase
// Deprecated
//
//goland:noinspection GoDeprecation
func DatasourceDatabase() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceDatabaseRead,
		Description: `The Database data source provides information about the existing Aiven Database.

~>**Deprecated** The Database data source is deprecated, please use service-specific data sources instead, for example: ` + "`aiven_pg_database`, `aiven_mysql_database` etc.",
		DeprecationMessage: "`aiven_database` data source is deprecated. Please use service-specific data sources instead of this one, for example: `aiven_pg_database`, `aiven_mysql_database` etc.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenDatabaseSchema,
			"project", "service_name", "database_name"),
	}
}

func datasourceDatabaseRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	databaseName := d.Get("database_name").(string)

	databases, err := client.Databases.List(projectName, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, db := range databases {
		if db.DatabaseName == databaseName {
			d.SetId(schemautil.BuildResourceID(projectName, serviceName, databaseName))
			return resourceDatabaseRead(ctx, d, m)
		}
	}

	return diag.Errorf("database %s/%s/%s not found",
		projectName, serviceName, databaseName)
}
