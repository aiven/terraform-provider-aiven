package kafkaschema

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/aiven/go-client-codegen/handler/kafkaschemaregistry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var aivenKafkaSchemaConfigurationSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,

	"compatibility_level": {
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringInSlice(kafkaschemaregistry.CompatibilityTypeChoices(), false),
		DiffSuppressFunc: func(_, _, newValue string, _ *schema.ResourceData) bool {
			// When a compatibility level is not set to any value and consequently is null (empty string).
			// Allow ignoring those.
			return newValue == ""
		},
		Description: userconfig.Desc("Kafka Schemas compatibility level.").PossibleValuesString(kafkaschemaregistry.CompatibilityTypeChoices()...).Build(),
	},
}

func ResourceKafkaSchemaConfiguration() *schema.Resource {
	return &schema.Resource{
		Description:   "The Kafka Schema Configuration resource allows the creation and management of Aiven Kafka Schema Configurations.",
		CreateContext: resourceKafkaSchemaConfigurationCreate,
		UpdateContext: resourceKafkaSchemaConfigurationUpdate,
		ReadContext:   resourceKafkaSchemaConfigurationRead,
		DeleteContext: resourceKafkaSchemaConfigurationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenKafkaSchemaConfigurationSchema,
	}
}

func resourceKafkaSchemaConfigurationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	project, serviceName, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = m.(*aiven.Client).KafkaGlobalSchemaConfig.Update(
		ctx,
		project,
		serviceName,
		aiven.KafkaSchemaConfig{
			CompatibilityLevel: d.Get("compatibility_level").(string),
		})
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceKafkaSchemaConfigurationRead(ctx, d, m)
}

// resourceKafkaSchemaConfigurationCreate Kafka Schemas global configuration cannot be created but only updated
func resourceKafkaSchemaConfigurationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)

	_, err := m.(*aiven.Client).KafkaGlobalSchemaConfig.Update(
		ctx,
		project,
		serviceName,
		aiven.KafkaSchemaConfig{
			CompatibilityLevel: d.Get("compatibility_level").(string),
		})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(project, serviceName))

	return resourceKafkaSchemaConfigurationRead(ctx, d, m)
}

func resourceKafkaSchemaConfigurationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	project, serviceName, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	r, err := m.(*aiven.Client).KafkaGlobalSchemaConfig.Get(ctx, project, serviceName)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	if err := d.Set("project", project); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("compatibility_level", r.CompatibilityLevel); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// resourceKafkaSchemaConfigurationDelete Kafka Schemas configuration cannot be deleted, therefore
// on delete event configuration will be set to the default setting
func resourceKafkaSchemaConfigurationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	project, serviceName, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = m.(*aiven.Client).KafkaGlobalSchemaConfig.Update(
		ctx,
		project,
		serviceName,
		aiven.KafkaSchemaConfig{
			CompatibilityLevel: "BACKWARD",
		})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
