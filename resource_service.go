package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jelmersnoeck/aiven"
)

func resourceService() *schema.Resource {
	return &schema.Resource{
		Create: resourceServiceCreate,
		Read:   resourceServiceRead,
		Update: resourceServiceUpdate,
		Delete: resourceServiceDelete,
		Exists: resourceServiceExists,
		Importer: &schema.ResourceImporter{
			State: resourceServiceState,
		},

		Schema: map[string]*schema.Schema{
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
				Required:    true,
				Description: "Service type code",
				ForceNew:    true,
			},
			"service_uri": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URI for connecting to the service. Service specific info is under \"kafka\", \"pg\", etc.",
			},
			"service_host": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Service hostname",
			},
			"service_port": {
				Type:        schema.TypeString,
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
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "Cassandra specific user configurable settings",
				Elem: &schema.Resource{
					Schema: GenerateTerraformUserConfigSchema(
						userConfigSchemas["service"]["cassandra"].(map[string]interface{})),
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
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "Elasticsearch specific user configurable settings",
				Elem: &schema.Resource{
					Schema: GenerateTerraformUserConfigSchema(
						userConfigSchemas["service"]["elasticsearch"].(map[string]interface{})),
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
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "Grafana specific user configurable settings",
				Elem: &schema.Resource{
					Schema: GenerateTerraformUserConfigSchema(
						userConfigSchemas["service"]["grafana"].(map[string]interface{})),
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
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "InfluxDB specific user configurable settings",
				Elem: &schema.Resource{
					Schema: GenerateTerraformUserConfigSchema(
						userConfigSchemas["service"]["influxdb"].(map[string]interface{})),
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
						},
						"rest_uri": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The Kafka REST URI, if any",
							Optional:    true,
						},
						"schema_registry_uri": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The Schema Registry URI, if any",
							Optional:    true,
						},
					},
				},
			},
			"kafka_user_config": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "Kafka specific user configurable settings",
				Elem: &schema.Resource{
					Schema: GenerateTerraformUserConfigSchema(
						userConfigSchemas["service"]["kafka"].(map[string]interface{})),
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
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "PostgreSQL specific user configurable settings",
				Elem: &schema.Resource{
					Schema: GenerateTerraformUserConfigSchema(
						userConfigSchemas["service"]["pg"].(map[string]interface{})),
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
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "Redis specific user configurable settings",
				Elem: &schema.Resource{
					Schema: GenerateTerraformUserConfigSchema(
						userConfigSchemas["service"]["redis"].(map[string]interface{})),
				},
			},
		},
	}
}

func resourceServiceCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)
	serviceType := d.Get("service_type").(string)
	userConfig := ConvertTerraformUserConfigToAPICompatibleFormat("service", serviceType, d)
	service, err := client.Services.Create(
		d.Get("project").(string),
		aiven.CreateServiceRequest{
			Cloud:       d.Get("cloud_name").(string),
			Plan:        d.Get("plan").(string),
			ServiceName: d.Get("service_name").(string),
			ServiceType: serviceType,
			UserConfig:  userConfig,
		},
	)

	if err != nil {
		return err
	}

	err = resourceServiceWait(d, m, "create")

	if err != nil {
		return err
	}

	d.SetId(buildResourceID(d.Get("project").(string), service.Name))

	return copyServicePropertiesFromAPIResponseToTerraform(d, service, d.Get("project").(string))
}

func resourceServiceRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName, serviceName := splitResourceID2(d.Id())
	service, err := client.Services.Get(projectName, serviceName)
	if err != nil {
		return err
	}

	return copyServicePropertiesFromAPIResponseToTerraform(d, service, projectName)
}

func resourceServiceUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName, serviceName := splitResourceID2(d.Id())
	userConfig := ConvertTerraformUserConfigToAPICompatibleFormat("service", d.Get("service_type").(string), d)
	service, err := client.Services.Update(
		projectName,
		serviceName,
		aiven.UpdateServiceRequest{
			Cloud:      d.Get("cloud_name").(string),
			Plan:       d.Get("plan").(string),
			Powered:    true,
			UserConfig: userConfig,
		},
	)
	if err != nil {
		return err
	}

	err = resourceServiceWait(d, m, "update")

	if err != nil {
		return err
	}

	return copyServicePropertiesFromAPIResponseToTerraform(d, service, projectName)
}

func resourceServiceDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName, serviceName := splitResourceID2(d.Id())
	return client.Services.Delete(projectName, serviceName)
}

func resourceServiceExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*aiven.Client)

	projectName, serviceName := splitResourceID2(d.Id())
	_, err := client.Services.Get(projectName, serviceName)
	return resourceExists(err)
}

func resourceServiceState(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*aiven.Client)

	if len(strings.Split(d.Id(), "/")) != 2 {
		return nil, fmt.Errorf("Invalid identifier %v, expected <project_name>/<service_name>", d.Id())
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

func resourceServiceWait(d *schema.ResourceData, m interface{}, operation string) error {
	w := &ServiceChangeWaiter{
		Client:      m.(*aiven.Client),
		Operation:   operation,
		Project:     d.Get("project").(string),
		ServiceName: d.Get("service_name").(string),
	}

	_, err := w.Conf().WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for Aiven service to be RUNNING: %s", err)
	}

	return nil
}

func copyServicePropertiesFromAPIResponseToTerraform(
	d *schema.ResourceData,
	service *aiven.Service,
	project string,
) error {
	d.Set("cloud_name", service.CloudName)
	d.Set("service_name", service.Name)
	d.Set("state", service.State)
	d.Set("plan", service.Plan)
	d.Set("service_type", service.Type)
	d.Set("service_uri", service.URI)
	d.Set("project", project)
	userConfig := ConvertAPIUserConfigToTerraformCompatibleFormat("service", service.Type, service.UserConfig)
	d.Set(service.Type+"_user_config", userConfig)

	params := service.URIParams
	d.Set("service_host", params["host"])
	port, _ := strconv.ParseInt(params["port"], 10, 32)
	d.Set("service_port", port)
	password, passwordOK := params["password"]
	username, usernameOK := params["username"]
	if passwordOK {
		d.Set("service_password", password)
	}
	if usernameOK {
		d.Set("service_username", username)
	}

	copyConnectionInfoFromAPIResponseToTerraform(d, service.Type, service.ConnectionInfo)

	return nil
}

func copyConnectionInfoFromAPIResponseToTerraform(
	d *schema.ResourceData,
	serviceType string,
	connectionInfo aiven.ConnectionInfo,
) {
	// Need to set empty value for all services or all Terraform keeps on showing there's
	// a change in the computed values that don't match actual service type
	d.Set("cassandra", []map[string]interface{}{})
	d.Set("elasticsearch", []map[string]interface{}{})
	d.Set("grafana", []map[string]interface{}{})
	d.Set("influxdb", []map[string]interface{}{})
	d.Set("kafka", []map[string]interface{}{})
	d.Set("pg", []map[string]interface{}{})
	d.Set("redis", []map[string]interface{}{})

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
	default:
		panic(fmt.Sprintf("Unsupported service type %v", serviceType))
	}
	d.Set(serviceType, []map[string]interface{}{props})
}
