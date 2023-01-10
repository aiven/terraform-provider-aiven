package stateupgrader

import (
	v0 "github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/v0"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Cassandra() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    v0.ResourceCassandra().CoreConfigSchema().ImpliedType(),
			Upgrade: v0.ResourceCassandraStateUpgrade,
			Version: 0,
		},
	}
}

func Flink() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    v0.ResourceFlink().CoreConfigSchema().ImpliedType(),
			Upgrade: v0.ResourceFlinkStateUpgrade,
			Version: 0,
		},
	}
}

func Grafana() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    v0.ResourceGrafana().CoreConfigSchema().ImpliedType(),
			Upgrade: v0.ResourceGrafanaStateUpgrade,
			Version: 0,
		},
	}
}

func InfluxDB() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    v0.ResourceInfluxDB().CoreConfigSchema().ImpliedType(),
			Upgrade: v0.ResourceInfluxDBStateUpgrade,
			Version: 0,
		},
	}
}

func Kafka() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    v0.ResourceKafka().CoreConfigSchema().ImpliedType(),
			Upgrade: v0.ResourceKafkaStateUpgrade,
			Version: 0,
		},
	}
}

func KafkaConnect() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    v0.ResourceKafkaConnect().CoreConfigSchema().ImpliedType(),
			Upgrade: v0.ResourceKafkaConnectStateUpgrade,
			Version: 0,
		},
	}
}

func KafkaMirrormaker() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    v0.ResourceKafkaMirrormaker().CoreConfigSchema().ImpliedType(),
			Upgrade: v0.ResourceKafkaMirrormakerStateUpgrade,
			Version: 0,
		},
	}
}

func M3Aggregator() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    v0.ResourceM3Aggregator().CoreConfigSchema().ImpliedType(),
			Upgrade: v0.ResourceM3AggregatorStateUpgrade,
			Version: 0,
		},
	}
}

func M3DB() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    v0.ResourceM3DBResource().CoreConfigSchema().ImpliedType(),
			Upgrade: v0.ResourceM3DBStateUpgrade,
			Version: 0,
		},
	}
}

func MySQL() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    v0.ResourceMySQLResource().CoreConfigSchema().ImpliedType(),
			Upgrade: v0.ResourceMySQLStateUpgrade,
			Version: 0,
		},
	}
}

func Opensearch() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    v0.ResourceOpensearch().CoreConfigSchema().ImpliedType(),
			Upgrade: v0.ResourceOpensearchStateUpgrade,
			Version: 0,
		},
	}
}

func PG() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    v0.ResourcePG().CoreConfigSchema().ImpliedType(),
			Upgrade: v0.ResourcePGStateUpgrade,
			Version: 0,
		},
	}
}

func Redis() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    v0.ResourceRedis().CoreConfigSchema().ImpliedType(),
			Upgrade: v0.ResourceRedisStateUpgrade,
			Version: 0,
		},
	}
}

func ServiceIntegration() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    v0.ResourceServiceIntegration().CoreConfigSchema().ImpliedType(),
			Upgrade: v0.ResourceServiceIntegrationStateUpgrade,
			Version: 0,
		},
	}
}

func ServiceIntegrationEndpoint() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    v0.ResourceServiceIntegrationEndpoint().CoreConfigSchema().ImpliedType(),
			Upgrade: v0.ResourceServiceIntegrationEndpointStateUpgrade,
			Version: 0,
		},
	}
}
