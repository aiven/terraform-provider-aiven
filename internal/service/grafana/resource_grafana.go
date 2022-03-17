package grafana

import (
	"time"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func grafanaSchema() map[string]*schema.Schema {
	s := schemautil.ServiceCommonSchema()
	s[schemautil.ServiceTypeGrafana] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Grafana server provided values",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
	}
	s[schemautil.ServiceTypeGrafana+"_user_config"] = schemautil.GenerateServiceUserConfigurationSchema(schemautil.ServiceTypeGrafana)

	return s
}

func ResourceGrafana() *schema.Resource {
	return &schema.Resource{
		Description:   "The Grafana resource allows the creation and management of Aiven Grafana services.",
		CreateContext: schemautil.ResourceServiceCreateWrapper(schemautil.ServiceTypeGrafana),
		ReadContext:   schemautil.ResourceServiceRead,
		UpdateContext: schemautil.ResourceServiceUpdate,
		DeleteContext: schemautil.ResourceServiceDelete,
		CustomizeDiff: customdiff.Sequence(
			schemautil.SetServiceTypeIfEmpty(schemautil.ServiceTypeGrafana),
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
			StateContext: schemautil.ResourceServiceState,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: grafanaSchema(),
	}
}
