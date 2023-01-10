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

func aivenFlinkSchema() map[string]*schema.Schema {
	aivenFlinkSchema := schemautil.ServiceCommonSchema()
	aivenFlinkSchema[schemautil.ServiceTypeFlink] = &schema.Schema{
		Type:        schema.TypeList,
		MaxItems:    1,
		Computed:    true,
		Description: "Flink server provided values",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"host_ports": {
					Type:        schema.TypeList,
					Computed:    true,
					Description: "Host and Port of a Flink server",
					Optional:    true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
			},
		},
	}
	aivenFlinkSchema[schemautil.ServiceTypeFlink+"_user_config"] = dist.ServiceTypeFlink()

	return aivenFlinkSchema
}

func ResourceFlinkResourceV0() *schema.Resource {
	return &schema.Resource{
		Description:   "The Flink resource allows the creation and management of Aiven Flink services.",
		CreateContext: schemautil.ResourceServiceCreateWrapper(schemautil.ServiceTypeFlink),
		ReadContext:   schemautil.ResourceServiceRead,
		UpdateContext: schemautil.ResourceServiceUpdate,
		DeleteContext: schemautil.ResourceServiceDelete,
		CustomizeDiff: customdiff.Sequence(
			schemautil.SetServiceTypeIfEmpty(schemautil.ServiceTypeFlink),
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

		Schema: aivenFlinkSchema(),
	}
}

func ResourceFlinkStateUpgradeV0(
	_ context.Context,
	rawState map[string]interface{},
	_ interface{},
) (map[string]interface{}, error) {
	userConfigSlice, ok := rawState["flink_user_config"].([]interface{})
	if !ok {
		return rawState, nil
	}

	userConfig, ok := userConfigSlice[0].(map[string]interface{})
	if !ok {
		return rawState, nil
	}

	err := typeupgrader.Map(userConfig, map[string]string{
		"execution_checkpointing_interval_ms":        "int",
		"execution_checkpointing_timeout_ms":         "int",
		"number_of_task_slots":                       "int",
		"parallelism_default":                        "int",
		"restart_strategy_delay_sec":                 "int",
		"restart_strategy_failure_rate_interval_min": "int",
		"restart_strategy_max_failures":              "int",
	})
	if err != nil {
		return rawState, err
	}

	privateLinkAccessSlice, ok := userConfig["privatelink_access"].([]interface{})
	if ok && len(privateLinkAccessSlice) > 0 {
		privateLinkAccess, ok := privateLinkAccessSlice[0].(map[string]interface{})
		if ok {
			err := typeupgrader.Map(privateLinkAccess, map[string]string{
				"flink":      "bool",
				"prometheus": "bool",
			})
			if err != nil {
				return rawState, err
			}
		}
	}

	return rawState, nil
}