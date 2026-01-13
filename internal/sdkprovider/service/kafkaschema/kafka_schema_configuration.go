package kafkaschema

import (
	"context"
	"path/filepath"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/kafkaschemaregistry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"golang.org/x/sync/singleflight"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var schemaRegistryCheckGroup singleflight.Group

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
		CreateContext: common.WithGenClientDiag(resourceKafkaSchemaConfigurationCreate),
		UpdateContext: common.WithGenClientDiag(resourceKafkaSchemaConfigurationUpdate),
		ReadContext:   common.WithGenClientDiag(resourceKafkaSchemaConfigurationRead),
		DeleteContext: common.WithGenClientDiag(resourceKafkaSchemaConfigurationDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenKafkaSchemaConfigurationSchema,
	}
}

func resourceKafkaSchemaConfigurationUpdate(ctx context.Context, d *schema.ResourceData, client avngen.Client) diag.Diagnostics {
	project, serviceName, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = client.ServiceSchemaRegistryGlobalConfigPut(
		ctx,
		project,
		serviceName,
		&kafkaschemaregistry.ServiceSchemaRegistryGlobalConfigPutIn{
			Compatibility: kafkaschemaregistry.CompatibilityType(d.Get("compatibility_level").(string)),
		})
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceKafkaSchemaConfigurationRead(ctx, d, client)
}

// resourceKafkaSchemaConfigurationCreate Kafka Schemas global configuration cannot be created but only updated
func resourceKafkaSchemaConfigurationCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) diag.Diagnostics {
	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)

	_, err := client.ServiceSchemaRegistryGlobalConfigPut(
		ctx,
		project,
		serviceName,
		&kafkaschemaregistry.ServiceSchemaRegistryGlobalConfigPutIn{
			Compatibility: kafkaschemaregistry.CompatibilityType(d.Get("compatibility_level").(string)),
		})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(project, serviceName))

	return resourceKafkaSchemaConfigurationRead(ctx, d, client)
}

func resourceKafkaSchemaConfigurationRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) diag.Diagnostics {
	project, serviceName, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// if Schema Registry is disabled, the resource is considered deleted
	enabled, err := isSchemaRegistryEnabled(ctx, client, project, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}
	if !enabled {
		d.SetId("")
		return nil
	}

	compatibilityLevel, err := client.ServiceSchemaRegistryGlobalConfigGet(ctx, project, serviceName)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	if err := d.Set("project", project); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("compatibility_level", string(compatibilityLevel)); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// resourceKafkaSchemaConfigurationDelete Kafka Schemas configuration cannot be deleted, therefore
// on delete event configuration will be set to the default setting
func resourceKafkaSchemaConfigurationDelete(ctx context.Context, d *schema.ResourceData, client avngen.Client) diag.Diagnostics {
	project, serviceName, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	enabled, err := isSchemaRegistryEnabled(ctx, client, project, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}
	if !enabled {
		return nil
	}

	_, err = client.ServiceSchemaRegistryGlobalConfigPut(
		ctx,
		project,
		serviceName,
		&kafkaschemaregistry.ServiceSchemaRegistryGlobalConfigPutIn{
			Compatibility: kafkaschemaregistry.CompatibilityTypeBackward,
		})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// isSchemaRegistryEnabled checks if the Schema Registry is enabled
func isSchemaRegistryEnabled(ctx context.Context, client avngen.Client, project, serviceName string) (bool, error) {
	key := filepath.Join(project, serviceName)

	v, err, _ := schemaRegistryCheckGroup.Do(key, func() (any, error) {
		service, err := client.ServiceGet(ctx, project, serviceName)
		if err != nil {
			return false, err
		}

		if v, ok := service.UserConfig["schema_registry"]; ok {
			if val, ok := v.(bool); ok {
				return val, nil
			}
		}

		return false, nil
	})

	if err != nil {
		return false, err
	}

	return v.(bool), nil
}
