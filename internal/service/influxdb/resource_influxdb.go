package influxdb

import (
	"time"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func influxDBSchema() map[string]*schema.Schema {
	s := schemautil.ServiceCommonSchema()
	s[schemautil.ServiceTypeInfluxDB] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "InfluxDB server provided values",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"database_name": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Name of the default InfluxDB database",
				},
			},
		},
	}
	s[schemautil.ServiceTypeInfluxDB+"_user_config"] = schemautil.GenerateServiceUserConfigurationSchema(schemautil.ServiceTypeInfluxDB)

	return s
}

func ResourceInfluxDB() *schema.Resource {
	return &schema.Resource{
		Description:   "The InfluxDB resource allows the creation and management of Aiven InfluxDB services.",
		CreateContext: schemautil.ResourceServiceCreateWrapper(schemautil.ServiceTypeInfluxDB),
		ReadContext:   schemautil.ResourceServiceRead,
		UpdateContext: schemautil.ResourceServiceUpdate,
		DeleteContext: schemautil.ResourceServiceDelete,
		CustomizeDiff: customdiff.Sequence(
			schemautil.SetServiceTypeIfEmpty(schemautil.ServiceTypeInfluxDB),
			customdiff.If(schemautil.ResourceShouldNotExist, schemautil.CustomizeDiffCheckProjectExistence),
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

		Schema: influxDBSchema(),
	}
}
