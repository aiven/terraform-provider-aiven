// Package opensearch implements the Aiven OpenSearch service.
package opensearch

import (
	"context"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

// DatasourceOpenSearchSecurityPluginConfig defines the OpenSearch Security Plugin Config data source.
func DatasourceOpenSearchSecurityPluginConfig() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceOpenSearchSecurityPluginConfigRead,
		Description: "The OpenSearch Security Plugin Config data source provides information about an existing Aiven" +
			" OpenSearch Security Plugin Config.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(
			aivenOpenSearchSecurityPluginConfigSchema, "project", "service_name",
		),
	}
}

// datasourceOpenSearchSecurityPluginConfigRead reads the configuration of an existing OpenSearch Security Plugin
// Config.
func datasourceOpenSearchSecurityPluginConfigRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)

	serviceName := d.Get("service_name").(string)

	if _, err := client.OpenSearchSecurityPluginHandler.Get(projectName, serviceName); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(projectName, serviceName))

	return resourceOpenSearchSecurityPluginConfigRead(ctx, d, m)
}
