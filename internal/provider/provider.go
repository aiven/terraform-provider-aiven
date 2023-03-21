// Package provider is the implementation of the Aiven provider.
package provider

import (
	"context"
	"os"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AivenProvider is the provider implementation for Aiven.
type AivenProvider struct {
	version string
}

var _ provider.Provider = &AivenProvider{}

// AivenProviderModel is the configuration for the Aiven provider.
type AivenProviderModel struct {
	// APIToken is the Aiven API token.
	APIToken types.String `tfsdk:"api_token"`
}

// Metadata returns information about the provider.
func (p *AivenProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "aiven"

	resp.Version = p.version
}

// Schema returns the schema for this provider's configuration.
func (p *AivenProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_token": schema.StringAttribute{
				Description: "Aiven authentication token. " +
					"Can also be set with the AIVEN_TOKEN environment variable.",
				MarkdownDescription: "Aiven authentication token. " +
					"Can also be set with the `AIVEN_TOKEN` environment variable.",
				Required:  true,
				Sensitive: true,
			},
		},
		MarkdownDescription: "The Aiven provider is used to interact with the various services that Aiven offers. " +
			"The provider needs to be configured with the proper credentials before it can be used.",
	}
}

// Configure configures the provider.
func (p *AivenProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data AivenProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.APIToken.IsNull() {
		token, ok := os.LookupEnv("AIVEN_TOKEN")
		if !ok {
			resp.Diagnostics.AddError(
				"Aiven API token not set",
				"Aiven API token was not set in the provider configuration or the AIVEN_TOKEN environment variable.",
			)

			return
		}

		data.APIToken = types.StringValue(token)
	}

	client, err := common.NewCustomAivenClient(data.APIToken.String(), req.TerraformVersion, p.version)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Aiven client", err.Error())

		return
	}

	resp.DataSourceData = client

	resp.ResourceData = client
}

// Resources returns the resources supported by this provider.
func (p *AivenProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

// DataSources returns the data sources supported by this provider.
func (p *AivenProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

// New returns a new provider factory for the Aiven provider.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &AivenProvider{
			version: version,
		}
	}
}
