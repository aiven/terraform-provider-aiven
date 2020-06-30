package aiven

import (
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func datasourceKafkaSchemaConfiguration() *schema.Resource {
	return &schema.Resource{
		Read:   datasourceKafkaSchemasConfigurationRead,
		Schema: resourceSchemaAsDatasourceSchema(aivenKafkaSchemaSchema, "project", "service_name"),
	}
}

func datasourceKafkaSchemasConfigurationRead(d *schema.ResourceData, m interface{}) error {
	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)

	_, err := m.(*aiven.Client).KafkaGlobalSchemaConfig.Get(projectName, serviceName)
	if err != nil {
		return err
	}

	d.SetId(buildResourceID(projectName, serviceName))

	return resourceKafkaSchemaConfigurationRead(d, m)
}
