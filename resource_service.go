package main

import (
	"fmt"
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
				Description: "Target cloud",
			},
			"group_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Service group name",
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
			"user_config": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Service type-specific settings",
			},
		},
	}
}

func resourceServiceCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)
	service, err := client.Services.Create(
		d.Get("project").(string),
		aiven.CreateServiceRequest{
			Cloud:       d.Get("cloud").(string),
			GroupName:   d.Get("group_name").(string),
			Plan:        d.Get("plan").(string),
			ServiceName: d.Get("service_name").(string),
			ServiceType: d.Get("service_type").(string),
			UserConfig:  transformUserConfig(d),
		},
	)

	if err != nil {
		d.SetId("")
		return err
	}

	err = resourceServiceWait(d, m)

	if err != nil {
		d.SetId("")
		return err
	}

	d.SetId(service.Name + "!")

	return resourceServiceRead(d, m)
}

func resourceServiceRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	service, err := client.Services.Get(
		d.Get("project").(string),
		d.Get("service_name").(string),
	)
	if err != nil {
		return err
	}

	d.Set("name", service.Name)
	d.Set("state", service.State)
	d.Set("plan", service.Plan)

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

	return nil
}

func resourceServiceUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	_, err := client.Services.Update(
		d.Get("project").(string),
		d.Get("service_name").(string),
		aiven.UpdateServiceRequest{
			Cloud:      d.Get("cloud").(string),
			GroupName:  d.Get("group_name").(string),
			Plan:       d.Get("plan").(string),
			Powered:    true,
			UserConfig: transformUserConfig(d),
		},
	)
	if err != nil {
		return err
	}

	err = resourceServiceWait(d, m)

	if err != nil {
		return err
	}

	return resourceServiceRead(d, m)
}

func resourceServiceDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	return client.Services.Delete(
		d.Get("project").(string),
		d.Get("service_name").(string),
	)
}

func resourceServiceWait(d *schema.ResourceData, m interface{}) error {
	w := &ServiceChangeWaiter{
		Client:      m.(*aiven.Client),
		Project:     d.Get("project").(string),
		ServiceName: d.Get("service_name").(string),
	}

	_, err := w.Conf().WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for Aiven service to be RUNNING: %s", err)
	}

	return nil
}

// Aiven requires field types on received JSON to be correctly typed. If the service type is known
// transform the Terraform type (one of string, list, or map) into the appropriate Go type and
// return the modified user config. If the  service does not have a special handler the user config
// is returned as-is with the default Terraform types associated.
func transformUserConfig(d *schema.ResourceData) map[string]interface{} {
	serviceType := d.Get("service_type").(string)
	userConfig := d.Get("user_config").(map[string]interface{})

	if serviceType == "kafka" {
		userConfig = transformKafkaUserConfig(userConfig)
	}

	return userConfig
}

// Transform the kafka service user config from Terraform's built-in types to the appropriate
// golang types.
func transformKafkaUserConfig(userConfig map[string]interface{}) map[string]interface{} {
	newUserConfig := make(map[string]interface{})
	kafkaConnectInterface, ok := userConfig["kafka_connect"]
	if ok {
		newUserConfig["kafka_connect"] = stringToBool(kafkaConnectInterface.(string))
	}
	kafkaRestInterface, ok := userConfig["kafka_rest"]
	if ok {
		newUserConfig["kafka_rest"] = stringToBool(kafkaRestInterface.(string))
	}
	schemaRegistryInterface, ok := userConfig["schema_registry"]
	if ok {
		newUserConfig["schema_registry"] = stringToBool(schemaRegistryInterface.(string))
	}
	return newUserConfig
}

func stringToBool(maybeBool string) bool {
	return maybeBool == "1" || strings.ToLower(maybeBool) == "true"
}
