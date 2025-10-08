package schemautil

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aiven/aiven-go-client/v2"
	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/aiven/go-client-codegen/handler/staticip"
	"github.com/docker/go-units"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/samber/lo"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/userconfig/converters"
)

// defaultTimeout is the default timeout for service operations. This is not a const because it can be changed during
// compile time with -ldflags "-X github.com/aiven/terraform-provider-aiven/internal/schemautil.defaultTimeout=30".
var defaultTimeout time.Duration = 20

func DefaultResourceTimeouts() *schema.ResourceTimeout {
	return &schema.ResourceTimeout{
		Create: schema.DefaultTimeout(defaultTimeout * time.Minute),
		Update: schema.DefaultTimeout(defaultTimeout * time.Minute),
		Delete: schema.DefaultTimeout(defaultTimeout * time.Minute),
		// DEPRECATED: Default timeout is deprecated.
		// The Plugin Framework does not support the Default timeout field.
		// This field will be removed in a future major version.
		// See: https://developer.hashicorp.com/terraform/plugin/framework/resources/timeouts
		Default: schema.DefaultTimeout(defaultTimeout * time.Minute),
		Read:    schema.DefaultTimeout(defaultTimeout * time.Minute),
	}
}

// GetDefaultTimeout returns the default timeout for service operations.
func GetDefaultTimeout() time.Duration {
	return defaultTimeout * time.Minute
}

const (
	ServiceTypeAlloyDBOmni      = "alloydbomni"
	ServiceTypePG               = "pg"
	ServiceTypeCassandra        = "cassandra"
	ServiceTypeOpenSearch       = "opensearch"
	ServiceTypeGrafana          = "grafana"
	ServiceTypeInfluxDB         = "influxdb"
	ServiceTypeRedis            = "redis"
	ServiceTypeMySQL            = "mysql"
	ServiceTypeKafka            = "kafka"
	ServiceTypeKafkaConnect     = "kafka_connect"
	ServiceTypeKafkaMirrormaker = "kafka_mirrormaker"
	ServiceTypeM3               = "m3db"
	ServiceTypeM3Aggregator     = "m3aggregator"
	ServiceTypeFlink            = "flink"
	ServiceTypeClickhouse       = "clickhouse"
	ServiceTypeDragonfly        = "dragonfly"
	ServiceTypeThanos           = "thanos"
	ServiceTypeValkey           = "valkey"
)

var TechEmailsResourceSchema = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"email": {
			Type:        schema.TypeString,
			Description: "An email address to contact for technical issues",
			Required:    true,
		},
	},
}

func ServiceCommonSchemaWithUserConfig(kind string) map[string]*schema.Schema {
	s := ServiceCommonSchema()
	converters.SetUserConfig(converters.ServiceUserConfig, kind, s)

	// Assigns the integration types that are allowed to be set when creating a service
	integrations := getBootstrapIntegrationTypes(kind)
	integrationType := s["service_integrations"].Elem.(*schema.Resource).Schema["integration_type"]
	switch len(integrations) {
	case 0:
		// Disables the service integrations field if there are no integrations supported
		integrationType.ValidateFunc = func(v any, _ string) ([]string, []error) {
			return nil, []error{fmt.Errorf("service integration %s can't be specified here", v)}
		}
	default:
		integrationType.Description = userconfig.Desc(integrationType.Description).PossibleValuesString(FlattenToString(integrations)...).Build()
		integrationType.ValidateFunc = validation.StringInSlice(FlattenToString(integrations), false)
	}

	return s
}

// getBootstrapIntegrationTypes returns the integration types that are allowed to be set when creating a service.
func getBootstrapIntegrationTypes(kind string) []service.IntegrationType {
	list := make([]service.IntegrationType, 0)

	switch kind {
	case ServiceTypeMySQL, ServiceTypePG, ServiceTypeAlloyDBOmni, ServiceTypeRedis, ServiceTypeValkey:
		list = append(list, service.IntegrationTypeReadReplica)
	}

	if kind == ServiceTypePG {
		list = append(list, service.IntegrationTypeDisasterRecovery)
	}

	return list
}

const diskSpaceDeprecation = "Please use `additional_disk_space` to specify the space to be added to the default disk space defined by the plan."

func ServiceCommonSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"project": CommonSchemaProjectReference,
		"cloud_name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The cloud provider and region the service is hosted in. The format is `provider-region`, for example: `google-europe-west1`. The [available cloud regions](https://aiven.io/docs/platform/reference/list_of_clouds) can differ per project and service. Changing this value [migrates the service to another cloud provider or region](https://aiven.io/docs/platform/howto/migrate-services-cloud-region). The migration runs in the background and includes a DNS update to redirect traffic to the new region. Most services experience no downtime, but some databases may have a brief interruption during DNS propagation.",
			DiffSuppressFunc: func(_, _, newValue string, _ *schema.ResourceData) bool {
				// This is a workaround for a bug when migrating from V3 to V4 Aiven Provider.
				// The bug is that the cloud_name is not set in the state file, but it is set
				// on the API side. This causes a diff during plan, and it will not disappear
				// even after consequent applies. This is because the state is not updated
				// with the cloud_name value.
				return newValue == ""
			},
		},
		"plan": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Defines what kind of computing resources are allocated for the service. It can be changed after creation, though there are some restrictions when going to a smaller plan such as the new plan must have sufficient amount of disk space to store all current data and switching to a plan with fewer nodes might not be supported. The basic plan names are `hobbyist`, `startup-x`, `business-x` and `premium-x` where `x` is (roughly) the amount of memory on each node (also other attributes like number of CPUs and amount of disk space varies but naming is based on memory). The available options can be seen from the [Aiven pricing page](https://aiven.io/pricing).",
		},
		"service_name": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "Specifies the actual name of the service. The name cannot be changed later without destroying and re-creating the service so name should be picked based on intended service usage rather than current attributes.",
		},
		"service_type": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Aiven internal service type code",
		},
		"project_vpc_id": {
			Type:     schema.TypeString,
			Optional: true,
			Description: "Specifies the VPC the service should run in. " +
				"If the value is not set, the service runs on the Public Internet. " +
				"When set, the value should be given as a reference to set up dependencies correctly, " +
				"and the VPC must be in the same cloud and region as the service itself. " +
				"The service can be freely moved to and from VPC after creation, but doing so triggers migration to new servers, " +
				"so the operation can take a significant amount of time to complete if the service has a lot of data.",
		},
		"maintenance_window_dow": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Day of week when maintenance operations should be performed. One monday, tuesday, wednesday, etc.",
			DiffSuppressFunc: func(_, _, newValue string, _ *schema.ResourceData) bool {
				return newValue == ""
			},
			// There is also `never` value, which can't be set, but can be received from the backend.
			// Sending `never` is suppressed in GetMaintenanceWindow function,
			// but then we need to not let to set `never` manually
			ValidateFunc: validation.StringInSlice([]string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}, false),
		},
		"maintenance_window_time": {
			Type:         schema.TypeString,
			Optional:     true,
			Description:  "Time of day when maintenance operations should be performed. UTC time in HH:mm:ss format.",
			ValidateFunc: ValidateMaintenanceWindowTime,
			DiffSuppressFunc: func(_, _, newValue string, _ *schema.ResourceData) bool {
				return newValue == ""
			},
		},
		"maintenance_window_enabled": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Indicates whether the maintenance window is currently enabled for this service.",
		},
		"termination_protection": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Prevents the service from being deleted. It is recommended to set this to `true` for all production services to prevent unintentional service deletion. This does not shield against deleting databases or topics but for services with backups much of the content can at least be restored from backup in case accidental deletion is done.",
		},
		"disk_space": {
			Type:          schema.TypeString,
			Optional:      true,
			Description:   "Service disk space. Possible values depend on the service type, the cloud provider and the project. Therefore, reducing will result in the service rebalancing. " + diskSpaceDeprecation,
			ValidateFunc:  ValidateHumanByteSizeString,
			ConflictsWith: []string{"additional_disk_space"},
			Deprecated:    diskSpaceDeprecation,
		},
		"disk_space_used": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The disk space that the service is currently using. This is the sum of `disk_space` and `additional_disk_space` in human-readable format (for example: `90GiB`).",
		},
		"disk_space_default": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The default disk space of the service, possible values depend on the service type, the cloud provider and the project. Its also the minimum value for `disk_space`",
		},
		"additional_disk_space": {
			Type:          schema.TypeString,
			Optional:      true,
			Computed:      true,
			Description:   "Add [disk storage](https://aiven.io/docs/platform/howto/add-storage-space) in increments of 30  GiB to scale your service. The maximum value depends on the service type and cloud provider. Removing additional storage causes the service nodes to go through a rolling restart, and there might be a short downtime for services without an autoscaler integration or high availability capabilities. The field can be safely removed when autoscaler is enabled without causing any changes.",
			ValidateFunc:  ValidateHumanByteSizeString,
			ConflictsWith: []string{"disk_space"},
		},
		"disk_space_step": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The default disk space step of the service, possible values depend on the service type, the cloud provider and the project. `disk_space` needs to increment from `disk_space_default` by increments of this size.",
		},
		"disk_space_cap": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The maximum disk space of the service, possible values depend on the service type, the cloud provider and the project.",
		},
		"service_uri": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "URI for connecting to the service. Service specific info is under \"kafka\", \"pg\", etc.",
			Sensitive:   true,
		},
		"service_host": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The hostname of the service.",
		},
		"service_port": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The port of the service",
		},
		"service_password": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Password used for connecting to the service, if applicable",
			Sensitive:   true,
		},
		"service_username": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Username used for connecting to the service, if applicable",
		},
		"state": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Service state. Possible values are `POWEROFF`, `REBALANCING`, `REBUILDING` or `RUNNING`. Services cannot be powered on or off with Terraform. To power a service on or off, [use the Aiven Console or Aiven CLI](https://aiven.io/docs/platform/concepts/service-power-cycle).",
		},
		"service_integrations": {
			Type:        schema.TypeSet,
			Optional:    true,
			Computed:    true,
			Description: "Service integrations to specify when creating a service. Not applied after initial service creation",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"source_service_name": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Name of the source service",
					},
					"integration_type": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Type of the service integration",
					},
				},
			},
		},
		"static_ips": {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "Static IPs that are going to be associated with this service. Please assign a value using the 'toset' function. Once a static ip resource is in the 'assigned' state it cannot be unbound from the node again",
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"components": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "Service component information objects",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"component": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Service component name",
					},
					"host": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Host name for connecting to the service component",
					},
					"port": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "Port number for connecting to the service component",
					},
					"connection_uri": {
						Type:     schema.TypeString,
						Computed: true,
						Description: "Connection info for connecting to the service component." +
							" This is a combination of host and port.",
					},
					"kafka_authentication_method": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Kafka authentication method. This is a value specific to the 'kafka' service component",
					},
					"kafka_ssl_ca": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Kafka certificate used. The possible values are `letsencrypt` and `project_ca`.",
					},
					"route": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Network access route",
					},
					"ssl": {
						Type:     schema.TypeBool,
						Computed: true,
						Description: "Whether the endpoint is encrypted or accepts plaintext. By default endpoints are " +
							"always encrypted and this property is only included for service components they may " +
							"disable encryption",
					},
					"usage": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "DNS usage name",
					},
				},
			},
		},
		"tag": {
			Description: "Tags are key-value pairs that allow you to categorize services.",
			Type:        schema.TypeSet,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"key": {
						Description: "Service tag key",
						Type:        schema.TypeString,
						Required:    true,
					},
					"value": {
						Description: "Service tag value",
						Type:        schema.TypeString,
						Required:    true,
					},
				},
			},
		},
		"tech_emails": {
			Type:        schema.TypeSet,
			Elem:        TechEmailsResourceSchema,
			Optional:    true,
			Description: " The email addresses for [service contacts](https://aiven.io/docs/platform/howto/technical-emails), who will receive important alerts and updates about this service. You can also set email contacts at the project level.",
		},
	}
}

func ResourceServiceCreateWrapper(serviceType string) schema.CreateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		if err := d.Set("service_type", serviceType); err != nil {
			return diag.Errorf("error setting service_type: %s", err)
		}
		return resourceServiceCreate(ctx, d, m)
	}
}

func ResourceServiceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, err := SplitResourceID2(d.Id())
	if err != nil {
		return diag.Errorf("error splitting service ID: %s", err)
	}

	avnGen, err := common.GenClient()
	if err != nil {
		return diag.FromErr(err)
	}

	s, err := avnGen.ServiceGet(ctx, projectName, serviceName, service.ServiceGetIncludeSecrets(true))
	if err != nil {
		if err = ResourceReadHandleNotFound(err, d); err != nil {
			return diag.Errorf("unable to GET service %s: %s", d.Id(), err)
		}
		return nil
	}

	// The GET with include_secrets=true must not return redacted creds
	err = ContainsRedactedCreds(s.UserConfig)
	if err != nil {
		return diag.FromErr(err)
	}

	servicePlanParams, err := GetServicePlanParametersFromServiceResponse(ctx, client, projectName, s)
	if err != nil {
		return diag.Errorf("unable to get service plan parameters: %s", err)
	}

	err = copyServicePropertiesFromAPIResponseToTerraform(d, s, servicePlanParams, projectName)
	if err != nil {
		return diag.Errorf("unable to copy api response into terraform schema: %s", err)
	}

	serviceIps, err := ServiceStaticIpsList(ctx, avnGen, projectName, serviceName)
	if err != nil {
		return diag.Errorf("unable to currently allocated static ips: %s", err)
	}
	if err = d.Set("static_ips", serviceIps); err != nil {
		return diag.Errorf("unable to set static ips field in schema: %s", err)
	}

	tags, err := avnGen.ProjectServiceTagsList(ctx, projectName, serviceName)
	if err != nil {
		return diag.Errorf("unable to get service tags: %s", err)
	}

	if err := d.Set("tag", SetTagsTerraformProperties(tags)); err != nil {
		return diag.Errorf("unable to set tag's in schema: %s", err)
	}

	if err := d.Set("tech_emails", getTechnicalEmailsForTerraform(s)); err != nil {
		return diag.Errorf("unable to set tech_emails in schema: %s", err)
	}

	var diags diag.Diagnostics
	for _, v := range s.ServiceNotifications {
		if v.Type == service.ServiceNotificationTypeServiceEndOfLife {
			const detail = "See the [documentation](%s) for more information on end of life for Aiven services."
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  v.Message,
				Detail:   fmt.Sprintf(detail, lo.FromPtr(v.Metadata.EndOfLifeHelpArticleUrl)),
			})
		}
	}

	if timeoutWarning := common.CheckDeprecatedTimeoutDefault(d); timeoutWarning.Severity != 0 {
		diags = append(diags, timeoutWarning)
	}

	return diags
}

func resourceServiceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)
	avnGen, err := common.GenClient()
	if err != nil {
		return diag.FromErr(err)
	}

	serviceType := d.Get("service_type").(string)
	project := d.Get("project").(string)

	diskSpace, err := getDiskSpaceFromStateOrDiff(ctx, d, client)
	if err != nil {
		return diag.Errorf("error getting default disc space: %s", err)
	}

	vpcID, err := GetProjectVPCIdPointer(d)
	if err != nil {
		return diag.Errorf("error getting project VPC ID: %s", err)
	}

	cuc, err := ExpandService(serviceType, d)
	if err != nil {
		return diag.FromErr(err)
	}

	technicalEmails, err := getContactEmailListForAPI(d)
	if err != nil {
		return diag.FromErr(err)
	}

	cloud := d.Get("cloud_name").(string)
	terminationProtection := d.Get("termination_protection").(bool)
	staticIps := FlattenToString(d.Get("static_ips").(*schema.Set).List())
	serviceIntegrations := GetAPIServiceIntegrations(d)

	var diskSpaceMb *int
	if diskSpace > 0 {
		diskSpaceMb = &diskSpace
	}

	serviceCreate := &service.ServiceCreateIn{
		Cloud:                 &cloud,
		Plan:                  d.Get("plan").(string),
		ProjectVpcId:          vpcID,
		ServiceIntegrations:   &serviceIntegrations,
		Maintenance:           GetMaintenanceWindow(d),
		ServiceName:           d.Get("service_name").(string),
		ServiceType:           serviceType,
		TerminationProtection: &terminationProtection,
		DiskSpaceMb:           diskSpaceMb,
		UserConfig:            &cuc,
		StaticIps:             &staticIps,
		TechEmails:            technicalEmails,
	}

	if _, err = avnGen.ServiceCreate(ctx, project, serviceCreate); err != nil {
		return diag.Errorf("error creating a service: %s", err)
	}

	// Create already takes care of static ip associations, no need to explictely associate them here
	s, err := WaitForServiceCreation(ctx, d, avnGen)
	if err != nil {
		return diag.Errorf("error waiting for service creation: %s", err)
	}

	_, err = client.ServiceTags.Set(ctx, project, d.Get("service_name").(string), aiven.ServiceTagsRequest{
		Tags: GetTagsFromSchema(d),
	})
	if err != nil {
		return diag.Errorf("error setting service tags: %s", err)
	}

	d.SetId(BuildResourceID(project, s.ServiceName))

	return ResourceServiceRead(ctx, d, m)
}

func ResourceServiceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)
	avnGen, err := common.GenClient()
	if err != nil {
		return diag.FromErr(err)
	}

	var karapace *bool
	if v := d.Get("karapace"); d.HasChange("karapace") && v != nil {
		if k, ok := v.(bool); ok && k {
			karapace = &k
		}
	}

	projectName, serviceName, err := SplitResourceID2(d.Id())
	if err != nil {
		return diag.Errorf("error splitting service id (%s): %s", d.Id(), err)
	}

	ass, dis, err := DiffStaticIps(ctx, d, avnGen)
	if err != nil {
		return diag.Errorf("error diff static ips: %s", err)
	}

	// associate first, so that we can enable `static_ips` for a preexisting common
	for _, aip := range ass {
		if _, err := avnGen.ProjectStaticIPAssociate(ctx, projectName, aip, &staticip.ProjectStaticIpassociateIn{
			ServiceName: serviceName,
		}); err != nil {
			return diag.Errorf("error associating Static IP (%s) to a service: %s", aip, err)
		}
	}

	var vpcID *string
	vpcID, err = GetProjectVPCIdPointer(d)
	if err != nil {
		return diag.Errorf("error getting project VPC ID: %s", err)
	}

	serviceType := d.Get("service_type").(string)
	cuc, err := ExpandService(serviceType, d)
	if err != nil {
		return diag.FromErr(err)
	}
	technicalEmails, err := getContactEmailListForAPI(d)
	if err != nil {
		return diag.FromErr(err)
	}

	cloud := d.Get("cloud_name").(string)
	plan := d.Get("plan").(string)
	powered := true
	terminationProtection := d.Get("termination_protection").(bool)

	// Sends disk size only when there is no autoscaler enabled
	var diskSpaceMb *int
	s, err := avnGen.ServiceGet(ctx, projectName, serviceName)
	if err != nil {
		return nil
	}

	if len(flattenIntegrations(s.ServiceIntegrations, service.IntegrationTypeAutoscaler)) == 0 {
		diskSpace, err := getDiskSpaceFromStateOrDiff(ctx, d, client)
		if err != nil {
			return diag.Errorf("error getting default disc space: %s", err)
		}

		if diskSpace > 0 {
			diskSpaceMb = &diskSpace
		}
	}

	serviceUpdate := &service.ServiceUpdateIn{
		Cloud:                 &cloud,
		Plan:                  &plan,
		Maintenance:           GetMaintenanceWindow(d),
		ProjectVpcId:          vpcID,
		Powered:               &powered,
		TerminationProtection: &terminationProtection,
		DiskSpaceMb:           diskSpaceMb,
		Karapace:              karapace,
		UserConfig:            &cuc,
		TechEmails:            technicalEmails,
	}

	if _, err := avnGen.ServiceUpdate(ctx, projectName, serviceName, serviceUpdate); err != nil {
		return diag.Errorf("error updating (%s) service: %s", serviceName, err)
	}

	if _, err = WaitForServiceUpdate(ctx, d, avnGen); err != nil {
		return diag.Errorf("error waiting for service (%s) update: %s", serviceName, err)
	}

	if len(dis) > 0 {
		for _, dip := range dis {
			if _, err := avnGen.ProjectStaticIPDissociate(ctx, projectName, dip); err != nil {
				return diag.Errorf("error dissociating Static IP (%s) from the service (%s): %s", dip, serviceName, err)
			}
		}
		if err = WaitStaticIpsDissociation(ctx, d, m); err != nil {
			return diag.Errorf("error waiting for Static IPs dissociation: %s", err)
		}
	}

	_, err = client.ServiceTags.Set(ctx, projectName, serviceName, aiven.ServiceTagsRequest{
		Tags: GetTagsFromSchema(d),
	})
	if err != nil {
		return diag.Errorf("error setting service tags: %s", err)
	}

	return ResourceServiceRead(ctx, d, m)
}

// getDiskSpaceFromStateOrDiff three cases:
// 1. disk_space is set
// 2. plan disk space
// 3. plan disk space + additional_disk_space
func getDiskSpaceFromStateOrDiff(ctx context.Context, d ResourceStateOrResourceDiff, client *aiven.Client) (int, error) {
	if v, ok := d.GetOk("disk_space"); ok {
		return ConvertToDiskSpaceMB(v.(string)), nil
	}

	// Get service plan specific defaults
	plan, err := GetServicePlanParametersFromSchema(ctx, client, d)
	if err != nil {
		if aiven.IsNotFound(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("unable to get service plan parameters: %w", err)
	}

	// Adds additional_disk_space only if it is in the config
	diskSpace := plan.DiskSizeMBDefault
	if HasConfigValue(d, "additional_disk_space") {
		diskSpace += ConvertToDiskSpaceMB(d.Get("additional_disk_space").(string))
	}

	return diskSpace, nil
}

func getTechnicalEmailsForTerraform(s *service.ServiceGetOut) *schema.Set {
	if len(s.TechEmails) == 0 {
		return nil
	}

	techEmails := make([]interface{}, len(s.TechEmails))
	for i, e := range s.TechEmails {
		techEmails[i] = map[string]interface{}{"email": e.Email}
	}

	return schema.NewSet(schema.HashResource(TechEmailsResourceSchema), techEmails)
}

// flattenIntegrations converts the service integrations into a list of maps
func flattenIntegrations(integrations []service.ServiceIntegrationOut, kinds ...service.IntegrationType) []map[string]interface{} {
	result := make([]map[string]any, 0)
	if len(integrations) == 0 || len(kinds) == 0 {
		return result
	}

	for _, v := range integrations {
		if slices.Contains(kinds, v.IntegrationType) {
			result = append(result, map[string]any{
				"integration_type":    v.IntegrationType,
				"source_service_name": v.SourceService,
			})
		}
	}
	return result
}

func ResourceServiceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, err := SplitResourceID2(d.Id())
	if err != nil {
		return diag.Errorf("error splitting service ID: %s", err)
	}

	if err := client.Services.Delete(ctx, projectName, serviceName); err != nil && !aiven.IsNotFound(err) {
		return diag.Errorf("error deleting a service: %s", err)
	}

	// Delete already takes care of static IPs disassociation; no need to explicitly disassociate them here

	if err := WaitForDeletion(ctx, d, m); err != nil {
		return diag.Errorf("error waiting for service deletion: %s", err)
	}
	return nil
}

func copyServicePropertiesFromAPIResponseToTerraform(
	d ResourceData,
	s *service.ServiceGetOut,
	servicePlanParams PlanParameters,
	project string,
) error {
	serviceType := d.Get("service_type").(string)
	if _, ok := d.GetOk("service_type"); !ok {
		serviceType = s.ServiceType
	}

	if err := d.Set("cloud_name", s.CloudName); err != nil {
		return err
	}
	if err := d.Set("service_name", s.ServiceName); err != nil {
		return err
	}
	if err := d.Set("state", s.State); err != nil {
		return err
	}
	if err := d.Set("plan", s.Plan); err != nil {
		return err
	}
	if err := d.Set("service_type", serviceType); err != nil {
		return err
	}
	if err := d.Set("termination_protection", s.TerminationProtection); err != nil {
		return err
	}
	if err := d.Set("maintenance_window_dow", s.Maintenance.Dow); err != nil {
		return err
	}
	if err := d.Set("maintenance_window_time", s.Maintenance.Time); err != nil {
		return err
	}
	if err := d.Set("maintenance_window_enabled", s.Maintenance.Enabled); err != nil {
		return err
	}

	diskSpace := 0
	if s.DiskSpaceMb != nil {
		diskSpace = *s.DiskSpaceMb
	}

	additionalDiskSpace := diskSpace - servicePlanParams.DiskSizeMBDefault
	if err := d.Set("additional_disk_space", HumanReadableByteSize(additionalDiskSpace*units.MiB)); err != nil {
		return err
	}

	_, isDiskSpaceSet := d.GetOk("disk_space")
	if isDiskSpaceSet && diskSpace > 0 {
		if err := d.Set("disk_space", HumanReadableByteSize(diskSpace*units.MiB)); err != nil {
			return err
		}
	}

	if err := d.Set("disk_space_used", HumanReadableByteSize(diskSpace*units.MiB)); err != nil {
		return err
	}
	if err := d.Set("disk_space_default", HumanReadableByteSize(servicePlanParams.DiskSizeMBDefault*units.MiB)); err != nil {
		return err
	}
	if err := d.Set("disk_space_step", HumanReadableByteSize(servicePlanParams.DiskSizeMBStep*units.MiB)); err != nil {
		return err
	}
	if err := d.Set("disk_space_cap", HumanReadableByteSize(servicePlanParams.DiskSizeMBMax*units.MiB)); err != nil {
		return err
	}
	if err := d.Set("service_uri", s.ServiceUri); err != nil {
		return err
	}
	if err := d.Set("project", project); err != nil {
		return err
	}

	if err := d.Set("tech_emails", getTechnicalEmailsForTerraform(s)); err != nil {
		return err
	}

	if s.ProjectVpcId != "" {
		// Historically, the project VPC ID was stored as a full resource ID.
		if err := d.Set("project_vpc_id", BuildResourceID(project, s.ProjectVpcId)); err != nil {
			return err
		}
	} else {
		if err := d.Set("project_vpc_id", nil); err != nil {
			return err
		}
	}

	err := FlattenService(serviceType, d, s.UserConfig)
	if err != nil {
		return err
	}

	params := s.ServiceUriParams
	if err := d.Set("service_host", params["host"]); err != nil {
		return err
	}

	port, _ := strconv.ParseInt(fmt.Sprintf("%v", params["port"]), 10, 32)
	if err := d.Set("service_port", port); err != nil {
		return err
	}

	password, passwordOK := params["password"]
	username, usernameOK := params["user"]
	if passwordOK {
		if err := d.Set("service_password", password); err != nil {
			return err
		}
	}
	if usernameOK {
		if err := d.Set("service_username", username); err != nil {
			return err
		}
	}

	// for some services, for example Kafka URIParams does not provide default user credentials
	if !passwordOK || !usernameOK {
		for _, u := range s.Users {
			if u.Username == "avnadmin" {
				if err := d.Set("service_username", u.Username); err != nil {
					return err
				}
				if err := d.Set("service_password", u.Password); err != nil {
					return err
				}
			}
		}
	}

	if err := d.Set("components", FlattenServiceComponents(s)); err != nil {
		return fmt.Errorf("cannot set `components` : %w", err)
	}

	// Handle service integrations
	integrations := flattenIntegrations(s.ServiceIntegrations, getBootstrapIntegrationTypes(serviceType)...)
	if err := d.Set("service_integrations", integrations); err != nil {
		return err
	}

	return copyConnectionInfoFromAPIResponseToTerraform(d, serviceType, s.ConnectionInfo, s.Metadata)
}

func FlattenServiceComponents(r *service.ServiceGetOut) []map[string]interface{} {
	components := make([]map[string]interface{}, len(r.Components))

	for i, c := range r.Components {
		component := map[string]interface{}{
			"component":                   c.Component,
			"host":                        c.Host,
			"port":                        c.Port,
			"connection_uri":              fmt.Sprintf("%s:%d", c.Host, c.Port),
			"kafka_authentication_method": c.KafkaAuthenticationMethod,
			"kafka_ssl_ca":                c.KafkaSslCa,
			"route":                       c.Route,
			// By default, endpoints are always encrypted and
			// this property is only included for service components that may disable encryption.
			"ssl":   PointerValueOrDefault(c.Ssl, true),
			"usage": c.Usage,
		}
		components[i] = component
	}

	return components
}

// TODO: This uses an untyped map in the final resource's schema, which might be unclear to the end users.
//
//	We should change this in the next major version.
func copyConnectionInfoFromAPIResponseToTerraform(
	d ResourceData,
	serviceType string,
	connectionInfo *service.ConnectionInfoOut,
	metadata map[string]any,
) error {
	props := make(map[string]any)

	switch serviceType {
	case ServiceTypeKafka:
		// KafkaHosts will be renamed to KafkaURIs in the next major version of the Go client.
		// That's why we name the props key `uris` here.
		props["uris"] = connectionInfo.Kafka
		setProp(props, "access_cert", connectionInfo.KafkaAccessCert)
		setProp(props, "access_key", connectionInfo.KafkaAccessKey)
		setProp(props, "connect_uri", connectionInfo.KafkaConnectUri)
		setProp(props, "rest_uri", connectionInfo.KafkaRestUri)
		setProp(props, "schema_registry_uri", connectionInfo.SchemaRegistryUri)
	case ServiceTypeAlloyDBOmni, ServiceTypePG:
		// For compatibility with the old schema, we only set the first URI.
		// TODO: Remove this block in the next major version. Keep `uris` key only, see below.
		if len(connectionInfo.Pg) > 0 {
			props["uri"] = connectionInfo.Pg[0]
		}

		props["uris"] = connectionInfo.Pg
		params := make([]map[string]any, len(connectionInfo.PgParams))
		for i, p := range connectionInfo.PgParams {
			port, err := strconv.ParseInt(p.Port, 10, 32)
			if err != nil {
				return err
			}

			if i == 0 {
				// For compatibility with the old schema, we only set the first params.
				// TODO: Remove this block in the next major version. Keep `params` key only.
				props["host"] = p.Host
				props["port"] = port
				props["sslmode"] = p.Sslmode
				props["user"] = p.User
				props["password"] = p.Password
				props["dbname"] = p.Dbname
			}

			params[i] = map[string]any{
				"host":          p.Host,
				"port":          port,
				"sslmode":       p.Sslmode,
				"user":          p.User,
				"password":      p.Password,
				"database_name": p.Dbname,
			}
		}

		props["params"] = params
		props["standby_uris"] = connectionInfo.PgStandby
		props["syncing_uris"] = connectionInfo.PgSyncing
		setProp(props, "replica_uri", connectionInfo.PgReplicaUri)

		// TODO: This isn't in the connection info, but it's in the metadata.
		//  We should move this to the other part of the schema in the next major version.
		props["max_connections"] = metadata["max_connections"]
	case ServiceTypeThanos:
		props["uris"] = connectionInfo.Thanos
		setProp(props, "query_frontend_uri", connectionInfo.QueryFrontendUri)
		setProp(props, "query_uri", connectionInfo.QueryUri)
		setProp(props, "receiver_remote_write_uri", connectionInfo.ReceiverRemoteWriteUri)
	case ServiceTypeMySQL:
		props["uris"] = connectionInfo.Mysql

		params := make([]map[string]any, len(connectionInfo.MysqlParams))
		for i, p := range connectionInfo.MysqlParams {
			port, err := strconv.ParseInt(p.Port, 10, 32)
			if err != nil {
				return err
			}

			params[i] = map[string]any{
				"host":          p.Host,
				"port":          port,
				"sslmode":       p.SslMode,
				"user":          p.User,
				"password":      p.Password,
				"database_name": p.Dbname,
			}
		}

		props["params"] = params
		props["standby_uris"] = connectionInfo.MysqlStandby
		setProp(props, "replica_uri", connectionInfo.MysqlReplicaUri)
	case ServiceTypeOpenSearch:
		props["uris"] = connectionInfo.Opensearch

		// TODO: Remove `opensearch_` prefix in the next major version.
		setProp(props, "opensearch_dashboards_uri", connectionInfo.OpensearchDashboardsUri)
		setProp(props, "username", connectionInfo.OpensearchUsername)
		setProp(props, "password", connectionInfo.OpensearchPassword)
	case ServiceTypeCassandra:
		// CassandraHosts will be renamed to CassandraURIs in the next major version of the Go client.
		// That's why we name the props key `uris` here.
		props["uris"] = connectionInfo.Cassandra
	case ServiceTypeRedis, ServiceTypeDragonfly:
		props["uris"] = connectionInfo.Redis
		props["slave_uris"] = connectionInfo.RedisSlave
		setProp(props, "replica_uri", connectionInfo.RedisReplicaUri)
		setProp(props, "password", connectionInfo.RedisPassword)
	case ServiceTypeValkey:
		props["uris"] = connectionInfo.Valkey
		props["slave_uris"] = connectionInfo.ValkeySlave
		setProp(props, "replica_uri", connectionInfo.ValkeyReplicaUri)
		setProp(props, "password", connectionInfo.ValkeyPassword)
	case ServiceTypeInfluxDB:
		props["uris"] = connectionInfo.Influxdb
		setProp(props, "username", connectionInfo.InfluxdbUsername)
		setProp(props, "password", connectionInfo.InfluxdbPassword)
		setProp(props, "database_name", connectionInfo.InfluxdbDbname)
	case ServiceTypeGrafana:
		props["uris"] = connectionInfo.Grafana
	case ServiceTypeM3:
		props["uris"] = connectionInfo.M3Db
		setProp(props, "http_cluster_uri", connectionInfo.HttpClusterUri)
		setProp(props, "http_node_uri", connectionInfo.HttpNodeUri)
		setProp(props, "influxdb_uri", connectionInfo.InfluxdbUri)
		setProp(props, "prometheus_remote_read_uri", connectionInfo.PrometheusRemoteReadUri)
		setProp(props, "prometheus_remote_write_uri", connectionInfo.PrometheusRemoteWriteUri)
	case ServiceTypeM3Aggregator:
		props["uris"] = connectionInfo.M3Aggregator
		setProp(props, "aggregator_http_uri", connectionInfo.AggregatorHttpUri)
	case ServiceTypeClickhouse:
		props["uris"] = connectionInfo.Clickhouse
	case ServiceTypeFlink:
		// TODO: Rename `host_ports` to `uris` in the next major version.
		props["host_ports"] = connectionInfo.Flink
	default:
		// Doesn't have connection info
		return nil
	}

	return d.Set(serviceType, []map[string]any{props})
}

func setProp[T comparable](m map[string]any, k string, v *T) {
	if v != nil {
		m[k] = *v
	}
}

// NewNotFound creates a new not found error
// There are lots of endpoints that return a list of objects which might not contain the object we are looking for.
// In this case, we should still return 404.
func NewNotFound(msg string, args ...any) error {
	return aiven.Error{Status: http.StatusNotFound, Message: fmt.Sprintf(msg, args...)}
}

func IsNotFound(err error) bool {
	return aiven.IsNotFound(err) || avngen.IsNotFound(err)
}

func OmitNotFound(err error) error {
	if IsNotFound(err) {
		return nil
	}
	return err
}

// IsUnknownRole checks if the database returned an error because of an unknown role
// to make deletions idempotent
func IsUnknownRole(err error) bool {
	var oldError aiven.Error
	var newError avngen.Error
	var msg string
	switch {
	case errors.As(err, &oldError):
		msg = oldError.Message
	case errors.As(err, &newError):
		msg = newError.Message
	}
	return strings.Contains(msg, "Code: 511")
}

// IsUnknownResource is a function to handle errors that we want to treat as "Not Found"
func IsUnknownResource(err error) bool {
	return IsNotFound(err) || IsUnknownRole(err)
}

func ResourceReadHandleNotFound(err error, d ResourceData) error {
	if err != nil && IsUnknownResource(err) && !d.IsNewResource() {
		d.SetId("")
		return nil
	}
	return err
}

func DatasourceServiceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	d.SetId(BuildResourceID(projectName, serviceName))

	services, err := client.Services.List(ctx, projectName)
	if err != nil {
		return diag.Errorf("error getting a list of services: %s", err)
	}

	for _, s := range services {
		if s.Name == serviceName {
			return ResourceServiceRead(ctx, d, m)
		}
	}

	return diag.Errorf("common %s/%s not found", projectName, serviceName)
}

func getContactEmailListForAPI(d ResourceData) (*[]service.TechEmailIn, error) {
	if valuesInterface, ok := d.GetOk("tech_emails"); ok {
		var emails []service.TechEmailIn
		err := Remarshal(valuesInterface.(*schema.Set).List(), &emails)
		if err != nil {
			return nil, err // Handle error appropriately
		}
		return &emails, nil
	}
	return &[]service.TechEmailIn{}, nil
}

func ExpandService(name string, d ResourceData) (map[string]any, error) {
	return converters.Expand(converters.ServiceUserConfig, name, d)
}

func FlattenService(name string, d ResourceData, dto map[string]any) error {
	return converters.Flatten(converters.ServiceUserConfig, name, d, dto)
}

const redactedSubstr = `\u003credacted\u003e`

var errContainsRedactedCreds = fmt.Errorf("unexpected redacted credentials")

// ContainsRedactedCreds looks for redactedSubstr in the given config
func ContainsRedactedCreds(config map[string]any) error {
	b, err := json.Marshal(&config)
	if err != nil {
		return err
	}

	if bytes.Contains(b, []byte(redactedSubstr)) {
		return errContainsRedactedCreds
	}
	return nil
}

var (
	servicePoweredMap    sync.Map
	servicePoweredGroup  DoOnce
	ErrServicePoweredOff = fmt.Errorf("the service is powered off")
)

// ServicePoweredForget for test purposes only.
func ServicePoweredForget(project, serviceName string) {
	key := filepath.Join(project, serviceName)
	servicePoweredGroup.Forget(key)
	servicePoweredMap.Delete(key)
}

// CheckServiceIsPowered checks if a service is powered on before performing operations that require it.
// Some operations like database management require the service to be running.
// For example, `ServiceDatabaseList` returns a generic 503 error when the service is powered off:
// "503: An error occurred. Please try again later."
// This function provides a clearer error message by explicitly checking the service power state first.
func CheckServiceIsPowered(ctx context.Context, client avngen.Client, project, serviceName string) error {
	key := filepath.Join(project, serviceName)
	err := servicePoweredGroup.Do(key, func() error {
		s, err := client.ServiceGet(ctx, project, serviceName)
		if err == nil {
			servicePoweredMap.Store(key, s.State != service.ServiceStateTypePoweroff)
		}
		return err
	})
	if err != nil {
		return err
	}

	isOn, _ := servicePoweredMap.Load(key)
	if !isOn.(bool) {
		return ErrServicePoweredOff
	}
	return nil
}
