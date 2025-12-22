package kafka

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceKafkaConnector() *schema.Resource {
	return &schema.Resource{
		ReadContext: common.WithGenClientDiag(datasourceKafkaConnectorRead),
		Description: "Gets information about an Aiven for Apache Kafka® connector.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenKafkaConnectorSchema,
			"project", "service_name", "connector_name"),
	}
}

func datasourceKafkaConnectorRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) diag.Diagnostics {
	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	connectorName := d.Get("connector_name").(string)

	connectors, err := client.ServiceKafkaConnectList(ctx, projectName, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, con := range connectors {
		if con.Name == connectorName {
			d.SetId(schemautil.BuildResourceID(projectName, serviceName, connectorName))
			return resourceKafkaConnectorRead(ctx, d, client)
		}
	}

	return diag.Errorf("kafka connector %s not found", connectorName)
}
