package kafka

import (
	"context"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceKafkaSchema() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceKafkaSchemaRead,
		Description: "The Kafka Schema data source provides information about the existing Aiven Kafka Schema.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenKafkaSchemaSchema,
			"project", "service_name", "subject_name"),
	}
}

func datasourceKafkaSchemaRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	subjectName := d.Get("subject_name").(string)

	subjects, err := m.(*aiven.Client).KafkaSubjectSchemas.List(projectName, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, subject := range subjects.Subjects {
		if subject == subjectName {
			d.SetId(schemautil.BuildResourceID(projectName, serviceName, subjectName))
			return resourceKafkaSchemaRead(ctx, d, m)
		}
	}

	return diag.Errorf("kafka schema subject %s/%s/%s not found",
		projectName, serviceName, subjectName)
}
