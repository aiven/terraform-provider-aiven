package aiven

import (
	"context"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceKafkaSchemaConfiguration() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceKafkaSchemasConfigurationRead,
		Description: "The Kafka Schema Configuration data source provides information about the existing Aiven Kafka Schema Configuration.",
		Schema: resourceSchemaAsDatasourceSchema(aivenKafkaSchemaSchema,
			"project", "service_name"),
	}
}

func datasourceKafkaSchemasConfigurationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)

	_, err := m.(*aiven.Client).KafkaGlobalSchemaConfig.Get(projectName, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(buildResourceID(projectName, serviceName))

	return resourceKafkaSchemaConfigurationRead(ctx, d, m)
}
