// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"strings"
	"time"
)

var aivenKafkaTopicSchema = map[string]*schema.Schema{
	"project": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Project to link the kafka topic to",
		ForceNew:    true,
	},
	"service_name": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Service to link the kafka topic to",
		ForceNew:    true,
	},
	"topic_name": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Topic name",
		ForceNew:    true,
	},
	"partitions": {
		Type:        schema.TypeInt,
		Required:    true,
		Description: "Number of partitions to create in the topic",
	},
	"replication": {
		Type:        schema.TypeInt,
		Required:    true,
		Description: "Replication factor for the topic",
	},
	"retention_bytes": {
		Type:             schema.TypeInt,
		Optional:         true,
		Description:      "Retention bytes",
		Deprecated:       "use config.retention_bytes instead",
		DiffSuppressFunc: emptyObjectDiffSuppressFunc,
	},
	"retention_hours": {
		Type:             schema.TypeInt,
		Optional:         true,
		Description:      "Retention period (hours)",
		ValidateFunc:     validation.IntAtLeast(-1),
		Deprecated:       "use config.retention_ms instead",
		DiffSuppressFunc: emptyObjectDiffSuppressFunc,
	},
	"minimum_in_sync_replicas": {
		Type:             schema.TypeInt,
		Optional:         true,
		Description:      "Minimum required nodes in-sync replicas (ISR) to produce to a partition",
		Deprecated:       "use config.min_insync_replicas instead",
		DiffSuppressFunc: emptyObjectDiffSuppressFunc,
	},
	"cleanup_policy": {
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "Topic cleanup policy. Allowed values: delete, compact",
		Deprecated:       "use config.cleanup_policy instead",
		DiffSuppressFunc: emptyObjectDiffSuppressFunc,
	},
	"termination_protection": {
		Type:     schema.TypeBool,
		Optional: true,
		Default:  false,
		Description: `It is a Terraform client-side deletion protection, which prevents a Kafka 
			topic from being deleted. It is recommended to enable this for any production Kafka 
			topic containing critical data.`,
	},
	"tag": {
		Type:        schema.TypeSet,
		Description: "Kafka Topic tag",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"key": {
					Type:         schema.TypeString,
					Description:  "Topic tag key",
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 64),
				},
				"value": {
					Type:         schema.TypeString,
					Description:  "Topic tag value",
					Optional:     true,
					ValidateFunc: validation.StringLenBetween(0, 256),
				},
			},
		},
	},
	"config": {
		Type:             schema.TypeList,
		Description:      "Kafka topic configuration",
		Optional:         true,
		MaxItems:         1,
		DiffSuppressFunc: emptyObjectDiffSuppressFunc,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"cleanup_policy": {
					Type:             schema.TypeString,
					Description:      "cleanup.policy value",
					Optional:         true,
					DiffSuppressFunc: emptyObjectDiffSuppressFunc,
				},
				"compression_type": {
					Type:             schema.TypeString,
					Description:      "compression.type value",
					Optional:         true,
					DiffSuppressFunc: emptyObjectDiffSuppressFunc,
				},
				"delete_retention_ms": {
					Type:             schema.TypeString,
					Description:      "delete.retention.ms value",
					Optional:         true,
					DiffSuppressFunc: emptyObjectDiffSuppressFunc,
				},
				"file_delete_delay_ms": {
					Type:             schema.TypeString,
					Description:      "file.delete.delay.ms value",
					Optional:         true,
					DiffSuppressFunc: emptyObjectDiffSuppressFunc,
				},
				"flush_messages": {
					Type:             schema.TypeString,
					Description:      "flush.messages value",
					Optional:         true,
					DiffSuppressFunc: emptyObjectDiffSuppressFunc,
				},
				"flush_ms": {
					Type:             schema.TypeString,
					Description:      "flush.ms value",
					Optional:         true,
					DiffSuppressFunc: emptyObjectDiffSuppressFunc,
				},
				"index_interval_bytes": {
					Type:             schema.TypeString,
					Description:      "index.interval.bytes value",
					Optional:         true,
					DiffSuppressFunc: emptyObjectDiffSuppressFunc,
				},
				"max_compaction_lag_ms": {
					Type:             schema.TypeString,
					Description:      "max.compaction.lag.ms value",
					Optional:         true,
					DiffSuppressFunc: emptyObjectDiffSuppressFunc,
				},
				"max_message_bytes": {
					Type:             schema.TypeString,
					Description:      "max.message.bytes value",
					Optional:         true,
					DiffSuppressFunc: emptyObjectDiffSuppressFunc,
				},
				"message_downconversion_enable": {
					Type:             schema.TypeString,
					Description:      "message.downconversion.enable value",
					Optional:         true,
					DiffSuppressFunc: emptyObjectDiffSuppressFunc,
				},
				"message_format_version": {
					Type:             schema.TypeString,
					Description:      "message.format.version value",
					Optional:         true,
					DiffSuppressFunc: emptyObjectDiffSuppressFunc,
				},
				"message_timestamp_difference_max_ms": {
					Type:             schema.TypeString,
					Description:      "message.timestamp.difference.max.ms value",
					Optional:         true,
					DiffSuppressFunc: emptyObjectDiffSuppressFunc,
				},
				"message_timestamp_type": {
					Type:             schema.TypeString,
					Description:      "message.timestamp.type value",
					Optional:         true,
					DiffSuppressFunc: emptyObjectDiffSuppressFunc,
				},
				"min_cleanable_dirty_ratio": {
					Type:             schema.TypeString,
					Description:      "min.cleanable.dirty.ratio value",
					Optional:         true,
					DiffSuppressFunc: emptyObjectDiffSuppressFunc,
				},
				"min_compaction_lag_ms": {
					Type:             schema.TypeString,
					Description:      "min.compaction.lag.ms value",
					Optional:         true,
					DiffSuppressFunc: emptyObjectDiffSuppressFunc,
				},
				"min_insync_replicas": {
					Type:             schema.TypeString,
					Description:      "min.insync.replicas value",
					Optional:         true,
					DiffSuppressFunc: emptyObjectDiffSuppressFunc,
				},
				"preallocate": {
					Type:             schema.TypeString,
					Description:      "preallocate value",
					Optional:         true,
					DiffSuppressFunc: emptyObjectDiffSuppressFunc,
				},
				"retention_bytes": {
					Type:             schema.TypeString,
					Description:      "retention.bytes value",
					Optional:         true,
					DiffSuppressFunc: emptyObjectDiffSuppressFunc,
				},
				"retention_ms": {
					Type:             schema.TypeString,
					Description:      "retention.ms value",
					Optional:         true,
					DiffSuppressFunc: emptyObjectDiffSuppressFunc,
				},
				"segment_bytes": {
					Type:             schema.TypeString,
					Description:      "segment.bytes value",
					Optional:         true,
					DiffSuppressFunc: emptyObjectDiffSuppressFunc,
				},
				"segment_index_bytes": {
					Type:             schema.TypeString,
					Description:      "segment.index.bytes value",
					Optional:         true,
					DiffSuppressFunc: emptyObjectDiffSuppressFunc,
				},
				"segment_jitter_ms": {
					Type:             schema.TypeString,
					Description:      "segment.jitter.ms value",
					Optional:         true,
					DiffSuppressFunc: emptyObjectDiffSuppressFunc,
				},
				"segment_ms": {
					Type:             schema.TypeString,
					Description:      "segment.ms value",
					Optional:         true,
					DiffSuppressFunc: emptyObjectDiffSuppressFunc,
				},
				"unclean_leader_election_enable": {
					Type:             schema.TypeString,
					Description:      "unclean.leader.election.enable value",
					Optional:         true,
					DiffSuppressFunc: emptyObjectDiffSuppressFunc,
				},
			},
		},
	},
}

func resourceKafkaTopic() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKafkaTopicCreate,
		ReadContext:   resourceKafkaTopicRead,
		UpdateContext: resourceKafkaTopicUpdate,
		DeleteContext: resourceKafkaTopicDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceKafkaTopicState,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Read:   schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},
		Schema: aivenKafkaTopicSchema,
	}
}

func resourceKafkaTopicCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	topicName := d.Get("topic_name").(string)
	partitions := d.Get("partitions").(int)
	replication := d.Get("replication").(int)

	createRequest := aiven.CreateKafkaTopicRequest{
		CleanupPolicy:         optionalStringPointer(d, "cleanup_policy"),
		MinimumInSyncReplicas: optionalIntPointer(d, "minimum_in_sync_replicas"),
		Partitions:            &partitions,
		Replication:           &replication,
		RetentionBytes:        optionalIntPointer(d, "retention_bytes"),
		RetentionHours:        optionalIntPointer(d, "retention_hours"),
		TopicName:             topicName,
		Config:                getKafkaTopicConfig(d),
		Tags:                  getTags(d),
	}

	w := &KafkaTopicCreateWaiter{
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

	d.SetId(buildResourceID(project, serviceName, topicName))

	_, err = getTopic(ctx, d, m, true)
	if err != nil {
		return diag.FromErr(err)
	}

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
		DeleteRetentionMs:               parseOptionalStringToInt64(configRaw["delete_retention_ms"]),
		FileDeleteDelayMs:               parseOptionalStringToInt64(configRaw["file_delete_delay_ms"]),
		FlushMessages:                   parseOptionalStringToInt64(configRaw["flush_messages"]),
		FlushMs:                         parseOptionalStringToInt64(configRaw["flush_ms"]),
		IndexIntervalBytes:              parseOptionalStringToInt64(configRaw["index_interval_bytes"]),
		MaxCompactionLagMs:              parseOptionalStringToInt64(configRaw["max_compaction_lag_ms"]),
		MaxMessageBytes:                 parseOptionalStringToInt64(configRaw["max_message_bytes"]),
		MessageDownconversionEnable:     parseOptionalStringToBool(configRaw["message_downconversion_enable"]),
		MessageFormatVersion:            configRaw["message_format_version"].(string),
		MessageTimestampDifferenceMaxMs: parseOptionalStringToInt64(configRaw["message_timestamp_difference_max_ms"]),
		MessageTimestampType:            configRaw["message_timestamp_type"].(string),
		MinCleanableDirtyRatio:          parseOptionalStringToFloat64(configRaw["min_cleanable_dirty_ratio"]),
		MinCompactionLagMs:              parseOptionalStringToInt64(configRaw["min_compaction_lag_ms"]),
		MinInsyncReplicas:               parseOptionalStringToInt64(configRaw["min_insync_replicas"]),
		Preallocate:                     parseOptionalStringToBool(configRaw["preallocate"]),
		RetentionBytes:                  parseOptionalStringToInt64(configRaw["retention_bytes"]),
		RetentionMs:                     parseOptionalStringToInt64(configRaw["retention_ms"]),
		SegmentBytes:                    parseOptionalStringToInt64(configRaw["segment_bytes"]),
		SegmentIndexBytes:               parseOptionalStringToInt64(configRaw["segment_index_bytes"]),
		SegmentJitterMs:                 parseOptionalStringToInt64(configRaw["segment_jitter_ms"]),
		SegmentMs:                       parseOptionalStringToInt64(configRaw["segment_ms"]),
		UncleanLeaderElectionEnable:     parseOptionalStringToBool(configRaw["unclean_leader_election_enable"]),
	}
}

func resourceKafkaTopicRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	project, serviceName, topicName := splitResourceID3(d.Id())
	topic, err := getTopic(ctx, d, m, false)
	if err != nil {
		return diag.FromErr(err)
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
	if _, ok := d.GetOk("cleanup_policy"); ok {
		if err := d.Set("cleanup_policy", topic.Config.CleanupPolicy.Value); err != nil {
			return diag.FromErr(err)
		}
	}
	if _, ok := d.GetOk("minimum_in_sync_replicas"); ok {
		if err := d.Set("minimum_in_sync_replicas", topic.Config.MinInsyncReplicas.Value); err != nil {
			return diag.FromErr(err)
		}
	}
	if _, ok := d.GetOk("retention_bytes"); ok {
		if err := d.Set("retention_bytes", topic.Config.RetentionBytes.Value); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("config", flattenKafkaTopicConfig(topic)); err != nil {
		return diag.FromErr(err)
	}

	if _, ok := d.GetOk("retention_hours"); ok {
		// it could be -1, which means infinite retention
		if topic.Config.RetentionMs.Value != -1 {
			if err := d.Set("retention_hours", topic.Config.RetentionMs.Value/(1000*60*60)); err != nil {
				return diag.FromErr(err)
			}
		} else {
			if err := d.Set("retention_hours", topic.Config.RetentionMs.Value); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if err := d.Set("termination_protection", d.Get("termination_protection")); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("tag", flattenKafkaTopicTags(topic.Tags)); err != nil {
		return diag.Errorf("error setting Kafka Topic Tags for resource %s: %s", d.Id(), err)
	}

	return nil
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

func getTopic(ctx context.Context, d *schema.ResourceData, m interface{}, ignore404 bool) (aiven.KafkaTopic, error) {
	project, serviceName, topicName := splitResourceID3(d.Id())

	w := &KafkaTopicAvailabilityWaiter{
		Client:      m.(*aiven.Client),
		Project:     project,
		ServiceName: serviceName,
		TopicName:   topicName,
		Ignore404:   ignore404,
	}

	timeout := d.Timeout(schema.TimeoutRead)
	topic, err := w.Conf(timeout).WaitForStateContext(ctx)
	if err != nil {
		return aiven.KafkaTopic{}, fmt.Errorf("error waiting for Aiven Kafka topic to be ACTIVE: %s", err)
	}

	return topic.(aiven.KafkaTopic), nil
}

func resourceKafkaTopicUpdate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	partitions := d.Get("partitions").(int)
	projectName, serviceName, topicName := splitResourceID3(d.Id())
	err := client.KafkaTopics.Update(
		projectName,
		serviceName,
		topicName,
		aiven.UpdateKafkaTopicRequest{
			MinimumInSyncReplicas: optionalIntPointer(d, "minimum_in_sync_replicas"),
			Partitions:            &partitions,
			Replication:           optionalIntPointer(d, "replication"),
			RetentionBytes:        optionalIntPointer(d, "retention_bytes"),
			RetentionHours:        optionalIntPointer(d, "retention_hours"),
			Config:                getKafkaTopicConfig(d),
			Tags:                  getTags(d),
		},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKafkaTopicDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, topicName := splitResourceID3(d.Id())

	if d.Get("termination_protection").(bool) {
		return diag.Errorf("cannot delete kafka topic when termination_protection is enabled")
	}

	waiter := KafkaTopicDeleteWaiter{
		Client:      client,
		ProjectName: projectName,
		ServiceName: serviceName,
		TopicName:   topicName,
	}

	timeout := d.Timeout(schema.TimeoutDelete)
	_, err := waiter.Conf(timeout).WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("error waiting for Aiven Kafka Topic to be DELETED: %s", err)
	}

	return nil
}

func resourceKafkaTopicState(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	if len(strings.Split(d.Id(), "/")) != 3 {
		return nil, fmt.Errorf("invalid identifier %v, expected <project_name>/<service_name>/<topic_name>", d.Id())
	}

	di := resourceKafkaTopicRead(ctx, d, m)
	if di.HasError() {
		return nil, fmt.Errorf("cannot get kafka topic: %v", di)
	}

	return []*schema.ResourceData{d}, nil
}

func flattenKafkaTopicConfig(t aiven.KafkaTopic) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"cleanup_policy":                      toOptionalString(t.Config.CleanupPolicy.Value),
			"compression_type":                    toOptionalString(t.Config.CompressionType.Value),
			"delete_retention_ms":                 toOptionalString(t.Config.DeleteRetentionMs.Value),
			"file_delete_delay_ms":                toOptionalString(t.Config.FileDeleteDelayMs.Value),
			"flush_messages":                      toOptionalString(t.Config.FlushMessages.Value),
			"flush_ms":                            toOptionalString(t.Config.FlushMs.Value),
			"index_interval_bytes":                toOptionalString(t.Config.IndexIntervalBytes.Value),
			"max_compaction_lag_ms":               toOptionalString(t.Config.MaxCompactionLagMs.Value),
			"max_message_bytes":                   toOptionalString(t.Config.MaxMessageBytes.Value),
			"message_downconversion_enable":       toOptionalString(t.Config.MessageDownconversionEnable.Value),
			"message_format_version":              toOptionalString(t.Config.MessageFormatVersion.Value),
			"message_timestamp_difference_max_ms": toOptionalString(t.Config.MessageTimestampDifferenceMaxMs.Value),
			"message_timestamp_type":              toOptionalString(t.Config.MessageTimestampType.Value),
			"min_cleanable_dirty_ratio":           toOptionalString(t.Config.MinCleanableDirtyRatio.Value),
			"min_compaction_lag_ms":               toOptionalString(t.Config.MinCompactionLagMs.Value),
			"min_insync_replicas":                 toOptionalString(t.Config.MinInsyncReplicas.Value),
			"preallocate":                         toOptionalString(t.Config.Preallocate.Value),
			"retention_bytes":                     toOptionalString(t.Config.RetentionBytes.Value),
			"retention_ms":                        toOptionalString(t.Config.RetentionMs.Value),
			"segment_bytes":                       toOptionalString(t.Config.SegmentBytes.Value),
			"segment_index_bytes":                 toOptionalString(t.Config.SegmentIndexBytes.Value),
			"segment_jitter_ms":                   toOptionalString(t.Config.SegmentJitterMs.Value),
			"segment_ms":                          toOptionalString(t.Config.SegmentMs.Value),
			"unclean_leader_election_enable":      toOptionalString(t.Config.UncleanLeaderElectionEnable.Value),
		},
	}
}

// KafkaTopicDeleteWaiter is used to wait for Kafka Topic to be deleted.
type KafkaTopicDeleteWaiter struct {
	Client      *aiven.Client
	ProjectName string
	ServiceName string
	TopicName   string
}

// RefreshFunc will call the Aiven client and refresh it's state.
func (w *KafkaTopicDeleteWaiter) RefreshFunc() resource.StateRefreshFunc {
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
func (w *KafkaTopicDeleteWaiter) Conf(timeout time.Duration) *resource.StateChangeConf {
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
