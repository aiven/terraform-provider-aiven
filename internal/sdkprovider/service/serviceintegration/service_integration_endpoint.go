package serviceintegration

import (
	"context"
	"fmt"
	"slices"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/samber/lo"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/userconfig/converters"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/userconfig/serviceintegrationendpoint"
)

func serviceEndpointTypeChoices() []string {
	// These endpoints are not supposed to be exposed in TF
	ignore := []string{
		"autoscaler_service",
	}
	return lo.Filter(service.EndpointTypeChoices(), func(s string, _ int) bool {
		return !slices.Contains(ignore, s)
	})
}

func hasEndpointConfig[T string | service.EndpointType](kind T) bool {
	return slices.Contains(serviceintegrationendpoint.UserConfigTypes(), string(kind))
}

func aivenServiceIntegrationEndpointSchema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"project": {
			Description: "Project the service integration endpoint is in.",
			ForceNew:    true,
			Required:    true,
			Type:        schema.TypeString,
		},
		"endpoint_name": {
			ForceNew:    true,
			Description: "Name of the service integration endpoint.",
			Required:    true,
			Type:        schema.TypeString,
		},
		"endpoint_type": {
			Description:  userconfig.Desc("The type of service integration endpoint.").PossibleValuesString(serviceEndpointTypeChoices()...).Build(),
			ForceNew:     true,
			Required:     true,
			Type:         schema.TypeString,
			ValidateFunc: validation.StringInSlice(serviceEndpointTypeChoices(), false),
		},
		"endpoint_config": {
			Description: "Backend configuration for the endpoint.",
			Computed:    true,
			Type:        schema.TypeMap,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
	}

	// Adds user configs
	for _, k := range serviceEndpointTypeChoices() {
		if hasEndpointConfig(k) {
			converters.SetUserConfig(converters.ServiceIntegrationEndpointUserConfig, k, s)
		}
	}
	return s
}

func ResourceServiceIntegrationEndpoint() *schema.Resource {
	return &schema.Resource{
		Description: `Creates and manages an integration endpoint.

Integration endpoints let you send data like metrics and logs from Aiven services to external systems. The ` + "`autoscaler`" + ` endpoint lets you automatically scale the disk space on your services.

After creating an endpoint, use the [service integration resource](https://registry.terraform.io/providers/aiven/aiven/latest/docs/resources/service_integration) to connect it to a service.
		`,
		CreateContext: common.WithGenClient(resourceServiceIntegrationEndpointCreate),
		ReadContext:   common.WithGenClient(resourceServiceIntegrationEndpointRead),
		UpdateContext: common.WithGenClient(resourceServiceIntegrationEndpointUpdate),
		DeleteContext: common.WithGenClient(resourceServiceIntegrationEndpointDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema:         aivenServiceIntegrationEndpointSchema(),
		SchemaVersion:  1,
		StateUpgraders: stateupgrader.ServiceIntegrationEndpoint(),
	}
}

func resourceServiceIntegrationEndpointCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName := d.Get("project").(string)
	endpointType := d.Get("endpoint_type").(string)

	req := &service.ServiceIntegrationEndpointCreateIn{
		EndpointName: d.Get("endpoint_name").(string),
		EndpointType: service.EndpointType(endpointType),
		UserConfig:   make(map[string]interface{}),
	}

	if hasEndpointConfig(endpointType) {
		uc, err := converters.Expand(converters.ServiceIntegrationEndpointUserConfig, endpointType, d)
		if err != nil {
			return err
		}
		if uc != nil {
			req.UserConfig = uc
		}
	}

	endpoint, err := client.ServiceIntegrationEndpointCreate(ctx, projectName, req)
	if err != nil {
		return err
	}

	d.SetId(schemautil.BuildResourceID(projectName, endpoint.EndpointId))
	return resourceServiceIntegrationEndpointRead(ctx, d, client)
}

func resourceServiceIntegrationEndpointRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName, endpointID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return err
	}

	endpoint, err := client.ServiceIntegrationEndpointGet(ctx, projectName, endpointID, service.ServiceIntegrationEndpointGetIncludeSecrets(true))
	if err != nil {
		return schemautil.ResourceReadHandleNotFound(err, d)
	}

	// The GET with include_secrets=true must not return redacted creds
	err = schemautil.ContainsRedactedCreds(endpoint.EndpointConfig)
	if err != nil {
		return err
	}

	err = copyServiceIntegrationEndpointPropertiesFromAPIResponseToTerraform(d, endpoint, projectName)
	if err != nil {
		return err
	}

	return nil
}

func resourceServiceIntegrationEndpointUpdate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName, endpointID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return err
	}

	endpointType := d.Get("endpoint_type").(string)
	req := &service.ServiceIntegrationEndpointUpdateIn{
		UserConfig: make(map[string]interface{}),
	}

	if hasEndpointConfig(endpointType) {
		uc, err := converters.Expand(converters.ServiceIntegrationEndpointUserConfig, endpointType, d)
		if err != nil {
			return err
		}
		if uc != nil {
			req.UserConfig = uc
		}
	}

	_, err = client.ServiceIntegrationEndpointUpdate(ctx, projectName, endpointID, req)
	if err != nil {
		return err
	}

	return resourceServiceIntegrationEndpointRead(ctx, d, client)
}

func resourceServiceIntegrationEndpointDelete(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName, endpointID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return err
	}

	err = client.ServiceIntegrationEndpointDelete(ctx, projectName, endpointID)
	if common.IsCritical(err) {
		return err
	}

	return nil
}

func copyServiceIntegrationEndpointPropertiesFromAPIResponseToTerraform(
	d *schema.ResourceData,
	endpoint *service.ServiceIntegrationEndpointGetOut,
	project string,
) error {
	if err := d.Set("project", project); err != nil {
		return err
	}
	if err := d.Set("endpoint_name", endpoint.EndpointName); err != nil {
		return err
	}
	endpointType := string(endpoint.EndpointType)
	if err := d.Set("endpoint_type", endpointType); err != nil {
		return err
	}

	// Must coerse all values into strings
	endpointConfig := map[string]string{}
	if len(endpoint.EndpointConfig) > 0 {
		for key, value := range endpoint.EndpointConfig {
			endpointConfig[key] = fmt.Sprintf("%v", value)
		}
	}

	if err := d.Set("endpoint_config", endpointConfig); err != nil {
		return err
	}

	if hasEndpointConfig(endpointType) {
		err := converters.Flatten(converters.ServiceIntegrationEndpointUserConfig, endpointType, d, endpoint.UserConfig)
		if err != nil {
			return err
		}
	}
	return nil
}
