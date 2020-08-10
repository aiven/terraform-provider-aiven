package aiven

import (
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func datasourceKafkaSchema() *schema.Resource {
	return &schema.Resource{
		Read: datasourceKafkaSchemaRead,
		Schema: resourceSchemaAsDatasourceSchema(aivenKafkaSchemaSchema,
			"project", "service_name", "subject_name"),
	}
}

func datasourceKafkaSchemaRead(d *schema.ResourceData, m interface{}) error {
	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	subjectName := d.Get("subject_name").(string)

	subjects, err := m.(*aiven.Client).KafkaSubjectSchemas.List(projectName, serviceName)
	if err != nil {
		return err
	}

	for _, subject := range subjects.Subjects {
		if subject == subjectName {
			d.SetId(buildResourceID(projectName, serviceName, subjectName))
			return resourceKafkaSchemaRead(d, m)
		}
	}

	return fmt.Errorf("kafka schema subject %s/%s/%s not found",
		projectName, serviceName, subjectName)
}
