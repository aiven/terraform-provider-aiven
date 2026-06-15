// Package topic implements the aiven_kafka_topic resource and data source.
//
// Resource design notes:
//
//  1. createView and updateView do not call readView afterwards because the read
//     path is a heavy operation — Kafka fetches values directly
//     from the broker, which can take up to a minute per topic. The state values are
//     therefore the values that were sent to the API. Computed config fields default to
//     null in the state and are only filled in on the next `terraform refresh`.
//
//  2. `partitions` and `replication` are modelled as required (rather than
//     computed+optional like in the API). The SDKv2-based resource did the same to
//     avoid a Read after create — keeping it that way preserves the existing user
//     experience and state shape.
//
//  3. When a Kafka service is powered off and has no backups, all topics are deleted.
//     See https://aiven.io/docs/platform/concepts/service-power-cycle#power-off-a-service .
//     The YAML sets `removeMissing: true` so that the framework drops the resource
//     from state on a 404 during refresh; the next apply will recreate the topic.
package topic

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/kafkatopic"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/kafkatopicrepository"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

// deleteView delegates to kafkatopicrepository.Delete so the seenTopics cache
// is invalidated alongside the API call. Skipping the cache eviction would
// cause modifyPlan's Exists() check to falsely report "topic conflict, already
// exists" for a name destroyed and re-created in the same provider process.
func deleteView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	return kafkatopicrepository.New(client).Delete(
		ctx,
		d.Get("project").(string),
		d.Get("service_name").(string),
		d.Get("topic_name").(string),
	)
}

var (
	errTopicAlreadyExists          = errors.New("topic conflict, already exists")
	errLocalRetentionBytesOverflow = errors.New("local_retention_bytes must not be more than retention_bytes value")
	errPartitionsCannotDecrease    = errors.New("number of partitions cannot be decreased")
)

// createView creates a topic.
//
// We do NOT call readView afterwards (see file header). The adapter writes the
// plan values into state; computed config fields that were not user-set stay null
// until the next refresh.
func createView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	req := new(kafkatopic.ServiceKafkaTopicCreateIn)
	if err := d.Expand(req, expandConfig, adapter.RenameFields(map[string]string{"tag": "tags"})); err != nil {
		return err
	}

	err := kafkatopicrepository.New(client).Create(
		ctx,
		d.Get("project").(string),
		d.Get("service_name").(string),
		req,
	)
	if err != nil {
		return err
	}

	return d.SetID(
		d.Get("project").(string),
		d.Get("service_name").(string),
		d.Get("topic_name").(string),
	)
}

func updateView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	req := new(kafkatopic.ServiceKafkaTopicUpdateIn)
	if err := d.Expand(req, expandConfig, adapter.RenameFields(map[string]string{"tag": "tags"})); err != nil {
		return err
	}

	return kafkatopicrepository.New(client).Update(
		ctx,
		d.Get("project").(string),
		d.Get("service_name").(string),
		d.Get("topic_name").(string),
		req,
	)
}

// readView is used both by the resource (with `removeMissing: true`) and by
// the data source. On 404 the resource path lets the framework drop the
// resource from state; the data source has no such handling and surfaces the
// error directly to the user.
func readView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	rsp, err := kafkatopicrepository.New(client).Read(
		ctx,
		d.Get("project").(string),
		d.Get("service_name").(string),
		d.Get("topic_name").(string),
	)
	if err != nil {
		return err
	}

	return d.Flatten(
		rsp,
		flattenConfig(rsp),
		flattenPartitions(rsp),
		adapter.RenameFields(map[string]string{"tags": "tag"}),
	)
}

// validateConfig enforces local_retention_bytes <= retention_bytes. Negative
// retention_bytes values are sentinels (-1 = infinite, -2 = "match overall
// retention time"); both bypass the comparison. The `alsoRequires` schema
// validator handles the "one set without the other" case; cross-state checks
// (partition decrease, name conflict) live in modifyPlan since ValidateConfig
// has no prior state.
func validateConfig(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	// Both fields are typed as string in the schema (legacy, see expandConfig).
	// Missing values come back as "" — nothing to validate then.
	localStr, _ := d.Get("config.0.local_retention_bytes").(string)
	retStr, _ := d.Get("config.0.retention_bytes").(string)
	if localStr == "" || retStr == "" {
		return nil
	}

	// Use 64-bit parsing because byte counts routinely exceed 2^31 (a few
	// gigabytes), which would overflow strconv.Atoi on 32-bit platforms.
	localBytes, err := strconv.ParseInt(localStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid local_retention_bytes %q: %w", localStr, err)
	}
	retBytes, err := strconv.ParseInt(retStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid retention_bytes %q: %w", retStr, err)
	}

	if retBytes >= 0 && retBytes < localBytes {
		return fmt.Errorf("%w (local=%d, retention=%d)", errLocalRetentionBytesOverflow, localBytes, retBytes)
	}

	return nil
}

// modifyPlan runs plan-time checks that need prior state:
//
//   - partition count cannot decrease on existing resources
//   - a topic with the same name must not already exist on the service (new
//     resources only — existing ones are reconciled by Read)
//
// Config-only checks (e.g. retention byte relationship) are in validateConfig.
func modifyPlan(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	if !d.IsNewResource() && d.HasChange("partitions") {
		// The adapter stores SchemaTypeInt values as plain `int` (see
		// marshalling.fromTFValueAny), so assert against `int`, not `int64`.
		oldP, _ := d.GetState("partitions").(int)
		newP, _ := d.Get("partitions").(int)
		if newP < oldP {
			return errPartitionsCannotDecrease
		}
	}

	if d.IsNewResource() {
		exists, err := kafkatopicrepository.New(client).Exists(
			ctx,
			d.Get("project").(string),
			d.Get("service_name").(string),
			d.Get("topic_name").(string),
		)
		if err != nil {
			return fmt.Errorf("failed to check whether topic exists: %w", err)
		}
		if exists {
			return fmt.Errorf("%w: %q", errTopicAlreadyExists, d.Get("topic_name").(string))
		}
	}
	return nil
}

// flattenConfig keeps only user-set overrides (SourceTypeTopicConfig) to avoid
// polluting state with service defaults, and omits the block when empty since
// the framework has no computed blocks.
// todo: remove in v5.0.0, use attribute.Computed instead.
func flattenConfig(rsp *kafkatopic.ServiceKafkaTopicGetOut) adapter.MapModifier {
	return func(d adapter.ResourceData, dto map[string]any) error {
		rspConfig := new(kafkaTopicConfig)
		if err := schemautil.Remarshal(rsp, rspConfig); err != nil {
			return err
		}

		userConfig := make(map[string]any, len(rspConfig.Config))
		for k, v := range rspConfig.Config {
			if d.IsResource() && v.Source != kafkatopic.SourceTypeTopicConfig {
				// Keeps user config values for resources only
				continue
			}
			userConfig[k] = v.Value
			// Legacy string-typed integers; todo: remove in v5.0.0.
			if stringIntegers[k] {
				userConfig[k] = fmt.Sprint(v.Value)
			}
		}

		// If there are no user-set config values, remove the block from the state.
		if len(userConfig) > 0 {
			dto["config"] = userConfig
		} else {
			delete(dto, "config")
		}
		return nil
	}
}

// flattenPartitions collapses the API's partition list into a count, as the schema expects.
func flattenPartitions(rsp *kafkatopic.ServiceKafkaTopicGetOut) adapter.MapModifier {
	return func(_ adapter.ResourceData, dto map[string]any) error {
		dto["partitions"] = len(rsp.Partitions)
		return nil
	}
}

// kafkaTopicConfig mirrors the {source, synonyms, value} shape of each property
// in the V2-list response config object; only `value` is needed for the schema.
type kafkaTopicConfig struct {
	Config map[string]struct {
		Source kafkatopic.SourceType `json:"source"`
		Value  any                   `json:"value"`
	} `json:"config"`
}

// stringIntegers lists every config field whose schema type is `string` even
// though the underlying value is numeric. Used to coerce values both ways
// across the adapter boundary.
// todo: remove in v5.0.0, legacy.
var stringIntegers = map[string]bool{
	"delete_retention_ms":                 true,
	"file_delete_delay_ms":                true,
	"flush_messages":                      true,
	"flush_ms":                            true,
	"index_interval_bytes":                true,
	"local_retention_bytes":               true,
	"local_retention_ms":                  true,
	"max_compaction_lag_ms":               true,
	"max_message_bytes":                   true,
	"message_timestamp_after_max_ms":      true,
	"message_timestamp_before_max_ms":     true,
	"message_timestamp_difference_max_ms": true,
	"min_compaction_lag_ms":               true,
	"min_insync_replicas":                 true,
	"retention_bytes":                     true,
	"retention_ms":                        true,
	"segment_bytes":                       true,
	"segment_index_bytes":                 true,
	"segment_jitter_ms":                   true,
	"segment_ms":                          true,
}

// expandConfig expands the config block by converting string values to integers.
// todo: remove in v5.0.0, legacy.
func expandConfig(_ adapter.ResourceData, dto map[string]any) error {
	config, ok := dto["config"]
	if !ok {
		return nil
	}

	configMap, ok := config.(map[string]any)
	if !ok {
		return fmt.Errorf("config is not a map[string]any")
	}

	for k, v := range configMap {
		if !stringIntegers[k] {
			continue
		}
		// fmt.Sprint copes with both the legacy string form and any future
		// numeric/json.Number schema migration without panicking.
		intValue, err := strconv.Atoi(fmt.Sprint(v))
		if err != nil {
			return fmt.Errorf("failed to parse %s as an integer: %w", k, err)
		}
		configMap[k] = intValue
	}

	return nil
}
