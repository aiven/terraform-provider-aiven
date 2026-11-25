package kafkaschema

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/kafkaschemaregistry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

// datasourceKafkaSchemaConfigurationSchema has the fields of the aiven_kafka_schema resource this data source
// was built from. Only project, service_name and compatibility_level are set.
var datasourceKafkaSchemaConfigurationSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,
	"subject_name": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: userconfig.Desc("The Kafka Schema Subject name.").ForceNew().Build(),
	},
	"schema": {
		Type:     schema.TypeString,
		Required: true,
		Description: "Kafka Schema configuration. Should be a valid Avro, JSON, or Protobuf schema," +
			" depending on the schema type.",
	},
	"schema_type": {
		Type:     schema.TypeString,
		Optional: true,
		Description: userconfig.
			Desc("Kafka Schema configuration type. Defaults to AVRO").
			PossibleValuesString(kafkaschemaregistry.SchemaTypeChoices()...).Build(),
	},
	"version": {
		Type:        schema.TypeInt,
		Computed:    true,
		Description: "Kafka Schema configuration version.",
	},
	"compatibility_level": aivenKafkaSchemaConfigurationSchema["compatibility_level"],
	"references": {
		Type:        schema.TypeSet,
		Optional:    true,
		Description: "Schema references.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name":    {Type: schema.TypeString, Required: true},
				"subject": {Type: schema.TypeString, Required: true},
				"version": {Type: schema.TypeInt, Required: true},
			},
		},
	},
}

func DatasourceKafkaSchemaConfiguration() *schema.Resource {
	return &schema.Resource{
		ReadContext: common.WithGenClientDiag(datasourceKafkaSchemasConfigurationRead),
		Description: "The Kafka Schema Configuration data source provides information about the existing Aiven Kafka Schema Configuration.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(datasourceKafkaSchemaConfigurationSchema,
			"project", "service_name"),
	}
}

func datasourceKafkaSchemasConfigurationRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) diag.Diagnostics {
	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)

	_, err := client.ServiceSchemaRegistryGlobalConfigGet(ctx, projectName, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(projectName, serviceName))

	return resourceKafkaSchemaConfigurationRead(ctx, d, client)
}
