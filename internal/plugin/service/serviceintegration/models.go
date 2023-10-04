package serviceintegration

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	idProjectIndex       = 0
	idIntegrationIDIndex = 1
)

// Plugin framework doesn't support embedded structs
// https://github.com/hashicorp/terraform-plugin-framework/issues/242
// We use resourceModel as base model, and copy state to/from dataSourceModel for datasource
type resourceModel struct {
	Timeouts                               timeouts.Value `tfsdk:"timeouts"`
	ID                                     types.String   `tfsdk:"id" copier:"ID"`
	Project                                types.String   `tfsdk:"project" copier:"Project"`
	IntegrationID                          types.String   `tfsdk:"integration_id" copier:"IntegrationID"`
	DestinationEndpointID                  types.String   `tfsdk:"destination_endpoint_id" copier:"DestinationEndpointID"`
	DestinationServiceName                 types.String   `tfsdk:"destination_service_name" copier:"DestinationServiceName"`
	IntegrationType                        types.String   `tfsdk:"integration_type" copier:"IntegrationType"`
	SourceEndpointID                       types.String   `tfsdk:"source_endpoint_id" copier:"SourceEndpointID"`
	SourceServiceName                      types.String   `tfsdk:"source_service_name" copier:"SourceServiceName"`
	ClickhouseKafkaUserConfig              types.Set      `tfsdk:"clickhouse_kafka_user_config" copier:"ClickhouseKafkaUserConfig"`
	ClickhousePostgresqlUserConfig         types.Set      `tfsdk:"clickhouse_postgresql_user_config" copier:"ClickhousePostgresqlUserConfig"`
	DatadogUserConfig                      types.Set      `tfsdk:"datadog_user_config" copier:"DatadogUserConfig"`
	ExternalAwsCloudwatchMetricsUserConfig types.Set      `tfsdk:"external_aws_cloudwatch_metrics_user_config" copier:"ExternalAwsCloudwatchMetricsUserConfig"`
	KafkaConnectUserConfig                 types.Set      `tfsdk:"kafka_connect_user_config" copier:"KafkaConnectUserConfig"`
	KafkaLogsUserConfig                    types.Set      `tfsdk:"kafka_logs_user_config" copier:"KafkaLogsUserConfig"`
	KafkaMirrormakerUserConfig             types.Set      `tfsdk:"kafka_mirrormaker_user_config" copier:"KafkaMirrormakerUserConfig"`
	LogsUserConfig                         types.Set      `tfsdk:"logs_user_config" copier:"LogsUserConfig"`
	MetricsUserConfig                      types.Set      `tfsdk:"metrics_user_config" copier:"MetricsUserConfig"`
}

type dataSourceModel struct {
	ID                                     types.String `tfsdk:"id" copier:"ID"`
	Project                                types.String `tfsdk:"project" copier:"Project"`
	IntegrationID                          types.String `tfsdk:"integration_id" copier:"IntegrationID"`
	DestinationEndpointID                  types.String `tfsdk:"destination_endpoint_id" copier:"DestinationEndpointID"`
	DestinationServiceName                 types.String `tfsdk:"destination_service_name" copier:"DestinationServiceName"`
	IntegrationType                        types.String `tfsdk:"integration_type" copier:"IntegrationType"`
	SourceEndpointID                       types.String `tfsdk:"source_endpoint_id" copier:"SourceEndpointID"`
	SourceServiceName                      types.String `tfsdk:"source_service_name" copier:"SourceServiceName"`
	ClickhouseKafkaUserConfig              types.Set    `tfsdk:"clickhouse_kafka_user_config" copier:"ClickhouseKafkaUserConfig"`
	ClickhousePostgresqlUserConfig         types.Set    `tfsdk:"clickhouse_postgresql_user_config" copier:"ClickhousePostgresqlUserConfig"`
	DatadogUserConfig                      types.Set    `tfsdk:"datadog_user_config" copier:"DatadogUserConfig"`
	ExternalAwsCloudwatchMetricsUserConfig types.Set    `tfsdk:"external_aws_cloudwatch_metrics_user_config" copier:"ExternalAwsCloudwatchMetricsUserConfig"`
	KafkaConnectUserConfig                 types.Set    `tfsdk:"kafka_connect_user_config" copier:"KafkaConnectUserConfig"`
	KafkaLogsUserConfig                    types.Set    `tfsdk:"kafka_logs_user_config" copier:"KafkaLogsUserConfig"`
	KafkaMirrormakerUserConfig             types.Set    `tfsdk:"kafka_mirrormaker_user_config" copier:"KafkaMirrormakerUserConfig"`
	LogsUserConfig                         types.Set    `tfsdk:"logs_user_config" copier:"LogsUserConfig"`
	MetricsUserConfig                      types.Set    `tfsdk:"metrics_user_config" copier:"MetricsUserConfig"`
}

func (p *resourceModel) getID() string {
	i := p.IntegrationID.ValueString()
	if i != "" {
		return i
	}
	return getIDIndex(p.ID.ValueString(), idIntegrationIDIndex)
}

func (p *resourceModel) getProject() string {
	project := p.Project.ValueString()
	if project != "" {
		return project
	}
	return getIDIndex(p.ID.ValueString(), idProjectIndex)
}

func getIDIndex(s string, i int) string {
	list := strings.Split(s, "/")
	if i < len(list) {
		return list[i]
	}
	return ""
}

func getEndpointIDPointer(s string) *string {
	id := getIDIndex(s, idIntegrationIDIndex)
	if s == "" {
		return nil
	}
	return &id
}

func getProjectPointer(s string) *string {
	id := getIDIndex(s, idProjectIndex)
	if s == "" {
		return nil
	}
	return &id
}

func newEndpointID(project string, s *string) types.String {
	if s != nil {
		v := fmt.Sprintf("%s/%s", project, *s)
		s = &v
	}
	return types.StringPointerValue(s)
}
