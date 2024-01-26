package dragonfly

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/dist"
)

func dragonflySchema() map[string]*schema.Schema {
	s := schemautil.ServiceCommonSchema()
	s[schemautil.ServiceTypeDragonfly] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Dragonfly server provided values",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
	}
	s[schemautil.ServiceTypeDragonfly+"_user_config"] = dist.ServiceTypeDragonfly()

	return s
}

func ResourceDragonfly() *schema.Resource {
	return &schema.Resource{
		Description:   "The Dragonfly resource allows the creation and management of Aiven Dragonfly services.",
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
