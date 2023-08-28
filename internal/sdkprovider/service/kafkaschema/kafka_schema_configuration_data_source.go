package kafkaschema

import (
	"context"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceKafkaSchemaConfiguration() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceKafkaSchemasConfigurationRead,
		Description: "The Kafka Schema Configuration data source provides information about the existing Aiven Kafka Schema Configuration.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenKafkaSchemaSchema,
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

	d.SetId(schemautil.BuildResourceID(projectName, serviceName))

	return resourceKafkaSchemaConfigurationRead(ctx, d, m)
}
