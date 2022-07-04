package cassandra

import (
	"time"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func cassandraSchema() map[string]*schema.Schema {
	s := schemautil.ServiceCommonSchema()
	s[schemautil.ServiceTypeCassandra] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Cassandra server provided values",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
	}
	s[schemautil.ServiceTypeCassandra+"_user_config"] = schemautil.GenerateServiceUserConfigurationSchema(schemautil.ServiceTypeCassandra)

	return s
}

func ResourceCassandra() *schema.Resource {
	return &schema.Resource{
		Description:   "The Cassandra resource allows the creation and management of Aiven Cassandra services.",
		CreateContext: schemautil.ResourceServiceCreateWrapper(schemautil.ServiceTypeCassandra),
		ReadContext:   schemautil.ResourceServiceRead,
		UpdateContext: schemautil.ResourceServiceUpdate,
		DeleteContext: schemautil.ResourceServiceDelete,
		CustomizeDiff: customdiff.Sequence(
			schemautil.SetServiceTypeIfEmpty(schemautil.ServiceTypeCassandra),
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

		Schema: cassandraSchema(),
	}
}
