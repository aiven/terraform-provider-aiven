package aiven

import (
	"context"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceKafkaSchema() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceKafkaSchemaRead,
		Schema: resourceSchemaAsDatasourceSchema(aivenKafkaSchemaSchema,
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
			d.SetId(buildResourceID(projectName, serviceName, subjectName))
			return resourceKafkaSchemaRead(ctx, d, m)
		}
	}

	return diag.Errorf("kafka schema subject %s/%s/%s not found",
		projectName, serviceName, subjectName)
}
