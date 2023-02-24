package kafka

import (
	"context"
	"time"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/typeupgrader"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var aivenKafkaTopicSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,

	"topic_name": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: userconfig.Desc("The name of the topic.").ForceNew().Build(),
	},
	"partitions": {
		Type:        schema.TypeInt,
		Required:    true,
		Description: "The number of partitions to create in the topic.",
	},
	"replication": {
		Type:        schema.TypeInt,
		Required:    true,
		Description: "The replication factor for the topic.",
	},
	"termination_protection": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "It is a Terraform client-side deletion protection, which prevents a Kafka topic from being deleted. It is recommended to enable this for any production Kafka topic containing critical data.",
	},
	"tag": {
		Type:        schema.TypeSet,
		Description: "Kafka Topic tag.",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"key": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 64),
					Description:  userconfig.Desc("Topic tag key.").MaxLen(64).Build(),
				},
				"value": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringLenBetween(0, 256),
					Description:  userconfig.Desc("Topic tag value.").MaxLen(256).Build(),
				},
			},
		},
	},
	"config": {
		Type:             schema.TypeList,
		Description:      "Kafka topic configuration",
		Optional:         true,
		MaxItems:         1,
		DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"cleanup_policy": {
					Type:             schema.TypeString,
					Description:      "cleanup.policy value",
					Optional:         true,
					DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
				},
				"compression_type": {
					Type:             schema.TypeString,
					Description:      "compression.type value",
					Optional:         true,
					DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
				},
				"delete_retention_ms": {
					Type:             schema.TypeString,
					Description:      "delete.retention.ms value",
					Optional:         true,
					DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
				},
				"file_delete_delay_ms": {
					Type:             schema.TypeString,
					Description:      "file.delete.delay.ms value",
					Optional:         true,
					DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
				},
				"flush_messages": {
					Type:             schema.TypeString,
					Description:      "flush.messages value",
					Optional:         true,
					DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
				},
				"flush_ms": {
					Type:             schema.TypeString,
					Description:      "flush.ms value",
					Optional:         true,
					DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
				},
				"index_interval_bytes": {
					Type:             schema.TypeString,
					Description:      "index.interval.bytes value",
					Optional:         true,
					DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
				},
				"max_compaction_lag_ms": {
					Type:             schema.TypeString,
					Description:      "max.compaction.lag.ms value",
					Optional:         true,
					DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
				},
				"max_message_bytes": {
					Type:             schema.TypeString,
					Description:      "max.message.bytes value",
					Optional:         true,
					DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
				},
				"message_downconversion_enable": {
					Type:             schema.TypeString,
					Description:      "message.downconversion.enable value",
					Optional:         true,
					DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
				},
				"message_format_version": {
					Type:             schema.TypeString,
					Description:      "message.format.version value",
					Optional:         true,
					DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
				},
				"message_timestamp_difference_max_ms": {
					Type:             schema.TypeString,
					Description:      "message.timestamp.difference.max.ms value",
					Optional:         true,
					DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
				},
				"message_timestamp_type": {
					Type:             schema.TypeString,
					Description:      "message.timestamp.type value",
					Optional:         true,
					DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
				},
				"min_cleanable_dirty_ratio": {
					Type:             schema.TypeString,
					Description:      "min.cleanable.dirty.ratio value",
					Optional:         true,
					DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
				},
				"min_compaction_lag_ms": {
					Type:             schema.TypeString,
					Description:      "min.compaction.lag.ms value",
					Optional:         true,
					DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
				},
				"min_insync_replicas": {
					Type:             schema.TypeString,
					Description:      "min.insync.replicas value",
					Optional:         true,
					DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
				},
				"preallocate": {
					Type:             schema.TypeString,
					Description:      "preallocate value",
					Optional:         true,
					DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
				},
				"retention_bytes": {
					Type:             schema.TypeString,
					Description:      "retention.bytes value",
					Optional:         true,
					DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
				},
				"retention_ms": {
					Type:             schema.TypeString,
					Description:      "retention.ms value",
					Optional:         true,
					DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
				},
				"segment_bytes": {
					Type:             schema.TypeString,
					Description:      "segment.bytes value",
					Optional:         true,
					DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
				},
				"segment_index_bytes": {
					Type:             schema.TypeString,
					Description:      "segment.index.bytes value",
					Optional:         true,
					DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
				},
				"segment_jitter_ms": {
					Type:             schema.TypeString,
					Description:      "segment.jitter.ms value",
					Optional:         true,
					DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
				},
				"segment_ms": {
					Type:             schema.TypeString,
					Description:      "segment.ms value",
					Optional:         true,
					DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
				},
				"unclean_leader_election_enable": {
					Type:             schema.TypeString,
					Description:      "unclean.leader.election.enable value",
					Optional:         true,
					DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
				},
			},
		},
	},
}

func ResourceKafkaTopic() *schema.Resource {
	return &schema.Resource{
		Description: "The Kafka Topic resource allows the creation and management of Aiven Kafka Topics.",
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Read:   schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},
		Schema: aivenKafkaTopicSchema,
	}
}

func ResourceKafkaTopicStateUpgrade(
	_ context.Context,
	rawState map[string]interface{},
	_ interface{},
) (map[string]interface{}, error) {
	configSlice, ok := rawState["config"].([]interface{})
	if !ok {
		return rawState, nil
	}

	if len(configSlice) == 0 {
		return rawState, nil
	}

	config, ok := configSlice[0].(map[string]interface{})
	if !ok {
		return rawState, nil
	}

	err := typeupgrader.Map(config, map[string]string{
		"message_downconversion_enable":  "bool",
		"min_cleanable_dirty_ratio":      "float",
		"preallocate":                    "bool",
		"unclean_leader_election_enable": "bool",
	})
	if err != nil {
		return rawState, err
	}

	return rawState, nil
}
