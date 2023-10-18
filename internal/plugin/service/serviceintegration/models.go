package serviceintegration

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceModel struct {
	Timeouts                               timeouts.Value `tfsdk:"timeouts"`
	ID                                     types.String   `tfsdk:"id"`
	Project                                types.String   `tfsdk:"project"`
	IntegrationID                          types.String   `tfsdk:"integration_id"`
	DestinationEndpointID                  types.String   `tfsdk:"destination_endpoint_id"`
	DestinationServiceName                 types.String   `tfsdk:"destination_service_name"`
	IntegrationType                        types.String   `tfsdk:"integration_type"`
	SourceEndpointID                       types.String   `tfsdk:"source_endpoint_id"`
	SourceServiceName                      types.String   `tfsdk:"source_service_name"`
	ClickhouseKafkaUserConfig              types.Set      `tfsdk:"clickhouse_kafka_user_config"`
	ClickhousePostgresqlUserConfig         types.Set      `tfsdk:"clickhouse_postgresql_user_config"`
	DatadogUserConfig                      types.Set      `tfsdk:"datadog_user_config"`
	ExternalAwsCloudwatchMetricsUserConfig types.Set      `tfsdk:"external_aws_cloudwatch_metrics_user_config"`
	KafkaConnectUserConfig                 types.Set      `tfsdk:"kafka_connect_user_config"`
	KafkaLogsUserConfig                    types.Set      `tfsdk:"kafka_logs_user_config"`
	KafkaMirrormakerUserConfig             types.Set      `tfsdk:"kafka_mirrormaker_user_config"`
	LogsUserConfig                         types.Set      `tfsdk:"logs_user_config"`
	MetricsUserConfig                      types.Set      `tfsdk:"metrics_user_config"`
}

type dataSourceModel struct {
	ID                                     types.String `tfsdk:"id"`
	Project                                types.String `tfsdk:"project"`
	IntegrationID                          types.String `tfsdk:"integration_id"`
	DestinationEndpointID                  types.String `tfsdk:"destination_endpoint_id"`
	DestinationServiceName                 types.String `tfsdk:"destination_service_name"`
	IntegrationType                        types.String `tfsdk:"integration_type"`
	SourceEndpointID                       types.String `tfsdk:"source_endpoint_id"`
	SourceServiceName                      types.String `tfsdk:"source_service_name"`
	ClickhouseKafkaUserConfig              types.Set    `tfsdk:"clickhouse_kafka_user_config"`
	ClickhousePostgresqlUserConfig         types.Set    `tfsdk:"clickhouse_postgresql_user_config"`
	DatadogUserConfig                      types.Set    `tfsdk:"datadog_user_config"`
	ExternalAwsCloudwatchMetricsUserConfig types.Set    `tfsdk:"external_aws_cloudwatch_metrics_user_config"`
	KafkaConnectUserConfig                 types.Set    `tfsdk:"kafka_connect_user_config"`
	KafkaLogsUserConfig                    types.Set    `tfsdk:"kafka_logs_user_config"`
	KafkaMirrormakerUserConfig             types.Set    `tfsdk:"kafka_mirrormaker_user_config"`
	LogsUserConfig                         types.Set    `tfsdk:"logs_user_config"`
	MetricsUserConfig                      types.Set    `tfsdk:"metrics_user_config"`
}

func newEndpointID(project string, s *string) types.String {
	if s != nil {
		v := fmt.Sprintf("%s/%s", project, *s)
		s = &v
	}
	return types.StringPointerValue(s)
}
