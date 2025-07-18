package pg

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourcePGDatabase() *schema.Resource {
	return &schema.Resource{
		ReadContext: common.WithGenClient(datasourceDatabaseRead),
		Description: "Gets information about a database in an Aiven for PostgreSQL® service.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenPGDatabaseSchema,
			"project", "service_name", "database_name"),
	}
}

func datasourceDatabaseRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	databaseName := d.Get("database_name").(string)

	d.SetId(schemautil.BuildResourceID(projectName, serviceName, databaseName))
	return resourcePGDatabaseRead(ctx, d, client)
}
