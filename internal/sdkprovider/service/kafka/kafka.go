package kafka

import (
	"context"
	"strconv"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader"
)

func aivenKafkaSchema() map[string]*schema.Schema {
	aivenKafkaSchema := schemautil.ServiceCommonSchemaWithUserConfig(schemautil.ServiceTypeKafka)
	aivenKafkaSchema["karapace"] = &schema.Schema{
		Type:             schema.TypeBool,
		Optional:         true,
		Description:      "Switch the service to use [Karapace](https://aiven.io/docs/products/kafka/karapace) for schema registry and REST proxy.",
		DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
		Deprecated:       "Usage of this field is discouraged.",
	}
	aivenKafkaSchema["default_acl"] = &schema.Schema{
		Type:             schema.TypeBool,
		Optional:         true,
		ForceNew:         true,
		Default:          true,
		DiffSuppressFunc: schemautil.CreateOnlyDiffSuppressFunc,
		Description:      "Create a default wildcard Kafka ACL.",
	}
	aivenKafkaSchema[schemautil.ServiceTypeKafka] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Sensitive:   true,
		Description: "Kafka server connection details.",
		MaxItems:    1,
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"uris": {
					Type:        schema.TypeList,
					Computed:    true,
					Description: "Kafka server URIs.",
					Optional:    true,
					Sensitive:   true,
					Elem: &schema.Schema{
						Type:      schema.TypeString,
						Sensitive: true,
					},
				},
				"access_cert": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The Kafka client certificate.",
					Sensitive:   true,
				},
				"access_key": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The Kafka client certificate key.",
					Sensitive:   true,
				},
				"connect_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The Kafka Connect URI.",
					Sensitive:   true,
				},
				"rest_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The Kafka REST URI.",
					Sensitive:   true,
				},
				"schema_registry_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The Schema Registry URI.",
					Sensitive:   true,
				},
			},
		},
	}

	return aivenKafkaSchema
}

func ResourceKafka() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates and manages an [Aiven for Apache Kafka®](https://aiven.io/docs/products/kafka) service.",
		CreateContext: resourceKafkaCreate,
		ReadContext:   resourceKafkaRead,
		UpdateContext: schemautil.ResourceServiceUpdate,
		DeleteContext: schemautil.ResourceServiceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenKafkaSchema(),
		CustomizeDiff: customdiff.Sequence(
			schemautil.CustomizeDiffGenericService(schemautil.ServiceTypeKafka),

			// if a kafka_version is >= 3.0 then this schema field is not applicable
			customdiff.ComputedIf("karapace", func(ctx context.Context, d *schema.ResourceDiff, m interface{}) bool {
				project := d.Get("project").(string)
				serviceName := d.Get("service_name").(string)
				client := m.(*aiven.Client)

				kafka, err := client.Services.Get(ctx, project, serviceName)
				if err != nil {
					return false
				}

				if v, ok := kafka.UserConfig["kafka_version"]; ok {
					if version, err := strconv.ParseFloat(v.(string), 64); err == nil {
						if version >= 3 {
							return true
						}
					}
				}

				return false
			}),
		),
		SchemaVersion:  1,
		StateUpgraders: stateupgrader.Kafka(),
	}
}

func resourceKafkaCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if di := schemautil.ResourceServiceCreateWrapper(schemautil.ServiceTypeKafka)(ctx, d, m); di.HasError() {
		return di
	}

	// if default_acl=false delete default wildcard Kafka ACL and ACLs for Schema Registry that are automatically created
	if !d.Get("default_acl").(bool) {
		client := m.(*aiven.Client)
		project := d.Get("project").(string)
		serviceName := d.Get("service_name").(string)

		const (
			defaultACLId                    = "default"
			defaultSchemaRegistryACLConfig  = "default-sr-admin-config"
			defaultSchemaRegistryACLSubject = "default-sr-admin-subject"
		)

		if err := client.KafkaACLs.Delete(ctx, project, serviceName, defaultACLId); err != nil && !aiven.IsNotFound(err) {
			return diag.Errorf("cannot delete default wildcard kafka acl: %s", err)
		}

		var defaultSchemaACLLs = []string{
			defaultSchemaRegistryACLConfig,
			defaultSchemaRegistryACLSubject,
		}
		for _, acl := range defaultSchemaACLLs {
			if err := client.KafkaSchemaRegistryACLs.Delete(ctx, project, serviceName, acl); err != nil && !aiven.IsNotFound(err) {
				return diag.Errorf("cannot delete `%s` kafka ACL for Schema Registry: %s", acl, err)
			}
		}
	}

	return nil
}

func resourceKafkaRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, service, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	kafka, err := client.Services.Get(ctx, project, service)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	var diags diag.Diagnostics
	var kafkaVersion float64
	var schemaRegistry bool
	var kafkaRest bool

	if v, ok := kafka.UserConfig["kafka_version"]; ok {
		if version, err := strconv.ParseFloat(v.(string), 64); err == nil {
			kafkaVersion = version
		}
	}

	if v, ok := kafka.UserConfig["schema_registry"]; ok && v.(bool) {
		schemaRegistry = true
	}

	if v, ok := kafka.UserConfig["kafka_rest"]; ok && v.(bool) {
		kafkaRest = true
	}

	// Checking is Confluent SR/REST -> Karapace migration is available
	if kafkaVersion < 3.0 &&
		((schemaRegistry && !kafka.Features.Karapace) ||
			(kafkaRest && !kafka.Features.KafkaRest)) {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary: "You are using Confluent Schema Registry v5.0 that is no longer supported " +
				"on Kafka v3.0. Please switch to Karapace, a drop-in open source replacement " +
				"before proceeding with the upgrade. To do that use aiven_kafka.karapace=true " +
				"that will switch the service to use Karapace for schema registry and REST proxy. " +
				"More information about Karpace is available in our documentation: https://aiven.io/docs/products/kafka/karapace",
		})
	}

	return append(diags, schemautil.ResourceServiceRead(ctx, d, m)...)
}
