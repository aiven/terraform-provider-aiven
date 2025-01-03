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

var aivenKafkaTopicConfigSchema = map[string]*schema.Schema{
	"cleanup_policy": {
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringInSlice(kafkatopic.CleanupPolicyTypeChoices(), false),
		Description:  userconfig.Desc("cleanup.policy value").PossibleValuesString(kafkatopic.CleanupPolicyTypeChoices()...).Build(),
	},
	"compression_type": {
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringInSlice(kafkatopic.CompressionTypeChoices(), false),
		Description:  userconfig.Desc("compression.type value").PossibleValuesString(kafkatopic.CompressionTypeChoices()...).Build(),
	},
	"delete_retention_ms": {
		Type:        schema.TypeString,
		Description: "delete.retention.ms value",
		Optional:    true,
	},
	"file_delete_delay_ms": {
		Type:        schema.TypeString,
		Description: "file.delete.delay.ms value",
		Optional:    true,
	},
	"flush_messages": {
		Type:        schema.TypeString,
		Description: "flush.messages value",
		Optional:    true,
	},
	"flush_ms": {
		Type:        schema.TypeString,
		Description: "flush.ms value",
		Optional:    true,
	},
	"index_interval_bytes": {
		Type:        schema.TypeString,
		Description: "index.interval.bytes value",
		Optional:    true,
	},
	"max_compaction_lag_ms": {
		Type:        schema.TypeString,
		Description: "max.compaction.lag.ms value",
		Optional:    true,
	},
	"max_message_bytes": {
		Type:        schema.TypeString,
		Description: "max.message.bytes value",
		Optional:    true,
	},
	"message_downconversion_enable": {
		Type:        schema.TypeBool,
		Description: "message.downconversion.enable value",
		Optional:    true,
	},
	"message_format_version": {
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringInSlice(kafkatopic.MessageFormatVersionTypeChoices(), false),
		Description:  userconfig.Desc("message.format.version value").PossibleValuesString(kafkatopic.MessageFormatVersionTypeChoices()...).Build(),
	},
	"message_timestamp_difference_max_ms": {
		Type:        schema.TypeString,
		Description: "message.timestamp.difference.max.ms value",
		Optional:    true,
	},
	"message_timestamp_type": {
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringInSlice(kafkatopic.MessageTimestampTypeChoices(), false),
		Description:  userconfig.Desc("message.timestamp.type value").PossibleValuesString(kafkatopic.MessageTimestampTypeChoices()...).Build(),
	},
	"min_cleanable_dirty_ratio": {
		Type:        schema.TypeFloat,
		Description: "min.cleanable.dirty.ratio value",
		Optional:    true,
	},
	"min_compaction_lag_ms": {
		Type:        schema.TypeString,
		Description: "min.compaction.lag.ms value",
		Optional:    true,
	},
	"min_insync_replicas": {
		Type:        schema.TypeString,
		Description: "min.insync.replicas value",
		Optional:    true,
	},
	"preallocate": {
		Type:        schema.TypeBool,
		Description: "preallocate value",
		Optional:    true,
	},
	"retention_bytes": {
		Type:        schema.TypeString,
		Description: "retention.bytes value",
		Optional:    true,
	},
	"retention_ms": {
		Type:        schema.TypeString,
		Description: "retention.ms value",
		Optional:    true,
	},
	"segment_bytes": {
		Type:        schema.TypeString,
		Description: "segment.bytes value",
		Optional:    true,
	},
	"segment_index_bytes": {
		Type:        schema.TypeString,
		Description: "segment.index.bytes value",
		Optional:    true,
	},
	"segment_jitter_ms": {
		Type:        schema.TypeString,
		Description: "segment.jitter.ms value",
		Optional:    true,
	},
	"segment_ms": {
		Type:        schema.TypeString,
		Description: "segment.ms value",
		Optional:    true,
	},
	"unclean_leader_election_enable": {
		Type:        schema.TypeBool,
		Description: "unclean.leader.election.enable value; This field is deprecated and no longer functional.",
		Optional:    true,
		Deprecated:  "This field is deprecated and no longer functional.",
	},
	"remote_storage_enable": {
		Type:        schema.TypeBool,
		Description: "remote.storage.enable value",
		Optional:    true,
	},
	"local_retention_bytes": {
		Type:        schema.TypeString,
		Description: "local.retention.bytes value",
		Optional:    true,
	},
	"local_retention_ms": {
		Type:        schema.TypeString,
		Description: "local.retention.ms value",
		Optional:    true,
	},
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
			Schema: aivenKafkaTopicConfigSchema,
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
	configs := d.GetRawConfig().AsValueMap()[configField]
	if configs.IsNull() || len(configs.AsValueSlice()) == 0 {
		return aiven.KafkaTopicConfig{}, nil
	}

	config := make(map[string]any)
	for k, v := range configs.AsValueSlice()[0].AsValueMap() {
		if v.IsNull() {
			continue
		}

		// Converts values to types that are expected by the API
		kind := aivenKafkaTopicConfigSchema[k].Type
		value := d.Get(fmt.Sprintf("%s.0.%s", configField, k))

		// These are known types
		switch kind {
		case schema.TypeBool, schema.TypeInt, schema.TypeFloat:
			config[k] = value
			continue
		}

		// These are known string types.
		// Add here new string fields
		switch k {
		case "cleanup_policy", "compression_type", "message_format_version", "message_timestamp_type":
			config[k] = value
			continue
		}

		// Legacy integer fields
		var err error
		config[k], err = strconv.ParseInt(value.(string), 10, 64)
		if err != nil {
			return aiven.KafkaTopicConfig{}, err
		}
	}

	// Converts to json and loads values to the struct
	b, err := json.Marshal(config)
	if err != nil {
		return aiven.KafkaTopicConfig{}, err
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
	for k, v := range source {
		if aivenKafkaTopicConfigSchema[k].Type == schema.TypeString {
			config[k] = schemautil.ToOptionalString(v.Value)
		} else {
			config[k] = v.Value
		}
	}

	return []map[string]interface{}{config}, nil
}
