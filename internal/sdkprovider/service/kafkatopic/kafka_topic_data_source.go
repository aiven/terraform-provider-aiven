package kafkatopic

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceKafkaTopic() *schema.Resource {
	return &schema.Resource{
		ReadContext: common.WithGenClientDiag(datasourceKafkaTopicRead),
		Description: "Gets information about an Aiven for Apache Kafka® topic.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenKafkaTopicSchema,
			"project", "service_name", "topic_name"),
	}
}

func datasourceKafkaTopicRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) diag.Diagnostics {
	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	topicName := d.Get("topic_name").(string)

	d.SetId(schemautil.BuildResourceID(projectName, serviceName, topicName))
	return resourceKafkaTopicReadDatasource(ctx, d, client)
}
