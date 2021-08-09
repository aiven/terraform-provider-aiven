// Copyright (c) 2019 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceServiceIntegrationEndpoint() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceServiceIntegrationEndpointRead,
		Schema: resourceSchemaAsDatasourceSchema(aivenServiceIntegrationEndpointSchema,
			"project", "endpoint_name"),
	}
}

func datasourceServiceIntegrationEndpointRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	endpointName := d.Get("endpoint_name").(string)

	endpoints, err := client.ServiceIntegrationEndpoints.List(projectName)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, endpoint := range endpoints {
		if endpoint.EndpointName == endpointName {
			d.SetId(buildResourceID(projectName, endpoint.EndpointID))
			return resourceServiceIntegrationEndpointRead(ctx, d, m)
		}
	}

	return diag.Errorf("endpoint \"%s\" not found", endpointName)
}
