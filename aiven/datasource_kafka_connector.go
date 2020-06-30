package aiven

import (
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func datasourceKafkaConnector() *schema.Resource {
	return &schema.Resource{
		Read:   datasourceKafkaConnectorRead,
		Schema: resourceSchemaAsDatasourceSchema(aivenKafkaConnectorSchema, "project", "service_name", "connector_name"),
	}
}

func datasourceKafkaConnectorRead(d *schema.ResourceData, m interface{}) error {
	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	connectorName := d.Get("connector_name").(string)

	cons, err := m.(*aiven.Client).KafkaConnectors.List(projectName, serviceName)
	if err != nil {
		return err
	}

	for _, con := range cons.Connectors {
		if con.Name == connectorName {
			d.SetId(buildResourceID(projectName, serviceName, connectorName))
			return resourceKafkaConnectorRead(d, m)
		}
	}

	return fmt.Errorf("kafka connector %s not found", connectorName)
}
