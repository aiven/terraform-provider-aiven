package service_integration

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aiven/terraform-provider-aiven/internal/meta"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/templates"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const serviceIntegrationEndpointRegExp = "^[a-zA-Z0-9_-]*\\/{1}[a-zA-Z0-9_-]*$"

var aivenServiceIntegrationSchema = map[string]*schema.Schema{
	"integration_id": {
		Description: "Service Integration Id at aiven",
		Computed:    true,
		Type:        schema.TypeString,
	},
	"destination_endpoint_id": {
		Description: "Destination endpoint for the integration (if any)",
		ForceNew:    true,
		Optional:    true,
		Type:        schema.TypeString,
		ValidateFunc: validation.StringMatch(regexp.MustCompile(serviceIntegrationEndpointRegExp),
			"endpoint id should have the following format: project_name/endpoint_id"),
	},
	"destination_service_name": {
		Description: "Destination service for the integration (if any)",
		ForceNew:    true,
		Optional:    true,
		Type:        schema.TypeString,
	},
	"integration_type": {
		Description: "Type of the service integration",
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
	},
	"logs_user_config": {
		Description: "Log integration specific user configurable settings",
		Elem: &schema.Resource{
			Schema: schemautil.GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("integration")["logs"].(map[string]interface{})),
		},
		MaxItems: 1,
		Optional: true,
		Type:     schema.TypeList,
	},
	"mirrormaker_user_config": {
		Description: "Mirrormaker 1 integration specific user configurable settings",
		Elem: &schema.Resource{
			Schema: schemautil.GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("integration")["mirrormaker"].(map[string]interface{})),
		},
		MaxItems: 1,
		Optional: true,
		Type:     schema.TypeList,
	},
	"kafka_mirrormaker_user_config": {
		Description: "Mirrormaker 2 integration specific user configurable settings",
		Elem: &schema.Resource{
			Schema: schemautil.GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("integration")["kafka_mirrormaker"].(map[string]interface{})),
		},
		MaxItems: 1,
		Optional: true,
		Type:     schema.TypeList,
	},
	"kafka_connect_user_config": {
		Description: "Kafka Connect specific user configurable settings",
		Elem: &schema.Resource{
			Schema: schemautil.GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("integration")["kafka_connect"].(map[string]interface{})),
		},
		MaxItems: 1,
		Optional: true,
		Type:     schema.TypeList,
	},
	"kafka_logs_user_config": {
		Description: "Kafka Logs specific user configurable settings",
		Elem: &schema.Resource{
			Schema: schemautil.GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("integration")["kafka_logs"].(map[string]interface{})),
		},
		MaxItems: 1,
		Optional: true,
		Type:     schema.TypeList,
	},
	"metrics_user_config": {
		Description: "Metrics specific user configurable settings",
		Elem: &schema.Resource{
			Schema: schemautil.GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("integration")["metrics"].(map[string]interface{})),
		},
		MaxItems: 1,
		Optional: true,
		Type:     schema.TypeList,
	},
	"project": {
		Description: "Project the integration belongs to",
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
	},
	"source_endpoint_id": {
		Description: "Source endpoint for the integration (if any)",
		ForceNew:    true,
		Optional:    true,
		Type:        schema.TypeString,
		ValidateFunc: validation.StringMatch(regexp.MustCompile(serviceIntegrationEndpointRegExp),
			"endpoint id should have the following format: project_name/endpoint_id"),
	},
	"source_service_name": {
		Description: "Source service for the integration (if any)",
		ForceNew:    true,
		Optional:    true,
		Type:        schema.TypeString,
	},
}

func ResourceServiceIntegration() *schema.Resource {
	return &schema.Resource{
		Description:   "The Service Integration resource allows the creation and management of Aiven Service Integrations.",
		CreateContext: resourceServiceIntegrationCreate,
		ReadContext:   resourceServiceIntegrationRead,
		UpdateContext: resourceServiceIntegrationUpdate,
		DeleteContext: resourceServiceIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceServiceIntegrationState,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: aivenServiceIntegrationSchema,
	}
}

func plainEndpointID(fullEndpointID *string) *string {
	if fullEndpointID == nil {
		return nil
	}
	_, endpointID := schemautil.SplitResourceID2(*fullEndpointID)
	return &endpointID
}

func resourceServiceIntegrationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*meta.Meta).Client

	projectName := d.Get("project").(string)
	integrationType := d.Get("integration_type").(string)

	// read_replicas can be only be created alongside the service. also the only way to promote the replica
	// is to delete the service integration that was created so we should make it least painful to do so.
	// for now we support to seemlessly import preexisting 'read_replica' service integrations in the resource create
	// all other integrations should be imported using `terraform import`
	if integrationType == "read_replica" {
		if preexisting, err := resourceServiceIntegrationCheckForPreexistingResource(ctx, d, m); err != nil {
			return diag.Errorf("unable to search for possible preexisting 'read_replica' service integration: %s", err)
		} else if preexisting != nil {
			d.SetId(schemautil.BuildResourceID(projectName, preexisting.ServiceIntegrationID))
			return resourceServiceIntegrationRead(ctx, d, m)
		}
	}

	integration, err := client.ServiceIntegrations.Create(
		projectName,
		aiven.CreateServiceIntegrationRequest{
			DestinationEndpointID: plainEndpointID(schemautil.OptionalStringPointer(d, "destination_endpoint_id")),
			DestinationService:    schemautil.OptionalStringPointer(d, "destination_service_name"),
			IntegrationType:       integrationType,
			SourceEndpointID:      plainEndpointID(schemautil.OptionalStringPointer(d, "source_endpoint_id")),
			SourceService:         schemautil.OptionalStringPointer(d, "source_service_name"),
			UserConfig:            resourceServiceIntegrationUserConfigFromSchemaToAPI(d),
		},
	)
	if err != nil {
		return diag.Errorf("error creating serivce integration: %s", err)
	}
	d.SetId(schemautil.BuildResourceID(projectName, integration.ServiceIntegrationID))

	if err = resourceServiceIntegrationWaitUntilActive(ctx, d, m); err != nil {
		return diag.Errorf("unable to wait for service integration to become active: %s", err)
	}
	return resourceServiceIntegrationRead(ctx, d, m)
}

func resourceServiceIntegrationRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*meta.Meta).Client

	projectName, integrationID := schemautil.SplitResourceID2(d.Id())
	integration, err := client.ServiceIntegrations.Get(projectName, integrationID)
	if err != nil {
		err = schemautil.ResourceReadHandleNotFound(err, d, m)
		if err != nil {
			return diag.Errorf("cannot get service integration: %s; id: %s", err, integrationID)
		}
		return nil
	}

	if err = resourceServiceIntegrationCopyAPIResponseToTerraform(d, integration, projectName); err != nil {
		return diag.Errorf("cannot copy api response into terraform schema: %s", err)
	}

	return nil
}

func resourceServiceIntegrationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*meta.Meta).Client

	projectName, integrationID := schemautil.SplitResourceID2(d.Id())

	_, err := client.ServiceIntegrations.Update(
		projectName,
		integrationID,
		aiven.UpdateServiceIntegrationRequest{
			UserConfig: resourceServiceIntegrationUserConfigFromSchemaToAPI(d),
		},
	)
	if err != nil {
		return diag.Errorf("unable to update service integration: %s", err)
	}
	if err = resourceServiceIntegrationWaitUntilActive(ctx, d, m); err != nil {
		return diag.Errorf("unable to wait for service integration to become active: %s", err)
	}

	return resourceServiceIntegrationRead(ctx, d, m)
}

func resourceServiceIntegrationDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*meta.Meta).Client

	projectName, integrationID := schemautil.SplitResourceID2(d.Id())
	err := client.ServiceIntegrations.Delete(projectName, integrationID)
	if err != nil && !aiven.IsNotFound(err) {
		return diag.Errorf("cannot delete service integration: %s", err)
	}

	return nil
}

func resourceServiceIntegrationState(_ context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	m.(*meta.Meta).Import = true

	client := m.(*meta.Meta).Client

	if len(strings.Split(d.Id(), "/")) != 2 {
		return nil, fmt.Errorf("invalid identifier %v, expected <project_name>/<integration_id>", d.Id())
	}

	projectName, integrationID := schemautil.SplitResourceID2(d.Id())
	integration, err := client.ServiceIntegrations.Get(projectName, integrationID)
	if err != nil {
		return nil, err
	}

	if err = resourceServiceIntegrationCopyAPIResponseToTerraform(d, integration, projectName); err != nil {
		return nil, fmt.Errorf("cannot copy api response into terraform schema: %s", err)
	}

	return []*schema.ResourceData{d}, nil
}

func resourceServiceIntegrationCheckForPreexistingResource(_ context.Context, d *schema.ResourceData, m interface{}) (*aiven.ServiceIntegration, error) {
	client := m.(*meta.Meta).Client

	projectName := d.Get("project").(string)
	integrationType := d.Get("integration_type").(string)
	sourceServiceName := d.Get("source_service_name").(string)
	destinationServiceName := d.Get("destination_service_name").(string)

	integrations, err := client.ServiceIntegrations.List(projectName, sourceServiceName)
	if err != nil && !aiven.IsNotFound(err) {
		return nil, fmt.Errorf("unable to get list of service integrations: %s", err)
	}

	for i := range integrations {
		integration := integrations[i]
		if integration.SourceService == nil || integration.DestinationService == nil || integration.ServiceIntegrationID == "" {
			continue
		}

		if integration.IntegrationType == integrationType &&
			*integration.SourceService == sourceServiceName &&
			*integration.DestinationService == destinationServiceName {
			return integration, nil
		}
	}
	return nil, nil
}

func resourceServiceIntegrationWaitUntilActive(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	const (
		active    = "ACTIVE"
		notActive = "NOTACTIVE"
	)
	client := m.(*meta.Meta).Client

	projectName, integrationID := schemautil.SplitResourceID2(d.Id())

	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{notActive},
		Target:  []string{active},
		Refresh: func() (interface{}, string, error) {
			log.Println("[DEBUG] Service Integration: waiting until active")

			ii, err := client.ServiceIntegrations.Get(projectName, integrationID)
			if err != nil {
				// Sometimes Aiven API retrieves 404 error even when a successful service integration is created
				if aiven.IsNotFound(err) {
					log.Println("[DEBUG] Service Integration: not yet found")
					return nil, notActive, nil
				}
				return nil, "", err
			}
			if !ii.Active {
				log.Println("[DEBUG] Service Integration: not yet active")
				return nil, notActive, nil
			}

			if ii.IntegrationType == "kafka_connect" && ii.DestinationService != nil {
				if _, err := client.KafkaConnectors.List(projectName, *ii.DestinationService); err != nil {
					log.Println("[DEBUG] Service Integration: error listing kafka connectors: ", err)
					return nil, notActive, nil
				}
			}
			return ii, active, nil
		},
		Delay:                     2 * time.Second,
		Timeout:                   d.Timeout(schema.TimeoutCreate),
		MinTimeout:                2 * time.Second,
		ContinuousTargetOccurence: 10,
	}
	if _, err := stateChangeConf.WaitForStateContext(ctx); err != nil {
		return err
	}
	return nil
}

func resourceServiceIntegrationUserConfigFromSchemaToAPI(d *schema.ResourceData) map[string]interface{} {
	const (
		userConfigTypeIntegrations = "integration"
	)
	integrationType := d.Get("integration_type").(string)
	return schemautil.ConvertTerraformUserConfigToAPICompatibleFormat(userConfigTypeIntegrations, integrationType, true, d)
}

func resourceServiceIntegrationCopyAPIResponseToTerraform(
	d *schema.ResourceData,
	integration *aiven.ServiceIntegration,
	project string,
) error {
	if err := d.Set("project", project); err != nil {
		return err
	}

	if integration.DestinationEndpointID != nil {
		if err := d.Set("destination_endpoint_id", schemautil.BuildResourceID(project, *integration.DestinationEndpointID)); err != nil {
			return err
		}
	} else if integration.DestinationService != nil {
		if err := d.Set("destination_service_name", *integration.DestinationService); err != nil {
			return err
		}
	}
	if integration.SourceEndpointID != nil {
		if err := d.Set("source_endpoint_id", schemautil.BuildResourceID(project, *integration.SourceEndpointID)); err != nil {
			return err
		}
	} else if integration.SourceService != nil {
		if err := d.Set("source_service_name", *integration.SourceService); err != nil {
			return err
		}
	}
	if err := d.Set("integration_id", integration.ServiceIntegrationID); err != nil {
		return err
	}
	integrationType := integration.IntegrationType
	if err := d.Set("integration_type", integrationType); err != nil {
		return err
	}

	userConfig := schemautil.ConvertAPIUserConfigToTerraformCompatibleFormat("integration", integrationType, integration.UserConfig)
	if len(userConfig) > 0 {
		d.Set(integrationType+"_user_config", userConfig)
	}

	return nil
}
