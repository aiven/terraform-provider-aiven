package v0

import (
	"context"
	"time"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/typeupgrader"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/v0/dist"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func aivenM3AggregatorSchema() map[string]*schema.Schema {
	schemaM3 := schemautil.ServiceCommonSchema()
	schemaM3[schemautil.ServiceTypeM3Aggregator] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "M3 aggregator specific server provided values",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
	}
	schemaM3[schemautil.ServiceTypeM3Aggregator+"_user_config"] = dist.ServiceTypeM3aggregator()

	return schemaM3
}
func ResourceM3AggregatorResourceV0() *schema.Resource {
	return &schema.Resource{
		Description:   "The M3 Aggregator resource allows the creation and management of Aiven M3 Aggregator services.",
		CreateContext: schemautil.ResourceServiceCreateWrapper(schemautil.ServiceTypeM3Aggregator),
		ReadContext:   schemautil.ResourceServiceRead,
		UpdateContext: schemautil.ResourceServiceUpdate,
		DeleteContext: schemautil.ResourceServiceDelete,
		CustomizeDiff: customdiff.Sequence(
			schemautil.SetServiceTypeIfEmpty(schemautil.ServiceTypeM3Aggregator),
			schemautil.CustomizeDiffDisallowMultipleManyToOneKeys,
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

		Schema: aivenM3AggregatorSchema(),
	}
}

func ResourceM3AggregatorStateUpgradeV0(
	_ context.Context,
	rawState map[string]interface{},
	_ interface{},
) (map[string]interface{}, error) {
	userConfigSlice, ok := rawState["m3aggregator_user_config"].([]interface{})
	if !ok {
		return rawState, nil
	}

	userConfig, ok := userConfigSlice[0].(map[string]interface{})
	if !ok {
		return rawState, nil
	}

	err := typeupgrader.Map(userConfig, map[string]string{
		"static_ips": "bool",
	})
	if err != nil {
		return rawState, err
	}

	return rawState, nil
}