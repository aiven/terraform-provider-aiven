package toapi

func GetToAPIBlueprint(t string) map[string]interface{} {
	switch t {
	case "cassandra":
		return ServiceTypeCassandraToAPIBlueprint()
	case "clickhouse":
		return ServiceTypeClickhouseToAPIBlueprint()
	case "elasticsearch":
		return ServiceTypeElasticsearchToAPIBlueprint()
	case "flink":
		return ServiceTypeFlinkToAPIBlueprint()
	case "grafana":
		return ServiceTypeGrafanaToAPIBlueprint()
	case "influxdb":
		return ServiceTypeInfluxdbToAPIBlueprint()
	case "kafka":
		return ServiceTypeKafkaToAPIBlueprint()
	case "kafka_connect":
		return ServiceTypeKafkaConnectToAPIBlueprint()
	case "kafka_mirrormaker":
		return ServiceTypeKafkaMirrormakerToAPIBlueprint()
	case "m3aggregator":
		return ServiceTypeM3aggregatorToAPIBlueprint()
	case "m3db":
		return ServiceTypeM3dbToAPIBlueprint()
	case "mysql":
		return ServiceTypeMysqlToAPIBlueprint()
	case "opensearch":
		return ServiceTypeOpensearchToAPIBlueprint()
	case "pg":
		return ServiceTypePgToAPIBlueprint()
	case "redis":
		return ServiceTypeRedisToAPIBlueprint()
	}

	return nil
}
