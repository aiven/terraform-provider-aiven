// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/aiven/templates"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"strconv"
	"strings"
	"time"
)

const (
	ServiceTypePG               = "pg"
	ServiceTypeCassandra        = "cassandra"
	ServiceTypeElasticsearch    = "elasticsearch"
	ServiceTypeGrafana          = "grafana"
	ServiceTypeInfluxDB         = "influxdb"
	ServiceTypeRedis            = "redis"
	ServiceTypeMySQL            = "mysql"
	ServiceTypeKafka            = "kafka"
	ServiceTypeKafkaConnect     = "kafka_connect"
	ServiceTypeKafkaMirrormaker = "kafka_mirrormaker"
	ServiceTypeM3               = "m3db"
	ServiceTypeM3Aggregator     = "m3aggregator"
)

func availableServiceTypes() []string {
	return []string{
		ServiceTypePG,
		ServiceTypeCassandra,
		ServiceTypeElasticsearch,
		ServiceTypeGrafana,
		ServiceTypeInfluxDB,
		ServiceTypeRedis,
		ServiceTypeMySQL,
		ServiceTypeKafka,
		ServiceTypeKafkaConnect,
		ServiceTypeKafkaMirrormaker,
		ServiceTypeM3,
		ServiceTypeM3Aggregator,
	}
}

func serviceCommonSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"project": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Target project",
			ForceNew:    true,
		},
		"cloud_name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Cloud the service runs in",
		},
		"plan": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Subscription plan",
		},
		"service_name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Service name",
			ForceNew:    true,
		},
		"service_type": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Aiven internal service type code",
		},
		"project_vpc_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Identifier of the VPC the service should be in, if any",
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
			Description: "Prevent service from being deleted. It is recommended to have this enabled for all services.",
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
			Description: "Service hostname",
		},
		"service_port": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "Service port",
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
			Description: "Service state",
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
						Description: "Type of the service integration. The only supported value at the moment is 'read_replica'",
					},
				},
			},
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
	}
}

var aivenServiceSchema = map[string]*schema.Schema{
	"project": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Target project",
		ForceNew:    true,
	},
	"cloud_name": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Cloud the service runs in",
	},
	"plan": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Subscription plan",
	},
	"service_name": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Service name",
		ForceNew:    true,
	},
	"service_type": {
		Type:         schema.TypeString,
		Required:     true,
		Description:  "Service type code",
		ForceNew:     true,
		ValidateFunc: validation.StringInSlice(availableServiceTypes(), false),
	},
	"project_vpc_id": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Identifier of the VPC the service should be in, if any",
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
		Description: "Prevent service from being deleted. It is recommended to have this enabled for all services.",
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
		Description: "Service hostname",
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
					Description: "Type of the service integration. The only supported value at the moment is 'read_replica'",
				},
			},
		},
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
	"service_port": {
		Type:        schema.TypeInt,
		Computed:    true,
		Description: "Service port",
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
		Description: "Service state",
	},
	"cassandra": {
		Type:        schema.TypeList,
		MaxItems:    1,
		Computed:    true,
		Description: "Cassandra specific server provided values",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
	},
	"cassandra_user_config": {
		Type:             schema.TypeList,
		MaxItems:         1,
		Optional:         true,
		Description:      "Cassandra specific user configurable settings",
		DiffSuppressFunc: emptyObjectDiffSuppressFunc,
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("service")[ServiceTypeCassandra].(map[string]interface{})),
		},
	},
	"elasticsearch": {
		Type:        schema.TypeList,
		MaxItems:    1,
		Computed:    true,
		Description: "Elasticsearch specific server provided values",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"kibana_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "URI for Kibana frontend",
					Sensitive:   true,
				},
			},
		},
	},
	"elasticsearch_user_config": {
		Type:             schema.TypeList,
		MaxItems:         1,
		Optional:         true,
		Description:      "Elasticsearch specific user configurable settings",
		DiffSuppressFunc: emptyObjectDiffSuppressFunc,
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("service")[ServiceTypeElasticsearch].(map[string]interface{})),
		},
	},
	"grafana": {
		Type:        schema.TypeList,
		MaxItems:    1,
		Computed:    true,
		Description: "Grafana specific server provided values",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
	},
	"grafana_user_config": {
		Type:             schema.TypeList,
		MaxItems:         1,
		Optional:         true,
		Description:      "Grafana specific user configurable settings",
		DiffSuppressFunc: emptyObjectDiffSuppressFunc,
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("service")[ServiceTypeGrafana].(map[string]interface{})),
		},
	},
	"influxdb": {
		Type:        schema.TypeList,
		MaxItems:    1,
		Computed:    true,
		Description: "InfluxDB specific server provided values",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"database_name": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Name of the default InfluxDB database",
				},
			},
		},
	},
	"influxdb_user_config": {
		Type:             schema.TypeList,
		MaxItems:         1,
		Optional:         true,
		Description:      "InfluxDB specific user configurable settings",
		DiffSuppressFunc: emptyObjectDiffSuppressFunc,
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("service")[ServiceTypeInfluxDB].(map[string]interface{})),
		},
	},
	"kafka": {
		Type:        schema.TypeList,
		MaxItems:    1,
		Computed:    true,
		Description: "Kafka specific server provided values",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"access_cert": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The Kafka client certificate",
					Optional:    true,
					Sensitive:   true,
				},
				"access_key": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The Kafka client certificate key",
					Optional:    true,
					Sensitive:   true,
				},
				"connect_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The Kafka Connect URI, if any",
					Optional:    true,
					Sensitive:   true,
				},
				"rest_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The Kafka REST URI, if any",
					Optional:    true,
					Sensitive:   true,
				},
				"schema_registry_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The Schema Registry URI, if any",
					Optional:    true,
					Sensitive:   true,
				},
			},
		},
	},
	"kafka_user_config": {
		Type:             schema.TypeList,
		MaxItems:         1,
		Optional:         true,
		Description:      "Kafka specific user configurable settings",
		DiffSuppressFunc: emptyObjectDiffSuppressFunc,
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("service")[ServiceTypeKafka].(map[string]interface{})),
		},
	},
	"kafka_connect": {
		Type:        schema.TypeList,
		MaxItems:    1,
		Computed:    true,
		Description: "Kafka Connect specific server provided values",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
	},
	"kafka_connect_user_config": {
		Type:             schema.TypeList,
		MaxItems:         1,
		Optional:         true,
		Description:      "Kafka Connect specific user configurable settings",
		DiffSuppressFunc: emptyObjectDiffSuppressFunc,
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("service")[ServiceTypeKafkaConnect].(map[string]interface{})),
		},
	},
	"mysql": {
		Type:        schema.TypeList,
		MaxItems:    1,
		Computed:    true,
		Description: "MySQL specific server provided values",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
	},
	"mysql_user_config": {
		Type:             schema.TypeList,
		MaxItems:         1,
		Optional:         true,
		Description:      "MySQL specific user configurable settings",
		DiffSuppressFunc: emptyObjectDiffSuppressFunc,
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("service")[ServiceTypeMySQL].(map[string]interface{})),
		},
	},
	"kafka_mirrormaker": {
		Type:        schema.TypeList,
		MaxItems:    1,
		Computed:    true,
		Description: "Kafka MirrorMaker 2 specific server provided values",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
	},
	"kafka_mirrormaker_user_config": {
		Type:             schema.TypeList,
		MaxItems:         1,
		Optional:         true,
		Description:      "Kafka MirrorMaker 2 specific user configurable settings",
		DiffSuppressFunc: emptyObjectDiffSuppressFunc,
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("service")[ServiceTypeKafkaMirrormaker].(map[string]interface{})),
		},
	},
	"pg": {
		Type:        schema.TypeList,
		MaxItems:    1,
		Computed:    true,
		Description: "PostgreSQL specific server provided values",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"replica_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "PostgreSQL replica URI for services with a replica",
					Sensitive:   true,
				},
				"uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "PostgreSQL master connection URI",
					Optional:    true,
					Sensitive:   true,
				},
				"dbname": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Primary PostgreSQL database name",
				},
				"host": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "PostgreSQL master node host IP or name",
				},
				"password": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "PostgreSQL admin user password",
					Sensitive:   true,
				},
				"port": {
					Type:        schema.TypeInt,
					Computed:    true,
					Description: "PostgreSQL port",
				},
				"sslmode": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "PostgreSQL sslmode setting (currently always \"require\")",
				},
				"user": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "PostgreSQL admin user name",
				},
			},
		},
	},
	"pg_user_config": {
		Type:             schema.TypeList,
		MaxItems:         1,
		Optional:         true,
		Description:      "PostgreSQL specific user configurable settings",
		DiffSuppressFunc: emptyObjectDiffSuppressFunc,
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("service")[ServiceTypePG].(map[string]interface{})),
		},
	},
	"redis": {
		Type:        schema.TypeList,
		MaxItems:    1,
		Computed:    true,
		Description: "Redis specific server provided values",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
	},
	"redis_user_config": {
		Type:             schema.TypeList,
		MaxItems:         1,
		Optional:         true,
		Description:      "Redis specific user configurable settings",
		DiffSuppressFunc: emptyObjectDiffSuppressFunc,
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("service")[ServiceTypeRedis].(map[string]interface{})),
		},
	},
}

func resourceService() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServiceCreateWrapper("service"),
		ReadContext:   resourceServiceRead,
		UpdateContext: resourceServiceUpdate,
		DeleteContext: resourceServiceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceServiceState,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: aivenServiceSchema,
	}
}

func resourceServiceCreateWrapper(serviceType string) schema.CreateContextFunc {
	if serviceType == "service" {
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

func resourceServiceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)
	serviceType := d.Get("service_type").(string)
	userConfig := ConvertTerraformUserConfigToAPICompatibleFormat("service", serviceType, true, d)
	vpcID := d.Get("project_vpc_id").(string)
	var apiServiceIntegrations []aiven.NewServiceIntegration
	tfServiceIntegrations := d.Get("service_integrations")
	if tfServiceIntegrations != nil {
		tfServiceIntegrationList := tfServiceIntegrations.([]interface{})
		for _, definition := range tfServiceIntegrationList {
			definitionMap := definition.(map[string]interface{})
			sourceService := definitionMap["source_service_name"].(string)
			apiIntegration := aiven.NewServiceIntegration{
				IntegrationType: definitionMap["integration_type"].(string),
				SourceService:   &sourceService,
				UserConfig:      make(map[string]interface{}),
			}
			apiServiceIntegrations = append(apiServiceIntegrations, apiIntegration)
		}
	}
	project := d.Get("project").(string)
	var vpcIDPointer *string
	if len(vpcID) > 0 {
		_, vpcID := splitResourceID2(vpcID)
		vpcIDPointer = &vpcID
	}

	_, err := client.Services.Create(
		project,
		aiven.CreateServiceRequest{
			Cloud:                 d.Get("cloud_name").(string),
			MaintenanceWindow:     getMaintenanceWindow(d),
			Plan:                  d.Get("plan").(string),
			ProjectVPCID:          vpcIDPointer,
			ServiceIntegrations:   apiServiceIntegrations,
			ServiceName:           d.Get("service_name").(string),
			ServiceType:           serviceType,
			TerminationProtection: d.Get("termination_protection").(bool),
			UserConfig:            userConfig,
		},
	)

	if err != nil && !isServiceAlreadyExists(err) {
		return diag.FromErr(err)
	}

	service, err := resourceServiceWait(ctx, d, m, "create")
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(buildResourceID(d.Get("project").(string), service.Name))

	err = copyServicePropertiesFromAPIResponseToTerraform(d, service, d.Get("project").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceServiceRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName := splitResourceID2(d.Id())
	service, err := client.Services.Get(projectName, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	err = copyServicePropertiesFromAPIResponseToTerraform(d, service, projectName)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceServiceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	if d.HasChanges("service_integrations") && len(d.Get("service_integrations").([]interface{})) != 0 {
		return diag.Errorf("service_integrations field can only be set during creation of a service")
	}

	projectName, serviceName := splitResourceID2(d.Id())
	userConfig := ConvertTerraformUserConfigToAPICompatibleFormat("service", d.Get("service_type").(string), false, d)
	vpcID := d.Get("project_vpc_id").(string)
	var vpcIDPointer *string
	if len(vpcID) > 0 {
		_, vpcID := splitResourceID2(vpcID)
		vpcIDPointer = &vpcID
	}
	_, err := client.Services.Update(
		projectName,
		serviceName,
		aiven.UpdateServiceRequest{
			Cloud:                 d.Get("cloud_name").(string),
			MaintenanceWindow:     getMaintenanceWindow(d),
			Plan:                  d.Get("plan").(string),
			ProjectVPCID:          vpcIDPointer,
			Powered:               true,
			TerminationProtection: d.Get("termination_protection").(bool),
			UserConfig:            userConfig,
		},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	service, err := resourceServiceWait(ctx, d, m, "update")
	if err != nil {
		return diag.FromErr(err)
	}

	err = copyServicePropertiesFromAPIResponseToTerraform(d, service, projectName)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceServiceDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName := splitResourceID2(d.Id())

	err := client.Services.Delete(projectName, serviceName)
	if err != nil && !aiven.IsNotFound(err) {
		return diag.FromErr(err)
	}

	return nil
}

func resourceServiceState(_ context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*aiven.Client)

	if len(strings.Split(d.Id(), "/")) != 2 {
		return nil, fmt.Errorf("invalid identifier %v, expected <project_name>/<service_name>", d.Id())
	}

	projectName, serviceName := splitResourceID2(d.Id())
	service, err := client.Services.Get(projectName, serviceName)
	if err != nil {
		return nil, err
	}

	err = copyServicePropertiesFromAPIResponseToTerraform(d, service, projectName)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func resourceServiceWait(ctx context.Context, d *schema.ResourceData, m interface{}, operation string) (*aiven.Service, error) {
	var timeout time.Duration
	if operation == "create" {
		timeout = d.Timeout(schema.TimeoutCreate)
	} else {
		timeout = d.Timeout(schema.TimeoutUpdate)
	}

	w := &ServiceChangeWaiter{
		Client:      m.(*aiven.Client),
		Operation:   operation,
		Project:     d.Get("project").(string),
		ServiceName: d.Get("service_name").(string),
	}

	service, err := w.Conf(timeout).WaitForStateContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("error waiting for Aiven service to be RUNNING: %s", err)
	}

	return service.(*aiven.Service), nil
}

func getMaintenanceWindow(d *schema.ResourceData) *aiven.MaintenanceWindow {
	dow := d.Get("maintenance_window_dow").(string)
	t := d.Get("maintenance_window_time").(string)
	if len(dow) > 0 && len(t) > 0 {
		return &aiven.MaintenanceWindow{DayOfWeek: dow, TimeOfDay: t}
	}
	return nil
}

func copyServicePropertiesFromAPIResponseToTerraform(
	d *schema.ResourceData,
	service *aiven.Service,
	project string,
) error {
	if err := d.Set("cloud_name", service.CloudName); err != nil {
		return err
	}
	if err := d.Set("service_name", service.Name); err != nil {
		return err
	}
	if err := d.Set("state", service.State); err != nil {
		return err
	}
	if err := d.Set("plan", service.Plan); err != nil {
		return err
	}
	if err := d.Set("service_type", service.Type); err != nil {
		return err
	}
	if err := d.Set("termination_protection", service.TerminationProtection); err != nil {
		return err
	}
	if err := d.Set("maintenance_window_dow", service.MaintenanceWindow.DayOfWeek); err != nil {
		return err
	}
	if err := d.Set("maintenance_window_time", service.MaintenanceWindow.TimeOfDay); err != nil {
		return err
	}
	if err := d.Set("service_uri", service.URI); err != nil {
		return err
	}
	if err := d.Set("project", project); err != nil {
		return err
	}

	if service.ProjectVPCID != nil {
		if err := d.Set("project_vpc_id", buildResourceID(project, *service.ProjectVPCID)); err != nil {
			return err
		}
	}

	userConfig := ConvertAPIUserConfigToTerraformCompatibleFormat("service", service.Type, service.UserConfig)
	if err := d.Set(service.Type+"_user_config", userConfig); err != nil {
		return fmt.Errorf("cannot set `%s_user_config` : %s;"+
			"Please make sure that all Aiven services have unique service names", service.Type, err)
	}

	params := service.URIParams
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

	if err := d.Set("components", flattenServiceComponents(service)); err != nil {
		return fmt.Errorf("cannot set `components` : %s", err)
	}

	return copyConnectionInfoFromAPIResponseToTerraform(d, service.Type, service.ConnectionInfo)
}

func flattenServiceComponents(r *aiven.Service) []map[string]interface{} {
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
	case "redis":
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

func isServiceAlreadyExists(err error) bool {
	if e, ok := err.(aiven.Error); ok {
		if e.Status == 409 && strings.Contains(err.Error(), "Service name is already in use in this project") {
			return true
		}
	}
	return false
}
