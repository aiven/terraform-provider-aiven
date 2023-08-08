package flink

import (
	"context"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceFlinkApplication() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceFlinkApplicationRead,
		Description: "The Flink Application data source provides information about the existing Aiven Flink Application.",
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(aivenFlinkApplicationSchema, "project", "service_name", "name"),
	}
}

func datasourceFlinkApplicationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	applicationName := d.Get("name").(string)

	a, err := client.FlinkApplications.List(project, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, app := range a.Applications {
		if app.Name == applicationName {
			d.SetId(schemautil.BuildResourceID(project, serviceName, app.ID))
			return resourceFlinkApplicationRead(ctx, d, m)
		}
	}

	return diag.Errorf("flink application %s not found", applicationName)
}
