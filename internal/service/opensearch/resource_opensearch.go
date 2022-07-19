package opensearch

import (
	"time"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func opensearchSchema() map[string]*schema.Schema {
	s := schemautil.ServiceCommonSchema()
	s[schemautil.ServiceTypeOpensearch] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Opensearch server provided values",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"opensearch_dashboards_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "URI for Opensearch dashboard frontend",
					Sensitive:   true,
				},
			},
		},
	}
	s[schemautil.ServiceTypeOpensearch+"_user_config"] = schemautil.GenerateServiceUserConfigurationSchema(schemautil.ServiceTypeOpensearch)

	return s
}

func ResourceOpensearch() *schema.Resource {
	return &schema.Resource{
		Description:   "The Opensearch resource allows the creation and management of Aiven Opensearch services.",
		CreateContext: schemautil.ResourceServiceCreateWrapper(schemautil.ServiceTypeOpensearch),
		ReadContext:   schemautil.ResourceServiceRead,
		UpdateContext: schemautil.ResourceServiceUpdate,
		DeleteContext: schemautil.ResourceServiceDelete,
		CustomizeDiff: customdiff.Sequence(
			schemautil.SetServiceTypeIfEmpty(schemautil.ServiceTypeOpensearch),
			customdiff.IfValueChange("tag",
				schemautil.TagsShouldNotBeEmpty,
				schemautil.CustomizeDiffCheckUniqueTag,
			),
			customdiff.IfValueChange("disk_space",
				schemautil.DiskSpaceShouldNotBeEmpty,
				schemautil.CustomizeDiffCheckDiskSpace,
			),
			customdiff.IfValueChange("service_integrations",
				schemautil.ServiceIntegrationShouldNotBeEmpty,
				schemautil.CustomizeDiffServiceIntegrationAfterCreation,
			),
			customdiff.Sequence(
				schemautil.CustomizeDiffCheckPlanAndStaticIpsCannotBeModifiedTogether,
				schemautil.CustomizeDiffCheckStaticIpDisassociation,
			),
		),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: opensearchSchema(),
	}
}
