package schemautil

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/aiven/terraform-provider-aiven/internal/meta"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil/templates"

	"github.com/aiven/aiven-go-client"
	"github.com/docker/go-units"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	ServiceTypePG               = "pg"
	ServiceTypeCassandra        = "cassandra"
	ServiceTypeElasticsearch    = "elasticsearch"
	ServiceTypeOpensearch       = "opensearch"
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
)

func ServiceCommonSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"project": CommonSchemaProjectReference,

		"cloud_name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Defines where the cloud provider and region where the service is hosted in. This can be changed freely after service is created. Changing the value will trigger a potentially lengthy migration process for the service. Format is cloud provider name (`aws`, `azure`, `do` `google`, `upcloud`, etc.), dash, and the cloud provider specific region name. These are documented on each Cloud provider's own support articles, like [here for Google](https://cloud.google.com/compute/docs/regions-zones/) and [here for AWS](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.RegionsAndAvailabilityZones.html).",
		},
		"plan": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Defines what kind of computing resources are allocated for the service. It can be changed after creation, though there are some restrictions when going to a smaller plan such as the new plan must have sufficient amount of disk space to store all current data and switching to a plan with fewer nodes might not be supported. The basic plan names are `hobbyist`, `startup-x`, `business-x` and `premium-x` where `x` is (roughly) the amount of memory on each node (also other attributes like number of CPUs and amount of disk space varies but naming is based on memory). The available options can be seem from the [Aiven pricing page](https://aiven.io/pricing).",
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
			DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
				return new == ""
			},
		},
		"maintenance_window_time": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Time of day when maintenance operations should be performed. UTC time in HH:mm:ss format.",
			DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
				return new == ""
			},
		},
		"termination_protection": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Prevents the service from being deleted. It is recommended to set this to `true` for all production services to prevent unintentional service deletion. This does not shield against deleting databases or topics but for services with backups much of the content can at least be restored from backup in case accidental deletion is done.",
		},
		"disk_space": {
			Type:         schema.TypeString,
			Optional:     true,
			Description:  "The disk space of the service, possible values depend on the service type, the cloud provider and the project. Reducing will result in the service rebalancing.",
			ValidateFunc: ValidateHumanByteSizeString,
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
			Type:        schema.TypeList,
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
						Description: "DNS name for connecting to the service component",
					},
					"kafka_authentication_method": {
						Type:        schema.TypeString,
						Computed:    true,
						Optional:    true,
						Description: "Kafka authentication method. This is a value specific to the 'kafka' service component",
					},
					"port": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "Port number for connecting to the service component",
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
	}
}

func ResourceServiceCreateWrapper(serviceType string) schema.CreateContextFunc {
	if serviceType == templates.UserConfigSchemaService {
		return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
			// Need to set empty value for all services or all Terraform keeps on showing there's
			// a change in the computed values that don't match actual service type
			if err := d.Set(ServiceTypeCassandra, []map[string]interface{}{}); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set(ServiceTypeElasticsearch, []map[string]interface{}{}); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set(ServiceTypeGrafana, []map[string]interface{}{}); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set(ServiceTypeInfluxDB, []map[string]interface{}{}); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set(ServiceTypeKafka, []map[string]interface{}{}); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set(ServiceTypeKafkaConnect, []map[string]interface{}{}); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set(ServiceTypeKafkaMirrormaker, []map[string]interface{}{}); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set(ServiceTypeMySQL, []map[string]interface{}{}); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set(ServiceTypePG, []map[string]interface{}{}); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set(ServiceTypeRedis, []map[string]interface{}{}); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set(ServiceTypeOpensearch, []map[string]interface{}{}); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set(ServiceTypeFlink, []map[string]interface{}{}); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set(ServiceTypeClickhouse, []map[string]interface{}{}); err != nil {
				return diag.FromErr(err)
			}
			return resourceServiceCreate(ctx, d, m)
		}
	}

	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		if err := d.Set("service_type", serviceType); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set(serviceType, []map[string]interface{}{}); err != nil {
			return diag.FromErr(err)
		}

		return resourceServiceCreate(ctx, d, m)
	}
}

func ResourceServiceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*meta.Meta).Client

	projectName, serviceName := SplitResourceID2(d.Id())
	s, err := client.Services.Get(projectName, serviceName)
	if err != nil {
		if err = ResourceReadHandleNotFound(err, d, m); err != nil {
			return diag.FromErr(fmt.Errorf("unable to GET service %s: %s", d.Id(), err))
		}
		return nil
	}
	servicePlanParams, err := GetServicePlanParametersFromServiceResponse(ctx, client, projectName, s)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to get service plan parameters: %w", err))
	}

	err = copyServicePropertiesFromAPIResponseToTerraform(d, s, servicePlanParams, projectName)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to copy api response into terraform schema: %w", err))
	}

	allocatedStaticIps, err := CurrentlyAllocatedStaticIps(ctx, projectName, serviceName, m)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to currently allocated static ips: %w", err))
	}
	if err = d.Set("static_ips", allocatedStaticIps); err != nil {
		return diag.FromErr(fmt.Errorf("unable to set static ips field in schema: %w", err))
	}

	t, err := client.ServiceTags.Get(projectName, serviceName)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to get service tags: %w", err))
	}

	if err := d.Set("tag", SetTagsTerraformProperties(t.Tags)); err != nil {
		return diag.FromErr(fmt.Errorf("unable to set tag's in schema: %w", err))
	}

	return nil
}

func resourceServiceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*meta.Meta).Client

	serviceType := d.Get("service_type").(string)
	project := d.Get("project").(string)

	// During the creation of service, if disc_space is not set by a TF user,
	// we transfer 0 values in API creation request, which makes Aiven provision
	// a default disc space values for a common
	var diskSpace int
	if ds, ok := d.GetOk("disk_space"); ok {
		diskSpace = ConvertToDiskSpaceMB(ds.(string))
	}

	_, err := client.Services.Create(
		project,
		aiven.CreateServiceRequest{
			Cloud:                 d.Get("cloud_name").(string),
			Plan:                  d.Get("plan").(string),
			ProjectVPCID:          GetProjectVPCIdPointer(d),
			ServiceIntegrations:   GetAPIServiceIntegrations(d),
			MaintenanceWindow:     GetMaintenanceWindow(d),
			ServiceName:           d.Get("service_name").(string),
			ServiceType:           serviceType,
			TerminationProtection: d.Get("termination_protection").(bool),
			DiskSpaceMB:           diskSpace,
			UserConfig:            ConvertTerraformUserConfigToAPICompatibleFormat(templates.UserConfigSchemaService, serviceType, true, d),
			StaticIPs:             FlattenToString(d.Get("static_ips").([]interface{})),
		},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	// Create already takes care of static ip associations, no need to explictely associate them here

	s, err := WaitForServiceCreation(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = client.ServiceTags.Set(project, d.Get("service_name").(string), aiven.ServiceTagsRequest{
		Tags: GetTagsFromSchema(d),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(BuildResourceID(project, s.Name))

	return ResourceServiceRead(ctx, d, m)
}

func ResourceServiceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*meta.Meta).Client

	var karapace *bool
	if v := d.Get("karapace"); d.HasChange("karapace") && v != nil {
		if k, ok := v.(bool); ok && k {
			karapace = &k
		}
	}

	// On service update, we send a default disc space value for a common
	// if the TF user does not specify it
	diskSpace, err := getDefaultDiskSpaceIfNotSet(ctx, d, client)
	if err != nil {
		return diag.FromErr(err)
	}

	projectName, serviceName := SplitResourceID2(d.Id())

	ass, dis, err := DiffStaticIps(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	// associate first, so that we can enable `static_ips` for a preexisting common
	for _, aip := range ass {
		if err := client.StaticIPs.Associate(projectName, aip, aiven.AssociateStaticIPRequest{ServiceName: serviceName}); err != nil {
			return diag.FromErr(err)
		}
	}

	if _, err := client.Services.Update(
		projectName,
		serviceName,
		aiven.UpdateServiceRequest{
			Cloud:                 d.Get("cloud_name").(string),
			Plan:                  d.Get("plan").(string),
			MaintenanceWindow:     GetMaintenanceWindow(d),
			ProjectVPCID:          GetProjectVPCIdPointer(d),
			Powered:               true,
			TerminationProtection: d.Get("termination_protection").(bool),
			DiskSpaceMB:           diskSpace,
			Karapace:              karapace,
			UserConfig:            ConvertTerraformUserConfigToAPICompatibleFormat(templates.UserConfigSchemaService, d.Get("service_type").(string), false, d),
		},
	); err != nil {
		return diag.FromErr(err)
	}

	if _, err = WaitForServiceUpdate(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	if len(dis) > 0 {
		for _, dip := range dis {
			if err := client.StaticIPs.Dissociate(projectName, dip); err != nil {
				return diag.FromErr(err)
			}
		}
		if err = WaitStaticIpsDissassociation(ctx, d, m); err != nil {
			return diag.FromErr(err)
		}
	}

	_, err = client.ServiceTags.Set(projectName, serviceName, aiven.ServiceTagsRequest{
		Tags: GetTagsFromSchema(d),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceServiceRead(ctx, d, m)
}

func getDefaultDiskSpaceIfNotSet(ctx context.Context, d *schema.ResourceData, client *aiven.Client) (int, error) {
	var diskSpace int
	if ds, ok := d.GetOk("disk_space"); !ok {
		// get service plan specific defaults
		servicePlanParams, err := GetServicePlanParametersFromSchema(ctx, client, d)
		if err != nil {
			return 0, fmt.Errorf("unable to get service plan parameters: %w", err)
		}

		diskSpace = servicePlanParams.DiskSizeMBDefault
	} else {
		diskSpace = ConvertToDiskSpaceMB(ds.(string))
	}

	return diskSpace, nil
}

func ResourceServiceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*meta.Meta).Client

	projectName, serviceName := SplitResourceID2(d.Id())

	if err := client.Services.Delete(projectName, serviceName); err != nil && !aiven.IsNotFound(err) {
		return diag.FromErr(err)
	}

	// Delete already takes care of static ip disassociation, no need to explictely dissasociate them here

	if err := WaitForDeletion(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func ResourceServiceState(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	m.(*meta.Meta).Import = true

	client := m.(*meta.Meta).Client

	if len(strings.Split(d.Id(), "/")) != 2 {
		return nil, fmt.Errorf("invalid identifier %v, expected <project_name>/<service_name>", d.Id())
	}

	projectName, serviceName := SplitResourceID2(d.Id())
	s, err := client.Services.Get(projectName, serviceName)
	if err != nil {
		return nil, fmt.Errorf("unable to GET service %s: %s", d.Id(), err)
	}
	servicePlanParams, err := GetServicePlanParametersFromServiceResponse(ctx, client, projectName, s)
	if err != nil {
		return nil, fmt.Errorf("unable to get service plan parameters: %w", err)
	}

	err = copyServicePropertiesFromAPIResponseToTerraform(d, s, servicePlanParams, projectName)
	if err != nil {
		return nil, fmt.Errorf("unable to copy api response into terraform schema: %w", err)
	}

	allocatedStaticIps, err := CurrentlyAllocatedStaticIps(ctx, projectName, serviceName, m)
	if err != nil {
		return nil, fmt.Errorf("unable to currently allocated static ips: %w", err)
	}
	if err = d.Set("static_ips", allocatedStaticIps); err != nil {
		return nil, fmt.Errorf("unable to set static ips field in schema: %w", err)
	}

	t, err := client.ServiceTags.Get(projectName, serviceName)
	if err != nil {
		return nil, fmt.Errorf("unable to get service tags: %w", err)
	}

	if err := d.Set("tag", SetTagsTerraformProperties(t.Tags)); err != nil {
		return nil, fmt.Errorf("unable to set tag's in schema: %w", err)
	}

	return []*schema.ResourceData{d}, nil
}

func copyServicePropertiesFromAPIResponseToTerraform(
	d *schema.ResourceData,
	s *aiven.Service,
	servicePlanParams PlanParameters,
	project string,
) error {
	serviceType := d.Get("service_type").(string)
	if _, ok := d.GetOk("service_type"); !ok {
		serviceType = s.Type
	}

	if err := d.Set("cloud_name", s.CloudName); err != nil {
		return err
	}
	if err := d.Set("service_name", s.Name); err != nil {
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
	if err := d.Set("maintenance_window_dow", s.MaintenanceWindow.DayOfWeek); err != nil {
		return err
	}
	if err := d.Set("maintenance_window_time", s.MaintenanceWindow.TimeOfDay); err != nil {
		return err
	}
	if err := d.Set("disk_space_used", HumanReadableByteSize(s.DiskSpaceMB*units.MiB)); err != nil {
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
	if err := d.Set("service_uri", s.URI); err != nil {
		return err
	}
	if err := d.Set("project", project); err != nil {
		return err
	}

	if s.ProjectVPCID != nil {
		if err := d.Set("project_vpc_id", BuildResourceID(project, *s.ProjectVPCID)); err != nil {
			return err
		}
	}
	userConfig := ConvertAPIUserConfigToTerraformCompatibleFormat(templates.UserConfigSchemaService, serviceType, s.UserConfig)
	if err := d.Set(serviceType+"_user_config", NormalizeIpFilter(d.Get(serviceType+"_user_config"), userConfig)); err != nil {
		return fmt.Errorf("cannot set `%s_user_config` : %s; Please make sure that all Aiven services have unique s names", serviceType, err)
	}

	params := s.URIParams
	if err := d.Set("service_host", params["host"]); err != nil {
		return err
	}

	port, _ := strconv.ParseInt(params["port"], 10, 32)
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
		return fmt.Errorf("cannot set `components` : %s", err)
	}

	return copyConnectionInfoFromAPIResponseToTerraform(d, serviceType, s.ConnectionInfo)
}

func FlattenServiceComponents(r *aiven.Service) []map[string]interface{} {
	var components []map[string]interface{}

	for _, c := range r.Components {
		component := map[string]interface{}{
			"component": c.Component,
			"host":      c.Host,
			"port":      c.Port,
			"route":     c.Route,
			"usage":     c.Usage,
		}
		components = append(components, component)
	}

	return components
}

func copyConnectionInfoFromAPIResponseToTerraform(
	d *schema.ResourceData,
	serviceType string,
	connectionInfo aiven.ConnectionInfo,
) error {
	props := make(map[string]interface{})

	switch serviceType {
	case "cassandra":
	case "opensearch":
		props["opensearch_dashboards_uri"] = connectionInfo.OpensearchDashboardsURI
	case "elasticsearch":
		props["kibana_uri"] = connectionInfo.KibanaURI
	case "grafana":
	case "influxdb":
		props["database_name"] = connectionInfo.InfluxDBDatabaseName
	case "kafka":
		props["access_cert"] = connectionInfo.KafkaAccessCert
		props["access_key"] = connectionInfo.KafkaAccessKey
		props["connect_uri"] = connectionInfo.KafkaConnectURI
		props["rest_uri"] = connectionInfo.KafkaRestURI
		props["schema_registry_uri"] = connectionInfo.SchemaRegistryURI
	case "kafka_connect":
	case "mysql":
	case "pg":
		if connectionInfo.PostgresURIs != nil && len(connectionInfo.PostgresURIs) > 0 {
			props["uri"] = connectionInfo.PostgresURIs[0]
		}
		if connectionInfo.PostgresParams != nil && len(connectionInfo.PostgresParams) > 0 {
			params := connectionInfo.PostgresParams[0]
			props["dbname"] = params.DatabaseName
			props["host"] = params.Host
			props["password"] = params.Password
			port, err := strconv.ParseInt(params.Port, 10, 32)
			if err == nil {
				props["port"] = int(port)
			}
			props["sslmode"] = params.SSLMode
			props["user"] = params.User
		}
		props["replica_uri"] = connectionInfo.PostgresReplicaURI
	case "clickhouse":
	case "redis":
	case "flink":
		props["host_ports"] = connectionInfo.FlinkHostPorts
	case "kafka_mirrormaker":
	case "m3db":
	case "m3aggregator":
	default:
		panic(fmt.Sprintf("Unsupported service type %v", serviceType))
	}

	if err := d.Set(serviceType, []map[string]interface{}{props}); err != nil {
		return err
	}

	return nil
}

// utility to check if the database returned an error because of an unknown role
// to make deletions idempotent
func IsUnknownRole(err error) bool {
	aivenError, ok := err.(aiven.Error)
	if !ok {
		return false
	}
	return strings.Contains(aivenError.Message, "Code: 511")
}

// this function is a place to bundle errors that we want to treat as "not found"
func IsUnknownResource(err error) bool {
	return aiven.IsNotFound(err) || IsUnknownRole(err)
}

func ResourceReadHandleNotFound(err error, d *schema.ResourceData, m interface{}) error {
	if err != nil && IsUnknownResource(err) && !m.(*meta.Meta).Import {
		d.SetId("")
		return nil
	}

	return err
}

func DatasourceServiceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*meta.Meta).Client

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	d.SetId(BuildResourceID(projectName, serviceName))

	services, err := client.Services.List(projectName)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, service := range services {
		if service.Name == serviceName {
			return ResourceServiceRead(ctx, d, m)
		}
	}

	return diag.Errorf("common %s/%s not found", projectName, serviceName)
}
