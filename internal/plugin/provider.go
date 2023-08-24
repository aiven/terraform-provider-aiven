// Package provider is the implementation of the Aiven provider.
package plugin

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/organization"
)

// AivenProvider is the provider implementation for Aiven.
type AivenProvider struct {
	// version is the version of the provider.
	version string
}

var _ provider.Provider = &AivenProvider{}

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

	client, err := common.NewCustomAivenClient(data.APIToken.String(), req.TerraformVersion, p.version)
	if err != nil {
		resp.Diagnostics.AddError(errmsg.SummaryConstructingClient, err.Error())

		return
	}

	resp.DataSourceData = client

	resp.ResourceData = client
}

// Resources returns the resources supported by this provider.
func (p *AivenProvider) Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		organization.NewOrganizationResource,
	}
}

// DataSources returns the data sources supported by this provider.
func (p *AivenProvider) DataSources(context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		organization.NewOrganizationDataSource,
	}
}

// New returns a new provider factory for the Aiven provider.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &AivenProvider{
			version: version,
		}
	}
}
