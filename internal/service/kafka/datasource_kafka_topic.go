package kafka

import (
	"context"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceKafkaTopic() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceKafkaTopicRead,
		Description: "The Kafka Topic data source provides information about the existing Aiven Kafka Topic.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenKafkaTopicSchema,
			"project", "service_name", "topic_name"),
	}
}

func datasourceKafkaTopicRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	topicName := d.Get("topic_name").(string)

	d.SetId(schemautil.BuildResourceID(projectName, serviceName, topicName))
	return resourceKafkaTopicReadDatasource(ctx, d, m)
}
