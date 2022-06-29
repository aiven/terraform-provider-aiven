package clickhouse

import (
	"context"

	"github.com/aiven/terraform-provider-aiven/internal/meta"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceClickhouseDatabase() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceClickhouseDatabaseRead,
		Description: "The Clickhouse database data source provides information about the existing Aiven Clickhouse Database.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenClickhouseDatabaseSchema,
			"project", "service_name", "name"),
	}
}

func datasourceClickhouseDatabaseRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*meta.Meta).Client

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	databaseName := d.Get("name").(string)

	r, err := client.ClickhouseDatabase.List(projectName, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, db := range r.Databases {
		if db.Name == databaseName {
			d.SetId(schemautil.BuildResourceID(projectName, serviceName, databaseName))
			return resourceClickhouseDatabaseRead(ctx, d, m)
		}
	}

	return diag.Errorf("clickhouse database %s/%s/%s not found",
		projectName, serviceName, databaseName)
}
