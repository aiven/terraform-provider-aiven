package serviceintegration

import (
	"context"
	"fmt"
	"regexp"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

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
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

var endpointIDValidator = stringvalidator.RegexMatches(
	regexp.MustCompile(`^[a-zA-Z0-9_-]*/[a-zA-Z0-9_-]*$`),
	"endpoint id should have the following format: project_name/endpoint_id",
)

var (
	_ resource.Resource                = &serviceIntegrationResource{}
	_ resource.ResourceWithConfigure   = &serviceIntegrationResource{}
	_ resource.ResourceWithImportState = &serviceIntegrationResource{}
)

func NewServiceIntegrationResource() resource.Resource {
	return &serviceIntegrationResource{}
}

type serviceIntegrationResource struct {
	client *aiven.Client
}

func (s *serviceIntegrationResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	s.client = req.ProviderData.(*aiven.Client)
}

func (s *serviceIntegrationResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aiven_service_integration"
}

func (s *serviceIntegrationResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = util.GeneralizeSchema(ctx, schema.Schema{
		Description: "The Service Integration resource allows the creation and management of Aiven Service Integrations.",
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Optional: true,
				Validators: []validator.String{
					endpointIDValidator,
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("destination_endpoint_id"),
						path.MatchRoot("destination_service_name"),
					),
				},
			},
			"destination_service_name": schema.StringAttribute{
				Description: "Destination service for the integration (if any)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Optional: true,
			},
			"integration_type": schema.StringAttribute{
				Description: "Type of the service integration. Possible values: " + schemautil.JoinQuoted(integrationTypes(), ", ", "`"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf(integrationTypes()...),
				},
			},
			"project": schema.StringAttribute{
				Description: "Project the integration belongs to",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Required: true,
			},
			"source_endpoint_id": schema.StringAttribute{
				Description: "Source endpoint for the integration (if any)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Optional: true,
				Validators: []validator.String{
					endpointIDValidator,
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("source_endpoint_id"),
						path.MatchRoot("source_service_name"),
					),
				},
			},
			"source_service_name": schema.StringAttribute{
				Description: "Source service for the integration (if any)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Optional: true,
			},
		},
		Blocks: map[string]schema.Block{
			"clickhouse_kafka_user_config":                clickhousekafka.NewResourceSchema(),
			"clickhouse_postgresql_user_config":           clickhousepostgresql.NewResourceSchema(),
			"datadog_user_config":                         datadog.NewResourceSchema(),
			"external_aws_cloudwatch_metrics_user_config": externalawscloudwatchmetrics.NewResourceSchema(),
			"kafka_connect_user_config":                   kafkaconnect.NewResourceSchema(),
			"kafka_logs_user_config":                      kafkalogs.NewResourceSchema(),
			"kafka_mirrormaker_user_config":               kafkamirrormaker.NewResourceSchema(),
			"logs_user_config":                            logs.NewResourceSchema(),
			"metrics_user_config":                         metrics.NewResourceSchema(),
		},
	})
}

func (s *serviceIntegrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var o resourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &o)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// read_replicas can be only be created alongside the service. also the only way to promote the replica
	// is to delete the service integration that was created, so we should make it least painful to do so.
	// for now, we support to seemlessly import preexisting 'read_replica' service integrations in the resource create
	// all other integrations should be imported using `terraform import`
	if o.IntegrationType.ValueString() == readReplicaType {
		if preexisting, err := getSIByName(ctx, s.client, &o); err != nil {
			resp.Diagnostics.AddError("unable to search for possible preexisting 'read_replica' service integration", err.Error())
			return
		} else if preexisting != nil {
			o.IntegrationID = types.StringValue(preexisting.ServiceIntegrationID)
			s.read(ctx, &resp.Diagnostics, &resp.State, &o)
			return
		}
	}

	userConfig, err := expandUserConfig(ctx, &resp.Diagnostics, &o, true)
	if err != nil {
		resp.Diagnostics.AddError("Failed to expand user config", err.Error())
		return
	}
	createReq := aiven.CreateServiceIntegrationRequest{
		DestinationProject:    getProjectPointer(o.DestinationEndpointID.ValueString()),
		DestinationEndpointID: getEndpointIDPointer(o.DestinationEndpointID.ValueString()),
		DestinationService:    o.DestinationServiceName.ValueStringPointer(),
		IntegrationType:       o.IntegrationType.ValueString(),
		SourceProject:         getProjectPointer(o.SourceEndpointID.ValueString()),
		SourceEndpointID:      getEndpointIDPointer(o.SourceEndpointID.ValueString()),
		SourceService:         o.SourceServiceName.ValueStringPointer(),
		UserConfig:            userConfig,
	}

	dto, err := s.client.ServiceIntegrations.Create(ctx, o.Project.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError(errmsg.SummaryErrorCreatingResource, err.Error())
		return
	}

	o.IntegrationID = types.StringValue(dto.ServiceIntegrationID)
	s.read(ctx, &resp.Diagnostics, &resp.State, &o)
}

func (s *serviceIntegrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var o resourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &o)...)
	if resp.Diagnostics.HasError() {
		return
	}
	s.read(ctx, &resp.Diagnostics, &resp.State, &o)
}

func (s *serviceIntegrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state resourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var o resourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &o)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Copies ID from the state
	o.IntegrationID = state.IntegrationID
	userConfig, err := expandUserConfig(ctx, &resp.Diagnostics, &o, false)
	if err != nil {
		resp.Diagnostics.AddError("Failed to expand user config", err.Error())
		return
	}

	_, err = s.client.ServiceIntegrations.Update(
		ctx,
		state.Project.ValueString(),
		state.IntegrationID.ValueString(),
		aiven.UpdateServiceIntegrationRequest{
			UserConfig: userConfig,
		},
	)

	if err != nil {
		resp.Diagnostics.AddError(errmsg.SummaryErrorUpdatingResource, err.Error())
		return
	}

	s.read(ctx, &resp.Diagnostics, &resp.State, &o)
}

func (s *serviceIntegrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var o resourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &o)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := s.client.ServiceIntegrations.Delete(ctx, o.Project.ValueString(), o.IntegrationID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errmsg.SummaryErrorDeletingResource, err.Error())
	}
}

func (s *serviceIntegrationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// read reads from API and saves to state
func (s *serviceIntegrationResource) read(ctx context.Context, diags *diag.Diagnostics, state *tfsdk.State, o *resourceModel) {
	dto, err := getSIByID(ctx, s.client, o)
	if err != nil {
		diags.AddError(errmsg.SummaryErrorReadingResource, err.Error())
		return
	}

	loadFromDTO(ctx, diags, o, dto)
	if diags.HasError() {
		return
	}
	diags.Append(state.Set(ctx, o)...)
}

// getSIByID gets ServiceIntegration by ID
func getSIByID(ctx context.Context, client *aiven.Client, o *resourceModel) (dto *aiven.ServiceIntegration, err error) {
	id := o.getID()
	project := o.getProject()
	if len(id)*len(project) == 0 {
		return nil, fmt.Errorf("no ID or project provided")
	}

	return dto, util.WaitActive(ctx, func() error {
		dto, err = client.ServiceIntegrations.Get(ctx, project, id)
		if err != nil {
			return err
		}
		if !dto.Active {
			return fmt.Errorf("service integration is not active")
		}
		return nil
	})
}

// getSIByName gets ServiceIntegration by name
func getSIByName(ctx context.Context, client *aiven.Client, o *resourceModel) (*aiven.ServiceIntegration, error) {
	project := o.Project.ValueString()
	integrationType := o.IntegrationType.ValueString()
	sourceServiceName := o.SourceServiceName.ValueString()
	destinationServiceName := o.DestinationServiceName.ValueString()

	integrations, err := client.ServiceIntegrations.List(ctx, project, sourceServiceName)
	if err != nil && !aiven.IsNotFound(err) {
		return nil, fmt.Errorf("unable to get list of service integrations: %s", err)
	}

	for _, i := range integrations {
		if i.SourceService == nil || i.DestinationService == nil || i.ServiceIntegrationID == "" {
			continue
		}

		if i.IntegrationType == integrationType &&
			*i.SourceService == sourceServiceName &&
			*i.DestinationService == destinationServiceName {
			return i, nil
		}
	}

	return nil, nil
}

// loadFromDTO loads API values to terraform object
func loadFromDTO(ctx context.Context, diags *diag.Diagnostics, o *resourceModel, dto *aiven.ServiceIntegration) {
	flattenUserConfig(ctx, diags, o, dto)
	if diags.HasError() {
		return
	}

	id := o.getID()
	project := o.getProject()
	o.ID = newEndpointID(project, &id)
	o.DestinationEndpointID = newEndpointID(project, dto.DestinationEndpointID)
	o.DestinationServiceName = types.StringPointerValue(dto.DestinationService)
	o.IntegrationType = types.StringValue(dto.IntegrationType)
	o.SourceEndpointID = newEndpointID(project, dto.SourceEndpointID)
	o.SourceServiceName = types.StringPointerValue(dto.SourceService)
}
