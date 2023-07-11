package stateupgrader

import (
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/v0/cassandra"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/v0/flink"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/v0/grafana"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/v0/influxdb"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/v0/kafka"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/v0/m3"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/v0/mysql"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/v0/opensearch"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/v0/pg"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/v0/redis"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/v0/serviceintegration"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Cassandra() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    cassandra.ResourceCassandra().CoreConfigSchema().ImpliedType(),
			Upgrade: cassandra.ResourceCassandraStateUpgrade,
			Version: 0,
		},
	}
}

func Flink() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    flink.ResourceFlink().CoreConfigSchema().ImpliedType(),
			Upgrade: flink.ResourceFlinkStateUpgrade,
			Version: 0,
		},
	}
}

func Grafana() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    grafana.ResourceGrafana().CoreConfigSchema().ImpliedType(),
			Upgrade: grafana.ResourceGrafanaStateUpgrade,
			Version: 0,
		},
	}
}

func InfluxDB() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    influxdb.ResourceInfluxDB().CoreConfigSchema().ImpliedType(),
			Upgrade: influxdb.ResourceInfluxDBStateUpgrade,
			Version: 0,
		},
	}
}

func Kafka() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    kafka.ResourceKafka().CoreConfigSchema().ImpliedType(),
			Upgrade: kafka.ResourceKafkaStateUpgrade,
			Version: 0,
		},
	}
}

func KafkaConnect() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    kafka.ResourceKafkaConnect().CoreConfigSchema().ImpliedType(),
			Upgrade: kafka.ResourceKafkaConnectStateUpgrade,
			Version: 0,
		},
	}
}

func KafkaMirrormaker() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    kafka.ResourceKafkaMirrormaker().CoreConfigSchema().ImpliedType(),
			Upgrade: kafka.ResourceKafkaMirrormakerStateUpgrade,
			Version: 0,
		},
	}
}

func KafkaTopic() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    kafka.ResourceKafkaTopic().CoreConfigSchema().ImpliedType(),
			Upgrade: kafka.ResourceKafkaTopicStateUpgrade,
			Version: 0,
		},
	}
}

func M3Aggregator() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    m3.ResourceM3Aggregator().CoreConfigSchema().ImpliedType(),
			Upgrade: m3.ResourceM3AggregatorStateUpgrade,
			Version: 0,
		},
	}
}

func M3DB() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    m3.ResourceM3DBResource().CoreConfigSchema().ImpliedType(),
			Upgrade: m3.ResourceM3DBStateUpgrade,
			Version: 0,
		},
	}
}

func MySQL() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    mysql.ResourceMySQLResource().CoreConfigSchema().ImpliedType(),
			Upgrade: mysql.ResourceMySQLStateUpgrade,
			Version: 0,
		},
	}
}

func OpenSearch() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    opensearch.ResourceOpenSearch().CoreConfigSchema().ImpliedType(),
			Upgrade: opensearch.ResourceOpenSearchStateUpgrade,
			Version: 0,
		},
	}
}

func PG() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    pg.ResourcePG().CoreConfigSchema().ImpliedType(),
			Upgrade: pg.ResourcePGStateUpgrade,
			Version: 0,
		},
	}
}

func Redis() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    redis.ResourceRedis().CoreConfigSchema().ImpliedType(),
			Upgrade: redis.ResourceRedisStateUpgrade,
			Version: 0,
		},
	}
}

func ServiceIntegration() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    serviceintegration.ResourceServiceIntegration().CoreConfigSchema().ImpliedType(),
			Upgrade: serviceintegration.ResourceServiceIntegrationStateUpgrade,
			Version: 0,
		},
	}
}

func ServiceIntegrationEndpoint() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type:    serviceintegration.ResourceServiceIntegrationEndpoint().CoreConfigSchema().ImpliedType(),
			Upgrade: serviceintegration.ResourceServiceIntegrationEndpointStateUpgrade,
			Version: 0,
		},
	}
}
