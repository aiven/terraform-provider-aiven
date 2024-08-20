package dragonfly

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func dragonflySchema() map[string]*schema.Schema {
	s := schemautil.ServiceCommonSchemaWithUserConfig(schemautil.ServiceTypeDragonfly)
	s[schemautil.ServiceTypeDragonfly] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Dragonfly server provided values",
		MaxItems:    1,
		Optional:    true,
		Sensitive:   true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"uris": {
					Type:        schema.TypeList,
					Computed:    true,
					Description: "Dragonfly server URIs.",
					Optional:    true,
					Sensitive:   true,
					Elem: &schema.Schema{
						Type:      schema.TypeString,
						Sensitive: true,
					},
				},
				"slave_uris": {
					Type:        schema.TypeList,
					Computed:    true,
					Description: "Dragonfly slave server URIs.",
					Optional:    true,
					Sensitive:   true,
					Elem: &schema.Schema{
						Type:      schema.TypeString,
						Sensitive: true,
					},
				},
				"replica_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Sensitive:   true,
					Description: "Dragonfly replica server URI.",
				},
				"password": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Dragonfly password.",
					Sensitive:   true,
				},
			},
		},
	}
	return s
}

func ResourceDragonfly() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates and manages an [Aiven for Dragonfly®](https://aiven.io/docs/products/dragonfly/concepts/overview) service.",
		CreateContext: schemautil.ResourceServiceCreateWrapper(schemautil.ServiceTypeDragonfly),
		ReadContext:   schemautil.ResourceServiceRead,
		UpdateContext: schemautil.ResourceServiceUpdate,
		DeleteContext: schemautil.ResourceServiceDelete,
		CustomizeDiff: customdiff.Sequence(
			schemautil.SetServiceTypeIfEmpty(schemautil.ServiceTypeDragonfly),
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

		Schema: dragonflySchema(),
	}
}
