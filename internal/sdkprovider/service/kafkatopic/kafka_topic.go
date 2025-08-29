package kafkatopic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/aiven/go-client-codegen/handler/kafkatopic"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/kafkatopicrepository"
)

const configField = "config"

var (
	errTopicAlreadyExists            = fmt.Errorf("topic conflict, already exists")
	errLocalRetentionBytesOverflow   = fmt.Errorf("local_retention_bytes must not be more than retention_bytes value")
	errLocalRetentionBytesDependency = fmt.Errorf("local_retention_bytes can't be set without retention_bytes")
)

func aivenKafkaTopicConfigSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"cleanup_policy": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringInSlice(kafkatopic.CleanupPolicyTypeChoices(), false),
			Description:  userconfig.Desc("The retention policy to use on old segments. Possible values include 'delete', 'compact', or a comma-separated list of them. The default policy ('delete') will discard old segments when their retention time or size limit has been reached. The 'compact' setting will enable log compaction on the topic.").PossibleValuesString(kafkatopic.CleanupPolicyTypeChoices()...).Build(),
		},
		"compression_type": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringInSlice(kafkatopic.CompressionTypeChoices(), false),
			Description:  userconfig.Desc("Specify the final compression type for a given topic. This configuration accepts the standard compression codecs ('gzip', 'snappy', 'lz4', 'zstd'). It additionally accepts 'uncompressed' which is equivalent to no compression; and 'producer' which means retain the original compression codec set by the producer.").PossibleValuesString(kafkatopic.CompressionTypeChoices()...).Build(),
		},
		"delete_retention_ms": {
			Type:        schema.TypeString,
			Description: "The amount of time to retain delete tombstone markers for log compacted topics. This setting also gives a bound on the time in which a consumer must complete a read if they begin from offset 0 to ensure that they get a valid snapshot of the final stage (otherwise delete tombstones may be collected before they complete their scan).",
			Optional:    true,
		},
		"file_delete_delay_ms": {
			Type:        schema.TypeString,
			Description: "The time to wait before deleting a file from the filesystem.",
			Optional:    true,
		},
		"flush_messages": {
			Type:        schema.TypeString,
			Description: "This setting allows specifying an interval at which we will force an fsync of data written to the log. For example if this was set to 1 we would fsync after every message; if it were 5 we would fsync after every five messages. In general we recommend you not set this and use replication for durability and allow the operating system's background flush capabilities as it is more efficient.",
			Optional:    true,
		},
		"flush_ms": {
			Type:        schema.TypeString,
			Description: "This setting allows specifying a time interval at which we will force an fsync of data written to the log. For example if this was set to 1000 we would fsync after 1000 ms had passed. In general we recommend you not set this and use replication for durability and allow the operating system's background flush capabilities as it is more efficient.",
			Optional:    true,
		},
		"index_interval_bytes": {
			Type:        schema.TypeString,
			Description: "This setting controls how frequently Kafka adds an index entry to its offset index. The default setting ensures that we index a message roughly every 4096 bytes. More indexing allows reads to jump closer to the exact position in the log but makes the index larger. You probably don't need to change this.",
			Optional:    true,
		},
		"max_compaction_lag_ms": {
			Type:        schema.TypeString,
			Description: "The maximum time a message will remain ineligible for compaction in the log. Only applicable for logs that are being compacted.",
			Optional:    true,
		},
		"max_message_bytes": {
			Type:        schema.TypeString,
			Description: "The largest record batch size allowed by Kafka (after compression if compression is enabled). If this is increased and there are consumers older than 0.10.2, the consumers' fetch size must also be increased so that the they can fetch record batches this large. In the latest message format version, records are always grouped into batches for efficiency. In previous message format versions, uncompressed records are not grouped into batches and this limit only applies to a single record in that case.",
			Optional:    true,
		},
		"message_downconversion_enable": {
			Type:        schema.TypeBool,
			Description: "This configuration controls whether down-conversion of message formats is enabled to satisfy consume requests. When set to false, broker will not perform down-conversion for consumers expecting an older message format. The broker responds with UNSUPPORTED_VERSION error for consume requests from such older clients. This configuration does not apply to any message format conversion that might be required for replication to followers.",
			Optional:    true,
		},
		"message_format_version": {
			Type:     schema.TypeString,
			Optional: true,
			// MessageFormatVersionTypeChoices has `None` value that available in the Response, we use it for the validation.
			// ConfigMessageFormatVersionTypeChoices has only valid values (wo `None), we expose it to the documentation.
			ValidateFunc: validation.StringInSlice(kafkatopic.MessageFormatVersionTypeChoices(), false),
			Description:  userconfig.Desc("Specify the message format version the broker will use to append messages to the logs. The value should be a valid ApiVersion. Some examples are: 0.8.2, 0.9.0.0, 0.10.0, check ApiVersion for more details. By setting a particular message format version, the user is certifying that all the existing messages on disk are smaller or equal than the specified version. Setting this value incorrectly will cause consumers with older versions to break as they will receive messages with a format that they don't understand. Deprecated in Kafka 4.0+: this configuration is removed and any supplied value will be ignored; for services upgraded to 4.0+, the returned value may be 'None'.").PossibleValuesString(kafkatopic.ConfigMessageFormatVersionTypeChoices()...).Build(),
		},
		"message_timestamp_difference_max_ms": {
			Type:        schema.TypeString,
			Description: "The maximum difference allowed between the timestamp when a broker receives a message and the timestamp specified in the message. If message.timestamp.type=CreateTime, a message will be rejected if the difference in timestamp exceeds this threshold. This configuration is ignored if message.timestamp.type=LogAppendTime.",
			Optional:    true,
		},
		"message_timestamp_type": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringInSlice(kafkatopic.MessageTimestampTypeChoices(), false),
			Description:  userconfig.Desc("Define whether the timestamp in the message is message create time or log append time.").PossibleValuesString(kafkatopic.MessageTimestampTypeChoices()...).Build(),
		},
		"min_cleanable_dirty_ratio": {
			Type:        schema.TypeFloat,
			Description: "This configuration controls how frequently the log compactor will attempt to clean the log (assuming log compaction is enabled). By default we will avoid cleaning a log where more than 50% of the log has been compacted. This ratio bounds the maximum space wasted in the log by duplicates (at 50% at most 50% of the log could be duplicates). A higher ratio will mean fewer, more efficient cleanings but will mean more wasted space in the log. If the max.compaction.lag.ms or the min.compaction.lag.ms configurations are also specified, then the log compactor considers the log to be eligible for compaction as soon as either: (i) the dirty ratio threshold has been met and the log has had dirty (uncompacted) records for at least the min.compaction.lag.ms duration, or (ii) if the log has had dirty (uncompacted) records for at most the max.compaction.lag.ms period.",
			Optional:    true,
		},
		"min_compaction_lag_ms": {
			Type:        schema.TypeString,
			Description: "The minimum time a message will remain uncompacted in the log. Only applicable for logs that are being compacted.",
			Optional:    true,
		},
		"min_insync_replicas": {
			Type:        schema.TypeString,
			Description: "When a producer sets acks to 'all' (or '-1'), this configuration specifies the minimum number of replicas that must acknowledge a write for the write to be considered successful. If this minimum cannot be met, then the producer will raise an exception (either NotEnoughReplicas or NotEnoughReplicasAfterAppend). When used together, min.insync.replicas and acks allow you to enforce greater durability guarantees. A typical scenario would be to create a topic with a replication factor of 3, set min.insync.replicas to 2, and produce with acks of 'all'. This will ensure that the producer raises an exception if a majority of replicas do not receive a write.",
			Optional:    true,
		},
		"preallocate": {
			Type:        schema.TypeBool,
			Description: "True if we should preallocate the file on disk when creating a new log segment.",
			Optional:    true,
		},
		"retention_bytes": {
			Type:        schema.TypeString,
			Description: "This configuration controls the maximum size a partition (which consists of log segments) can grow to before we will discard old log segments to free up space if we are using the 'delete' retention policy. By default there is no size limit only a time limit. Since this limit is enforced at the partition level, multiply it by the number of partitions to compute the topic retention in bytes.",
			Optional:    true,
		},
		"retention_ms": {
			Type:        schema.TypeString,
			Description: "This configuration controls the maximum time we will retain a log before we will discard old log segments to free up space if we are using the 'delete' retention policy. This represents an SLA on how soon consumers must read their data. If set to -1, no time limit is applied.",
			Optional:    true,
		},
		"segment_bytes": {
			Type:        schema.TypeString,
			Description: "This configuration controls the size of the index that maps offsets to file positions. We preallocate this index file and shrink it only after log rolls. You generally should not need to change this setting.",
			Optional:    true,
		},
		"segment_index_bytes": {
			Type:        schema.TypeString,
			Description: "This configuration controls the size of the index that maps offsets to file positions. We preallocate this index file and shrink it only after log rolls. You generally should not need to change this setting.",
			Optional:    true,
		},
		"segment_jitter_ms": {
			Type:        schema.TypeString,
			Description: "The maximum random jitter subtracted from the scheduled segment roll time to avoid thundering herds of segment rolling",
			Optional:    true,
		},
		"segment_ms": {
			Type:        schema.TypeString,
			Description: "This configuration controls the period of time after which Kafka will force the log to roll even if the segment file isn't full to ensure that retention can delete or compact old data. Setting this to a very low value has consequences, and the Aiven management plane ignores values less than 10 seconds.",
			Optional:    true,
		},
		"unclean_leader_election_enable": {
			Type:        schema.TypeBool,
			Description: "Indicates whether to enable replicas not in the ISR set to be elected as leader as a last resort, even though doing so may result in data loss.",
			Optional:    true,
		},
		"remote_storage_enable": {
			Type:        schema.TypeBool,
			Description: "Indicates whether tiered storage should be enabled.",
			Optional:    true,
		},
		"local_retention_bytes": {
			Type:        schema.TypeString,
			Description: "This configuration controls the maximum bytes tiered storage will retain segment files locally before it will discard old log segments to free up space. If set to -2, the limit is equal to overall retention time. If set to -1, no limit is applied but it's possible only if overall retention is also -1.",
			Optional:    true,
		},
		"local_retention_ms": {
			Type:        schema.TypeString,
			Description: "This configuration controls the maximum time tiered storage will retain segment files locally before it will discard old log segments to free up space. If set to -2, the time limit is equal to overall retention time. If set to -1, no time limit is applied but it's possible only if overall retention is also -1.",
			Optional:    true,
		},
		"inkless_enable": {
			Type:        schema.TypeBool,
			Description: "Creates a [diskless topic](https://aiven.io/docs/products/diskless). You can only do this when you create the topic and you cannot change it later. Diskless topics are only available for bring your own cloud (BYOC) services that have the feature enabled.",
			Optional:    true,
		},
	}
}

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
	"topic_description": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "The description of the topic",
	},
	"owner_user_group_id": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "The ID of the user group that owns the topic. Assigning ownership to decentralize topic management is part of [Aiven for Apache Kafka® governance](https://aiven.io/docs/products/kafka/concepts/governance-overview).",
	},
	"termination_protection": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "Prevents topics from being deleted by Terraform. It's recommended for topics containing critical data. **Topics can still be deleted in the Aiven Console.**",
	},
	"tag": {
		Type:        schema.TypeSet,
		Description: "Tags for the topic.",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"key": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 64),
					Description:  userconfig.Desc("Tag key.").MaxLen(64).Build(),
				},
				"value": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringLenBetween(0, 256),
					Description:  userconfig.Desc("Tag value.").MaxLen(256).Build(),
				},
			},
		},
	},
	configField: {
		Type:             schema.TypeList,
		Description:      "[Advanced parameters](https://aiven.io/docs/products/kafka/reference/advanced-params) to configure topics.",
		Optional:         true,
		MaxItems:         1,
		DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
		Elem: &schema.Resource{
			Schema: aivenKafkaTopicConfigSchema(),
		},
	},
}

func ResourceKafkaTopic() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates and manages an Aiven for Apache Kafka® [topic](https://aiven.io/docs/products/kafka/concepts).",
		CreateContext: resourceKafkaTopicCreate,
		ReadContext:   resourceKafkaTopicReadResource,
		UpdateContext: resourceKafkaTopicUpdate,
		DeleteContext: resourceKafkaTopicDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts:       schemautil.DefaultResourceTimeouts(),
		Schema:         aivenKafkaTopicSchema,
		SchemaVersion:  1,
		StateUpgraders: stateupgrader.KafkaTopic(),
		CustomizeDiff: func(ctx context.Context, d *schema.ResourceDiff, m interface{}) error {
			oldPartitions, newPartitions := d.GetChange("partitions")

			assertedOldPartitions, ok := oldPartitions.(int)
			if !ok {
				return nil
			}

			assertedNewPartitions, ok := newPartitions.(int)
			if !ok {
				return nil
			}

			if assertedOldPartitions > assertedNewPartitions {
				return errors.New("number of partitions cannot be decreased")
			}

			retentionBytes, rOk := d.GetOk("config.0.retention_bytes")
			localRetentionBytes, lOk := d.GetOk("config.0.local_retention_bytes")

			switch {
			case lOk && !rOk:
				return errLocalRetentionBytesDependency
			case lOk && rOk:
				r, err := strconv.ParseInt(retentionBytes.(string), 10, 64)
				if err != nil {
					return err
				}

				l, err := strconv.ParseInt(localRetentionBytes.(string), 10, 64)
				if err != nil {
					return err
				}

				if r < l {
					return errLocalRetentionBytesOverflow
				}
			}

			// Validates topic conflict for new topics
			if d.Id() != "" {
				return nil
			}

			// A new topic
			client := m.(*aiven.Client)
			project := d.Get("project").(string)
			serviceName := d.Get("service_name").(string)
			topicName := d.Get("topic_name").(string)
			exists, err := kafkatopicrepository.New(client.KafkaTopics).Exists(ctx, project, serviceName, topicName)
			if err != nil {
				return err
			}

			if exists {
				return fmt.Errorf("%w: %q", errTopicAlreadyExists, topicName)
			}

			return nil
		},
	}
}

func resourceKafkaTopicCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	topicName := d.Get("topic_name").(string)
	partitions := d.Get("partitions").(int)
	replication := d.Get("replication").(int)
	topicDescription := d.Get("topic_description").(string)
	ownerUserGroupID := d.Get("owner_user_group_id").(string)

	config, err := getKafkaTopicConfig(d)
	if err != nil {
		return diag.Errorf("config to json error: %s", err)
	}

	createRequest := aiven.CreateKafkaTopicRequest{
		Partitions:  &partitions,
		Replication: &replication,
		TopicName:   topicName,
		Config:      config,
		Tags:        getTags(d),
	}

	if topicDescription != "" {
		createRequest.TopicDescription = &topicDescription
	}

	if ownerUserGroupID != "" {
		createRequest.OwnerUserGroupId = &ownerUserGroupID
	}

	client := m.(*aiven.Client)
	err = kafkatopicrepository.New(client.KafkaTopics).Create(ctx, project, serviceName, createRequest)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(project, serviceName, topicName))

	// We do not call a Kafka Topic read here to speed up the performance.
	// However, in the case of Kafka Topic resource getting a computed field
	// in the future, a read operation should be called after creation.
	return nil
}

func getTags(d *schema.ResourceData) []aiven.KafkaTopicTag {
	tagSet := d.Get("tag").(*schema.Set)
	tags := make([]aiven.KafkaTopicTag, tagSet.Len())
	for i, tagD := range tagSet.List() {
		tagM := tagD.(map[string]interface{})
		tag := aiven.KafkaTopicTag{
			Key:   tagM["key"].(string),
			Value: tagM["value"].(string),
		}

		tags[i] = tag
	}

	return tags
}

// getKafkaTopicConfig converts schema.ResourceData into aiven.KafkaTopicConfig
// Takes manifest values only
func getKafkaTopicConfig(d *schema.ResourceData) (aiven.KafkaTopicConfig, error) {
	empty := aiven.KafkaTopicConfig{}
	configs := d.GetRawConfig().AsValueMap()[configField]
	if configs.IsNull() || len(configs.AsValueSlice()) == 0 {
		return empty, nil
	}

	config := make(map[string]any)
	for k, v := range configs.AsValueSlice()[0].AsValueMap() {
		if v.IsNull() {
			continue
		}

		key := fmt.Sprintf("%s.0.%s", configField, k)
		value, err := typedConfigValue(k, d.Get(key))
		if err != nil {
			return empty, fmt.Errorf("error converting config field %q: %w", k, err)
		}
		config[k] = value
	}

	// Converts to json and loads values to the struct
	b, err := json.Marshal(config)
	if err != nil {
		return empty, err
	}

	var result aiven.KafkaTopicConfig
	err = json.Unmarshal(b, &result)
	return result, err
}

func resourceKafkaTopicRead(ctx context.Context, d *schema.ResourceData, m interface{}, isResource bool) diag.Diagnostics {
	project, serviceName, topicName, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	client := m.(*aiven.Client)
	topic, err := kafkatopicrepository.New(client.KafkaTopics).Read(ctx, project, serviceName, topicName)
	// Topics are destroyed when kafka is off
	// https://aiven.io/docs/platform/concepts/service-power-cycle
	// So it's better to recreate them, than make user to clear the state manually
	// Recreates missing topics:
	// 1. if server returns 404
	// 2. is resource (not datasource)
	// 3. only for resources with existing state, not imports
	if err != nil {
		if !isResource {
			return diag.FromErr(err)
		}

		// This might help with knowing what ResourceReadHandleNotFound has set
		log.Printf("[DEBUG] KafkaTopic get error %s, known=%v, is_new=%v", err, schemautil.IsUnknownResource(err), d.IsNewResource())

		// Datasource sets id to find, this might drop id to empty
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	if err := d.Set("project", project); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("topic_name", topicName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("partitions", len(topic.Partitions)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("replication", topic.Replication); err != nil {
		return diag.FromErr(err)
	}

	config, err := FlattenKafkaTopicConfig(topic)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("config", config); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("termination_protection", d.Get("termination_protection")); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("tag", flattenKafkaTopicTags(topic.Tags)); err != nil {
		return diag.Errorf("error setting Kafka Topic Tags for resource %s: %s", d.Id(), err)
	}

	if err := d.Set("topic_description", topic.TopicDescription); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("owner_user_group_id", topic.OwnerUserGroupId); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKafkaTopicReadResource(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceKafkaTopicRead(ctx, d, m, true)
}

func resourceKafkaTopicReadDatasource(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceKafkaTopicRead(ctx, d, m, false)
}

func flattenKafkaTopicTags(list []aiven.KafkaTopicTag) []map[string]interface{} {
	tags := make([]map[string]interface{}, 0, len(list))
	for _, tagS := range list {
		tags = append(tags, map[string]interface{}{
			"key":   tagS.Key,
			"value": tagS.Value,
		})
	}

	return tags
}

func resourceKafkaTopicUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	partitions := d.Get("partitions").(int)
	projectName, serviceName, topicName, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	config, err := getKafkaTopicConfig(d)
	if err != nil {
		return diag.Errorf("config to json error: %s", err)
	}

	topicDescription := d.Get("topic_description").(string)
	ownerUserGroupID := d.Get("owner_user_group_id").(string)

	updateRequest := aiven.UpdateKafkaTopicRequest{
		Partitions:  &partitions,
		Replication: schemautil.OptionalIntPointer(d, "replication"),
		Config:      config,
		Tags:        getTags(d),
	}

	if topicDescription != "" {
		updateRequest.TopicDescription = &topicDescription
	}

	if ownerUserGroupID != "" {
		updateRequest.OwnerUserGroupId = &ownerUserGroupID
	}

	client := m.(*aiven.Client)
	err = kafkatopicrepository.New(client.KafkaTopics).Update(
		ctx,
		projectName,
		serviceName,
		topicName,
		updateRequest,
	)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKafkaTopicDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	projectName, serviceName, topicName, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.Get("termination_protection").(bool) {
		return diag.Errorf("cannot delete kafka topic when termination_protection is enabled")
	}

	err = kafkatopicrepository.New(m.(*aiven.Client).KafkaTopics).Delete(ctx, projectName, serviceName, topicName)
	if err != nil {
		return diag.Errorf("error waiting for Aiven Kafka Topic to be DELETED: %s", err)
	}

	return nil
}

func FlattenKafkaTopicConfig(t *aiven.KafkaTopic) ([]map[string]interface{}, error) {
	source := make(map[string]struct {
		Value any `json:"value"`
	})

	data, err := json.Marshal(t.Config)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &source)
	if err != nil {
		return nil, err
	}

	config := make(map[string]any)
	configSchema := aivenKafkaTopicConfigSchema()
	for k, v := range source {
		if configSchema[k].Type == schema.TypeString {
			config[k] = schemautil.ToOptionalString(v.Value)
		} else {
			config[k] = v.Value
		}
	}

	return []map[string]interface{}{config}, nil
}
