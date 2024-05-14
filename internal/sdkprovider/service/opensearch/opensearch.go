package opensearch

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader"
)

func opensearchSchema() map[string]*schema.Schema {
	s := schemautil.ServiceCommonSchemaWithUserConfig(schemautil.ServiceTypeOpenSearch)
	s[schemautil.ServiceTypeOpenSearch] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "OpenSearch server provided values",
		MaxItems:    1,
		Optional:    true,
		Sensitive:   true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"uris": {
					Type:        schema.TypeList,
					Computed:    true,
					Description: "OpenSearch server URIs.",
					Optional:    true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"opensearch_dashboards_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "URI for OpenSearch dashboard frontend",
					Sensitive:   true,
				},
				"kibana_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "URI for Kibana dashboard frontend",
				},
				"username": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "OpenSearch username",
				},
				"password": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "OpenSearch password",
					Sensitive:   true,
				},
			},
		},
	}
	return s
}

func ResourceOpenSearch() *schema.Resource {
	return &schema.Resource{
		Description:   "The OpenSearch resource allows the creation and management of Aiven OpenSearch services.",
		CreateContext: schemautil.ResourceServiceCreateWrapper(schemautil.ServiceTypeOpenSearch),
		ReadContext:   schemautil.ResourceServiceRead,
		UpdateContext: schemautil.ResourceServiceUpdate,
		DeleteContext: schemautil.ResourceServiceDelete,
		CustomizeDiff: customdiff.Sequence(
			schemautil.SetServiceTypeIfEmpty(schemautil.ServiceTypeOpenSearch),
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

		Schema:         opensearchSchema(),
		SchemaVersion:  1,
		StateUpgraders: stateupgrader.OpenSearch(),
	}
}
