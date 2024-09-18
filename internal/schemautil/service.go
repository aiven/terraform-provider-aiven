package schemautil

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/docker/go-units"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/userconfig/converters"
)

// defaultTimeout is the default timeout for service operations. This is not a const because it can be changed during
// compile time with -ldflags "-X github.com/aiven/terraform-provider-aiven/internal/schemautil.defaultTimeout=30".
var defaultTimeout time.Duration = 20

func DefaultResourceTimeouts() *schema.ResourceTimeout {
	return &schema.ResourceTimeout{
		Create:  schema.DefaultTimeout(defaultTimeout * time.Minute),
		Update:  schema.DefaultTimeout(defaultTimeout * time.Minute),
		Delete:  schema.DefaultTimeout(defaultTimeout * time.Minute),
		Default: schema.DefaultTimeout(defaultTimeout * time.Minute),
		Read:    schema.DefaultTimeout(defaultTimeout * time.Minute),
	}
}

const (
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
	return s
}

func ServiceCommonSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"project": CommonSchemaProjectReference,
		"cloud_name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Defines where the cloud provider and region where the service is hosted in. This can be changed freely after service is created. Changing the value will trigger a potentially lengthy migration process for the service. Format is cloud provider name (`aws`, `azure`, `do` `google`, `upcloud`, etc.), dash, and the cloud provider specific region name. These are documented on each Cloud provider's own support articles, like [here for Google](https://cloud.google.com/compute/docs/regions-zones/) and [here for AWS](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.RegionsAndAvailabilityZones.html).",
			DiffSuppressFunc: func(_, _, new string, _ *schema.ResourceData) bool {
				// This is a workaround for a bug when migrating from V3 to V4 Aiven Provider.
				// The bug is that the cloud_name is not set in the state file, but it is set
				// on the API side. This causes a diff during plan, and it will not disappear
				// even after consequent applies. This is because the state is not updated
				// with the cloud_name value.
				return new == ""
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
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Specifies the VPC the service should run in. If the value is not set the service is not run inside a VPC. When set, the value should be given as a reference to set up dependencies correctly and the VPC must be in the same cloud and region as the service itself. Project can be freely moved to and from VPC after creation but doing so triggers migration to new servers so the operation can take significant amount of time to complete if the service has a lot of data.",
		},
		"maintenance_window_dow": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Day of week when maintenance operations should be performed. One monday, tuesday, wednesday, etc.",
			DiffSuppressFunc: func(_, _, new string, _ *schema.ResourceData) bool {
				return new == ""
			},
			// There is also `never` value, which can't be set, but can be received from the backend.
			// Sending `never` is suppressed in GetMaintenanceWindow function,
			// but then we need to not let to set `never` manually
			ValidateFunc: validation.StringInSlice([]string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}, false),
		},
		"maintenance_window_time": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Time of day when maintenance operations should be performed. UTC time in HH:mm:ss format.",
			DiffSuppressFunc: func(_, _, new string, _ *schema.ResourceData) bool {
				return new == ""
			},
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
			Description:   "Service disk space. Possible values depend on the service type, the cloud provider and the project. Therefore, reducing will result in the service rebalancing.",
			ValidateFunc:  ValidateHumanByteSizeString,
			ConflictsWith: []string{"additional_disk_space"},
			Deprecated:    "This will be removed in v5.0.0. Please use `additional_disk_space` to specify the space to be added to the default `disk_space` defined by the plan.",
		},
		"disk_space_used": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Disk space that service is currently using",
		},
		"disk_space_default": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The default disk space of the service, possible values depend on the service type, the cloud provider and the project. Its also the minimum value for `disk_space`",
		},
		"additional_disk_space": {
			Type:          schema.TypeString,
			Optional:      true,
			Description:   "Add [disk storage](https://aiven.io/docs/platform/howto/add-storage-space) in increments of 30  GiB to scale your service. The maximum value depends on the service type and cloud provider. Removing additional storage causes the service nodes to go through a rolling restart and there might be a short downtime for services with no HA capabilities.",
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
			Description: "Service state. One of `POWEROFF`, `REBALANCING`, `REBUILDING` or `RUNNING`",
		},
		"service_integrations": {
			Type:        schema.TypeList,
			Optional:    true,
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
						Description: "Type of the service integration. The only supported value at the moment is `read_replica`",
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

	return nil
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

	_, err = client.Services.Create(
		ctx,
		project,
		aiven.CreateServiceRequest{
			Cloud:                 d.Get("cloud_name").(string),
			Plan:                  d.Get("plan").(string),
			ProjectVPCID:          vpcID,
			ServiceIntegrations:   GetAPIServiceIntegrations(d),
			MaintenanceWindow:     GetMaintenanceWindow(d),
			ServiceName:           d.Get("service_name").(string),
			ServiceType:           serviceType,
			TerminationProtection: d.Get("termination_protection").(bool),
			DiskSpaceMB:           diskSpace,
			UserConfig:            cuc,
			StaticIPs:             FlattenToString(d.Get("static_ips").(*schema.Set).List()),
			TechnicalEmails:       technicalEmails,
		},
	)
	if err != nil {
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

	// On service update, we send a default disc space value for a common
	// if the TF user does not specify it
	diskSpace, err := getDiskSpaceFromStateOrDiff(ctx, d, client)
	if err != nil {
		return diag.Errorf("error getting default disc space: %s", err)
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
		if err := client.StaticIPs.Associate(ctx, projectName, aip, aiven.AssociateStaticIPRequest{ServiceName: serviceName}); err != nil {
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

	if _, err := client.Services.Update(
		ctx,
		projectName,
		serviceName,
		aiven.UpdateServiceRequest{
			Cloud:                 d.Get("cloud_name").(string),
			Plan:                  d.Get("plan").(string),
			MaintenanceWindow:     GetMaintenanceWindow(d),
			ProjectVPCID:          vpcID,
			Powered:               true,
			TerminationProtection: d.Get("termination_protection").(bool),
			DiskSpaceMB:           diskSpace,
			Karapace:              karapace,
			UserConfig:            cuc,
			TechnicalEmails:       technicalEmails,
		},
	); err != nil {
		return diag.Errorf("error updating (%s) service: %s", serviceName, err)
	}

	if _, err = WaitForServiceUpdate(ctx, d, avnGen); err != nil {
		return diag.Errorf("error waiting for service (%s) update: %s", serviceName, err)
	}

	if len(dis) > 0 {
		for _, dip := range dis {
			if err := client.StaticIPs.Dissociate(ctx, projectName, dip); err != nil {
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

func getDiskSpaceFromStateOrDiff(ctx context.Context, d ResourceStateOrResourceDiff, client *aiven.Client) (int, error) {
	var diskSpace int

	// Get service plan specific defaults
	servicePlanParams, err := GetServicePlanParametersFromSchema(ctx, client, d)
	if err != nil {
		if aiven.IsNotFound(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("unable to get service plan parameters: %w", err)
	}

	// Use `additional_disk_space` if set
	if ads, ok := d.GetOk("additional_disk_space"); ok {
		diskSpace = servicePlanParams.DiskSizeMBDefault + ConvertToDiskSpaceMB(ads.(string))
	} else if ds, ok := d.GetOk("disk_space"); ok {
		// Use `disk_space` if set...
		diskSpace = ConvertToDiskSpaceMB(ds.(string))
	} else {
		// ... otherwise, use the default disk space
		diskSpace = servicePlanParams.DiskSizeMBDefault
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

func getReadReplicaIntegrationsForTerraform(integrations []service.ServiceIntegrationOut) ([]map[string]interface{}, error) {
	var readReplicaIntegrations []map[string]interface{}
	for _, integration := range integrations {
		if integration.IntegrationType == "read_replica" {
			integrationMap := map[string]interface{}{
				"integration_type":    integration.IntegrationType,
				"source_service_name": integration.SourceService,
			}
			readReplicaIntegrations = append(readReplicaIntegrations, integrationMap)
		}
	}
	return readReplicaIntegrations, nil
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
	d *schema.ResourceData,
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

	diskSpace := 0
	if s.DiskSpaceMb != nil {
		diskSpace = int(*s.DiskSpaceMb)
	}
	additionalDiskSpace := diskSpace - servicePlanParams.DiskSizeMBDefault

	_, isAdditionalDiskSpaceSet := d.GetOk("additional_disk_space")
	_, isDiskSpaceSet := d.GetOk("disk_space")

	// Handles two different cases:
	//
	// 1. During import when neither `additional_disk_space` nor `disk_space` are set
	// 2. During create / update when `additional_disk_space` is set
	if additionalDiskSpace > 0 && (!isDiskSpaceSet || isAdditionalDiskSpaceSet) {
		if err := d.Set("additional_disk_space", HumanReadableByteSize(additionalDiskSpace*units.MiB)); err != nil {
			return err
		}
		if err := d.Set("disk_space", nil); err != nil {
			return err
		}
	}

	if isDiskSpaceSet && diskSpace > 0 {
		if err := d.Set("disk_space", HumanReadableByteSize(diskSpace*units.MiB)); err != nil {
			return err
		}
		if err := d.Set("additional_disk_space", nil); err != nil {
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
		if err := d.Set("project_vpc_id", BuildResourceID(project, s.ProjectVpcId)); err != nil {
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

	// Handle read_replica service integrations
	readReplicaIntegrations, err := getReadReplicaIntegrationsForTerraform(s.ServiceIntegrations)
	if err != nil {
		return err
	}
	if err := d.Set("service_integrations", readReplicaIntegrations); err != nil {
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
	d *schema.ResourceData,
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
	case ServiceTypePG:
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
		setProp(props, "receiver_ingesting_remote_write_uri", connectionInfo.ReceiverIngestingRemoteWriteUri)
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

// IsUnknownRole checks if the database returned an error because of an unknown role
// to make deletions idempotent
func IsUnknownRole(err error) bool {
	var e aiven.Error
	return errors.As(err, &e) && strings.Contains(e.Message, "Code: 511")
}

// IsUnknownResource is a function to handle errors that we want to treat as "Not Found"
func IsUnknownResource(err error) bool {
	return aiven.IsNotFound(err) || IsUnknownRole(err)
}

func ResourceReadHandleNotFound(err error, d *schema.ResourceData) error {
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

func getContactEmailListForAPI(d *schema.ResourceData) (*[]aiven.ContactEmail, error) {
	if valuesInterface, ok := d.GetOk("tech_emails"); ok {
		var emails []aiven.ContactEmail
		err := Remarshal(valuesInterface.(*schema.Set).List(), &emails)
		if err != nil {
			return nil, err // Handle error appropriately
		}
		return &emails, nil
	}
	return &[]aiven.ContactEmail{}, nil
}

func ExpandService(name string, d *schema.ResourceData) (map[string]any, error) {
	return converters.Expand(converters.ServiceUserConfig, name, d)
}

func FlattenService(name string, d *schema.ResourceData, dto map[string]any) error {
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
