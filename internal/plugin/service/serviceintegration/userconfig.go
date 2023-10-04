package serviceintegration

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/userconfig/integration/clickhousekafka"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/userconfig/integration/clickhousepostgresql"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/userconfig/integration/datadog"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/userconfig/integration/externalawscloudwatchmetrics"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/userconfig/integration/kafkaconnect"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/userconfig/integration/kafkalogs"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/userconfig/integration/kafkamirrormaker"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/userconfig/integration/logs"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/userconfig/integration/metrics"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

const (
	clickhouseKafkaType              = "clickhouse_kafka"
	clickhousePostgresqlType         = "clickhouse_postgresql"
	datadogType                      = "datadog"
	externalAwsCloudwatchMetricsType = "external_aws_cloudwatch_metrics"
	kafkaConnectType                 = "kafka_connect"
	kafkaLogsType                    = "kafka_logs"
	kafkaMirrormakerType             = "kafka_mirrormaker"
	logsType                         = "logs"
	metricsType                      = "metrics"
	readReplicaType                  = "read_replica"
)

func integrationTypes() []string {
	return []string{
		"alertmanager",
		"cassandra_cross_service_cluster",
		clickhouseKafkaType,
		clickhousePostgresqlType,
		"dashboard",
		datadogType,
		"datasource",
		"external_aws_cloudwatch_logs",
		externalAwsCloudwatchMetricsType,
		"external_elasticsearch_logs",
		"external_google_cloud_logging",
		"external_opensearch_logs",
		"flink",
		"internal_connectivity",
		"jolokia",
		kafkaConnectType,
		kafkaLogsType,
		kafkaMirrormakerType,
		logsType,
		"m3aggregator",
		"m3coordinator",
		metricsType,
		"opensearch_cross_cluster_replication",
		"opensearch_cross_cluster_search",
		"prometheus",
		readReplicaType,
		"rsyslog",
		"schema_registry_proxy",
	}
}

// flattenUserConfig from aiven to terraform
func flattenUserConfig(ctx context.Context, diags *diag.Diagnostics, o *resourceModel, dto *aiven.ServiceIntegration) {
	if dto.UserConfig == nil {
		return
	}

	// We set user config from Aiven only if it's been set in TF
	// Otherwise it will produce invalid "after apply"
	switch {
	case isSet(o.ClickhouseKafkaUserConfig):
		o.ClickhouseKafkaUserConfig = clickhousekafka.Flatten(ctx, diags, dto.UserConfig)
	case isSet(o.ClickhousePostgresqlUserConfig):
		o.ClickhousePostgresqlUserConfig = clickhousepostgresql.Flatten(ctx, diags, dto.UserConfig)
	case isSet(o.DatadogUserConfig):
		o.DatadogUserConfig = datadog.Flatten(ctx, diags, dto.UserConfig)
	case isSet(o.ExternalAwsCloudwatchMetricsUserConfig):
		o.ExternalAwsCloudwatchMetricsUserConfig = externalawscloudwatchmetrics.Flatten(ctx, diags, dto.UserConfig)
	case isSet(o.KafkaConnectUserConfig):
		o.KafkaConnectUserConfig = kafkaconnect.Flatten(ctx, diags, dto.UserConfig)
	case isSet(o.KafkaLogsUserConfig):
		o.KafkaLogsUserConfig = kafkalogs.Flatten(ctx, diags, dto.UserConfig)
	case isSet(o.KafkaMirrormakerUserConfig):
		o.KafkaMirrormakerUserConfig = kafkamirrormaker.Flatten(ctx, diags, dto.UserConfig)
	case isSet(o.LogsUserConfig):
		o.LogsUserConfig = logs.Flatten(ctx, diags, dto.UserConfig)
	case isSet(o.MetricsUserConfig):
		o.MetricsUserConfig = metrics.Flatten(ctx, diags, dto.UserConfig)
	}
}

// expandUserConfig from terraform to aiven
func expandUserConfig(ctx context.Context, diags *diag.Diagnostics, o *resourceModel, create bool) (map[string]any, error) {
	var marshal func(any) (map[string]any, error)
	if create {
		marshal = schemautil.MarshalCreateUserConfig
	} else {
		marshal = schemautil.MarshalUpdateUserConfig
	}

	// If invalid integration type is set
	// This will send wrong config to Aiven
	// Which is sort of a validation too
	switch {
	case isSet(o.ClickhouseKafkaUserConfig):
		return marshal(clickhousekafka.Expand(ctx, diags, o.ClickhouseKafkaUserConfig))
	case isSet(o.ClickhousePostgresqlUserConfig):
		return marshal(clickhousepostgresql.Expand(ctx, diags, o.ClickhousePostgresqlUserConfig))
	case isSet(o.DatadogUserConfig):
		return marshal(datadog.Expand(ctx, diags, o.DatadogUserConfig))
	case isSet(o.ExternalAwsCloudwatchMetricsUserConfig):
		return marshal(externalawscloudwatchmetrics.Expand(ctx, diags, o.ExternalAwsCloudwatchMetricsUserConfig))
	case isSet(o.KafkaConnectUserConfig):
		return marshal(kafkaconnect.Expand(ctx, diags, o.KafkaConnectUserConfig))
	case isSet(o.KafkaLogsUserConfig):
		return marshal(kafkalogs.Expand(ctx, diags, o.KafkaLogsUserConfig))
	case isSet(o.KafkaMirrormakerUserConfig):
		return marshal(kafkamirrormaker.Expand(ctx, diags, o.KafkaMirrormakerUserConfig))
	case isSet(o.LogsUserConfig):
		return marshal(logs.Expand(ctx, diags, o.LogsUserConfig))
	case isSet(o.MetricsUserConfig):
		return marshal(metrics.Expand(ctx, diags, o.MetricsUserConfig))
	default:
		return nil, nil
	}
}

func isSet(o types.Set) bool {
	return !(o.IsUnknown() || o.IsNull())
}
