package serviceintegration

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/jinzhu/copier"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
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

var (
	_ datasource.DataSource              = &serviceIntegrationDataSource{}
	_ datasource.DataSourceWithConfigure = &serviceIntegrationDataSource{}
)

func NewServiceIntegrationDataSource() datasource.DataSource {
	return &serviceIntegrationDataSource{}
}

type serviceIntegrationDataSource struct {
	client *aiven.Client
}

func (s *serviceIntegrationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	s.client = req.ProviderData.(*aiven.Client)
}

func (s *serviceIntegrationDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "aiven_service_integration"
}

func (s *serviceIntegrationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The Service Integration data source provides information about the existing Aiven Service Integration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:   true,
				Validators: []validator.String{endpointIDValidator},
			},
			"integration_id": schema.StringAttribute{
				Description: "Service Integration Id at aiven",
				Computed:    true,
			},
			"destination_endpoint_id": schema.StringAttribute{
				Description: "Destination endpoint for the integration (if any)",
				Computed:    true,
				Validators:  []validator.String{endpointIDValidator},
			},
			"destination_service_name": schema.StringAttribute{
				Description: "Destination service for the integration (if any)",
				Required:    true,
			},
			"integration_type": schema.StringAttribute{
				Description: "Type of the service integration. Possible values: " + schemautil.JoinQuoted(integrationTypes(), ", ", "`"),
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(integrationTypes()...),
				},
			},
			"project": schema.StringAttribute{
				Description: "Project the integration belongs to",
				Required:    true,
			},
			"source_endpoint_id": schema.StringAttribute{
				Description: "Source endpoint for the integration (if any)",
				Computed:    true,
				Validators:  []validator.String{endpointIDValidator},
			},
			"source_service_name": schema.StringAttribute{
				Description: "Source service for the integration (if any)",
				Required:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"clickhouse_kafka_user_config":                clickhousekafka.NewDataSourceSchema(),
			"clickhouse_postgresql_user_config":           clickhousepostgresql.NewDataSourceSchema(),
			"datadog_user_config":                         datadog.NewDataSourceSchema(),
			"external_aws_cloudwatch_metrics_user_config": externalawscloudwatchmetrics.NewDataSourceSchema(),
			"kafka_connect_user_config":                   kafkaconnect.NewDataSourceSchema(),
			"kafka_logs_user_config":                      kafkalogs.NewDataSourceSchema(),
			"kafka_mirrormaker_user_config":               kafkamirrormaker.NewDataSourceSchema(),
			"logs_user_config":                            logs.NewDataSourceSchema(),
			"metrics_user_config":                         metrics.NewDataSourceSchema(),
		},
	}
}

// Read reads datasource
// All functions adapted for resourceModel, so we use it as donor
// Copies state from datasource to resource, then back, when things are done
func (s *serviceIntegrationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var o dataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &o)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var res resourceModel
	err := copier.Copy(&res, &o)
	if err != nil {
		resp.Diagnostics.AddError("data config copy error", err.Error())
	}

	dto, err := getSIByName(ctx, s.client, &res)
	if err != nil {
		resp.Diagnostics.AddError(errmsg.SummaryErrorReadingResource, err.Error())
		return
	}

	loadFromDTO(ctx, &resp.Diagnostics, &res, dto)
	if resp.Diagnostics.HasError() {
		return
	}

	err = copier.Copy(&o, &res)
	if err != nil {
		resp.Diagnostics.AddError("dto copy error", err.Error())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, o)...)
}
