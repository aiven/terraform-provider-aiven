package clickhouse

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func clickhouseSchema() map[string]*schema.Schema {
	s := schemautil.ServiceCommonSchemaWithUserConfig(schemautil.ServiceTypeClickhouse)
	s[schemautil.ServiceTypeClickhouse] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Values provided by the ClickHouse server.",
		MaxItems:    1,
		Optional:    true,
		Sensitive:   true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"uris": {
					Type:        schema.TypeList,
					Computed:    true,
					Description: "ClickHouse server URIs.",
					Optional:    true,
					Sensitive:   true,
					Elem: &schema.Schema{
						Type:      schema.TypeString,
						Sensitive: true,
					},
				},
			},
		},
	}
	s["service_integrations"] = &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Description: "Integrations with other services. Service integrations are only applied at service creation.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"source_service_name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Name of the source service.",
				},
				"integration_type": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Type of the service integration. Supported integrations are `clickhouse_kafka` and `clickhouse_postgresql`.",
				},
			},
		},
	}

	return s
}

func ResourceClickhouse() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates and manages an [Aiven for ClickHouse®](https://aiven.io/docs/products/clickhouse/concepts/features-overview) service.",
		CreateContext: schemautil.ResourceServiceCreateWrapper(schemautil.ServiceTypeClickhouse),
		ReadContext:   schemautil.ResourceServiceRead,
		UpdateContext: schemautil.ResourceServiceUpdate,
		DeleteContext: schemautil.ResourceServiceDelete,
		CustomizeDiff: customdiff.Sequence(
			schemautil.SetServiceTypeIfEmpty(schemautil.ServiceTypeClickhouse),
			schemautil.CustomizeDiffDisallowMultipleManyToOneKeys,
			customdiff.IfValueChange("tag",
				schemautil.TagsShouldNotBeEmpty,
				schemautil.CustomizeDiffCheckUniqueTag,
			),
			customdiff.IfValueChange("disk_space",
				schemautil.DiskSpaceShouldNotBeEmpty,
				schemautil.CustomizeDiffCheckDiskSpace,
			),
			customdiff.IfValueChange("additional_disk_space",
				schemautil.DiskSpaceShouldNotBeEmpty,
				schemautil.CustomizeDiffCheckDiskSpace,
			),
			customdiff.IfValueChange("service_integrations",
				schemautil.ServiceIntegrationShouldNotBeEmpty,
				schemautil.CustomizeDiffServiceIntegrationAfterCreation,
			),
			customdiff.Sequence(
				schemautil.CustomizeDiffCheckPlanAndStaticIpsCannotBeModifiedTogether,
				schemautil.CustomizeDiffCheckStaticIPDisassociation,
			),
		),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: clickhouseSchema(),
	}
}
