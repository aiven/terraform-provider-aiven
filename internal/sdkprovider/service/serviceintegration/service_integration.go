package serviceintegration

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"slices"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/samber/lo"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/userconfig/converters"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/userconfig/serviceintegration"
)

const serviceIntegrationEndpointRegExp = "^[a-zA-Z0-9_-]*\\/{1}[a-zA-Z0-9_-]*$"

func serviceIntegrationTypeChoices() []string {
	// These integrations are not supposed to be exposed in TF
	ignore := []string{
		"application_service_credential",
		"kafka_inkless_postgresql",
	}
	return lo.Filter(service.IntegrationTypeChoices(), func(s string, _ int) bool {
		return !slices.Contains(ignore, s)
	})
}

func hasIntegrationConfig[T string | service.EndpointType](kind T) bool {
	return slices.Contains(serviceintegration.UserConfigTypes(), string(kind))
}

func aivenServiceIntegrationSchema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"integration_id": {
			Description: "The ID of the Aiven service integration.",
			Computed:    true,
			Type:        schema.TypeString,
		},
		"destination_endpoint_id": {
			Description: "Destination endpoint for the integration.",
			ForceNew:    true,
			Optional:    true,
			Type:        schema.TypeString,
			ValidateFunc: validation.StringMatch(regexp.MustCompile(serviceIntegrationEndpointRegExp),
				"endpoint ID should have the following format: project_name/endpoint_id"),
		},
		"destination_service_name": {
			Description: "Destination service for the integration.",
			ForceNew:    true,
			Optional:    true,
			Type:        schema.TypeString,
		},
		"destination_service_project": {
			Description: "Destination service project name",
			ForceNew:    true,
			Optional:    true,
			Computed:    true,
			Type:        schema.TypeString,
		},
		"integration_type": {
			Description:  userconfig.Desc("Type of the service integration").PossibleValuesString(serviceIntegrationTypeChoices()...).Build(),
			ForceNew:     true,
			Required:     true,
			Type:         schema.TypeString,
			ValidateFunc: validation.StringInSlice(serviceIntegrationTypeChoices(), false),
		},
		"project": {
			Description: "Project the integration belongs to.",
			ForceNew:    true,
			Required:    true,
			Type:        schema.TypeString,
		},
		"source_endpoint_id": {
			Description: "Source endpoint for the integration.",
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
			Computed:    true,
			Type:        schema.TypeString,
		},
		"source_service_project": {
			Description: "Source service project name",
			ForceNew:    true,
			Optional:    true,
			Computed:    true,
			Type:        schema.TypeString,
		},
	}

	// Adds user configs
	for _, k := range serviceIntegrationTypeChoices() {
		if hasIntegrationConfig(k) {
			converters.SetUserConfig(converters.ServiceIntegrationUserConfig, k, s)
		}
	}
	return s
}

func ResourceServiceIntegration() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates and manages an Aiven [service integration](https://aiven.io/docs/platform/concepts/service-integration).",
		CreateContext: common.WithGenClient(resourceServiceIntegrationCreate),
		ReadContext:   common.WithGenClient(resourceServiceIntegrationRead),
		UpdateContext: common.WithGenClient(resourceServiceIntegrationUpdate),
		DeleteContext: common.WithGenClient(resourceServiceIntegrationDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema:         aivenServiceIntegrationSchema(),
		SchemaVersion:  1,
		StateUpgraders: stateupgrader.ServiceIntegration(),
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

func resourceServiceIntegrationCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName := d.Get("project").(string)
	integrationType := d.Get("integration_type").(string)

	// read_replicas can be only be created alongside the service. also the only way to promote the replica
	// is to delete the service integration that was created so we should make it least painful to do so.
	// for now we support to seemlessly import preexisting 'read_replica' service integrations in the resource create
	// all other integrations should be imported using `terraform import`
	if integrationType == "read_replica" {
		if preexisting, err := resourceServiceIntegrationCheckForPreexistingResource(ctx, d, client); err != nil {
			return fmt.Errorf("unable to search for possible preexisting 'read_replica' service integration: %w", err)
		} else if preexisting != nil {
			d.SetId(schemautil.BuildResourceID(projectName, preexisting.ServiceIntegrationId))
			return resourceServiceIntegrationRead(ctx, d, client)
		}
	}

	req := &service.ServiceIntegrationCreateIn{
		DestEndpointId:   plainEndpointID(schemautil.OptionalStringPointer(d, "destination_endpoint_id")),
		DestService:      schemautil.OptionalStringPointer(d, "destination_service_name"),
		IntegrationType:  service.IntegrationType(integrationType),
		SourceEndpointId: plainEndpointID(schemautil.OptionalStringPointer(d, "source_endpoint_id")),
		SourceService:    schemautil.OptionalStringPointer(d, "source_service_name"),
		DestProject:      schemautil.OptionalStringPointer(d, "destination_service_project"),
		SourceProject:    schemautil.OptionalStringPointer(d, "source_service_project"),
	}

	uc, err := converters.Expand(converters.ServiceIntegrationUserConfig, integrationType, d)
	if err != nil {
		return err
	}

	if uc != nil {
		req.UserConfig = &uc
	}

	res, err := client.ServiceIntegrationCreate(ctx, projectName, req)
	if err != nil {
		return fmt.Errorf("error creating service integration: %w", err)
	}
	d.SetId(schemautil.BuildResourceID(projectName, res.ServiceIntegrationId))

	if err = resourceServiceIntegrationWaitUntilActive(ctx, d, client, d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("unable to wait for service integration to become active: %w", err)
	}
	return resourceServiceIntegrationRead(ctx, d, client)
}

func resourceServiceIntegrationRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName, integrationID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return err
	}

	res, err := client.ServiceIntegrationGet(ctx, projectName, integrationID)
	if err != nil {
		err = schemautil.ResourceReadHandleNotFound(err, d)
		if err != nil {
			return fmt.Errorf("cannot get service integration: %w; id: %s", err, integrationID)
		}
		return nil
	}

	if err = resourceServiceIntegrationCopyAPIResponseToTerraform(d, res, projectName); err != nil {
		return fmt.Errorf("cannot copy api response into terraform schema: %w", err)
	}

	return nil
}

func resourceServiceIntegrationUpdate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName, integrationID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return err
	}

	integrationType := d.Get("integration_type").(string)
	userConfig, err := converters.Expand(converters.ServiceIntegrationUserConfig, integrationType, d)
	if err != nil {
		return err
	}

	if userConfig == nil {
		// Required by API
		userConfig = make(map[string]interface{})
	}

	_, err = client.ServiceIntegrationUpdate(
		ctx,
		projectName,
		integrationID,
		&service.ServiceIntegrationUpdateIn{
			UserConfig: userConfig,
		},
	)
	if err != nil {
		return fmt.Errorf("unable to update service integration: %w", err)
	}
	if err = resourceServiceIntegrationWaitUntilActive(ctx, d, client, d.Timeout(schema.TimeoutUpdate)); err != nil {
		return fmt.Errorf("unable to wait for service integration to become active: %w", err)
	}

	return resourceServiceIntegrationRead(ctx, d, client)
}

func resourceServiceIntegrationDelete(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName, integrationID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return err
	}

	err = client.ServiceIntegrationDelete(ctx, projectName, integrationID)
	if common.IsCritical(err) {
		return fmt.Errorf("cannot delete service integration: %w", err)
	}

	return nil
}

func resourceServiceIntegrationCheckForPreexistingResource(ctx context.Context, d *schema.ResourceData, client avngen.Client) (*service.ServiceIntegrationOut, error) {
	projectName := d.Get("project").(string)
	integrationType := d.Get("integration_type").(string)
	sourceServiceName := d.Get("source_service_name").(string)
	destinationServiceName := d.Get("destination_service_name").(string)

	integrations, err := client.ServiceIntegrationList(ctx, projectName, sourceServiceName)
	if common.IsCritical(err) {
		return nil, fmt.Errorf("unable to get list of service integrations: %w", err)
	}

	for i := range integrations {
		integration := integrations[i]
		if integration.SourceService == "" || integration.DestService == nil || integration.ServiceIntegrationId == "" {
			continue
		}

		if string(integration.IntegrationType) == integrationType &&
			integration.SourceService == sourceServiceName &&
			*integration.DestService == destinationServiceName {
			return &integration, nil
		}
	}
	return nil, nil
}

func resourceServiceIntegrationWaitUntilActive(ctx context.Context, d *schema.ResourceData, client avngen.Client, timeout time.Duration) error {
	const (
		active    = "ACTIVE"
		notActive = "NOTACTIVE"
	)
	projectName, integrationID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return err
	}

	stateChangeConf := &retry.StateChangeConf{
		Pending: []string{notActive},
		Target:  []string{active},
		Refresh: func() (interface{}, string, error) {
			log.Println("[DEBUG] Service Integration: waiting until active")

			ii, err := client.ServiceIntegrationGet(ctx, projectName, integrationID)
			if err != nil {
				// Sometimes Aiven API retrieves 404 error even when a successful service integration is created
				if avngen.IsNotFound(err) {
					log.Println("[DEBUG] Service Integration: not yet found")
					return nil, notActive, nil
				}
				return nil, "", err
			}
			if !ii.Active {
				log.Println("[DEBUG] Service Integration: not yet active")
				return nil, notActive, nil
			}

			if ii.IntegrationType == "kafka_connect" && ii.DestService != nil {
				if _, err := client.ServiceKafkaConnectList(ctx, projectName, *ii.DestService); err != nil {
					log.Println("[DEBUG] Service Integration: error listing kafka connectors: ", err)
					return nil, notActive, nil
				}
			}
			return ii, active, nil
		},
		Delay:                     common.DefaultStateChangeDelay,
		Timeout:                   timeout,
		MinTimeout:                common.DefaultStateChangeMinTimeout,
		ContinuousTargetOccurence: 10,
	}
	if _, err := stateChangeConf.WaitForStateContext(ctx); err != nil {
		return err
	}
	return nil
}

func resourceServiceIntegrationCopyAPIResponseToTerraform(
	d *schema.ResourceData,
	res *service.ServiceIntegrationGetOut,
	project string,
) error {
	if err := d.Set("project", project); err != nil {
		return err
	}

	if res.DestEndpointId != nil {
		if err := d.Set("destination_endpoint_id", schemautil.BuildResourceID(project, *res.DestEndpointId)); err != nil {
			return err
		}
	} else if res.DestService != nil {
		if err := d.Set("destination_service_name", *res.DestService); err != nil {
			return err
		}
	}
	if res.SourceEndpointId != nil {
		if err := d.Set("source_endpoint_id", schemautil.BuildResourceID(project, *res.SourceEndpointId)); err != nil {
			return err
		}
	} else if res.SourceService != "" {
		if err := d.Set("source_service_name", res.SourceService); err != nil {
			return err
		}
	}
	if err := d.Set("integration_id", res.ServiceIntegrationId); err != nil {
		return err
	}
	integrationType := res.IntegrationType
	if err := d.Set("integration_type", integrationType); err != nil {
		return err
	}

	if err := d.Set("destination_service_project", res.DestProject); err != nil {
		return err
	}
	if err := d.Set("source_service_project", res.SourceProject); err != nil {
		return err
	}

	return converters.Flatten(converters.ServiceIntegrationUserConfig, string(integrationType), d, res.UserConfig)
}
