package serviceintegration

import (
	"context"
	"fmt"
	"slices"

	"github.com/aiven/aiven-go-client/v2"
	codegenintegrations "github.com/aiven/go-client-codegen/handler/serviceintegration"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/userconfig/converters"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/userconfig/serviceintegrationendpoint"
)

func hasEndpointConfig(kind string) bool {
	return slices.Contains(serviceintegrationendpoint.UserConfigTypes(), kind)
}

func aivenServiceIntegrationEndpointSchema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"project": {
			Description: "Project the service integration endpoint belongs to",
			ForceNew:    true,
			Required:    true,
			Type:        schema.TypeString,
		},
		"endpoint_name": {
			ForceNew:    true,
			Description: "Name of the service integration endpoint",
			Required:    true,
			Type:        schema.TypeString,
		},
		"endpoint_type": {
			Description: "Type of the service integration endpoint. Possible values: " +
				schemautil.JoinQuoted(codegenintegrations.EndpointTypeChoices(), ", ", "`"),
			ForceNew:     true,
			Required:     true,
			Type:         schema.TypeString,
			ValidateFunc: validation.StringInSlice(codegenintegrations.EndpointTypeChoices(), false),
		},
		"endpoint_config": {
			Description: "Integration endpoint specific backend configuration",
			Computed:    true,
			Type:        schema.TypeMap,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
	}

	// Adds user configs
	for _, k := range serviceintegrationendpoint.UserConfigTypes() {
		converters.SetUserConfig(converters.ServiceIntegrationEndpointUserConfig, k, s)
	}
	return s
}

func ResourceServiceIntegrationEndpoint() *schema.Resource {
	return &schema.Resource{
		Description:   "The Service Integration Endpoint resource allows the creation and management of Aiven Service Integration Endpoints.",
		CreateContext: resourceServiceIntegrationEndpointCreate,
		ReadContext:   resourceServiceIntegrationEndpointRead,
		UpdateContext: resourceServiceIntegrationEndpointUpdate,
		DeleteContext: resourceServiceIntegrationEndpointDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema:         aivenServiceIntegrationEndpointSchema(),
		SchemaVersion:  1,
		StateUpgraders: stateupgrader.ServiceIntegrationEndpoint(),
	}
}

func resourceServiceIntegrationEndpointCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)
	projectName := d.Get("project").(string)
	endpointType := d.Get("endpoint_type").(string)

	req := aiven.CreateServiceIntegrationEndpointRequest{
		EndpointName: d.Get("endpoint_name").(string),
		EndpointType: endpointType,
		UserConfig:   make(map[string]interface{}),
	}

	if hasEndpointConfig(endpointType) {
		uc, err := converters.Expand(converters.ServiceIntegrationEndpointUserConfig, endpointType, d)
		if err != nil {
			return diag.FromErr(err)
		}
		if uc != nil {
			req.UserConfig = uc
		}
	}

	endpoint, err := client.ServiceIntegrationEndpoints.Create(ctx, projectName, req)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(projectName, endpoint.EndpointID))

	return resourceServiceIntegrationEndpointRead(ctx, d, m)
}

func resourceServiceIntegrationEndpointRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, endpointID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	endpoint, err := client.ServiceIntegrationEndpoints.Get(ctx, projectName, endpointID)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	err = copyServiceIntegrationEndpointPropertiesFromAPIResponseToTerraform(d, endpoint, projectName)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceServiceIntegrationEndpointUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, endpointID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	endpointType := d.Get("endpoint_type").(string)
	req := aiven.UpdateServiceIntegrationEndpointRequest{
		UserConfig: make(map[string]interface{}),
	}

	if hasEndpointConfig(endpointType) {
		uc, err := converters.Expand(converters.ServiceIntegrationEndpointUserConfig, endpointType, d)
		if err != nil {
			return diag.FromErr(err)
		}
		if uc != nil {
			req.UserConfig = uc
		}
	}

	_, err = client.ServiceIntegrationEndpoints.Update(ctx, projectName, endpointID, req)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceServiceIntegrationEndpointRead(ctx, d, m)
}

func resourceServiceIntegrationEndpointDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, endpointID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.ServiceIntegrationEndpoints.Delete(ctx, projectName, endpointID)
	if common.IsCritical(err) {
		return diag.FromErr(err)
	}

	return nil
}

func copyServiceIntegrationEndpointPropertiesFromAPIResponseToTerraform(
	d *schema.ResourceData,
	endpoint *aiven.ServiceIntegrationEndpoint,
	project string,
) error {
	if err := d.Set("project", project); err != nil {
		return err
	}
	if err := d.Set("endpoint_name", endpoint.EndpointName); err != nil {
		return err
	}
	endpointType := endpoint.EndpointType
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
