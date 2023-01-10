package stateupgrader

import (
	v0 "github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/v0"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Cassandra() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    v0.ResourceCassandraResourceV0().CoreConfigSchema().ImpliedType(),
			Upgrade: v0.ResourceCassandraStateUpgradeV0,
			Version: 0,
		},
	}
}

func Flink() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    v0.ResourceFlinkResourceV0().CoreConfigSchema().ImpliedType(),
			Upgrade: v0.ResourceFlinkStateUpgradeV0,
			Version: 0,
		},
	}
}

func Grafana() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    v0.ResourceGrafanaResourceV0().CoreConfigSchema().ImpliedType(),
			Upgrade: v0.ResourceGrafanaStateUpgradeV0,
			Version: 0,
		},
	}
}

func Kafka() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    v0.ResourceKafkaResourceV0().CoreConfigSchema().ImpliedType(),
			Upgrade: v0.ResourceKafkaStateUpgradeV0,
			Version: 0,
		},
	}
}

func KafkaConnect() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    v0.ResourceKafkaConnectResourceV0().CoreConfigSchema().ImpliedType(),
			Upgrade: v0.ResourceKafkaConnectStateUpgradeV0,
			Version: 0,
		},
	}
}

func KafkaMirrormaker() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    v0.ResourceKafkaMirrormakerResourceV0().CoreConfigSchema().ImpliedType(),
			Upgrade: v0.ResourceKafkaMirrormakerStateUpgradeV0,
			Version: 0,
		},
	}
}
