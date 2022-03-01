package redis

import (
	"time"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/service"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func redisSchema() map[string]*schema.Schema {
	s := service.ServiceCommonSchema()
	s[service.ServiceTypeRedis] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Redis server provided values",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
	}
	s[service.ServiceTypeRedis+"_user_config"] = schemautil.GenerateServiceUserConfigurationSchema(service.ServiceTypeRedis)

	return s
}

func ResourceRedis() *schema.Resource {
	return &schema.Resource{
		Description:   "The Redis resource allows the creation and management of Aiven Redis services.",
		CreateContext: service.ResourceServiceCreateWrapper(service.ServiceTypeRedis),
		ReadContext:   service.ResourceServiceRead,
		UpdateContext: service.ResourceServiceUpdate,
		DeleteContext: service.ResourceServiceDelete,
		CustomizeDiff: customdiff.Sequence(
			customdiff.Sequence(
				schemautil.SetServiceTypeIfEmpty(service.ServiceTypeRedis),
				customdiff.IfValueChange("disk_space",
					schemautil.DiskSpaceShouldNotBeEmpty,
					schemautil.CustomizeDiffCheckDiskSpace,
				),
			),
			customdiff.IfValueChange("service_integrations",
				schemautil.ServiceIntegrationShouldNotBeEmpty,
				schemautil.CustomizeDiffServiceIntegrationAfterCreation,
			),
			customdiff.Sequence(
				schemautil.CustomizeDiffCheckStaticIpDisassociation,
				schemautil.CustomizeDiffCheckPlanAndStaticIpsCannotBeModifiedTogether,
			),
		),
		Importer: &schema.ResourceImporter{
			StateContext: service.ResourceServiceState,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: redisSchema(),
	}
}
