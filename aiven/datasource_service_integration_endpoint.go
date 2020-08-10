// Copyright (c) 2019 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"fmt"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func datasourceServiceIntegrationEndpoint() *schema.Resource {
	return &schema.Resource{
		Read: datasourceServiceIntegrationEndpointRead,
		Schema: resourceSchemaAsDatasourceSchema(aivenServiceIntegrationEndpointSchema,
			"project", "endpoint_name"),
	}
}

func datasourceServiceIntegrationEndpointRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	endpointName := d.Get("endpoint_name").(string)

	endpoints, err := client.ServiceIntegrationEndpoints.List(projectName)
	if err != nil {
		return err
	}

	for _, endpoint := range endpoints {
		if endpoint.EndpointName == endpointName {
			d.SetId(buildResourceID(projectName, endpoint.EndpointID))
			return resourceServiceIntegrationEndpointRead(d, m)
		}
	}

	return fmt.Errorf("endpoint \"%s\" not found", endpointName)
}
