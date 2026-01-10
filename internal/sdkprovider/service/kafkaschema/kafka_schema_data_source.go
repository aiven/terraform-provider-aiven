package kafkaschema

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.org/x/exp/slices"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceKafkaSchema() *schema.Resource {
	return &schema.Resource{
		ReadContext: common.WithGenClientDiag(datasourceKafkaSchemaRead),
		Description: "The Kafka Schema data source provides information about the existing Aiven Kafka Schema.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenKafkaSchemaSchema,
			"project", "service_name", "subject_name"),
	}
}

func datasourceKafkaSchemaRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) diag.Diagnostics {
	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	subjectName := d.Get("subject_name").(string)

	subjects, err := client.ServiceSchemaRegistrySubjects(ctx, projectName, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	if slices.Contains(subjects, subjectName) {
		d.SetId(schemautil.BuildResourceID(projectName, serviceName, subjectName))
		return resourceKafkaSchemaRead(ctx, d, client)
	}

	return diag.Errorf("kafka schema subject %s/%s/%s not found",
		projectName, serviceName, subjectName)
}
