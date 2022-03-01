package clickhouse

import (
	"time"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/service"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func clickhouseSchema() map[string]*schema.Schema {
	s := service.ServiceCommonSchema()
	s[service.ServiceTypeClickhouse] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Clickhouse server provided values",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
	}
	s[service.ServiceTypeClickhouse+"_user_config"] = schemautil.GenerateServiceUserConfigurationSchema(service.ServiceTypeClickhouse)

	return s
}

func ResourceClickhouse() *schema.Resource {
	return &schema.Resource{
		Description:   "The Clickhouse resource allows the creation and management of Aiven Clickhouse services.",
		CreateContext: service.ResourceServiceCreateWrapper(service.ServiceTypeClickhouse),
		ReadContext:   service.ResourceServiceRead,
		UpdateContext: service.ResourceServiceUpdate,
		DeleteContext: service.ResourceServiceDelete,
		CustomizeDiff: customdiff.Sequence(
			schemautil.SetServiceTypeIfEmpty(service.ServiceTypeClickhouse),
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
			StateContext: service.ResourceServiceState,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: clickhouseSchema(),
	}
}
