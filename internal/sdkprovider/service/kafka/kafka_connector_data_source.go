package kafka

import (
	"context"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceKafkaConnector() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceKafkaConnectorRead,
		Description: "The Kafka connector data source provides information about the existing Aiven Kafka connector.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenKafkaConnectorSchema,
			"project", "service_name", "connector_name"),
	}
}

func datasourceKafkaConnectorRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	connectorName := d.Get("connector_name").(string)

	cons, err := m.(*aiven.Client).KafkaConnectors.List(projectName, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, con := range cons.Connectors {
		if con.Name == connectorName {
			d.SetId(schemautil.BuildResourceID(projectName, serviceName, connectorName))
			return resourceKafkaConnectorRead(ctx, d, m)
		}
	}

	return diag.Errorf("kafka connector %s not found", connectorName)
}
