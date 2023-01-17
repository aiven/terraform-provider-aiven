package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

var aivenKafkaTopicSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,

	"topic_name": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: schemautil.Complex("The name of the topic.").ForceNew().Build(),
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
					Description:  schemautil.Complex("Topic tag key.").MaxLen(64).Build(),
				},
				"value": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringLenBetween(0, 256),
					Description:  schemautil.Complex("Topic tag value.").MaxLen(256).Build(),
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
					Type:             schema.TypeBool,
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
					Type:             schema.TypeFloat,
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
					Type:             schema.TypeBool,
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
					Type:             schema.TypeBool,
					Description:      "unclean.leader.election.enable value",
					Optional:         true,
					DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
				},
			},
		},
	},
}

func ResourceKafkaTopic() *schema.Resource {
	_ = newTopicCache()

	return &schema.Resource{
		Description:   "The Kafka Topic resource allows the creation and management of Aiven Kafka Topics.",
		CreateContext: resourceKafkaTopicCreate,
		ReadContext:   resourceKafkaTopicReadResource,
		UpdateContext: resourceKafkaTopicUpdate,
		DeleteContext: resourceKafkaTopicDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Read:   schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},
		Schema:         aivenKafkaTopicSchema,
		SchemaVersion:  1,
		StateUpgraders: stateupgrader.KafkaTopic(),
	}
}

func resourceKafkaTopicCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	topicName := d.Get("topic_name").(string)
	partitions := d.Get("partitions").(int)
	replication := d.Get("replication").(int)

	createRequest := aiven.CreateKafkaTopicRequest{
		Partitions:  &partitions,
		Replication: &replication,
		TopicName:   topicName,
		Config:      getKafkaTopicConfig(d),
		Tags:        getTags(d),
	}

	w := &kafkaTopicCreateWaiter{
		Client:        m.(*aiven.Client),
		Project:       project,
		ServiceName:   serviceName,
		CreateRequest: createRequest,
	}

	timeout := d.Timeout(schema.TimeoutCreate)
	_, err := w.Conf(timeout).WaitForStateContext(ctx)
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
	var tags []aiven.KafkaTopicTag
	for _, tagD := range d.Get("tag").(*schema.Set).List() {
		tagM := tagD.(map[string]interface{})
		tag := aiven.KafkaTopicTag{
			Key:   tagM["key"].(string),
			Value: tagM["value"].(string),
		}

		tags = append(tags, tag)
	}

	return tags
}

func getKafkaTopicConfig(d *schema.ResourceData) aiven.KafkaTopicConfig {
	if len(d.Get("config").([]interface{})) == 0 {
		return aiven.KafkaTopicConfig{}
	}

	if d.Get("config").([]interface{})[0] == nil {
		return aiven.KafkaTopicConfig{}
	}

	configRaw := d.Get("config").([]interface{})[0].(map[string]interface{})

	return aiven.KafkaTopicConfig{
		CleanupPolicy:                   configRaw["cleanup_policy"].(string),
		CompressionType:                 configRaw["compression_type"].(string),
		DeleteRetentionMs:               schemautil.ParseOptionalStringToInt64(configRaw["delete_retention_ms"]),
		FileDeleteDelayMs:               schemautil.ParseOptionalStringToInt64(configRaw["file_delete_delay_ms"]),
		FlushMessages:                   schemautil.ParseOptionalStringToInt64(configRaw["flush_messages"]),
		FlushMs:                         schemautil.ParseOptionalStringToInt64(configRaw["flush_ms"]),
		IndexIntervalBytes:              schemautil.ParseOptionalStringToInt64(configRaw["index_interval_bytes"]),
		MaxCompactionLagMs:              schemautil.ParseOptionalStringToInt64(configRaw["max_compaction_lag_ms"]),
		MaxMessageBytes:                 schemautil.ParseOptionalStringToInt64(configRaw["max_message_bytes"]),
		MessageDownconversionEnable:     schemautil.OptionalBoolPointer(d, "config.0.message_downconversion_enable"),
		MessageFormatVersion:            configRaw["message_format_version"].(string),
		MessageTimestampDifferenceMaxMs: schemautil.ParseOptionalStringToInt64(configRaw["message_timestamp_difference_max_ms"]),
		MessageTimestampType:            configRaw["message_timestamp_type"].(string),
		MinCleanableDirtyRatio:          schemautil.OptionalFloatPointer(d, "config.0.min_cleanable_dirty_ratio"),
		MinCompactionLagMs:              schemautil.ParseOptionalStringToInt64(configRaw["min_compaction_lag_ms"]),
		MinInsyncReplicas:               schemautil.ParseOptionalStringToInt64(configRaw["min_insync_replicas"]),
		Preallocate:                     schemautil.OptionalBoolPointer(d, "config.0.preallocate"),
		RetentionBytes:                  schemautil.ParseOptionalStringToInt64(configRaw["retention_bytes"]),
		RetentionMs:                     schemautil.ParseOptionalStringToInt64(configRaw["retention_ms"]),
		SegmentBytes:                    schemautil.ParseOptionalStringToInt64(configRaw["segment_bytes"]),
		SegmentIndexBytes:               schemautil.ParseOptionalStringToInt64(configRaw["segment_index_bytes"]),
		SegmentJitterMs:                 schemautil.ParseOptionalStringToInt64(configRaw["segment_jitter_ms"]),
		SegmentMs:                       schemautil.ParseOptionalStringToInt64(configRaw["segment_ms"]),
		UncleanLeaderElectionEnable:     schemautil.OptionalBoolPointer(d, "config.0.unclean_leader_election_enable"),
	}
}

func resourceKafkaTopicRead(ctx context.Context, d *schema.ResourceData, m interface{}, isResource bool) diag.Diagnostics {
	project, serviceName, topicName, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	topic, err := getTopic(ctx, d, m)

	// Topics are destroyed when kafka is off
	// https://docs.aiven.io/docs/platform/concepts/service-power-cycle
	// So it's better to recreate them, than make user to clear the state manually
	// Recreates missing topics:
	// 1. if server returns 404
	// 2. is resource (not datasource)
	// 3. only for resources with existing state, not imports
	if err != nil {
		if !isResource {
			return diag.FromErr(err)
		}

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
	if err := d.Set("config", flattenKafkaTopicConfig(topic)); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("termination_protection", d.Get("termination_protection")); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("tag", flattenKafkaTopicTags(topic.Tags)); err != nil {
		return diag.Errorf("error setting Kafka Topic Tags for resource %s: %s", d.Id(), err)
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
	var tags []map[string]interface{}
	for _, tagS := range list {
		tags = append(tags, map[string]interface{}{
			"key":   tagS.Key,
			"value": tagS.Value,
		})
	}

	return tags
}

func getTopic(ctx context.Context, d *schema.ResourceData, m interface{}) (*aiven.KafkaTopic, error) {
	project, serviceName, topicName, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return nil, err
	}

	client, ok := m.(*aiven.Client)
	if !ok {
		return nil, fmt.Errorf("invalid Aiven client")
	}

	w, err := newKafkaTopicAvailabilityWaiter(client, project, serviceName, topicName)
	if err != nil {
		return nil, err
	}

	timeout := d.Timeout(schema.TimeoutRead)
	topic, err := w.Conf(timeout).WaitForStateContext(ctx)
	if err != nil {
		return nil, err
	}

	kt, ok := topic.(aiven.KafkaTopic)
	if !ok {
		return nil, fmt.Errorf("can't cast value to aiven.KafkaTopic")
	}
	return &kt, nil
}

func resourceKafkaTopicUpdate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	partitions := d.Get("partitions").(int)
	projectName, serviceName, topicName, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.KafkaTopics.Update(
		projectName,
		serviceName,
		topicName,
		aiven.UpdateKafkaTopicRequest{
			Partitions:  &partitions,
			Replication: schemautil.OptionalIntPointer(d, "replication"),
			Config:      getKafkaTopicConfig(d),
			Tags:        getTags(d),
		},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKafkaTopicDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, topicName, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.Get("termination_protection").(bool) {
		return diag.Errorf("cannot delete kafka topic when termination_protection is enabled")
	}

	waiter := TopicDeleteWaiter{
		Client:      client,
		ProjectName: projectName,
		ServiceName: serviceName,
		TopicName:   topicName,
	}

	timeout := d.Timeout(schema.TimeoutDelete)
	_, err = waiter.Conf(timeout).WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("error waiting for Aiven Kafka Topic to be DELETED: %s", err)
	}

	return nil
}

func flattenKafkaTopicConfig(t *aiven.KafkaTopic) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"cleanup_policy":                      schemautil.ToOptionalString(t.Config.CleanupPolicy.Value),
			"compression_type":                    schemautil.ToOptionalString(t.Config.CompressionType.Value),
			"delete_retention_ms":                 schemautil.ToOptionalString(t.Config.DeleteRetentionMs.Value),
			"file_delete_delay_ms":                schemautil.ToOptionalString(t.Config.FileDeleteDelayMs.Value),
			"flush_messages":                      schemautil.ToOptionalString(t.Config.FlushMessages.Value),
			"flush_ms":                            schemautil.ToOptionalString(t.Config.FlushMs.Value),
			"index_interval_bytes":                schemautil.ToOptionalString(t.Config.IndexIntervalBytes.Value),
			"max_compaction_lag_ms":               schemautil.ToOptionalString(t.Config.MaxCompactionLagMs.Value),
			"max_message_bytes":                   schemautil.ToOptionalString(t.Config.MaxMessageBytes.Value),
			"message_downconversion_enable":       t.Config.MessageDownconversionEnable.Value,
			"message_format_version":              schemautil.ToOptionalString(t.Config.MessageFormatVersion.Value),
			"message_timestamp_difference_max_ms": schemautil.ToOptionalString(t.Config.MessageTimestampDifferenceMaxMs.Value),
			"message_timestamp_type":              schemautil.ToOptionalString(t.Config.MessageTimestampType.Value),
			"min_cleanable_dirty_ratio":           t.Config.MinCleanableDirtyRatio.Value,
			"min_compaction_lag_ms":               schemautil.ToOptionalString(t.Config.MinCompactionLagMs.Value),
			"min_insync_replicas":                 schemautil.ToOptionalString(t.Config.MinInsyncReplicas.Value),
			"preallocate":                         t.Config.Preallocate.Value,
			"retention_bytes":                     schemautil.ToOptionalString(t.Config.RetentionBytes.Value),
			"retention_ms":                        schemautil.ToOptionalString(t.Config.RetentionMs.Value),
			"segment_bytes":                       schemautil.ToOptionalString(t.Config.SegmentBytes.Value),
			"segment_index_bytes":                 schemautil.ToOptionalString(t.Config.SegmentIndexBytes.Value),
			"segment_jitter_ms":                   schemautil.ToOptionalString(t.Config.SegmentJitterMs.Value),
			"segment_ms":                          schemautil.ToOptionalString(t.Config.SegmentMs.Value),
			"unclean_leader_election_enable":      t.Config.UncleanLeaderElectionEnable.Value,
		},
	}
}

// TopicDeleteWaiter is used to wait for Kafka Topic to be deleted.
type TopicDeleteWaiter struct {
	Client      *aiven.Client
	ProjectName string
	ServiceName string
	TopicName   string
}

// RefreshFunc will call the Aiven client and refresh it's state.
func (w *TopicDeleteWaiter) RefreshFunc() resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		err := w.Client.KafkaTopics.Delete(w.ProjectName, w.ServiceName, w.TopicName)
		if err != nil {
			if !aiven.IsNotFound(err) {
				return nil, "REMOVING", nil
			}
		}

		return aiven.KafkaTopic{}, "DELETED", nil
	}
}

// Conf sets up the configuration to refresh.
func (w *TopicDeleteWaiter) Conf(timeout time.Duration) *resource.StateChangeConf {
	log.Printf("[DEBUG] Delete waiter timeout %.0f minutes", timeout.Minutes())

	return &resource.StateChangeConf{
		Pending:    []string{"REMOVING"},
		Target:     []string{"DELETED"},
		Refresh:    w.RefreshFunc(),
		Delay:      1 * time.Second,
		Timeout:    timeout,
		MinTimeout: 1 * time.Second,
	}
}
