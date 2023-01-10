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

func InfluxDB() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    v0.ResourceInfluxDBResourceV0().CoreConfigSchema().ImpliedType(),
			Upgrade: v0.ResourceInfluxDBStateUpgradeV0,
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

func M3Aggregator() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    v0.ResourceM3AggregatorResourceV0().CoreConfigSchema().ImpliedType(),
			Upgrade: v0.ResourceM3AggregatorStateUpgradeV0,
			Version: 0,
		},
	}
}

func M3DB() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    v0.ResourceM3DBResourceV0().CoreConfigSchema().ImpliedType(),
			Upgrade: v0.ResourceM3DBStateUpgradeV0,
			Version: 0,
		},
	}
}

func MySQL() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    v0.ResourceMySQLResourceV0().CoreConfigSchema().ImpliedType(),
			Upgrade: v0.ResourceMySQLStateUpgradeV0,
			Version: 0,
		},
	}
}

func Opensearch() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    v0.ResourceOpensearchResourceV0().CoreConfigSchema().ImpliedType(),
			Upgrade: v0.ResourceOpensearchStateUpgradeV0,
			Version: 0,
		},
	}
}

func PG() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    v0.ResourcePGResourceV0().CoreConfigSchema().ImpliedType(),
			Upgrade: v0.ResourcePGStateUpgradeV0,
			Version: 0,
		},
	}
}

func Redis() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    v0.ResourceRedisResourceV0().CoreConfigSchema().ImpliedType(),
			Upgrade: v0.ResourceRedisStateUpgradeV0,
			Version: 0,
		},
	}
}
