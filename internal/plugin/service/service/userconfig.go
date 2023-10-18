package service

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/userconfig/integration/kafkaconnect"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/userconfig/integration/kafkamirrormaker"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/userconfig/service/cassandra"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/userconfig/service/clickhouse"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/userconfig/service/elasticsearch"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/userconfig/service/flink"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/userconfig/service/grafana"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/userconfig/service/influxdb"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/userconfig/service/kafka"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/userconfig/service/m3aggregator"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/userconfig/service/m3db"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/userconfig/service/mysql"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/userconfig/service/opensearch"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/userconfig/service/pg"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/userconfig/service/redis"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

const (
	TypePg               string = "pg"
	TypeCassandra        string = "cassandra"
	TypeElasticsearch    string = "elasticsearch"
	TypeOpenSearch       string = "opensearch"
	TypeGrafana          string = "grafana"
	TypeInfluxDB         string = "influxdb"
	TypeRedis            string = "redis"
	TypeMySQL            string = "mysql"
	TypeKafka            string = "kafka"
	TypeKafkaConnect     string = "kafka_connect"
	TypeKafkaMirrormaker string = "kafka_mirrormaker"
	TypeM3               string = "m3db"
	TypeM3Aggregator     string = "m3aggregator"
	TypeFlink            string = "flink"
	TypeClickhouse       string = "clickhouse"
)

// flattenUserConfig from aiven to terraform
func flattenUserConfig(ctx context.Context, diags diag.Diagnostics, r *Resource, dto *aiven.Service) {
	if dto.UserConfig == nil {
		return
	}

	switch r.ServiceType.ValueString() {
	case TypePg:
		r.PgUserConfig = pg.Flatten(ctx, diags, dto.UserConfig)
	case TypeCassandra:
		r.CassandraUserConfig = cassandra.Flatten(ctx, diags, dto.UserConfig)
	case TypeElasticsearch:
		r.ElasticsearchUserConfig = elasticsearch.Flatten(ctx, diags, dto.UserConfig)
	case TypeOpenSearch:
		r.OpenSearchUserConfig = opensearch.Flatten(ctx, diags, dto.UserConfig)
	case TypeGrafana:
		r.GrafanaUserConfig = grafana.Flatten(ctx, diags, dto.UserConfig)
	case TypeInfluxDB:
		r.InfluxdbUserConfig = influxdb.Flatten(ctx, diags, dto.UserConfig)
	case TypeRedis:
		r.RedisUserConfig = redis.Flatten(ctx, diags, dto.UserConfig)
	case TypeMySQL:
		r.MysqlUserConfig = mysql.Flatten(ctx, diags, dto.UserConfig)
	case TypeKafka:
		r.KafkaUserConfig = kafka.Flatten(ctx, diags, dto.UserConfig)
	case TypeKafkaConnect:
		r.KafkaConnectUserConfig = kafkaconnect.Flatten(ctx, diags, dto.UserConfig)
	case TypeKafkaMirrormaker:
		r.KafkaMirrormakerUserConfig = kafkamirrormaker.Flatten(ctx, diags, dto.UserConfig)
	case TypeM3:
		r.M3dbUserConfig = m3db.Flatten(ctx, diags, dto.UserConfig)
	case TypeM3Aggregator:
		r.M3aggregatorUserConfig = m3aggregator.Flatten(ctx, diags, dto.UserConfig)
	case TypeFlink:
		r.FlinkUserConfig = flink.Flatten(ctx, diags, dto.UserConfig)
	case TypeClickhouse:
		r.ClickhouseUserConfig = clickhouse.Flatten(ctx, diags, dto.UserConfig)
	}
}

// expandUserConfig from terraform to aiven
func expandUserConfig(ctx context.Context, diags diag.Diagnostics, o *Resource, create bool) map[string]any {
	var config any

	switch {
	case schemautil.HasValue(o.PgUserConfig):
		config = pg.Expand(ctx, diags, o.PgUserConfig)
	case schemautil.HasValue(o.CassandraUserConfig):
		config = cassandra.Expand(ctx, diags, o.CassandraUserConfig)
	case schemautil.HasValue(o.ElasticsearchUserConfig):
		config = elasticsearch.Expand(ctx, diags, o.ElasticsearchUserConfig)
	case schemautil.HasValue(o.OpenSearchUserConfig):
		config = opensearch.Expand(ctx, diags, o.OpenSearchUserConfig)
	case schemautil.HasValue(o.GrafanaUserConfig):
		config = grafana.Expand(ctx, diags, o.GrafanaUserConfig)
	case schemautil.HasValue(o.InfluxdbUserConfig):
		config = influxdb.Expand(ctx, diags, o.InfluxdbUserConfig)
	case schemautil.HasValue(o.RedisUserConfig):
		config = redis.Expand(ctx, diags, o.RedisUserConfig)
	case schemautil.HasValue(o.MysqlUserConfig):
		config = mysql.Expand(ctx, diags, o.MysqlUserConfig)
	case schemautil.HasValue(o.KafkaUserConfig):
		config = kafka.Expand(ctx, diags, o.KafkaUserConfig)
	case schemautil.HasValue(o.KafkaConnectUserConfig):
		config = kafkaconnect.Expand(ctx, diags, o.KafkaConnectUserConfig)
	case schemautil.HasValue(o.KafkaMirrormakerUserConfig):
		config = kafkamirrormaker.Expand(ctx, diags, o.KafkaMirrormakerUserConfig)
	case schemautil.HasValue(o.M3dbUserConfig):
		config = m3db.Expand(ctx, diags, o.M3dbUserConfig)
	case schemautil.HasValue(o.M3aggregatorUserConfig):
		config = m3aggregator.Expand(ctx, diags, o.M3aggregatorUserConfig)
	case schemautil.HasValue(o.FlinkUserConfig):
		config = flink.Expand(ctx, diags, o.FlinkUserConfig)
	case schemautil.HasValue(o.ClickhouseUserConfig):
		config = clickhouse.Expand(ctx, diags, o.ClickhouseUserConfig)
	}

	if diags.HasError() {
		return nil
	}

	dict, err := schemautil.MarshalUserConfig(config, create)
	if err != nil {
		diags.AddError("Failed to expand user config", err.Error())
		return nil
	}
	return dict
}
