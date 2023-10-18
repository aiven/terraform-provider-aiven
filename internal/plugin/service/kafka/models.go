package kafka

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type kafkaModel struct {
	Timeouts              timeouts.Value `tfsdk:"timeouts"`
	ID                    types.String   `tfsdk:"id"`
	AdditionalDiskSpace   types.String   `tfsdk:"additional_disk_space"`
	CloudName             types.String   `tfsdk:"cloud_name"`
	Components            types.Set      `tfsdk:"components"`
	DefaultACL            types.Bool     `tfsdk:"default_acl"`
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
	ServiceName           types.String   `tfsdk:"service_name"`
	ServicePassword       types.String   `tfsdk:"service_password"`
	ServicePort           types.Int64    `tfsdk:"service_port"`
	ServiceType           types.String   `tfsdk:"service_type"`
	ServiceURI            types.String   `tfsdk:"service_uri"`
	ServiceUsername       types.String   `tfsdk:"service_username"`
	ServiceIntegrations   types.Set      `tfsdk:"service_integrations"`
	State                 types.String   `tfsdk:"state"`
	StaticIPs             types.Set      `tfsdk:"static_ips"`
	Tag                   types.Set      `tfsdk:"tag"`
	TerminationProtection types.Bool     `tfsdk:"termination_protection"`
	KafkaUserConfig       types.Set      `tfsdk:"kafka_user_config"`
}
