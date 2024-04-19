package kafka

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/typeupgrader"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/v0/dist"
)

func aivenKafkaConnectSchema() map[string]*schema.Schema {
	kafkaConnectSchema := schemautil.ServiceCommonSchema()
	kafkaConnectSchema[schemautil.ServiceTypeKafkaConnect] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Sensitive:   true,
		Description: "Kafka Connect server provided values",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
	}
	kafkaConnectSchema[schemautil.ServiceTypeKafkaConnect+"_user_config"] = dist.ServiceTypeKafkaConnect()

	return kafkaConnectSchema
}

func ResourceKafkaConnect() *schema.Resource {
	return &schema.Resource{
		Description:   "The Kafka Connect resource allows the creation and management of Aiven Kafka Connect services.",
		CreateContext: schemautil.ResourceServiceCreateWrapper(schemautil.ServiceTypeKafkaConnect),
		ReadContext:   schemautil.ResourceServiceRead,
		UpdateContext: schemautil.ResourceServiceUpdate,
		DeleteContext: schemautil.ResourceServiceDelete,
		CustomizeDiff: customdiff.Sequence(
			schemautil.SetServiceTypeIfEmpty(schemautil.ServiceTypeKafkaConnect),
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
				schemautil.CustomizeDiffCheckStaticIPDisassociation,
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

		Schema: aivenKafkaConnectSchema(),
	}
}

func ResourceKafkaConnectStateUpgrade(
	_ context.Context,
	rawState map[string]any,
	_ any,
) (map[string]any, error) {
	userConfigSlice, ok := rawState["kafka_connect_user_config"].([]any)
	if !ok {
		return rawState, nil
	}

	if len(userConfigSlice) == 0 {
		return rawState, nil
	}

	userConfig, ok := userConfigSlice[0].(map[string]any)
	if !ok {
		return rawState, nil
	}

	err := typeupgrader.Map(userConfig, map[string]string{
		"static_ips": "bool",
	})
	if err != nil {
		return rawState, err
	}

	kafkaConnectSlice, ok := userConfig["kafka_connect"].([]any)
	if ok && len(kafkaConnectSlice) > 0 {
		kafkaConnect, ok := kafkaConnectSlice[0].(map[string]any)
		if ok {
			err = typeupgrader.Map(kafkaConnect, map[string]string{
				"consumer_fetch_max_bytes":           "int",
				"consumer_max_partition_fetch_bytes": "int",
				"consumer_max_poll_interval_ms":      "int",
				"consumer_max_poll_records":          "int",
				"offset_flush_interval_ms":           "int",
				"offset_flush_timeout_ms":            "int",
				"producer_max_request_size":          "int",
				"session_timeout_ms":                 "int",
			})
			if err != nil {
				return rawState, err
			}
		}
	}

	privateAccessSlice, ok := userConfig["private_access"].([]any)
	if ok && len(privateAccessSlice) > 0 {
		privateAccess, ok := privateAccessSlice[0].(map[string]any)
		if ok {
			err = typeupgrader.Map(privateAccess, map[string]string{
				"kafka_connect": "bool",
				"prometheus":    "bool",
			})
			if err != nil {
				return rawState, err
			}
		}
	}

	privateLinkAccessSlice, ok := userConfig["privatelink_access"].([]any)
	if ok && len(privateLinkAccessSlice) > 0 {
		privateLinkAccess, ok := privateLinkAccessSlice[0].(map[string]any)
		if ok {
			err := typeupgrader.Map(privateLinkAccess, map[string]string{
				"jolokia":       "bool",
				"kafka_connect": "bool",
				"prometheus":    "bool",
			})
			if err != nil {
				return rawState, err
			}
		}
	}

	publicAccessSlice, ok := userConfig["public_access"].([]any)
	if ok && len(publicAccessSlice) > 0 {
		publicAccess, ok := publicAccessSlice[0].(map[string]any)
		if ok {
			err := typeupgrader.Map(publicAccess, map[string]string{
				"kafka_connect": "bool",
				"prometheus":    "bool",
			})
			if err != nil {
				return rawState, err
			}
		}
	}

	return rawState, nil
}
