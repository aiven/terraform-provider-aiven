// Copyright (c) 2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceServiceIntegration() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceServiceIntegrationRead,
		Description: "The Service Integration data source provides information about the existing Aiven Service Integration.",
		Schema: resourceSchemaAsDatasourceSchema(aivenServiceIntegrationSchema,
			"project", "integration_type", "source_service_name", "destination_service_name"),
	}
}

func datasourceServiceIntegrationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	integrationType := d.Get("integration_type").(string)
	sourceServiceName := d.Get("source_service_name").(string)
	destinationServiceName := d.Get("destination_service_name").(string)

	integrations, err := client.ServiceIntegrations.List(projectName, sourceServiceName)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, i := range integrations {
		if i.SourceService == nil || i.DestinationService == nil {
			continue
		}

		if i.IntegrationType == integrationType &&
			*i.SourceService == sourceServiceName &&
			*i.DestinationService == destinationServiceName {

			d.SetId(buildResourceID(projectName, i.ServiceIntegrationID))
			return resourceServiceIntegrationRead(ctx, d, m)
		}
	}

	return diag.Errorf("service integration %s/%s/%s/%s not found",
		projectName, integrationType, sourceServiceName, destinationServiceName)
}
