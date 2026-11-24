package kafkatopic

import (
	"context"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/kafkatopic"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

// A topic config override can be reset outside Terraform. On the next refresh,
// Terraform should show that the override is gone instead of keeping the old
// value in state and hiding the drift from the user.
func TestKafkaTopicFlattenClearsStaleResourceConfig(t *testing.T) {
	t.Parallel()

	const (
		projectName = "project"
		serviceName = "kafka"
		topicName   = "topic"
	)

	d, err := adapter.NewResourceData(
		resourceSchemaInternal(),
		[]string{"project", "service_name", "topic_name"},
		adapter.WithTestState(map[string]any{
			"project":      projectName,
			"service_name": serviceName,
			"topic_name":   topicName,
			"partitions":   3,
			"replication":  2,
			"config": []any{
				map[string]any{"retention_ms": "604800000"},
			},
		}),
		adapter.WithTestConfig(map[string]any{
			"project":      projectName,
			"service_name": serviceName,
			"topic_name":   topicName,
			"partitions":   3,
			"replication":  2,
		}),
	)
	require.NoError(t, err)

	rsp := &kafkatopic.ServiceKafkaTopicGetOut{
		TopicName:      topicName,
		Partitions:     []kafkatopic.PartitionOut{{}, {}, {}},
		Replication:    2,
		Config:         kafkatopic.ConfigOut{},
		RetentionBytes: -1,
	}

	err = d.Flatten(
		rsp,
		flattenConfig(rsp),
		flattenPartitions(rsp),
		adapter.RenameFields(map[string]string{"tags": "tag"}),
	)
	require.NoError(t, err)

	_, ok := d.GetOk("config")
	require.False(t, ok)
}

// A user-set value equal to the Kafka default is reported by the API as DEFAULT_CONFIG.
// flattenConfig must still preserve it when it is present in prior state.
func TestKafkaTopicFlattenPreservesPriorConfigMatchingDefault(t *testing.T) {
	t.Parallel()

	const (
		projectName = "project"
		serviceName = "kafka"
		topicName   = "topic"
	)

	d, err := adapter.NewResourceData(
		resourceSchemaInternal(),
		[]string{"project", "service_name", "topic_name"},
		adapter.WithTestState(map[string]any{
			"project":      projectName,
			"service_name": serviceName,
			"topic_name":   topicName,
			"partitions":   3,
			"replication":  2,
			"config": []any{
				map[string]any{"retention_ms": "604800000"},
			},
		}),
	)
	require.NoError(t, err)

	rsp := &kafkatopic.ServiceKafkaTopicGetOut{
		TopicName:   topicName,
		Partitions:  []kafkatopic.PartitionOut{{}, {}, {}},
		Replication: 2,
		Config: kafkatopic.ConfigOut{
			RetentionMs: &kafkatopic.RetentionMsOut{
				Source: kafkatopic.SourceTypeDefaultConfig,
				Value:  604800000,
			},
		},
		RetentionBytes: -1,
	}

	err = d.Flatten(
		rsp,
		flattenConfig(rsp),
		flattenPartitions(rsp),
		adapter.RenameFields(map[string]string{"tags": "tag"}),
	)
	require.NoError(t, err)

	got, ok := d.GetOk("config.0.retention_ms")
	require.True(t, ok, "retention_ms should be preserved from prior state")
	require.Equal(t, "604800000", got)
}

// A topic with no config overrides in prior state decodes the config block to an
// empty list. priorConfigKeys must handle that without panicking on a 0-index.
func TestKafkaTopicFlattenNoConfigDoesNotPanic(t *testing.T) {
	t.Parallel()

	const (
		projectName = "project"
		serviceName = "kafka"
		topicName   = "topic"
	)

	d, err := adapter.NewResourceData(
		resourceSchemaInternal(),
		[]string{"project", "service_name", "topic_name"},
		adapter.WithTestState(map[string]any{
			"project":      projectName,
			"service_name": serviceName,
			"topic_name":   topicName,
			"partitions":   3,
			"replication":  2,
			"config":       []any{},
		}),
	)
	require.NoError(t, err)

	rsp := &kafkatopic.ServiceKafkaTopicGetOut{
		TopicName:      topicName,
		Partitions:     []kafkatopic.PartitionOut{{}, {}, {}},
		Replication:    2,
		Config:         kafkatopic.ConfigOut{},
		RetentionBytes: -1,
	}

	require.NotPanics(t, func() {
		err = d.Flatten(
			rsp,
			flattenConfig(rsp),
			flattenPartitions(rsp),
			adapter.RenameFields(map[string]string{"tags": "tag"}),
		)
	})
	require.NoError(t, err)
}

// Removing the config block keeps the topic settings: the mock fails on any call.
func TestUpdateWithoutConfig(t *testing.T) {
	t.Parallel()

	values := func() map[string]any {
		return map[string]any{
			"id":           "my-project/my-kafka/my-topic",
			"project":      "my-project",
			"service_name": "my-kafka",
			"topic_name":   "my-topic",
			"partitions":   3,
			"replication":  2,
		}
	}
	state := values()
	state["config"] = []any{map[string]any{"retention_ms": "100"}}
	d, err := adapter.NewResourceData(
		resourceSchemaInternal(),
		ResourceOptions.IdentityFields,
		adapter.WithTestPlan(values()),
		adapter.WithTestState(state),
	)
	require.NoError(t, err)
	err = ResourceOptions.Update(context.Background(), avngen.NewMockClient(t), d)
	require.ErrorIs(t, err, adapter.ErrNoChanges)
}

func topicValues() map[string]any {
	return map[string]any{
		"id":           "my-project/my-kafka/my-topic",
		"project":      "my-project",
		"service_name": "my-kafka",
		"topic_name":   "my-topic",
		"partitions":   3,
		"replication":  2,
	}
}

// updateTopic builds the request updateView sends. The topic repository is a process
// singleton that keeps the first client, so a mock client can't capture it.
func updateTopic(t *testing.T, state, plan, config map[string]any) (*kafkatopic.ServiceKafkaTopicUpdateIn, error) {
	t.Helper()

	d, err := adapter.NewResourceData(
		resourceSchemaInternal(),
		ResourceOptions.IdentityFields,
		adapter.WithTestPlan(plan),
		adapter.WithTestState(state),
		adapter.WithTestConfig(config),
	)
	require.NoError(t, err)

	req := new(kafkatopic.ServiceKafkaTopicUpdateIn)
	err = d.Expand(req, expandConfig, adapter.RenameFields(map[string]string{"tag": "tags"}))
	return req, err
}

// State values missing from the manifest aren't sent: pre-v4.55.2 providers
// could leave them unparsable (https://github.com/aiven/terraform-provider-aiven/issues/3082).
func TestUpdateSendsConfigFromManifestOnly(t *testing.T) {
	t.Parallel()

	staleConfig := map[string]any{
		"retention_bytes":    "1234",
		"flush_messages":     "9223372036854776000",
		"local_retention_ms": "",
	}

	state := topicValues()
	state["config"] = []any{staleConfig}

	plan := topicValues()
	plan["partitions"] = 4
	plan["config"] = []any{staleConfig}

	config := topicValues()
	config["config"] = []any{map[string]any{"retention_bytes": "1234"}}

	got, err := updateTopic(t, state, plan, config)
	require.NoError(t, err)

	require.NotNil(t, got.Config.RetentionBytes)
	require.Equal(t, 1234, *got.Config.RetentionBytes)
	require.Nil(t, got.Config.FlushMessages)
	require.Nil(t, got.Config.LocalRetentionMs)
}

// An invalid value in the manifest still fails.
func TestUpdateRejectsUserSetOverflowingConfigValue(t *testing.T) {
	t.Parallel()

	overflowed := map[string]any{"flush_messages": "9223372036854776000"}

	plan := topicValues()
	plan["partitions"] = 4
	plan["config"] = []any{overflowed}

	config := topicValues()
	config["config"] = []any{overflowed}

	_, err := updateTopic(t, topicValues(), plan, config)
	require.ErrorContains(t, err, "failed to parse flush_messages as an integer")
}
