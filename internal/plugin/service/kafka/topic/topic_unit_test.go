package topic

import (
	"testing"

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
