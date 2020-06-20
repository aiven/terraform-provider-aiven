// Copyright (c) 2020 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func datasourceServiceIntegration() *schema.Resource {
	return &schema.Resource{
		Read: datasourceServiceIntegrationRead,
		Schema: resourceSchemaAsDatasourceSchema(aivenServiceIntegrationSchema,
			"project", "integration_type", "source_service_name", "destination_service_name"),
	}
}

func datasourceServiceIntegrationRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	integrationType := d.Get("integration_type").(string)
	sourceServiceName := d.Get("source_service_name").(string)
	destinationServiceName := d.Get("destination_service_name").(string)

	integrations, err := client.ServiceIntegrations.List(projectName, sourceServiceName)
	if err != nil {
		return err
	}

	for _, i := range integrations {
		if i.SourceService == nil || i.DestinationService == nil {
			continue
		}

		if i.IntegrationType == integrationType &&
			*i.SourceService == sourceServiceName &&
			*i.DestinationService == destinationServiceName {

			d.SetId(buildResourceID(projectName, i.ServiceIntegrationID))
			return resourceServiceIntegrationRead(d, m)
		}
	}

	return fmt.Errorf("service integration %s/%s/%s/%s not found",
		projectName, integrationType, sourceServiceName, destinationServiceName)
}
