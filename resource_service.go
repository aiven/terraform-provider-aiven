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
			"cloud": {
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
			"hostname": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Service hostname",
			},
			"port": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Service port",
			},
			"state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Service state",
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
						"hosts": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "List of Kafka Hosts",
							Elem:        &schema.Schema{Type: schema.TypeString},
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
			Cloud:       d.Get("cloud").(string),
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
			Cloud:      d.Get("cloud").(string),
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
	d.Set("cloud", service.CloudName)
	d.Set("service_name", service.Name)
	d.Set("state", service.State)
	d.Set("plan", service.Plan)
	d.Set("service_type", service.Type)
	d.Set("project", project)
	userConfig := ConvertAPIUserConfigToTerraformCompatibleFormat("service", service.Type, service.UserConfig)
	d.Set(service.Type+"_user_config", userConfig)

	hn, err := service.Hostname()
	if err != nil {
		return err
	}
	port, err := service.Port()
	if err != nil {
		return err
	}

	d.Set("hostname", hn)
	d.Set("port", port)
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
	d.Set("kafka", []map[string]interface{}{})
	d.Set("pg", []map[string]interface{}{})

	props := make(map[string]interface{})
	switch serviceType {
	case "kafka":
		props["access_cert"] = connectionInfo.KafkaAccessCert
		props["access_key"] = connectionInfo.KafkaAccessKey
		props["connect_uri"] = connectionInfo.KafkaConnectURI
		props["hosts"] = connectionInfo.KafkaHosts
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
	default:
		panic(fmt.Sprintf("Unsupported service type %v", serviceType))
	}
	d.Set(serviceType, []map[string]interface{}{props})
}
