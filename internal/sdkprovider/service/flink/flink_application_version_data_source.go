package flink

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceFlinkApplicationVersion() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceFlinkApplicationVersionRead,
		Description: "Gets information about an Aiven for Apache Flink® application version.",
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(aivenFlinkApplicationVersionSchema, "project", "service_name", "application_id", "application_version_id"),
	}
}

func datasourceFlinkApplicationVersionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	appID := d.Get("application_id").(string)
	versionID := d.Get("application_version_id").(string)

	_, err := client.FlinkApplicationVersions.Get(ctx, project, serviceName, appID, versionID)
	if err != nil {
		if aiven.IsNotFound(err) {
			return diag.Errorf("flink application version %s not found", versionID)
		}
		return diag.Errorf("error getting flink application version %s: %s", versionID, err)
	}

	d.SetId(schemautil.BuildResourceID(project, serviceName, appID, versionID))

	return resourceFlinkApplicationVersionRead(ctx, d, m)
}
