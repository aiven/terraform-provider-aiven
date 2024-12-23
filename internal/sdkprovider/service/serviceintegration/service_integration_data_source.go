package serviceintegration

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceServiceIntegration() *schema.Resource {
	return &schema.Resource{
		ReadContext: common.WithGenClient(datasourceServiceIntegrationRead),
		Description: "Gets information about an Aiven service integration.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenServiceIntegrationSchema(),
			"project", "integration_type", "source_service_name", "destination_service_name"),
	}
}

func datasourceServiceIntegrationRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName := d.Get("project").(string)
	integrationType := d.Get("integration_type").(string)
	sourceServiceName := d.Get("source_service_name").(string)
	destinationServiceName := d.Get("destination_service_name").(string)

	integrations, err := client.ServiceIntegrationList(ctx, projectName, sourceServiceName)
	if err != nil {
		return fmt.Errorf("unable to list integrations for %s/%s: %w", projectName, sourceServiceName, err)
	}

	for _, i := range integrations {
		if i.SourceService == "" || i.DestService == nil {
			continue
		}

		if string(i.IntegrationType) == integrationType &&
			i.SourceService == sourceServiceName &&
			*i.DestService == destinationServiceName {

			d.SetId(schemautil.BuildResourceID(projectName, i.ServiceIntegrationId))
			return resourceServiceIntegrationRead(ctx, d, client)
		}
	}

	return fmt.Errorf("common integration %s/%s/%s/%s not found",
		projectName, integrationType, sourceServiceName, destinationServiceName)
}
