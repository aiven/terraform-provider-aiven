package thanos

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func thanosSchema() map[string]*schema.Schema {
	s := schemautil.ServiceCommonSchemaWithUserConfig(schemautil.ServiceTypeThanos)
	s[schemautil.ServiceTypeThanos] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Thanos server connection details.",
		MaxItems:    1,
		Optional:    true,
		Sensitive:   true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"uris": {
					Type:        schema.TypeList,
					Computed:    true,
					Description: "Thanos server URIs.",
					Optional:    true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"query_frontend_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Query frontend URI.",
					Sensitive:   true,
				},
				"query_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Query URI.",
					Sensitive:   true,
				},
				"receiver_ingesting_remote_write_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Receiver ingesting remote write URI.",
					Sensitive:   true,
				},
				"receiver_remote_write_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Receiver remote write URI.",
					Sensitive:   true,
				},
				"store_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Store URI.",
					Sensitive:   true,
				},
			},
		},
	}
	return s
}

func ResourceThanos() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates and manages an [Aiven for Metrics®](https://aiven.io/docs/products/metrics/concepts/metrics-overview) service.",
		CreateContext: schemautil.ResourceServiceCreateWrapper(schemautil.ServiceTypeThanos),
		ReadContext:   schemautil.ResourceServiceRead,
		UpdateContext: schemautil.ResourceServiceUpdate,
		DeleteContext: schemautil.ResourceServiceDelete,
		CustomizeDiff: customdiff.Sequence(
			schemautil.SetServiceTypeIfEmpty(schemautil.ServiceTypeThanos),
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
				schemautil.CustomizeDiffCheckStaticIPDisassociation,
				schemautil.CustomizeDiffCheckPlanAndStaticIpsCannotBeModifiedTogether,
			),
		),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: thanosSchema(),
	}
}
