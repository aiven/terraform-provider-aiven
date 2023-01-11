package v0

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/typeupgrader"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/v0/dist"
)

func aivenKafkaMirrormakerSchema() map[string]*schema.Schema {
	kafkaMMSchema := schemautil.ServiceCommonSchema()
	kafkaMMSchema[schemautil.ServiceTypeKafkaMirrormaker] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Kafka MirrorMaker 2 server provided values",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
	}
	kafkaMMSchema[schemautil.ServiceTypeKafkaMirrormaker+"_user_config"] = dist.ServiceTypeKafkaMirrormaker()

	return kafkaMMSchema
}

func ResourceKafkaMirrormaker() *schema.Resource {
	return &schema.Resource{
		Description:   "The Kafka MirrorMaker resource allows the creation and management of Aiven Kafka MirrorMaker 2 services.",
		CreateContext: schemautil.ResourceServiceCreateWrapper(schemautil.ServiceTypeKafkaMirrormaker),
		ReadContext:   schemautil.ResourceServiceRead,
		UpdateContext: schemautil.ResourceServiceUpdate,
		DeleteContext: schemautil.ResourceServiceDelete,
		CustomizeDiff: customdiff.Sequence(
			schemautil.SetServiceTypeIfEmpty(schemautil.ServiceTypeKafkaMirrormaker),
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

		Schema: aivenKafkaMirrormakerSchema(),
	}
}

func ResourceKafkaMirrormakerStateUpgrade(
	_ context.Context,
	rawState map[string]interface{},
	_ interface{},
) (map[string]interface{}, error) {
	userConfigSlice, ok := rawState["kafka_mirrormaker_user_config"].([]interface{})
	if !ok {
		return rawState, nil
	}

	if len(userConfigSlice) == 0 {
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

	kafkaMirrormakerSlice, ok := userConfig["kafka_mirrormaker"].([]interface{})
	if ok && len(kafkaMirrormakerSlice) > 0 {
		kafkaMirrormaker, ok := kafkaMirrormakerSlice[0].(map[string]interface{})
		if ok {
			err = typeupgrader.Map(kafkaMirrormaker, map[string]string{
				"emit_checkpoints_enabled":            "bool",
				"emit_checkpoints_interval_seconds":   "int",
				"refresh_groups_enabled":              "bool",
				"refresh_groups_interval_seconds":     "int",
				"refresh_topics_enabled":              "bool",
				"refresh_topics_interval_seconds":     "int",
				"sync_group_offsets_enabled":          "bool",
				"sync_group_offsets_interval_seconds": "int",
				"sync_topic_configs_enabled":          "bool",
				"tasks_max_per_cpu":                   "int",
			})
			if err != nil {
				return rawState, err
			}
		}
	}

	return rawState, nil
}
