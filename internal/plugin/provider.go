// Package plugin is the implementation of the Aiven provider.
package plugin

import (
	"context"
	"os"

	"github.com/aiven/aiven-go-client/v2"
	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/providerdata"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/externalidentity"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/governance/access"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/organization/address"
	applicationuser "github.com/aiven/terraform-provider-aiven/internal/plugin/service/organization/application_user"
	applicationusertoken "github.com/aiven/terraform-provider-aiven/internal/plugin/service/organization/application_user_token"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/organization/billinggroup"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/organization/billinggrouplist"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/organization/groupproject"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/organization/organization"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/organization/permission"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/organization/project"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/organization/usergroupmember"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/planlist"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
)

// AivenProvider is the provider implementation for Aiven.
type AivenProvider struct {
	// version is the version of the provider.
	version string

	// Client is the handwritten Aiven client
	Client *aiven.Client

	// GenClient is the generated Aiven client
	GenClient avngen.Client
}

// GetClient returns the handwritten Aiven client.
func (p *AivenProvider) GetClient() *aiven.Client {
	return p.Client
}

// GetGenClient returns the generated Aiven client.
func (p *AivenProvider) GetGenClient() avngen.Client {
	return p.GenClient
}

var (
	_ provider.Provider         = &AivenProvider{}
	_ providerdata.ProviderData = &AivenProvider{}
)

// AivenProviderModel is the configuration for the Aiven provider.
type AivenProviderModel struct {
	// APIToken is the Aiven API token.
	APIToken types.String `tfsdk:"api_token"`
}

// Metadata returns information about the provider.
func (p *AivenProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "aiven"

	resp.Version = p.version
}

// Schema returns the schema for this provider's configuration.
func (p *AivenProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_token": schema.StringAttribute{
				// Description should match the one in internal/sdkprovider/provider/provider.go.
				Description: "Aiven authentication token. " +
					"Can also be set with the AIVEN_TOKEN environment variable.",
				// TODO: MarkdownDescription is not supported by Terraform Plugin SDK, and is a feature
				//  that is only available in the Terraform Plugin Framework.
				//  We need to uncomment this once the Terraform Plugin SDK supports it (unlikely), or
				//  when we fully migrate to the Terraform Plugin Framework.
				//
				// MarkdownDescription: "Aiven authentication token. " +
				// 	"Can also be set with the `AIVEN_TOKEN` environment variable.",
				Optional:  true,
				Sensitive: true,
			},
		},
		// TODO: Description and MarkdownDescription are not supported by Terraform Plugin SDK, and are features
		//  that are only available in the Terraform Plugin Framework.
		//  We need to uncomment this once the Terraform Plugin SDK supports them (unlikely), or
		//  when we fully migrate to the Terraform Plugin Framework.
		//
		//  N.B. Description is supported by the Terraform Plugin SDK when used in the context of a resource or
		//  data source.
		//
		// MarkdownDescription: "The Aiven provider is used to interact with the various services that Aiven offers. " +
		// 	"The provider needs to be configured with the proper credentials before it can be used.",
	}
}

// Configure configures the provider.
func (p *AivenProvider) Configure(
	ctx context.Context,
	req provider.ConfigureRequest,
	resp *provider.ConfigureResponse,
) {
	var data AivenProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.APIToken.IsNull() {
		token, ok := os.LookupEnv("AIVEN_TOKEN")
		if !ok {
			resp.Diagnostics.AddError(errmsg.SummaryTokenMissing, errmsg.DetailTokenMissing)

			return
		}

		data.APIToken = types.StringValue(token)
	}

	// Initialize the handwritten client
	client, err := common.NewAivenClient(
		common.TokenOpt(data.APIToken.ValueString()),
		common.TFVersionOpt(req.TerraformVersion),
		common.BuildVersionOpt(p.version),
	)
	if err != nil {
		resp.Diagnostics.AddError(errmsg.SummaryConstructingClient, err.Error())
		return
	}
	p.Client = client

	// Initialize the generated client
	genClient, err := common.NewAivenGenClient(
		common.TokenOpt(data.APIToken.ValueString()),
		common.TFVersionOpt(req.TerraformVersion),
		common.BuildVersionOpt(p.version),
	)
	if err != nil {
		resp.Diagnostics.AddError(errmsg.SummaryConstructingClient, err.Error())
		return
	}
	p.GenClient = genClient

	// Pass the provider itself as the provider data
	resp.DataSourceData = p
	resp.ResourceData = p
}

// Resources returns the resources supported by this provider.
func (p *AivenProvider) Resources(context.Context) []func() resource.Resource {
	// List of resources that are currently available in the provider.
	resources := []func() resource.Resource{
		organization.NewResource,
		usergroupmember.NewResource,
		groupproject.NewResource,
		access.NewResource,
		applicationuser.NewResource,
		applicationusertoken.NewResource,
		permission.NewResource,
	}

	// Add to a list of resources that are currently in beta.
	if util.IsBeta() {
		betaResources := []func() resource.Resource{
			address.NewResource,
			billinggroup.NewResource,
			project.NewResource,
		}
		resources = append(resources, betaResources...)
	}

	return resources
}

// DataSources returns the data sources supported by this provider.
func (p *AivenProvider) DataSources(context.Context) []func() datasource.DataSource {
	// List of data sources that are currently available in the provider.
	dataSources := []func() datasource.DataSource{
		organization.NewDataSource,
		applicationuser.NewDataSource,
		planlist.NewDataSource,
	}

	// Add to a list of data sources that are currently in beta.
	if util.IsBeta() {
		betaDataSources := []func() datasource.DataSource{
			externalidentity.NewDataSource,
			address.NewDataSource,
			billinggroup.NewDataSource,
			billinggrouplist.NewDataSource,
			project.NewDataSource,
		}
		dataSources = append(dataSources, betaDataSources...)
	}

	return dataSources
}

// New returns a new provider factory for the Aiven provider.
func New(version string) provider.Provider {
	return &AivenProvider{
		version: version,
	}
}
