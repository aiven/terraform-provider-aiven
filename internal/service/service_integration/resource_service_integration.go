package service_integration

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/apiconvert"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/dist"

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
	"logs_user_config":              dist.IntegrationTypeLogs(),
	"mirrormaker_user_config":       dist.IntegrationTypeMirrormaker(),
	"kafka_mirrormaker_user_config": dist.IntegrationTypeKafkaMirrormaker(),
	"kafka_connect_user_config":     dist.IntegrationTypeKafkaConnect(),
	"kafka_logs_user_config":        dist.IntegrationTypeKafkaLogs(),
	"metrics_user_config":           dist.IntegrationTypeMetrics(),
	"datadog_user_config":           dist.IntegrationTypeDatadog(),
}

func ResourceServiceIntegration() *schema.Resource {
	return &schema.Resource{
		Description:   "The Service Integration resource allows the creation and management of Aiven Service Integrations.",
		CreateContext: resourceServiceIntegrationCreate,
		ReadContext:   resourceServiceIntegrationRead,
		UpdateContext: resourceServiceIntegrationUpdate,
		DeleteContext: resourceServiceIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
	_, endpointID, err := schemautil.SplitResourceID2(*fullEndpointID)
	if err != nil {
		return nil
	}
	return &endpointID
}

func resourceServiceIntegrationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

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
	client := m.(*aiven.Client)

	projectName, integrationID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	integration, err := client.ServiceIntegrations.Get(projectName, integrationID)
	if err != nil {
		err = schemautil.ResourceReadHandleNotFound(err, d)
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
	client := m.(*aiven.Client)

	projectName, integrationID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	userConfig := resourceServiceIntegrationUserConfigFromSchemaToAPI(d)
	if userConfig == nil {
		// Required by API
		userConfig = make(map[string]interface{})
	}

	_, err = client.ServiceIntegrations.Update(
		projectName,
		integrationID,
		aiven.UpdateServiceIntegrationRequest{
			UserConfig: userConfig,
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
	client := m.(*aiven.Client)

	projectName, integrationID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.ServiceIntegrations.Delete(projectName, integrationID)
	if err != nil && !aiven.IsNotFound(err) {
		return diag.Errorf("cannot delete service integration: %s", err)
	}

	return nil
}

func resourceServiceIntegrationCheckForPreexistingResource(_ context.Context, d *schema.ResourceData, m interface{}) (*aiven.ServiceIntegration, error) {
	client := m.(*aiven.Client)

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
	client := m.(*aiven.Client)

	projectName, integrationID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return err
	}

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
	integrationType := d.Get("integration_type").(string)
	return apiconvert.ToAPI(userconfig.IntegrationTypes, integrationType, d)
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

	userConfig := apiconvert.FromAPI(userconfig.IntegrationTypes, integrationType, integration.UserConfig)
	if len(userConfig) > 0 {
		if err := d.Set(integrationType+"_user_config", userConfig); err != nil {
			return err
		}
	}

	return nil
}
