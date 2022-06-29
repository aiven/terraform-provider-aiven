package mysql

import (
	"time"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func aivenMySQLSchema() map[string]*schema.Schema {
	schemaMySQL := schemautil.ServiceCommonSchema()
	schemaMySQL[schemautil.ServiceTypeMySQL] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "MySQL specific server provided values",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
	}
	schemaMySQL[schemautil.ServiceTypeMySQL+"_user_config"] = schemautil.GenerateServiceUserConfigurationSchema(schemautil.ServiceTypeMySQL)

	return schemaMySQL
}
func ResourceMySQL() *schema.Resource {
	return &schema.Resource{
		Description:   "The MySQL resource allows the creation and management of Aiven MySQL services.",
		CreateContext: schemautil.ResourceServiceCreateWrapper(schemautil.ServiceTypeMySQL),
		ReadContext:   schemautil.ResourceServiceRead,
		UpdateContext: schemautil.ResourceServiceUpdate,
		DeleteContext: schemautil.ResourceServiceDelete,
		CustomizeDiff: customdiff.Sequence(
			schemautil.SetServiceTypeIfEmpty(schemautil.ServiceTypeMySQL),
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

		Schema: aivenMySQLSchema(),
	}
}
