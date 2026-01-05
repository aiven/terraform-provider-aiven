package kafka

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader"
)

func aivenKafkaSchema() map[string]*schema.Schema {
	aivenKafkaSchema := schemautil.ServiceCommonSchemaWithUserConfig(schemautil.ServiceTypeKafka)
	aivenKafkaSchema["karapace"] = &schema.Schema{
		Type:             schema.TypeBool,
		Optional:         true,
		Description:      "Switch the service to use [Karapace](https://aiven.io/docs/products/kafka/karapace) for schema registry and REST proxy. This attribute is deprecated, use `schema_registry` and `kafka_rest` instead.",
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
		CreateContext: common.WithGenClientDiag(resourceKafkaCreate),
		ReadContext:   schemautil.ResourceServiceRead,
		UpdateContext: schemautil.ResourceServiceUpdate,
		DeleteContext: schemautil.ResourceServiceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenKafkaSchema(),
		CustomizeDiff: customdiff.Sequence(
			schemautil.CustomizeDiffGenericService(schemautil.ServiceTypeKafka),
		),
		SchemaVersion:  1,
		StateUpgraders: stateupgrader.Kafka(),
	}
}

func resourceKafkaCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) diag.Diagnostics {
	if di := schemautil.ResourceServiceCreateWrapper(schemautil.ServiceTypeKafka)(ctx, d, client); di.HasError() {
		return di
	}

	// if default_acl=false delete default wildcard Kafka ACL and ACLs for Schema Registry that are automatically created
	if !d.Get("default_acl").(bool) {
		project := d.Get("project").(string)
		serviceName := d.Get("service_name").(string)

		const (
			defaultACLId                    = "default"
			defaultSchemaRegistryACLConfig  = "default-sr-admin-config"
			defaultSchemaRegistryACLSubject = "default-sr-admin-subject"
		)

		_, err := client.ServiceKafkaAclDelete(ctx, project, serviceName, defaultACLId)
		if err != nil && !avngen.IsNotFound(err) {
			return diag.Errorf("cannot delete default wildcard kafka acl: %s", err)
		}

		defaultSchemaACLLs := []string{
			defaultSchemaRegistryACLConfig,
			defaultSchemaRegistryACLSubject,
		}
		for _, acl := range defaultSchemaACLLs {
			_, err := client.ServiceSchemaRegistryAclDelete(ctx, project, serviceName, acl)
			if err != nil && !avngen.IsNotFound(err) {
				return diag.Errorf("cannot delete `%s` kafka ACL for Schema Registry: %s", acl, err)
			}
		}
	}

	return nil
}
