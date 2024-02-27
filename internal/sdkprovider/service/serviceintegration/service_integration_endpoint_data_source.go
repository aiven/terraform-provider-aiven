package serviceintegration

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceServiceIntegrationEndpoint() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceServiceIntegrationEndpointRead,
		Description: "The Service Integration Endpoint data source provides information about the existing Aiven Service Integration Endpoint.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenServiceIntegrationEndpointSchema(),
			"project", "endpoint_name"),
	}
}

func datasourceServiceIntegrationEndpointRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	endpointName := d.Get("endpoint_name").(string)

	endpoints, err := client.ServiceIntegrationEndpoints.List(ctx, projectName)
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
