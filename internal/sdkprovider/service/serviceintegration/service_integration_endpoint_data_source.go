package serviceintegration

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceServiceIntegrationEndpoint() *schema.Resource {
	return &schema.Resource{
		ReadContext: common.WithGenClient(datasourceServiceIntegrationEndpointRead),
		Description: "Gets information about an integration endpoint.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenServiceIntegrationEndpointSchema(),
			"project", "endpoint_name"),
	}
}

func datasourceServiceIntegrationEndpointRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName := d.Get("project").(string)
	endpointName := d.Get("endpoint_name").(string)

	endpoints, err := client.ServiceIntegrationEndpointList(ctx, projectName)
	if err != nil {
		return err
	}

	for _, endpoint := range endpoints {
		if endpoint.EndpointName == endpointName {
			d.SetId(schemautil.BuildResourceID(projectName, endpoint.EndpointId))
			return resourceServiceIntegrationEndpointRead(ctx, d, client)
		}
	}

	return fmt.Errorf("endpoint \"%s\" not found", endpointName)
}
