package service

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/docker/go-units"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Plugin framework doesn't support embedded structs
// https://github.com/hashicorp/terraform-plugin-framework/issues/242
// We use resource as a base model, and copy state to/from dataSourceModel for datasource

type Resource struct {
	Timeouts              timeouts.Value `tfsdk:"timeouts"`
	ID                    types.String   `tfsdk:"id"`
	AdditionalDiskSpace   types.String   `tfsdk:"additional_disk_space"`
	CloudName             types.String   `tfsdk:"cloud_name"`
	Components            types.Set      `tfsdk:"components"`
	DiskSpaceCap          types.String   `tfsdk:"disk_space_cap"`
	DiskSpaceDefault      types.String   `tfsdk:"disk_space_default"`
	DiskSpaceStep         types.String   `tfsdk:"disk_space_step"`
	DiskSpaceUsed         types.String   `tfsdk:"disk_space_used"`
	Kafka                 types.Set      `tfsdk:"kafka"`
	MaintenanceWindowDow  types.String   `tfsdk:"maintenance_window_dow"`
	MaintenanceWindowTime types.String   `tfsdk:"maintenance_window_time"`
	Plan                  types.String   `tfsdk:"plan"`
	Project               types.String   `tfsdk:"project"`
	ProjectVPCId          types.String   `tfsdk:"project_vpc_id"`
	ServiceHost           types.String   `tfsdk:"service_host"`
	ServiceIntegrations   types.Set      `tfsdk:"service_integrations"`
	ServiceName           types.String   `tfsdk:"service_name"`
	ServicePassword       types.String   `tfsdk:"service_password"`
	ServicePort           types.Int64    `tfsdk:"service_port"`
	ServiceType           types.String   `tfsdk:"service_type"`
	ServiceURI            types.String   `tfsdk:"service_uri"`
	ServiceUsername       types.String   `tfsdk:"service_username"`
	State                 types.String   `tfsdk:"state"`
	StaticIPs             types.Set      `tfsdk:"static_ips"`
	Tag                   types.Set      `tfsdk:"tag"`
	TerminationProtection types.Bool     `tfsdk:"termination_protection"`

	// User configs
	PgUserConfig               types.Set `tfsdk:"pg_user_config"`
	CassandraUserConfig        types.Set `tfsdk:"cassandra_user_config"`
	ElasticsearchUserConfig    types.Set `tfsdk:"elasticsearch_user_config"`
	OpenSearchUserConfig       types.Set `tfsdk:"opensearch_user_config"`
	GrafanaUserConfig          types.Set `tfsdk:"grafana_user_config"`
	InfluxdbUserConfig         types.Set `tfsdk:"influxdb_user_config"`
	RedisUserConfig            types.Set `tfsdk:"redis_user_config"`
	MysqlUserConfig            types.Set `tfsdk:"mysql_user_config"`
	KafkaUserConfig            types.Set `tfsdk:"kafka_user_config"`
	KafkaConnectUserConfig     types.Set `tfsdk:"kafka_connect_user_config"`
	KafkaMirrormakerUserConfig types.Set `tfsdk:"kafka_mirrormaker_user_config"`
	M3dbUserConfig             types.Set `tfsdk:"m3db_user_config"`
	M3aggregatorUserConfig     types.Set `tfsdk:"m3aggregator_user_config"`
	FlinkUserConfig            types.Set `tfsdk:"flink_user_config"`
	ClickhouseUserConfig       types.Set `tfsdk:"clickhouse_user_config"`
}

type serviceIntegration struct {
	SourceServiceName types.String `tfsdk:"source_service_name"`
	IntegrationType   types.String `tfsdk:"integration_type"`
}

var serviceIntegrationAttrs = map[string]attr.Type{
	"source_service_name": types.StringType,
	"integration_type":    types.StringType,
}

func expandServiceIntegration(_ context.Context, _ diag.Diagnostics, o *serviceIntegration) *aiven.NewServiceIntegration {
	return &aiven.NewServiceIntegration{
		IntegrationType: o.IntegrationType.ValueString(),
		SourceService:   o.SourceServiceName.ValueStringPointer(),
	}
}

func flattenServiceIntegration(_ context.Context, _ diag.Diagnostics, o *aiven.ServiceIntegration) *serviceIntegration {
	return &serviceIntegration{
		SourceServiceName: types.StringPointerValue(o.SourceService),
		IntegrationType:   types.StringValue(o.IntegrationType),
	}
}

type tag struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

var tagAttrs = map[string]attr.Type{
	"key":   types.StringType,
	"value": types.StringType,
}

func flattenTags(ctx context.Context, diags diag.Diagnostics, src map[string]string) types.Set {
	tags := make([]*tag, 0, len(src))
	for k, v := range src {
		tags = append(tags, &tag{
			Key:   types.StringValue(k),
			Value: types.StringValue(v),
		})
	}
	oType := types.ObjectType{AttrTypes: tagAttrs}
	result, d := types.SetValueFrom(ctx, oType, tags)
	diags.Append(d...)
	if diags.HasError() {
		return types.SetValueMust(oType, []attr.Value{})
	}
	return result
}

func fromPointers[T any](collection []*T) []T {
	result := make([]T, 0)
	for i := range collection {
		result = append(result, *collection[i])
	}
	return result
}

func getMaintenanceWindow(o *Resource) *aiven.MaintenanceWindow {
	dow := o.MaintenanceWindowDow.ValueString()
	if dow == "never" {
		// `never` is not available in the API, but can be set by support
		// Sending this back to the backend will fail the validation
		return nil
	}
	t := o.MaintenanceWindowTime.ValueString()
	if len(dow) > 0 && len(t) > 0 {
		return &aiven.MaintenanceWindow{DayOfWeek: dow, TimeOfDay: t}
	}
	return nil
}

func inMiB(b string) int {
	if b == "" {
		return 0
	}
	bytes, _ := units.RAMInBytes(b)
	return int(bytes / units.MiB)
}
