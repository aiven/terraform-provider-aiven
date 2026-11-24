package kafkatopic

import (
	"context"
	"encoding/json"
	"maps"
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

// topicData returns the resource data with the managed config keys in the private state.
func topicData(t *testing.T, state, plan, config map[string]any, managed ...string) adapter.ResourceData {
	t.Helper()

	d, err := adapter.NewResourceData(
		resourceSchemaInternal(),
		ResourceOptions.IdentityFields,
		adapter.WithTestPlan(plan),
		adapter.WithTestState(state),
		adapter.WithTestConfig(config),
	)
	require.NoError(t, err)

	if len(managed) > 0 {
		b, err := json.Marshal(managed)
		require.NoError(t, err)
		require.NoError(t, d.SetPrivate(t.Context(), managedConfigKeysPrivate, b))
	}
	return d
}

// updateTopic builds the request updateView sends. The topic repository is a process
// singleton that keeps the first client, so a mock client can't capture it.
func updateTopic(t *testing.T, state, plan, config map[string]any, managed ...string) (*kafkatopic.ServiceKafkaTopicUpdateIn, error) {
	t.Helper()
	return updateRequest(t.Context(), topicData(t, state, plan, config, managed...))
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
	require.Equal(t, map[string]any{"retention_bytes": json.Number("1234")}, *got.Config)
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

// Large integers keep their precision in the config map.
func TestUpdateKeepsLargeConfigIntegers(t *testing.T) {
	t.Parallel()

	flush := map[string]any{"flush_messages": "9223372036854775807"}

	plan := topicValues()
	plan["config"] = []any{flush}

	config := topicValues()
	config["config"] = []any{flush}

	got, err := updateTopic(t, topicValues(), plan, config)
	require.NoError(t, err)

	b, err := json.Marshal(got.Config)
	require.NoError(t, err)
	require.JSONEq(t, `{"flush_messages": 9223372036854775807}`, string(b))
}

// Only the managed keys removed from the manifest are sent as null.
func TestUpdateResetsRemovedConfig(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		plan    map[string]any // The config block after modifyPlan, nil when removed.
		config  map[string]any // The config block in the manifest, nil when removed.
		managed []string
		want    string // The JSON config, empty for no request.
	}{
		{
			name:    "removed key",
			plan:    map[string]any{"segment_ms": "200"},
			config:  map[string]any{"segment_ms": "200"},
			managed: []string{"retention_ms", "segment_ms"},
			want:    `{"retention_ms": null, "segment_ms": 200}`,
		},
		{
			name:    "removed block",
			managed: []string{"retention_ms", "segment_ms"},
			want:    `{"retention_ms": null, "segment_ms": null}`,
		},
		{
			name:    "unmanaged keys are kept",
			managed: []string{"segment_ms"},
			want:    `{"segment_ms": null}`,
		},
		{
			name: "no managed keys",
		},
		{
			name:    "remote storage can't be reset",
			managed: []string{"remote_storage_enable"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			state := topicValues()
			state["config"] = []any{map[string]any{
				"retention_ms":          "100",
				"segment_ms":            "200",
				"remote_storage_enable": true,
			}}

			plan := topicValues()
			config := topicValues()
			if tc.plan != nil {
				plan["config"] = []any{tc.plan}
			}
			if tc.config != nil {
				config["config"] = []any{tc.config}
			}

			got, err := updateTopic(t, state, plan, config, tc.managed...)
			if tc.want == "" {
				require.ErrorIs(t, err, adapter.ErrNoChanges)
				return
			}

			require.NoError(t, err)
			b, err := json.Marshal(got.Config)
			require.NoError(t, err)
			require.JSONEq(t, tc.want, string(b))
		})
	}
}

// The plan shows the reset of a removed managed key and keeps the other values.
func TestModifyPlanRemovesConfig(t *testing.T) {
	t.Parallel()

	state := topicValues()
	state["config"] = []any{map[string]any{"retention_ms": "100", "segment_ms": "200"}}

	config := topicValues()
	config["config"] = []any{map[string]any{"segment_ms": "200"}}

	d := topicData(t, state, maps.Clone(state), config, "retention_ms", "segment_ms")
	require.NoError(t, planRemovedConfig(t.Context(), d))

	_, ok := d.GetOk("config.0.retention_ms")
	require.False(t, ok)
	require.Equal(t, "200", d.Get("config.0.segment_ms"))
}

// Create and update store the keys set in the manifest.
func TestSetManagedConfigKeys(t *testing.T) {
	t.Parallel()

	config := topicValues()
	config["config"] = []any{map[string]any{"segment_ms": "200", "retention_ms": "100"}}

	d := topicData(t, topicValues(), config, config)
	require.NoError(t, setManagedConfigKeys(t.Context(), d))

	b, err := d.GetPrivate(t.Context(), managedConfigKeysPrivate)
	require.NoError(t, err)
	require.JSONEq(t, `["retention_ms", "segment_ms"]`, string(b))

	d = topicData(t, topicValues(), topicValues(), topicValues(), "retention_ms")
	require.NoError(t, setManagedConfigKeys(t.Context(), d))

	b, err = d.GetPrivate(t.Context(), managedConfigKeysPrivate)
	require.NoError(t, err)
	require.Empty(t, b)
}
