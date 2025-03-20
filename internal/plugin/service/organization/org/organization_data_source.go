package org

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/diagnostics"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
	providertypes "github.com/aiven/terraform-provider-aiven/internal/plugin/types"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
)

var (
	_ datasource.DataSource              = &organizationDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationDataSource{}
)

// NewOrganizationDataSource is a constructor for the organization data source.
func NewOrganizationDataSource() datasource.DataSource {
	return &organizationDataSource{}
}

// organizationDataSource is the organization data source implementation.
type organizationDataSource struct {
	// client is the instance of the Aiven client to use.
	client avngen.Client

	// diag is the diagnostics helper.
	diag *diagnostics.DiagnosticsHelper
}

// organizationDataSourceModel is the model for the organization data source.
type organizationDataSourceModel struct {
	OrganizationModel
}

// TypeName returns the data source type name for the organization data source.
func (r *organizationDataSource) TypeName() string {
	return "organization_data_source"
}

// Metadata returns the metadata for the organization data source.
func (r *organizationDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

// Schema defines the schema for the organization data source.
func (r *organizationDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Gets information about an organization.",
		Attributes:  util.ResourceSchemaToDataSourceSchema(ResourceSchema(), []string{"id", "name"}, []string{}),
	}
}

// Configure sets up the organization data source.
func (r *organizationDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	p, ok := req.ProviderData.(providertypes.AivenClientProvider)
	if !ok {
		resp.Diagnostics.AddError(
			errmsg.SummaryUnexpectedProviderDataType,
			fmt.Sprintf(errmsg.DetailUnexpectedProviderDataType, req.ProviderData),
		)
		return
	}

	r.client = p.GetGenClient()
	r.diag = diagnostics.NewDiagnosticsHelper(r.TypeName())
}

// ConfigValidators returns the configuration validators for the organization data source.
func (r *organizationDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	validatable := path.Expressions{
		path.MatchRoot("id"),
		path.MatchRoot("name"),
	}

	return []datasource.ConfigValidator{
		datasourcevalidator.Conflicting(validatable...),
		datasourcevalidator.AtLeastOneOf(validatable...),
	}
}

// Read reads an organization data source.
func (r *organizationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config organizationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If ID is not set, find by name
	var organizationID string
	if !config.ID.IsNull() {
		organizationID = config.ID.ValueString()
	} else if !config.Name.IsNull() {
		// List organizations and find by name
		accounts, err := r.client.AccountList(ctx)
		if err != nil {
			r.diag.AddError(&resp.Diagnostics, "reading", err)
			return
		}

		var found bool
		for _, account := range accounts {
			if account.AccountName == config.Name.ValueString() {
				if found {
					resp.Diagnostics.AddError(
						"Multiple organizations found with the same name",
						fmt.Sprintf("Multiple organizations found with name %s, please use ID instead", config.Name.ValueString()),
					)
					return
				}
				organizationID = account.AccountId
				found = true
			}
		}

		if !found {
			resp.Diagnostics.AddError(
				"Organization not found",
				fmt.Sprintf("No organization found with name %s", config.Name.ValueString()),
			)
			return
		}
	}

	// Get the organization
	account, err := r.client.AccountGet(ctx, organizationID)
	if err != nil {
		r.diag.AddError(&resp.Diagnostics, "reading", err)
		return
	}

	// Set all fields
	state := organizationDataSourceModel{}
	state.ID = types.StringValue(account.OrganizationId)
	state.Name = types.StringValue(account.AccountName)
	state.TenantID = types.StringPointerValue(account.TenantId)
	state.CreateTime = types.StringValue(account.CreateTime.String())
	state.UpdateTime = types.StringValue(account.UpdateTime.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
