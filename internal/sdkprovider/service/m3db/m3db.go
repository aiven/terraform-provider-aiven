package m3db

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader"
)

func aivenM3DBSchema() map[string]*schema.Schema {
	s := schemautil.ServiceCommonSchemaWithUserConfig(schemautil.ServiceTypeM3)
	s[schemautil.ServiceTypeM3] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "M3DB server provided values",
		MaxItems:    1,
		Optional:    true,
		Sensitive:   true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"uris": {
					Type:        schema.TypeList,
					Computed:    true,
					Description: "M3DB server URIs.",
					Optional:    true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"http_cluster_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "M3DB cluster URI.",
				},
				"http_node_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "M3DB node URI.",
				},
				"influxdb_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "InfluxDB URI.",
				},
				"prometheus_remote_read_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Prometheus remote read URI.",
				},
				"prometheus_remote_write_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Prometheus remote write URI.",
				},
			},
		},
	}
	return s
}
func ResourceM3DB() *schema.Resource {
	return &schema.Resource{
		Description:   "The M3 DB resource allows the creation and management of Aiven M3 services.",
		CreateContext: schemautil.ResourceServiceCreateWrapper(schemautil.ServiceTypeM3),
		ReadContext:   schemautil.ResourceServiceRead,
		UpdateContext: schemautil.ResourceServiceUpdate,
		DeleteContext: schemautil.ResourceServiceDelete,
		CustomizeDiff: customdiff.Sequence(
			schemautil.SetServiceTypeIfEmpty(schemautil.ServiceTypeM3),
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

		Schema:         aivenM3DBSchema(),
		SchemaVersion:  1,
		StateUpgraders: stateupgrader.M3DB(),
	}
}
