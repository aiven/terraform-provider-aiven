package service_integration

import (
	"context"

	"github.com/aiven/terraform-provider-aiven/internal/meta"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceServiceIntegrationEndpoint() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceServiceIntegrationEndpointRead,
		Description: "The Service Integration Endpoint data source provides information about the existing Aiven Service Integration Endpoint.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenServiceIntegrationEndpointSchema,
			"project", "endpoint_name"),
	}
}

func datasourceServiceIntegrationEndpointRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*meta.Meta).Client

	projectName := d.Get("project").(string)
	endpointName := d.Get("endpoint_name").(string)

	endpoints, err := client.ServiceIntegrationEndpoints.List(projectName)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, endpoint := range endpoints {
		if endpoint.EndpointName == endpointName {
			d.SetId(schemautil.BuildResourceID(projectName, endpoint.EndpointID))
			return resourceServiceIntegrationEndpointRead(ctx, d, m)
		}
	}

	return diag.Errorf("endpoint \"%s\" not found", endpointName)
}
